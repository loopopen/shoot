package shoot

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestDoRetriesIdempotentRequest(t *testing.T) {
	const payload = `{"name":"shoot"}`
	var bodies []string
	statuses := []int{http.StatusBadGateway, http.StatusOK}

	client := &http.Client{Transport: roundTrip(func(req *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		bodies = append(bodies, string(body))
		code := statuses[0]
		if len(bodies) > 1 {
			code = statuses[1]
		}
		return &http.Response{
			StatusCode: code,
			Body:       io.NopCloser(strings.NewReader("upstream")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})}

	req, err := http.NewRequest(http.MethodPut, "https://example.com/items", bytes.NewReader([]byte(payload)))
	if err != nil {
		t.Fatal(err)
	}
	call, err := ApplyRequestOptions(req, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer call.Cancel()

	resp, err := Do(client, req, RetryConfig{
		Count:    1,
		Strategy: RetryConstantDelay(0),
	}, call)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if len(bodies) != 2 || bodies[0] != payload || bodies[1] != payload {
		t.Fatalf("bodies = %#v, want two copies of the payload", bodies)
	}
}

func TestDoDoesNotRetryWithoutPolicy(t *testing.T) {
	attempts := 0
	client := &http.Client{Transport: roundTrip(func(req *http.Request) (*http.Response, error) {
		attempts++
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Body:       io.NopCloser(strings.NewReader("no")),
			Header:     make(http.Header),
		}, nil
	})}
	req := newRetryRequest(t, http.MethodGet, nil)
	resp, err := Do(client, req, RetryConfig{}, Call{})
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}

func TestDoSkipsNonIdempotentMethod(t *testing.T) {
	attempts := 0
	client := &http.Client{Transport: roundTrip(func(req *http.Request) (*http.Response, error) {
		attempts++
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Body:       io.NopCloser(strings.NewReader("no")),
			Header:     make(http.Header),
		}, nil
	})}
	req := newRetryRequest(t, http.MethodPost, bytes.NewReader([]byte(`{}`)))
	resp, err := Do(client, req, RetryConfig{Count: 2, Strategy: RetryConstantDelay(0)}, Call{})
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}

func TestDoAllowsNonIdempotentOverride(t *testing.T) {
	attempts := 0
	client := &http.Client{Transport: roundTrip(func(req *http.Request) (*http.Response, error) {
		attempts++
		code := http.StatusBadGateway
		if attempts == 2 {
			code = http.StatusOK
		}
		return &http.Response{
			StatusCode: code,
			Body:       io.NopCloser(strings.NewReader("ok")),
			Header:     make(http.Header),
		}, nil
	})}
	req := newRetryRequest(t, http.MethodPost, nil)
	call, err := ApplyRequestOptions(req, 0, WithAllowNonIdempotentRetry(), WithRetryCount(1), WithRetryStrategy(RetryConstantDelay(0)))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := Do(client, req, RetryConfig{}, call)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
}

func TestDoClosesAbandonedResponseBeforeRetry(t *testing.T) {
	failed := &trackingBody{Reader: strings.NewReader("fail")}
	succeeded := &trackingBody{Reader: strings.NewReader("ok")}
	attempt := 0
	client := &http.Client{Transport: roundTrip(func(req *http.Request) (*http.Response, error) {
		attempt++
		if attempt == 1 {
			return &http.Response{StatusCode: http.StatusInternalServerError, Body: failed, Header: make(http.Header)}, nil
		}
		if !failed.closed {
			t.Errorf("failed response body was still open when the retry started")
		}
		return &http.Response{StatusCode: http.StatusOK, Body: succeeded, Header: make(http.Header)}, nil
	})}

	req := newRetryRequest(t, http.MethodGet, nil)
	resp, err := Do(client, req, RetryConfig{Count: 1, Strategy: RetryConstantDelay(0)}, Call{})
	if err != nil {
		t.Fatal(err)
	}
	if succeeded.closed {
		t.Fatal("successful response body was closed by Do")
	}
	resp.Body.Close()
	if !failed.closed {
		t.Fatal("failed response body was never closed")
	}
}

func TestDoHonorsParentContextDuringBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	client := &http.Client{Transport: roundTrip(func(req *http.Request) (*http.Response, error) {
		cancel()
		return &http.Response{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader("later")),
			Header:     make(http.Header),
		}, nil
	})}
	req := newRetryRequest(t, http.MethodGet, nil).WithContext(ctx)
	start := time.Now()
	resp, err := Do(client, req, RetryConfig{Count: 1, Strategy: RetryConstantDelay(300 * time.Millisecond)}, Call{})
	elapsed := time.Since(start)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if elapsed >= 150*time.Millisecond {
		t.Fatalf("Do took %s after cancel", elapsed)
	}
}

