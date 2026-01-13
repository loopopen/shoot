package dest

import (
	"github.com/lopolopen/shoot/samples/mapper-sample3/iface"
)

type Result[T any] struct {
	Data T
	Code int
}

type Data struct {
	X   iface.AIface
	Y   iface.AIface
	Age int32
}
