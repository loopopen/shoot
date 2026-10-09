package shoot

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"math"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultRetryWait    = 100 * time.Millisecond
	defaultRetryMaxWait = 2 * time.Second
)

// RetryCondition reports whether a finished attempt should be retried.
// resp is nil when the transport fails before a response arrives.
type RetryCondition func(resp *http.Response, err error) bool

// RetryDelayStrategy chooses how long to wait after a failed attempt.
// attempt is the 1-based index of the attempt that just failed.
type RetryDelayStrategy func(resp *http.Response, err error, attempt int) (time.Duration, error)

// RetryConfig is the client-level retry policy. Count is the number of
// additional attempts, so the total number of attempts is Count + 1.
// A zero Count disables retries. POST and PATCH are not retried unless
// AllowNonIdempotent is set.
type RetryConfig struct {
	Count              int
	Wait               time.Duration
	MaxWait            time.Duration
	AllowNonIdempotent bool
	DisableDefaults    bool
	Conditions         []RetryCondition
	Strategy           RetryDelayStrategy
}

// RetryConstantDelay waits the same duration before every retry.
func RetryConstantDelay(delay time.Duration) RetryDelayStrategy {
	return func(*http.Response, error, int) (time.Duration, error) {
		return delay, nil
	}
}

// Do sends req with client, retrying outside http.Client.Do.
//
// call comes from ApplyRequestOptions. Its timeout is applied separately to
// each attempt. The request context, including a WithDeadline budget, bounds
// every attempt and the waits between them. The returned response body stays
// open; closing it cancels the winning attempt. The caller still cancels the
// overall deadline with Call.Cancel.
func Do(client *http.Client, req *http.Request, base RetryConfig, call Call) (*http.Response, error) {
	if client == nil {
		return nil, errors.New("retry: client must not be nil")
	}
	if req == nil {
		return nil, errors.New("retry: request must not be nil")
	}

	policy := resolveRetry(base, call)
	parent := req.Context()
	attempts := 1
	if policy.Count > 0 && (policy.AllowNonIdempotent || idempotent(req.Method)) {
		attempts = policy.Count + 1
	}

	var resp *http.Response
	var err error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			cause := err
			if err = waitBeforeRetry(parent, resp, cause, attempt, policy); err != nil {
				return nil, err
			}
			resp = nil
			if err = rewindBody(req); err != nil {
				return nil, err
			}
		}

		attemptCtx, cancel := attemptContext(parent, call.attemptTimeout)
		resp, err = client.Do(req.WithContext(attemptCtx))
		if attempt+1 == attempts || !shouldRetry(resp, err, policy) {
			return finishAttempt(resp, err, cancel)
		}
		// Keep headers for Retry-After. The body is released before the wait.
		drainBody(resp)
		cancel()
		if err = parent.Err(); err != nil {
			return nil, err
		}
	}
	return resp, err
}

func resolveRetry(base RetryConfig, call Call) RetryConfig {
	policy := base
	if call.retryCount != nil {
		policy.Count = *call.retryCount
	}
	if policy.Count < 0 {
		policy.Count = 0
	}
	if call.retryWait != nil {
		policy.Wait = *call.retryWait
	}
	if call.retryMaxWait != nil {
		policy.MaxWait = *call.retryMaxWait
	}
	if call.allowNonIdempotent != nil {
		policy.AllowNonIdempotent = *call.allowNonIdempotent
	}
	if call.disableDefaults != nil {
		policy.DisableDefaults = *call.disableDefaults
	}
	if call.retryStrategy != nil {
		policy.Strategy = call.retryStrategy
	}
	if len(call.retryConditions) > 0 {
		policy.Conditions = append(append([]RetryCondition{}, call.retryConditions...), policy.Conditions...)
	}
	return policy
}

func idempotent(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

func attemptContext(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return parent, func() {}
	}
	return context.WithTimeout(parent, timeout)
}

