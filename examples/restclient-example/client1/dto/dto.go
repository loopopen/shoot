package dto

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
	language  string `shoot:"alias=lang,omitempty"`
	PageSize  int    `shoot:"alias=page_size"`
	PageIndex int
	// PubDate   time.Time `shoot:"dateformat=2006-01-02"` //todo: support this
	// Ignore    string    `shoot:"-"`                     //todo: support this
}

type QueryBooksResp struct {
	data []Book
}
