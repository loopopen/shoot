package ctor

type Conf struct {
	//@new
	name string
	host []string
	//@def=80
	port int
}
