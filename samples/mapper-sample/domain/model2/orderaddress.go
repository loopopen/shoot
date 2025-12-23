package model2

//go:generate go tool shoot new -getset -type=OrderAddress

type OrderAddress struct { //Value object
	city   string
	street string
	room   string
}
