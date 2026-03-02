package enums

//go:generate go tool shoot enum -json -text -file=$GOFILE

type WillBeIgnored = int

type Light int32

const (
	LightRed Light = iota
	LightYello
	LightGreen
)

type Status int32

const (
	StatusFailed  Status = -1
	StatusPending Status = iota
	StatusProcessing
	StatusSucceeded
)
