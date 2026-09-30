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

	//shoot: Get("/users/{id}")
	//shoot: alias={userID:id}
	GetUser(ctx context.Context, userID string) (*User, *shoot.Response, error)

	//shoot: Get("/users")
	//shoot: alias={pageSize:size},{pageIdx:page_idx}
	QueryUsers(ctx context.Context, key string, pageSize, pageIdx int) (*QueryUsersResp, *shoot.Response, error)

	//shoot: Get("/users")
	QueryUsers2(ctx context.Context, params map[string]string) (*QueryUsersResp, *shoot.Response, error)

	//shoot: Get("/users")
	QueryUsers3(ctx context.Context, params *map[string]string) (*QueryUsersResp, *shoot.Response, error)

	//shoot: Get("/users")
	QueryUsers4(ctx context.Context, req QueryUsersReq) (*QueryUsersResp, *shoot.Response, error)

	//shoot: Get("/users")
	QueryUsers5(ctx context.Context, req *QueryUsersReq) (*QueryUsersResp, *shoot.Response, error)

	//shoot: Get("/users")
	QueryUsers6(ctx context.Context, req *QueryUsersReq) ([]User, *shoot.Response, error)

	//shoot: Get("/users")
	QueryUsers7(ctx context.Context, req *dto.QueryUsersReq) ([]dto.User, *shoot.Response, error)

	//shoot: Get("/users")
	QueryUsers8(ctx context.Context, req *alias.QueryUsersReq) ([]alias.User, *shoot.Response, error)

	//shoot: Get("/users")
	QueryUsers9(ctx context.Context, req *QueryUsersReq) (*Rest[[]User], *shoot.Response, error)

	//shoot: Put("/users/{id}")
	UpdateUser(ctx context.Context, id int, user User) (*shoot.Response, error)

	//shoot: Post("/ping")
	PostNoBody() (*shoot.Response, error)

	//shoot: Post("/enabled")
	SetEnabled(ctx context.Context, enabled bool) (*shoot.Response, error)

	//shoot: Patch("/users/{id}/labels")
	ReplaceLabels(ctx context.Context, id int, labels map[string]string) (*shoot.Response, error)

	//shoot: Get("/download")
	Download(ctx context.Context) (*shoot.StreamResponse, error)
}
