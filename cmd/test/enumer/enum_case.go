package enumer

type HTTPStatus int

const (
	HTTPStatusOK HTTPStatus = iota + 1
	HTTPStatusNotFound
	HTTPStatusHTTPServerError
)
