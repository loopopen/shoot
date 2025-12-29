package dto

import (
	"mappersample/domain/enums"
	"mappersample/infra/mapper"

	"github.com/lopolopen/shoot/samples/mapper-sample2/domain/model"
)

type Order struct {
	*mapper.Mapper
	id           string
	amount       string
	status       enums.OrderStatus
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

func (o *Order) readDomain(model.OrderGetter) {
	o.y = ""
}
