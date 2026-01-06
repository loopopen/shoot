package embed

//go:generate go tool shoot new -getset -json -type=Order

type Order struct {
	*Base
	status string
}
