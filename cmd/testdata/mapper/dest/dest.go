package dest

import (
	"time"
)

type Decimal struct{}

type Order struct {
	ID        string
	Amount    Decimal
	OrderTime time.Time
	Addr0     OrderAddress
	Addr1     OrderAddress
	Addr2     *OrderAddress
	Addr3     *OrderAddress
}

type OrderAddress struct {
	City  string
	Other string
}

type Dest struct {
	ID       int
	DestName string
	X        string
}
