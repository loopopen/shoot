package shoot

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type trackingReadCloser struct {
	io.Reader
	closed bool
}

func (r *trackingReadCloser) Close() error {
	r.closed = true
	return nil
}

func TestResponseParsedBody(t *testing.T) {
	raw := &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Proto:      "HTTP/1.1",
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
	}
	response := NewResponse(raw, []byte("  body  "))

	if got := response.String(); got != "body" {
		t.Fatalf("String() = %q, want %q", got, "body")
	}
	if got := response.StatusCode(); got != http.StatusOK {
		t.Fatalf("StatusCode() = %d, want %d", got, http.StatusOK)
	}
	if !response.IsSuccess() || response.IsError() {
		t.Fatal("expected a successful response")
	}
	if got := response.Size(); got != 8 {
		t.Fatalf("Size() = %d, want 8", got)
	}
}

func TestResponseIsSuccess(t *testing.T) {
	tests := []struct {
		statusCode int
		want       bool
	}{
		{statusCode: http.StatusOK, want: true},
		{statusCode: http.StatusNoContent, want: true},
		{statusCode: http.StatusFound, want: false},
		{statusCode: http.StatusBadRequest, want: false},
		{statusCode: http.StatusInternalServerError, want: false},
	}

	for _, tt := range tests {
		response := NewResponse(&http.Response{StatusCode: tt.statusCode}, nil)
		if got := response.IsSuccess(); got != tt.want {
			t.Errorf("status %d: IsSuccess() = %t, want %t", tt.statusCode, got, tt.want)
		}
	}
}

func TestStreamResponse(t *testing.T) {
	rawBody := &trackingReadCloser{Reader: strings.NewReader("stream")}
	response := NewStreamResponse(&http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Body:       rawBody,
	})

	body, err := io.ReadAll(response.Body())
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if got := string(body); got != "stream" {
		t.Fatalf("raw body = %q, want %q", got, "stream")
	}
	if !response.IsSuccess() || response.Status() != "200 OK" {
		t.Fatal("expected a successful stream response")
	}
	if err := response.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if !rawBody.closed {
		t.Fatal("Close() did not close the underlying response body")
	}
}