func shouldRetry(resp *http.Response, err error, policy RetryConfig) bool {
	if !policy.DisableDefaults && defaultRetryCondition(resp, err) {
		return true
	}
	for _, condition := range policy.Conditions {
		if condition != nil && condition(resp, err) {
			return true
		}
	}
	return false
}

func defaultRetryCondition(resp *http.Response, err error) bool {
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return false
		}
		var certErr *tls.CertificateVerificationError
		if errors.As(err, &certErr) {
			return false
		}
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			msg := urlErr.Error()
			if strings.Contains(msg, "stopped after") && strings.Contains(msg, "redirects") {
				return false
			}
			if strings.Contains(msg, "unsupported protocol scheme") || strings.Contains(msg, "invalid header") {
				return false
			}
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return true
		}
		var netErr net.Error
		if errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary()) {
			return true
		}
		return false
	}
	if resp == nil {
		return false
	}
	code := resp.StatusCode
	return code == http.StatusTooManyRequests ||
		code == 0 ||
		(code >= http.StatusInternalServerError && code != http.StatusNotImplemented)
}

func waitBeforeRetry(parent context.Context, resp *http.Response, cause error, attempt int, policy RetryConfig) error {
	if err := parent.Err(); err != nil {
		return err
	}
	delay, err := nextDelay(resp, cause, attempt, policy)
	if err != nil {
		return err
	}
	if delay <= 0 {
		return parent.Err()
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-parent.Done():
		return parent.Err()
	case <-timer.C:
		return nil
	}
}

func nextDelay(resp *http.Response, err error, attempt int, policy RetryConfig) (time.Duration, error) {
	if delay, ok := retryAfter(resp); ok {
		return delay, nil
	}
	if policy.Strategy != nil {
		return policy.Strategy(resp, err, attempt)
	}
	return exponentialJitter(policy.Wait, policy.MaxWait, attempt), nil
}

func retryAfter(resp *http.Response) (time.Duration, bool) {
	if resp == nil {
		return 0, false
	}
	if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode != http.StatusServiceUnavailable {
		return 0, false
	}
	return parseRetryAfter(resp.Header.Get("Retry-After"))
}

func parseRetryAfter(value string) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		if seconds < 0 {
			return 0, false
		}
		return time.Duration(seconds) * time.Second, true
	}
	when, err := time.Parse(time.RFC1123, value)
	if err != nil {
		return 0, false
	}
	delay := time.Until(when)
	if delay < 0 {
		return 0, true
	}
	return delay, true
}

func exponentialJitter(min, max time.Duration, attempt int) time.Duration {
	if min <= 0 {
		min = defaultRetryWait
	}
	if max <= 0 {
		max = defaultRetryMaxWait
	}
	if attempt < 0 {
		attempt = 0
	}
	ceiling := math.Min(float64(max), float64(min)*math.Exp2(float64(attempt)))
	center := time.Duration(ceiling / 2)
	if center <= 0 {
		return time.Duration(ceiling)
	}
	span := int64(center)
	if span <= 0 {
		return center
	}
	return center + time.Duration(rand.Int64N(span))
}

func rewindBody(req *http.Request) error {
	if req.Body == nil || req.Body == http.NoBody {
		return nil
	}
	if req.GetBody == nil {
		return errors.New("retry: request body cannot be rewound")
	}
	fresh, err := req.GetBody()
	if err != nil {
		return err
	}
	_ = req.Body.Close()
	req.Body = fresh
	return nil
}

func drainBody(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}

func finishAttempt(resp *http.Response, err error, cancel context.CancelFunc) (*http.Response, error) {
	if err != nil || resp == nil || resp.Body == nil {
		drainBody(resp)
		cancel()
		return resp, err
	}
	resp.Body = &cancelOnClose{ReadCloser: resp.Body, cancel: cancel}
	return resp, nil
}

type cancelOnClose struct {
	io.ReadCloser
	cancel func()
	once   sync.Once
	err    error
}

func (b *cancelOnClose) Close() error {
	b.once.Do(func() {
		b.err = b.ReadCloser.Close()
		if b.cancel != nil {
			b.cancel()
		}
	})
	return b.err
}
