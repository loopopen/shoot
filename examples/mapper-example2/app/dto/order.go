package dto

import (
	"mapperexample2/domain/model"
	"mapperexample2/infra/mapper"
)

type Order struct {
	*mapper.Mapper
	id     string
	amount string
	// status       enums.OrderStatus
	orderingTime string `map:"orderTime"`
	address      *OrderAddress
	//shoot: get
	x string
	//shoot: set
	y string
}

func (o *Order) writeDomain(dest *model.Order) {
	// dest.SetId("...")
	// dest.SetStatus(0)
}

// func (o *Order) readDomain(model.OrderGetter) {
// 	// o.SetY("y")
// }