func TestDoUsesFreshAttemptTimeout(t *testing.T) {
	var deadlines []time.Duration
	client := &http.Client{Transport: roundTrip(func(req *http.Request) (*http.Response, error) {
		deadline, ok := req.Context().Deadline()
		if !ok {
			t.Error("attempt context has no deadline")
			return nil, errors.New("missing deadline")
		}
		deadlines = append(deadlines, time.Until(deadline))
		<-req.Context().Done()
		return nil, req.Context().Err()
	})}

	req := newRetryRequest(t, http.MethodGet, nil)
	const perAttempt = 40 * time.Millisecond
	start := time.Now()
	_, err := Do(client, req, RetryConfig{Count: 1, Strategy: RetryConstantDelay(0)}, Call{attemptTimeout: perAttempt})
	elapsed := time.Since(start)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want deadline exceeded", err)
	}
	if len(deadlines) != 2 {
		t.Fatalf("attempts = %d, want 2", len(deadlines))
	}
	for i, remaining := range deadlines {
		if remaining < 20*time.Millisecond || remaining > perAttempt {
			t.Fatalf("attempt %d remaining = %s, want about %s", i+1, remaining, perAttempt)
		}
	}
	if elapsed < 60*time.Millisecond || elapsed > 200*time.Millisecond {
		t.Fatalf("elapsed = %s, want two fresh %s attempts", elapsed, perAttempt)
	}
}

func TestDoParentDeadlineBoundsRetries(t *testing.T) {
	attempts := 0
	client := &http.Client{Transport: roundTrip(func(req *http.Request) (*http.Response, error) {
		attempts++
		<-req.Context().Done()
		return nil, req.Context().Err()
	})}
	parent, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	req := newRetryRequest(t, http.MethodGet, nil).WithContext(parent)
	start := time.Now()
	_, err := Do(client, req, RetryConfig{Count: 5, Strategy: RetryConstantDelay(0)}, Call{attemptTimeout: time.Second})
	elapsed := time.Since(start)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want deadline exceeded", err)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1 because the parent budget is shorter than one attempt", attempts)
	}
	if elapsed > 200*time.Millisecond {
		t.Fatalf("elapsed = %s, parent budget was not honored", elapsed)
	}
}

func TestDoHonorsRetryAfter(t *testing.T) {
	attempt := 0
	client := &http.Client{Transport: roundTrip(func(req *http.Request) (*http.Response, error) {
		attempt++
		header := make(http.Header)
		code := http.StatusTooManyRequests
		if attempt == 1 {
			header.Set("Retry-After", "0")
		} else {
			code = http.StatusOK
		}
		return &http.Response{
			StatusCode: code,
			Header:     header,
			Body:       io.NopCloser(strings.NewReader("ok")),
		}, nil
	})}
	req := newRetryRequest(t, http.MethodGet, nil)
	start := time.Now()
	resp, err := Do(client, req, RetryConfig{
		Count:    1,
		Strategy: RetryConstantDelay(time.Hour),
	}, Call{})
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if attempt != 2 {
		t.Fatalf("attempts = %d, want 2", attempt)
	}
	if time.Since(start) > time.Second {
		t.Fatal("Retry-After: 0 did not replace the one-hour strategy")
	}
}

