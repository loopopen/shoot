package dest

import "time"

//go:generate go run github.com/loopopen/shoot/cmd/shoot new -getset -type=Order2 -ver=test

type Order2 struct {
	id        string
	amount    Decimal
	orderTime time.Time
	addr0     OrderAddress
	addr1     OrderAddress
	addr2     *OrderAddress
	addr3     *OrderAddress
}
