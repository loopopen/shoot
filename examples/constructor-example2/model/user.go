package model

// Test case only, not the recommended invocation.
// go generate starts in this directory; the command below moves to the
// module root and points [dir] back at this package.

//go:generate go -C .. tool shoot new -getset -type=* ./model

type User struct {
	id   string
	name string
}

type Order struct {
	id     string
	amount int
}
