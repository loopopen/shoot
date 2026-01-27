package model

type UserAddress struct { //Entity
	ID        uint
	City      string
	Street    string
	Room      string
	Tag       string
	IsDefault bool
	OwnerID   uint
}

type OrderAddress struct { //Value object
	City   string
	Street string
	Room   string
}
