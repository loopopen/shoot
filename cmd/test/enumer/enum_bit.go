package enumer

type FormatStyle int32

const (
	None FormatStyle = 0
	Bold FormatStyle = 1 << (iota - 1)
	Italic
	Underline
	Strikethrough
)
