package enumer

type Wide uint64

const (
	WideSmall Wide = 1
	WideHigh  Wide = 1 << 63
)
