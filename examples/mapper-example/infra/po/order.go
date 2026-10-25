package po

import (
	"mapperexample/domain/enums"
	"mapperexample/domain/model"
	"mapperexample/infra/mapper"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Order struct {
	*mapper.SQLMapper
	*gorm.Model
	ID        string            `gorm:"primarykey"`                                    // 订单ID
	Amount    decimal.Decimal   `gorm:"type:decimal(20,2)"`                            // 订单金额
	Status    enums.OrderStatus `gorm:"type:enum('Pending', 'Completed', 'Canceled')"` // 订单状态
	City      string
	Street    string
	Room      string
	OrderTime time.Time // 下单时间
	// OrderTime sql.Null[time.Time]
}

func (po *Order) readDomain(model.Order) {
	po.Model = nil
}
