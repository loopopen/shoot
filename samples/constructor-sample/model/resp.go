package model

import (
	"encoding/json"
	"fmt"
)

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

func Test() {
	r := NewResp(200, "成功")
	// r.PageResult = NewPageResult(20, 1, 3)
	j, _ := json.Marshal(r)
	fmt.Println(string(j))
}
