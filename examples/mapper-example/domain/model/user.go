package model

type User struct { //Aggregate Root
	ID          uint
	FirstName   string
	LastName    string
	Email       string
	AddressList []*UserAddress
}
