package model

import (
	"mapperexample/domain/enums"
	"time"

	"github.com/shopspring/decimal"
)

type Order struct {
	ID        string
	Amount    decimal.Decimal
	Status    enums.OrderStatus
	OrderTime time.Time
	Address   OrderAddress
}
