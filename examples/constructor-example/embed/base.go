package embed

//go:generate go tool shoot new -getset -json -type=Base

type Base struct {
	id uint
}
