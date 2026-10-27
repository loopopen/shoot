package middleware

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type maxResponseBytesKey struct{}
type skipDefaultResponseLimitKey struct{}

// MaxResponseBytesError is returned when a response body exceeds its limit.
type MaxResponseBytesError struct {
	Limit int64
}

func (e *MaxResponseBytesError) Error() string {
	return fmt.Sprintf("response body exceeds %d bytes", e.Limit)
}

// ResponseLimitMiddleware limits a response body to defaultLimit bytes.
// Zero or negative means no client-level limit. A per-request limit stored
// on the context replaces it. SkipDefaultResponseLimit skips the
// client-level limit and still honors an explicit per-request limit.
func ResponseLimitMiddleware(defaultLimit int64) Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripper(func(req *http.Request) (*http.Response, error) {
			resp, err := next.RoundTrip(req)
			if err != nil || resp == nil || resp.Body == nil {
				return resp, err
			}
			limit, ok := responseBodyLimit(req, defaultLimit)
			if !ok {
				return resp, nil
			}
			if resp.ContentLength > limit {
				_ = resp.Body.Close()
				return nil, &MaxResponseBytesError{Limit: limit}
			}
			resp.Body = &limitedResponseBody{
				ReadCloser: http.MaxBytesReader(nil, resp.Body, limit),
				limit:      limit,
			}
			return resp, nil
		})
	}
}

// ContextWithMaxResponseBytes stores a per-request response body limit.
// Zero or negative removes the limit, including a client-level default.
func ContextWithMaxResponseBytes(ctx context.Context, n int64) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, maxResponseBytesKey{}, n)
}

// SkipDefaultResponseLimit marks a context so the client-level response
// limit does not apply. An explicit per-request limit still applies.
func SkipDefaultResponseLimit(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, skipDefaultResponseLimitKey{}, true)
}

func responseBodyLimit(req *http.Request, defaultLimit int64) (int64, bool) {
	if req != nil {
		if v := req.Context().Value(maxResponseBytesKey{}); v != nil {
			n, ok := v.(int64)
			if !ok || n <= 0 {
				return 0, false
			}
			return n, true
		}
		if req.Context().Value(skipDefaultResponseLimitKey{}) != nil {
			return 0, false
		}
	}
	if defaultLimit <= 0 {
		return 0, false
	}
	return defaultLimit, true
}

type limitedResponseBody struct {
	io.ReadCloser
	limit int64
}

func (b *limitedResponseBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		err = &MaxResponseBytesError{Limit: b.limit}
	}
	return n, err
}
