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
}

// todo1: model2.OrderSetter
// todo2: int is ok?
func (o *Order) writeDomain(dest *model.Order) {
}

func (o *Order) readDomain(dest *model.Order) {
}
