package src

//go:generate go tool shoot map -path=../dest -type=Result -r

type Result[T any] struct {
	Data T
	Code int
}
