package client2

import (
	"context"
	"net/http"
	"restclientexample/client2/dto"

	"github.com/lopolopen/shoot"
)

//go:generate go tool shoot rest -type=Client

type Client interface {
	shoot.RestClient[Client]

	//shoot: Put("/users/{id}")
	UpdateUser1(ctx context.Context, id int, user User) (*http.Response, error)

	//shoot: Put("/users/{id}")
	UpdateUser2(ctx context.Context, id int, user dto.User) (*http.Response, error)
}
