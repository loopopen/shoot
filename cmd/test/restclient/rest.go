package restclient

import (
	"context"
	"net/http"

	"github.com/lopolopen/shoot"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type QueryUsersReq struct {
	Name     string `shoot:"alias=name"`
	PageSize int    `shoot:"alias=size"`
	PageIdx  int    `shoot:"alias=page_idx"`
}

type QueryUsersResp struct{}

type Client interface {
	shoot.RestClient[Client]

	//shoot: Get("/users/{id}")
	//shoot: alias={userID:id}
	GetUser(ctx context.Context, userID string) (*User, *http.Response, error)

	//shoot: Get("/users")
	//shoot: alias={pageSize:size},{pageIdx:page_idx}
	QueryUsers(ctx context.Context, key string, pageSize, pageIdx int) (*QueryUsersResp, *http.Response, error)

	//shoot: Get("/users")
	QueryUsers2(ctx context.Context, params map[string]string) (*QueryUsersResp, *http.Response, error)

	//shoot: Get("/users")
	QueryUsers3(ctx context.Context, params *map[string]string) (*QueryUsersResp, *http.Response, error)

	//shoot: Get("/users")
	QueryUsers4(ctx context.Context, req QueryUsersReq) (*QueryUsersResp, *http.Response, error)

	//shoot: Get("/users")
	QueryUsers5(ctx context.Context, req *QueryUsersReq) (*QueryUsersResp, *http.Response, error)

	//shoot: Get("/users")
	QueryUsers6(ctx context.Context, req *QueryUsersReq) ([]User, *http.Response, error)

	//shoot: Put("/users/{id}")
	UpdateUser(ctx context.Context, id int, user User) (*http.Response, error)
}
