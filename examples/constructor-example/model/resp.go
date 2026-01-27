package model

//go:generate go tool shoot new -getset -json -file=$GOFILE

type PageResult struct {
	perPage   int
	page      int
	totalPage int
}

type Resp struct {
	code int
	msg  string
	*PageResult
}

type UserResp struct {
	Resp
	data *User
}
