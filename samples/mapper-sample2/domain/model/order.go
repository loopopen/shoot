package model

import (
	"mappersample/domain/enums"
	"time"

	"github.com/shopspring/decimal"
)

type Order struct {
	id        string
	amount    decimal.Decimal
	status    enums.OrderStatus
	orderTime time.Time
	address   OrderAddress
	x         string
	y         string
}
