package po

import (
	"mapperexample/domain/model"

	"gorm.io/gorm"
)

type UserAddress struct {
	*gorm.Model
	City      string `gorm:"size:64"`
	Street    string `gorm:"size:64"`
	Room      string `gorm:"size:64"`
	Tag       string `gorm:"size:32"`
	IsDefault bool   `gorm:"default:false"`
	UserID    uint   `gorm:"index" map:"OwnerID"`
}

func (po *UserAddress) writeDomain(dm *model.UserAddress) {
}

func (po *UserAddress) readDomain(dm *model.UserAddress) {
	po.Model = &gorm.Model{
		ID: dm.ID,
	}
}
