package dto2

import (
	"mappersample/domain/model2"
)

//go:generate go tool shoot new -getset -json -type=OrderAddress
//go:generate go tool shoot map -path=../../domain/model -alias=domain -type=OrderAddress

type OrderAddress struct {
	city    string
	street  string
	roomNum string `map:"room"`
}

func (o *OrderAddress) ToDomain() *model2.OrderAddress {
	if o == nil {
		return nil
	}
	orderAddress_ := new(model2.OrderAddress)
	orderAddress_.SetCity(o.city)
	orderAddress_.SetStreet(o.street)
	orderAddress_.SetRoom(o.roomNum)
	return orderAddress_
}

// FromDomain reads from type domain.OrderAddress, then writes back to receiver and returns it
func (o *OrderAddress) FromDomain(orderAddress_ *model2.OrderAddress) *OrderAddress {
	if orderAddress_ == nil {
		return o
	}
	if o == nil {
		o = new(OrderAddress)
	}
	o.city = orderAddress_.City()
	o.street = orderAddress_.Street()
	o.roomNum = orderAddress_.Room()
	return o
}

func (o *OrderAddress) writeDomain(dm *model2.OrderAddress) {
	dm.SetCity("")
}

func (o *OrderAddress) readDomain(*model2.OrderAddress) {
	o.city = ""
	o.SetCity("")
}
