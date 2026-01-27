package dto

import (
	"mapperexample/domain/model"
	"mapperexample/infra/mapper"
	"strings"
)

type User struct {
	*mapper.Mapper
	ID        uint          `json:"id"`
	FullName  string        `json:"fullName"`
	Email     string        `json:"email"`
	Addresses []UserAddress `json:"addresses" map:"AddressList"`
}

func (u *User) toDomain(user *model.User) { //or writeModel
	parts := strings.Fields(u.FullName)
	if len(parts) == 1 {
		user.FirstName = parts[0]
	}
	if len(parts) > 1 {
		user.LastName = parts[1]
	}
}

func (u *User) fromDomain(user model.User) { //or readModel
	u.FullName = user.FirstName + " " + user.LastName
}

// func (u *User) writeModel(user *model.User) {
// 	//err: found more than one manual write method: writeModel
// }

type UserAddress struct {
	ID        uint
	City      string
	Street    string
	Room      string
	Tag       string
	IsDefault bool
	UserID    uint `map:"OwnerID"`
}
