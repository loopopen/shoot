package dto

import "github.com/lopolopen/shoot/samples/mapper-sample2/domain/model"

type OrderAddress struct {
	city    string
	street  string
	roomNum string `map:"room"`
}

func (o *OrderAddress) writeDomain(dm *model.OrderAddress) {
	// dm.SetCity("")
}

func (o *OrderAddress) readDomain(*model.OrderAddress) {
	// o.city = ""
	// o.SetCity("")
}
