package dto2

import (
	"mappersample/domain/enums"
	"mappersample/domain/model2"
	"mappersample/infra/mapper"
)

//go:generate go tool shoot new -getset -json -type=Order
//go:generate go tool shoot map -path=../../domain/model2 -alias=domain -type=Order -v

type Order struct {
	*mapper.Mapper
	id           string
	amount       string
	status       enums.OrderStatus
	orderingTime string `map:"orderTime"`
	address      *OrderAddress
	address2     OrderAddress
	addrList     []*OrderAddress
	AddrList2    []*OrderAddress
	X            int
}

// todo1: model2.OrderSetter
// todo2: int is ok?
func (o *Order) writeDomain(dest *model2.Order) {
}

func (o *Order) readDomain(dest *model2.Order) {
}
