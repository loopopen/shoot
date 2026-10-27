package dto

import "time"

//go:generate go tool shoot new -getset -json -file=$GOFILE

type User struct {
	id   string
	name string
	age  string
}

type QueryUsersReq struct {
	name string
}

type QueryUsersResp struct {
	data []User
}

type Book struct {
	sn    string
	name  string
	price int
}

type QueryBooksReq struct {
	name      string
	language  string `rest:"alias=lang,omitempty"`
	PageSize  int    `rest:"alias=page_size"`
	PageIndex int
	Ignore    string    `rest:"-"`
	PubDate   time.Time `rest:"dateformat=2006-01-02"`
}

type QueryBooksResp struct {
	data []Book
}
