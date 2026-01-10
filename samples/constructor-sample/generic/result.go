package generic

//go:generate go tool shoot new -getset -tagcase=camel -type=Result

type Result[T any] struct {
	data T
	code int
	msg  string
}
