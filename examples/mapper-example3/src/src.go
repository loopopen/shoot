package src

import "mapperexample3/iface"

//go:generate go tool shoot map -path=../dest -type=Data

type Result[T any] struct {
	Data T
	Code int
}

type A struct{}

func (A) A() {}

type Mapper struct{}

func (Mapper) IntToInt32(x int) int32 {
	return 0
}

type Data struct {
	Mapper
	X   iface.ABIface
	Y   A
	Age int
}
