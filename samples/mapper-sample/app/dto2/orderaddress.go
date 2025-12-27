package dto2

import (
	"mappersample/domain/model2"
)

//go:generate go tool shoot new -getset -json -type=OrderAddress
//go:generate go tool shoot map -path=../../domain/model2 -alias=domain -type=OrderAddress

type OrderAddress struct {
	city    string
	street  string
	roomNum string `map:"room"`
}

func (o *OrderAddress) writeDomain(dm *model2.OrderAddress) {
	// dm.SetCity("")
}

func (o *OrderAddress) readDomain(*model2.OrderAddress) {
	// o.city = ""
	// o.SetCity("")
}