func TestDoUserConditionAndNotImplemented(t *testing.T) {
	t.Run("condition", func(t *testing.T) {
		attempt := 0
		client := &http.Client{Transport: roundTrip(func(req *http.Request) (*http.Response, error) {
			attempt++
			code := http.StatusNotFound
			if attempt == 2 {
				code = http.StatusOK
			}
			return &http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader("x")), Header: make(http.Header)}, nil
		})}
		req := newRetryRequest(t, http.MethodGet, nil)
		call, err := ApplyRequestOptions(req, 0, WithRetryCondition(func(resp *http.Response, err error) bool {
			return err == nil && resp != nil && resp.StatusCode == http.StatusNotFound
		}))
		if err != nil {
			t.Fatal(err)
		}
		resp, err := Do(client, req, RetryConfig{Count: 1, Strategy: RetryConstantDelay(0)}, call)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if attempt != 2 {
			t.Fatalf("attempts = %d, want 2", attempt)
		}
	})

	t.Run("501", func(t *testing.T) {
		attempt := 0
		client := &http.Client{Transport: roundTrip(func(req *http.Request) (*http.Response, error) {
			attempt++
			return &http.Response{StatusCode: http.StatusNotImplemented, Body: io.NopCloser(strings.NewReader("x")), Header: make(http.Header)}, nil
		})}
		req := newRetryRequest(t, http.MethodGet, nil)
		resp, err := Do(client, req, RetryConfig{Count: 2, Strategy: RetryConstantDelay(0)}, Call{})
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if attempt != 1 {
			t.Fatalf("attempts = %d, want 1", attempt)
		}
	})
}

func TestDoDoesNotRetryPermanentErrors(t *testing.T) {
	cases := []error{
		&url.Error{Op: "Get", URL: "https://example.com", Err: errors.New("stopped after 10 redirects")},
		&url.Error{Op: "Get", URL: "https://example.com", Err: errors.New("unsupported protocol scheme")},
		&tls.CertificateVerificationError{},
	}
	for _, cause := range cases {
		attempts := 0
		client := &http.Client{Transport: roundTrip(func(req *http.Request) (*http.Response, error) {
			attempts++
			return nil, cause
		})}
		req := newRetryRequest(t, http.MethodGet, nil)
		_, err := Do(client, req, RetryConfig{Count: 3, Strategy: RetryConstantDelay(0)}, Call{})
		if err == nil {
			t.Fatal("expected error")
		}
		if attempts != 1 {
			t.Fatalf("attempts = %d for %v, want 1", attempts, cause)
		}
	}
}

func TestDoRejectsBodyThatCannotBeRewound(t *testing.T) {
	attempt := 0
	client := &http.Client{Transport: roundTrip(func(req *http.Request) (*http.Response, error) {
		attempt++
		_, _ = io.ReadAll(req.Body)
		return &http.Response{StatusCode: http.StatusBadGateway, Body: io.NopCloser(strings.NewReader("x")), Header: make(http.Header)}, nil
	})}
	req := newRetryRequest(t, http.MethodPut, nil)
	req.Body = io.NopCloser(strings.NewReader("once"))
	req.GetBody = nil
	_, err := Do(client, req, RetryConfig{Count: 1, Strategy: RetryConstantDelay(0)}, Call{})
	if err == nil || !strings.Contains(err.Error(), "cannot be rewound") {
		t.Fatalf("err = %v", err)
	}
	if attempt != 1 {
		t.Fatalf("attempts = %d, want 1", attempt)
	}
}

func TestExponentialJitterStaysInRange(t *testing.T) {
	for i := 0; i < 20; i++ {
		delay := exponentialJitter(100*time.Millisecond, 2*time.Second, 1)
		if delay < 100*time.Millisecond || delay >= 200*time.Millisecond {
			t.Fatalf("delay = %s, want [100ms, 200ms)", delay)
		}
	}
}

func TestParseRetryAfter(t *testing.T) {
	delay, ok := parseRetryAfter("120")
	if !ok || delay != 120*time.Second {
		t.Fatalf("seconds = %s, %v", delay, ok)
	}
	if _, ok := parseRetryAfter("-1"); ok {
		t.Fatal("negative Retry-After was accepted")
	}
	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC1123)
	delay, ok = parseRetryAfter(past)
	if !ok || delay != 0 {
		t.Fatalf("past date = %s, %v", delay, ok)
	}
}

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newRetryRequest(t *testing.T, method string, body io.Reader) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, "https://example.com/items", body)
	if err != nil {
		t.Fatal(err)
	}
	return req
}

type trackingBody struct {
	io.Reader
	closed bool
}

func (b *trackingBody) Close() error {
	b.closed = true
	return nil
}
