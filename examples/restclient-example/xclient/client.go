package xclient

import (
	"context"
	"fmt"
	"net/http"
	"restclientexample/xclient/dto"

	"github.com/lopolopen/shoot"
)

//go:generate go tool shoot new -getset -json -file=$GOFILE
//go:generate go tool shoot enum -json -file=$GOFILE
//go:generate go tool shoot rest -type=Client

type Gender int32

const (
	Unknown Gender = iota
	Male
	Femal
)

type User struct {
	id     string
	name   string
	age    int
	gender Gender
}

type QueryBooksReq struct {
	name      *string
	language  string `shoot:"alias=lang,omitempty"`
	PageSize  int    `shoot:"alias=page_size"`
	PageIndex *int
	// PubDate   time.Time `shoot:"dateformat=2006-01-02"` //todo: support this
	// Ignore    string    `shoot:"-"`                     //todo: support this
}

type Client interface {
	//shoot: headers={Tenant-Id:123}
	shoot.RestClient[Client]

	//shoot: Get("/users/{id}")
	//shoot: alias={userID:id},{pageSize:size},{pageIdx:page_idx}
	GetUser(ctx context.Context, userID string, pageSize int, pageIdx *int) (*User, *http.Response, error)

	//shoot: Post("/users")
	QueryUsers(ctx context.Context, req dto.QueryUsersReq) (*dto.QueryUsersResp, *http.Response, error)

	// //shoot: Post("/users2")
	// //shoot: headers={Content-Type:application/x-www-form-urlencoded}                           //todo: support this
	// QueryUsers2(ctx context.Context, req dto.QueryUsersReq) (*dto.QueryUsersResp, *http.Response, error)

	//shoot: Get("/books")
	QueryBooks(ctx context.Context, req dto.QueryBooksReq) (*dto.QueryBooksResp, *http.Response, error)

	//shoot: Get("/books")
	QueryBooks0(ctx context.Context, req QueryBooksReq) (*dto.QueryBooksResp, *http.Response, error)

	//shoot: Get("/groups/{id}/books")
	//shoot: alias={groupID:id}
	QueryBooks2(ctx context.Context, groupID int, params map[string]interface{}) (*dto.Book, *http.Response, error) //todo: return array?

	//shoot: Put("/users/{id}")
	UpdateUser(ctx context.Context, id int, user User) (*http.Response, error)

	NoComment(ctx context.Context)
}

func (c *client) NoComment(ctx context.Context) {
	fmt.Println("NoComment called")
}
