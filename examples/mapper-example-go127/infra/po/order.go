package po

import (
	"database/sql"
	"examples/mapperexamplego127/infra/mapper"
	"time"
)

//go:generate go tool shoot map -path=../../domain/model -comment=force -gorm -type=*

type Order struct {
	*mapper.Mapper
	OrderedAt sql.Null[time.Time] `gorm:"comment:下单时间"` // 下单时间
}

type Order2 struct {
	OrderNo string `gorm:"size:64;comment:销售订单号"` // 销售订单号
	X       string `gorm:"-"`                     // X
}
