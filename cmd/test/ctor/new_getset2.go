package ctor

type Base struct {
	z string
	b int
	a string
}

type Son struct {
	Base
	k string
}

type SonOfSon struct {
	Base
	b1 int
	*Son
	a1 int
}
