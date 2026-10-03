package model

//go:generate go tool shoot new -getset -opt -type=Conf

type Conf struct {
	//@new
	name string
	host []string //todo: shoot: def=host1,host2
	//@def=80
	port int
	//@new;def="key"
	key1 string
	//@def="key"
	key2 string
}
