package model2

import (
	"mappersample/domain/enums"
	"time"

	"github.com/shopspring/decimal"
)

//go:generate go tool shoot new -getset -type=Order

type Order struct {
	id        string
	amount    decimal.Decimal
	status    enums.OrderStatus
	orderTime time.Time
	// OrderTime *time.Time
	address   OrderAddress
	address2  OrderAddress
	addrList  []*OrderAddress
	AddrList2 []*OrderAddress
	X         int
}
