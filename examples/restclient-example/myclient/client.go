package myclient

import (
	"context"
	"net/http"

	"github.com/lopolopen/shoot"
)

//go:generate go tool shoot new -getset -json -file=$GOFILE
//go:generate go tool shoot rest -type=Client

type KV struct {
	key   string
	value string
}

type Client interface {
	//shoot: headers={Authorization:Basic dXNlcm5hbWU6cGFzc3dvcmQ=}
	shoot.RestClient[Client]

	//shoot: Get("/get")
	Get(ctx context.Context, key string) (*KV, *http.Response, error)

	//shoot: Post("/set")
	Set(ctx context.Context, kv *KV) (*http.Response, error)
}
