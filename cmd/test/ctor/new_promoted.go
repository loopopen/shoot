package ctor

type PaymentBase[T any] struct {
	id T
}

type PaymentOrder struct {
	PaymentBase[uint]
	payNo string
}
