package restclient

import (
	"context"

	"github.com/loopopen/shoot"
	"github.com/loopopen/shoot/cmd/test/restclient/alias"
	"github.com/loopopen/shoot/cmd/test/restclient/dto"
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

type Rest[T any] struct {
	Data T      `json:"data"`
	Code string `json:"code"`
}

type Client interface {
	shoot.RestClient[Client]

	//@Get("/users/{id}")
	//@alias={userID:id}
	GetUser(ctx context.Context, userID string, opts ...shoot.RequestOption) (*User, *shoot.Response, error)

	//@Get("/users")
	//@alias={pageSize:size},{pageIdx:page_idx}
	QueryUsers(ctx context.Context, key string, pageSize, pageIdx int) (*QueryUsersResp, *shoot.Response, error)

	//@Get("/users")
	QueryUsers2(ctx context.Context, params map[string]string) (*QueryUsersResp, *shoot.Response, error)

	//@Get("/users")
	QueryUsers3(ctx context.Context, params *map[string]string) (*QueryUsersResp, *shoot.Response, error)

	//@Get("/users")
	QueryUsers4(ctx context.Context, req QueryUsersReq) (*QueryUsersResp, *shoot.Response, error)

	//@Get("/users")
	QueryUsers5(ctx context.Context, req *QueryUsersReq) (*QueryUsersResp, *shoot.Response, error)

	//@Get("/users")
	QueryUsers6(ctx context.Context, req *QueryUsersReq) ([]User, *shoot.Response, error)

	//@Get("/users")
	QueryUsers7(ctx context.Context, req *dto.QueryUsersReq) ([]dto.User, *shoot.Response, error)

	//@Get("/users")
	QueryUsers8(ctx context.Context, req *alias.QueryUsersReq) ([]alias.User, *shoot.Response, error)

	//@Get("/users")
	QueryUsers9(ctx context.Context, req *QueryUsersReq) (*Rest[[]User], *shoot.Response, error)

	//@Put("/users/{id}")
	UpdateUser(ctx context.Context, id int, user User, opts ...shoot.RequestOption) (*shoot.Response, error)

	//@Post("/ping")
	PostNoBody() (*shoot.Response, error)

	//@Post("/enabled")
	SetEnabled(ctx context.Context, enabled bool) (*shoot.Response, error)

	//@Patch("/users/{id}/labels")
	ReplaceLabels(ctx context.Context, id int, labels map[string]string) (*shoot.Response, error)

	//@Get("/download")
	Download(ctx context.Context) (*shoot.StreamResponse, error)
}
