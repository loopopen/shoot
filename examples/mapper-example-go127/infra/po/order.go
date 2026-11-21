package po

import (
	"database/sql"
	"examples/mapperexamplego127/infra/mapper"
	"time"
)

//go:generate go tool shoot map -path=../../domain/model -type=*

type Order struct {
	*mapper.Mapper
	OrderedAt sql.Null[time.Time] // 下单时间
}

type Order2 struct {
	OrderNo string // 销售订单号
}
