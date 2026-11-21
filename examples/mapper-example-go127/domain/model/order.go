package model

import "time"

type Order struct {
	OrderedAt *time.Time // 下单时间
}

//go:generate go tool shoot new -getset -type=Order2

// @getter
type Order2 struct {
	orderNo string // 销售订单号
	X       string // X
}
