package src

import (
	"time"

	"github.com/lopolopen/shoot/cmd/test/mapper/dest"
)

type Mapper struct {
	//shoot: map
}

func (Mapper) StringToDecimal(s string) dest.Decimal {
	return dest.Decimal{}
}

func (Mapper) DecimalToString(d dest.Decimal) string {
	return ""
}

func (Mapper) StringToTime(s string) time.Time {
	return time.Time{}
}

func (Mapper) TimeToString(t time.Time) string {
	return ""
}

type Order struct {
	Mapper
	ID           string
	Amount       string
	OrderingTime string `map:"OrderTime"`
	Addr0        OrderAddress
	Addr1        *OrderAddress
	Addr2        OrderAddress
	Addr3        *OrderAddress
}

type OrderAddress struct {
	City string
}

type Src struct {
	*Embed
	ID      int
	SrcName string `map:"DestName"`
}

type Embed struct {
	ID int
	*EmbedEmbed
}

type EmbedEmbed struct {
	X string
}

type Order2 struct {
	Mapper
	ID           string
	Amount       string
	OrderingTime string `map:"OrderTime"`
	Addr0        OrderAddress
	Addr1        *OrderAddress
	Addr2        OrderAddress
	Addr3        *OrderAddress
}
