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
	ID        string            `gorm:"primarykey"`
	Amount    decimal.Decimal   `gorm:"type:decimal(20,2)"`
	Status    enums.OrderStatus `gorm:"type:enum('Pending', 'Completed', 'Canceled')"`
	City      string
	Street    string
	Room      string
	OrderTime time.Time
	// OrderTime sql.Null[time.Time]
}

func (po *Order) readDomain(model.Order) {
	po.Model = nil
}
