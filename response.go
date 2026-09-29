package shoot

import (
	"io"
	"net/http"
	"strings"
)

// Response wraps the response returned by net/http.
//
// Body contains a cached copy of the response payload. RawResponse.Body has
// already been consumed and closed by the generated client.
type Response struct {
	RawResponse *http.Response
	body        []byte
}

// NewResponse constructs a Response for generated clients.
func NewResponse(rawResponse *http.Response, body []byte) *Response {
	return &Response{
		RawResponse: rawResponse,
		body:        body,
	}
}

// Body returns the response payload cached by the generated client.
func (r *Response) Body() []byte {
	if r == nil {
		return nil
	}
	return r.body
}

// String returns the cached response payload as a trimmed string.
func (r *Response) String() string {
	return strings.TrimSpace(string(r.Body()))
}

// RawBody returns the consumed and closed underlying response body.
// Use StreamResponse for responses that must be read incrementally.
func (r *Response) RawBody() io.ReadCloser {
	if r == nil || r.RawResponse == nil {
		return nil
	}
	return r.RawResponse.Body
}

// Status returns the HTTP status text, for example "200 OK".
func (r *Response) Status() string {
	if r == nil || r.RawResponse == nil {
		return ""
	}
	return r.RawResponse.Status
}

// StatusCode returns the HTTP status code.
func (r *Response) StatusCode() int {
	if r == nil || r.RawResponse == nil {
		return 0
	}
	return r.RawResponse.StatusCode
}

// Proto returns the HTTP protocol used for the response.
func (r *Response) Proto() string {
	if r == nil || r.RawResponse == nil {
		return ""
	}
	return r.RawResponse.Proto
}

// Header returns the response headers.
func (r *Response) Header() http.Header {
	if r == nil || r.RawResponse == nil {
		return http.Header{}
	}
	return r.RawResponse.Header
}

// Cookies returns the cookies included in the response.
func (r *Response) Cookies() []*http.Cookie {
	if r == nil || r.RawResponse == nil {
		return []*http.Cookie{}
	}
	return r.RawResponse.Cookies()
}

// Size returns the size of the cached response payload.
func (r *Response) Size() int64 {
	return int64(len(r.Body()))
}

// IsSuccess reports whether the status code is in the 2xx range.
func (r *Response) IsSuccess() bool {
	return r.StatusCode() >= http.StatusOK && r.StatusCode() < http.StatusMultipleChoices
}

// IsError reports whether the status code is 4xx or 5xx.
func (r *Response) IsError() bool {
	return r.StatusCode() >= http.StatusBadRequest
}

// Close closes the underlying response body. Generated clients have normally
// closed it already.
func (r *Response) Close() error {
	body := r.RawBody()
	if body == nil {
		return nil
	}
	return body.Close()
}

// StreamResponse wraps an HTTP response whose body has not been consumed.
// The caller must close it after reading, including when the request returns a
// non-nil error together with a StreamResponse.
type StreamResponse struct {
	RawResponse *http.Response
}

// NewStreamResponse constructs a StreamResponse for generated clients.
func NewStreamResponse(rawResponse *http.Response) *StreamResponse {
	return &StreamResponse{RawResponse: rawResponse}
}

// Body returns the underlying response stream. The caller must close it.
func (r *StreamResponse) Body() io.ReadCloser {
	if r == nil || r.RawResponse == nil {
		return nil
	}
	return r.RawResponse.Body
}

// Status returns the HTTP status text, for example "200 OK".
func (r *StreamResponse) Status() string {
	if r == nil || r.RawResponse == nil {
		return ""
	}
	return r.RawResponse.Status
}

// StatusCode returns the HTTP status code.
func (r *StreamResponse) StatusCode() int {
	if r == nil || r.RawResponse == nil {
		return 0
	}
	return r.RawResponse.StatusCode
}

// Proto returns the HTTP protocol used for the response.
func (r *StreamResponse) Proto() string {
	if r == nil || r.RawResponse == nil {
		return ""
	}
	return r.RawResponse.Proto
}

// Header returns the response headers.
func (r *StreamResponse) Header() http.Header {
	if r == nil || r.RawResponse == nil {
		return http.Header{}
	}
	return r.RawResponse.Header
}

// Cookies returns the cookies included in the response.
func (r *StreamResponse) Cookies() []*http.Cookie {
	if r == nil || r.RawResponse == nil {
		return []*http.Cookie{}
	}
	return r.RawResponse.Cookies()
}

// IsSuccess reports whether the status code is in the 2xx range.
func (r *StreamResponse) IsSuccess() bool {
	return r.StatusCode() >= http.StatusOK && r.StatusCode() < http.StatusMultipleChoices
}

// IsError reports whether the status code is 4xx or 5xx.
func (r *StreamResponse) IsError() bool {
	return r.StatusCode() >= http.StatusBadRequest
}

// Close closes the underlying response stream.
func (r *StreamResponse) Close() error {
	body := r.Body()
	if body == nil {
		return nil
	}
	return body.Close()
}
