package status

//go:generate go tool shoot enum -sql -file=$GOFILE

type Status int32

const (
	StatusPending Status = iota
	StatusPaid
)
