package src

//go:generate go tool shoot map -path=../dest -alias=model -type=Event

type Event struct {
	Elapsed int64
}
