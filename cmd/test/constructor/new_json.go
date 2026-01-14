package constructor

//go:generate go tool shoot new -getset -json -file=$GOFILE -ver=test

type BaseJSON struct {
	z string
	b int
	a string
}

type Other struct {
	Name string
	Age  int
}

type SonJSON struct {
	BaseJSON
	y     int
	x     int
	Other Other
}
