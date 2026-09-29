package client2

import (
	"context"
	"restclientexample/client2/dto"

	"github.com/loopopen/shoot"
)

//go:generate go tool shoot rest -type=Client

type Client interface {
	shoot.RestClient[Client]

	//shoot: Put("/users/{id}")
	UpdateUser1(ctx context.Context, id int, user User) (*shoot.Response, error)

	//shoot: Put("/users/{id}")
	UpdateUser2(ctx context.Context, id int, user dto.User) (*shoot.Response, error)
}
