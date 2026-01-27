package model

//go:generate go tool shoot new -exp -opt -getset -json -tagcase=camel -file=$GOFILE

type User struct {
	//shoot: get
	id     string
	name   string
	gender int
	age    int
	tel    string
}

type Book struct {
	//shoot: new
	name    string
	writers []string
	Remarks string
	owner   *User
}

type Address struct {
	Province string
	City     string
	District string
	Street   string
}
