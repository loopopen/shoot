package shoot

import "github.com/loopopen/shoot/constraints"

type defaulter interface {
	SetDefault()
}

// Option defines a functional option for configuring an instance of type T.
// PT is a pointer type that satisfies DefaultSetter[T], allowing default setup.
type Option[T any, PT constraints.NewShooter[T]] func(*T)

// NewWith creates a new instance of type T using the functional options pattern.
// It first initializes the instance with default values via SetDefault,
// then applies each provided Option in order.
// This pattern promotes clean, declarative construction of configurable types.
func NewWith[T any, PT constraints.NewShooter[T]](opts ...Option[T, PT]) *T {
	t := new(T)
	if d, ok := any(t).(defaulter); ok {
		d.SetDefault()
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}
