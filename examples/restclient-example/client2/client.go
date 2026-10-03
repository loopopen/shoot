package client2

import (
	"context"
	"restclientexample/client2/dto"

	"github.com/loopopen/shoot"
)

//go:generate go tool shoot rest -type=Client

type Client interface {
	shoot.RestClient[Client]

	//@Put("/users/{id}")
	UpdateUser1(ctx context.Context, id int, user User) (*shoot.Response, error)

	//@Put("/users/{id}")
	UpdateUser2(ctx context.Context, id int, user dto.User) (*shoot.Response, error)

	//@Post("/ping")
	PostNoBody() (*shoot.Response, error)

	//@Post("/enabled")
	SetEnabled(ctx context.Context, enabled bool) (*shoot.Response, error)

	//@Patch("/users/{id}/labels")
	ReplaceLabels(ctx context.Context, id int, labels map[string]string) (*shoot.Response, error)
}
