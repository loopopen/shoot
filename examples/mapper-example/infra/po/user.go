package po

import (
	"mapperexample/domain/model"

	"gorm.io/gorm"
)

type User struct {
	*gorm.Model
	FirstName   string         `gorm:"size:64"`
	LastName    string         `gorm:"size:64"`
	Email       string         `gorm:"size:128;uniqueIndex"`
	AddressList []*UserAddress `gorm:"-"`
}

func (po *User) writeDomain(dm *model.User) {

}

func (po *User) readDomain(dm model.User) {
	po.Model = &gorm.Model{
		ID: dm.ID,
	}
}
