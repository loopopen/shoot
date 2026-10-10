package src

//go:generate go tool shoot map -path=../dest -type=Record

type Label int

type Record struct {
	Label Label
}
