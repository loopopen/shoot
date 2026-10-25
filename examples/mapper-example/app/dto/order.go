package dto

import (
	"mapperexample/domain/enums"
	"mapperexample/domain/model"
	"mapperexample/infra/mapper"
)

type Order struct {
	mapper.Mapper
	ID           string            `json:"id"`                           // 订单ID
	Amount       string            `json:"amount"`                       // 订单金额
	Status       enums.OrderStatus `json:"status"`                       // 订单状态
	OrderingTime string            `json:"orderingTime" map:"OrderTime"` // 下单时间
	Address      *OrderAddress     `json:"address"`                      // 收货地址
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
