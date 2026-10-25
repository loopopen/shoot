package model

import (
	"mapperexample/domain/enums"
	"time"

	"github.com/shopspring/decimal"
)

type Order struct {
	ID        string            //订单ID
	Amount    decimal.Decimal   //订单金额
	Status    enums.OrderStatus //订单状态
	OrderTime time.Time         //下单时间
	Address   OrderAddress      //收货地址
}
