package middleware

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestMaxResponseBytes(t *testing.T) {
	tests := []struct {
		name           string
		defaultLimit   int64
		requestLimit   *int64
		skipDefault    bool
		body           string
		contentLength  int64
		want           string
		wantLimit      int64
		wantErr        bool
		wantBodyClosed bool
	}{
		{name: "default allows exact body", defaultLimit: 4, body: "abcd", contentLength: -1, want: "abcd"},
		{name: "default rejects one extra byte", defaultLimit: 4, body: "abcde", contentLength: -1, wantErr: true, wantLimit: 4},
		{name: "content length is rejected before read", defaultLimit: 4, body: "abcdefgh", contentLength: 8, wantErr: true, wantLimit: 4, wantBodyClosed: true},
		{name: "zero default does not limit", body: "abcdefgh", contentLength: -1, want: "abcdefgh"},
		{name: "request limit raises the default", defaultLimit: 4, requestLimit: int64ptr(8), body: "abcdefgh", contentLength: -1, want: "abcdefgh"},
		{name: "request limit removes the default", defaultLimit: 4, requestLimit: int64ptr(0), body: "abcdefgh", contentLength: 8, want: "abcdefgh"},
		{name: "skip default", defaultLimit: 4, skipDefault: true, body: "abcdefgh", contentLength: -1, want: "abcdefgh"},
		{name: "request limit still applies when default is skipped", defaultLimit: 4, skipDefault: true, requestLimit: int64ptr(3), body: "abcdefgh", contentLength: -1, wantErr: true, wantLimit: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "http://example.test/item", nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := req.Context()
			if tt.skipDefault {
				ctx = SkipDefaultResponseLimit(ctx)
			}
			if tt.requestLimit != nil {
				ctx = ContextWithMaxResponseBytes(ctx, *tt.requestLimit)
			}
			req = req.WithContext(ctx)
			payload := &recordingBody{Reader: strings.NewReader(tt.body)}
			transport := ResponseLimitMiddleware(tt.defaultLimit)(RoundTripper(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode:    http.StatusOK,
					Header:        make(http.Header),
					ContentLength: tt.contentLength,
					Body:          payload,
				}, nil
			}))
			resp, err := transport.RoundTrip(req)
			if err != nil || resp == nil {
				assertLimitErr(t, err, tt.wantLimit)
				if !tt.wantErr {
					t.Fatal(err)
				}
				if tt.wantBodyClosed && !payload.closed {
					t.Fatal("body was not closed")
				}
				if tt.wantBodyClosed && payload.reads != 0 {
					t.Fatalf("body reads = %d", payload.reads)
				}
				return
			}
			got, err := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if tt.wantErr {
				assertLimitErr(t, err, tt.wantLimit)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Fatalf("body = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSkipDefaultResponseLimitNilContext(t *testing.T) {
	ctx := SkipDefaultResponseLimit(nil)
	if ctx.Value(skipDefaultResponseLimitKey{}) == nil {
		t.Fatal("nil context was not marked")
	}
	if context.Background().Value(skipDefaultResponseLimitKey{}) != nil {
		t.Fatal("background context was marked")
	}
}

func assertLimitErr(t *testing.T, err error, limit int64) {
	t.Helper()
	var limitErr *MaxResponseBytesError
	if !errors.As(err, &limitErr) || limitErr.Limit != limit {
		t.Fatalf("error = %v, want limit %d", err, limit)
	}
}

func int64ptr(n int64) *int64 {
	return &n
}

type recordingBody struct {
	io.Reader
	reads  int
	closed bool
}

func (b *recordingBody) Read(p []byte) (int, error) {
	b.reads++
	return b.Reader.Read(p)
}

func (b *recordingBody) Close() error {
	b.closed = true
	return nil
}
