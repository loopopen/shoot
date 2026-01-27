package enums

//go:generate go tool shoot enum -json -text -file=$GOFILE

type WillBeIgnored = int

type Light int32

const (
	LightRed Light = iota
	LightYello
	LightGreen
)
