package dest

type Result[T any] struct {
	Data T
	Code int
}
