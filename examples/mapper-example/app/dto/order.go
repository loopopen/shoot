package dto

import (
	"mapperexample/domain/enums"
	"mapperexample/domain/model"
	"mapperexample/infra/mapper"
)

type Order struct {
	mapper.Mapper
	ID           string            `json:"id"`
	Amount       string            `json:"amount"`
	Status       enums.OrderStatus `json:"status"`
	OrderingTime string            `json:"orderingTime" map:"OrderTime"`
	Address      *OrderAddress     `json:"address"`
}

func (o *Order) writeDomain(dest *model.Order) {
}

func (o *Order) readDomain(dest *model.Order) {
}

type OrderAddress struct {
	City    string `json:"city"`
	Street  string `json:"street"`
	RoomNum string `json:"roomNum" map:"Room"`
}
