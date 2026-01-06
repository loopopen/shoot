package model

//go:generate go tool shoot new -getset -opt -type=Conf

type Conf struct {
	//shoot: new
	name string
	host []string //todo: shoot: def=host1,host2
	//shoot: def=80
	port int
	//shoot: new;def="key"
	key1 string
	//shoot: def="key"
	key2 string
}
