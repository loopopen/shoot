package shoot

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func newRequest(t *testing.T) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, "https://example.com?keep=yes&replace=old", nil)
	if err != nil {
		t.Fatal(err)
	}
	return req
}

func applyOptions(t *testing.T, req *http.Request, options ...RequestOption) func() {
	t.Helper()
	cleanup, err := ApplyRequestOptions(req, 0, options...)
	if err != nil {
		t.Fatal(err)
	}
	return cleanup
}

func TestRequestHeaderOptions(t *testing.T) {
	req := newRequest(t)
	req.Header.Add("Authorization", "old")

	cleanup := applyOptions(t, req,
		WithHeader("Authorization", "first"),
		WithHeaders(http.Header{"X-Value": {"one", "two"}}),
		WithBearerToken("token"),
		WithAccept("application/xml"),
		WithContentType("application/json"),
		WithUserAgent("shoot-test"),
	)
	defer cleanup()

	if got := req.Header.Values("Authorization"); len(got) != 1 || got[0] != "Bearer token" {
		t.Fatalf("Authorization values = %v, want [Bearer token]", got)
	}
	if got := req.Header.Values("X-Value"); len(got) != 2 || got[0] != "one" || got[1] != "two" {
		t.Fatalf("X-Value values = %v, want [one two]", got)
	}
	if got := req.Header.Get("Accept"); got != "application/xml" {
		t.Fatalf("Accept = %q, want application/xml", got)
	}
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
	if got := req.Header.Get("User-Agent"); got != "shoot-test" {
		t.Fatalf("User-Agent = %q, want shoot-test", got)
	}
}

func TestWithBasicAuth(t *testing.T) {
	req := newRequest(t)
	cleanup := applyOptions(t, req, WithBasicAuth("user", "password"))
	defer cleanup()

	username, password, ok := req.BasicAuth()
	if !ok || username != "user" || password != "password" {
		t.Fatalf("BasicAuth() = %q, %q, %v", username, password, ok)
	}
}

func TestRequestQueryOptions(t *testing.T) {
	req := newRequest(t)
	cleanup := applyOptions(t, req,
		WithQueryParam("replace", "new"),
		WithQueryParams(url.Values{
			"multi":   {"one", "two"},
			"replace": {"newest"},
		}),
	)
	defer cleanup()

	query := req.URL.Query()
	if got := query.Get("keep"); got != "yes" {
		t.Fatalf("keep = %q, want yes", got)
	}
	if got := query.Get("replace"); got != "newest" {
		t.Fatalf("replace = %q, want newest", got)
	}
	if got := query["multi"]; len(got) != 2 || got[0] != "one" || got[1] != "two" {
		t.Fatalf("multi = %v, want [one two]", got)
	}
}

func TestWithCookies(t *testing.T) {
	req := newRequest(t)
	cleanup := applyOptions(t, req,
		WithCookie(&http.Cookie{Name: "session", Value: "abc"}),
		WithCookies(&http.Cookie{Name: "theme", Value: "dark"}),
	)
	defer cleanup()

	cookies := req.Cookies()
	if len(cookies) != 2 || cookies[0].Name != "session" || cookies[1].Name != "theme" {
		t.Fatalf("cookies = %v", cookies)
	}
}

func TestWithTimeoutCleansUpContext(t *testing.T) {
	req := newRequest(t)
	cleanup := applyOptions(t, req, WithTimeout(time.Hour))
	if _, ok := req.Context().Deadline(); !ok {
		t.Fatal("request context has no deadline")
	}

	cleanup()
	select {
	case <-req.Context().Done():
		if req.Context().Err() != context.Canceled {
			t.Fatalf("context error = %v, want context.Canceled", req.Context().Err())
		}
	case <-time.After(time.Second):
		t.Fatal("request context was not canceled by cleanup")
	}
}

func TestClientDefaultTimeoutAppliesWithoutRequestOption(t *testing.T) {
	req := newRequest(t)
	started := time.Now()
	cleanup, err := ApplyRequestOptions(req, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	deadline, ok := req.Context().Deadline()
	if !ok {
		t.Fatal("request context has no deadline")
	}
	remaining := deadline.Sub(started)
	if remaining < 59*time.Minute || remaining > 61*time.Minute {
		t.Fatalf("remaining timeout = %v, want about 1h", remaining)
	}
}

func TestRequestTimeoutOverridesClientDefault(t *testing.T) {
	req := newRequest(t)
	started := time.Now()
	cleanup, err := ApplyRequestOptions(req, time.Minute, WithTimeout(2*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	deadline, ok := req.Context().Deadline()
	if !ok {
		t.Fatal("request context has no deadline")
	}
	remaining := deadline.Sub(started)
	if remaining < 119*time.Minute || remaining > 121*time.Minute {
		t.Fatalf("remaining timeout = %v, want about 2h", remaining)
	}
}

func TestRequestTimeoutCanDisableClientDefault(t *testing.T) {
	req := newRequest(t)
	cleanup, err := ApplyRequestOptions(req, time.Minute, WithTimeout(0))
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	if _, ok := req.Context().Deadline(); ok {
		t.Fatal("request context unexpectedly has a deadline")
	}
}

func TestUpstreamContextCanExpireBeforeRequestTimeout(t *testing.T) {
	parent, cancelParent := context.WithTimeout(context.Background(), time.Hour)
	defer cancelParent()
	req := newRequest(t).WithContext(parent)
	cleanup, err := ApplyRequestOptions(req, time.Minute, WithTimeout(2*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	got, ok := req.Context().Deadline()
	want, _ := parent.Deadline()
	if !ok || !got.Equal(want) {
		t.Fatalf("deadline = %v, %v; want upstream deadline %v", got, ok, want)
	}
}

func TestWithDeadline(t *testing.T) {
	req := newRequest(t)
	want := time.Now().Add(time.Hour)
	cleanup := applyOptions(t, req, WithDeadline(want))
	defer cleanup()

	got, ok := req.Context().Deadline()
	if !ok || !got.Equal(want) {
		t.Fatalf("deadline = %v, %v; want %v, true", got, ok, want)
	}
}

func TestWithRequestModifierError(t *testing.T) {
	req := newRequest(t)
	wantErr := context.Canceled
	cleanup, err := ApplyRequestOptions(
		req,
		0,
		WithTimeout(time.Hour),
		WithRequestModifier(func(*http.Request) error {
			return wantErr
		}),
	)
	defer cleanup()
	if err != wantErr {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
	if req.Context().Err() != nil {
		t.Fatalf("context error = %v, want nil", req.Context().Err())
	}
}
