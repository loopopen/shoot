package client2

import (
	"context"
	"net/http"

	"github.com/lopolopen/shoot"
)

//go:generate go tool shoot rest -type=Client

type Client interface {
	shoot.RestClient[Client]

	// //shoot: Get("/users/{id}")
	// //shoot: alias={userID:id},{pageSize:size},{pageIdx:page_idx}
	// GetUser(ctx context.Context, userID string, pageSize int, pageIdx *int) (*User, *http.Response, error)

	//shoot: Put("/users/{id}")
	UpdateUser(ctx context.Context, id int, user User) (*http.Response, error)
}
