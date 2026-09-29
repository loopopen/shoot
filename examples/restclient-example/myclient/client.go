package myclient

import (
	"context"

	"github.com/loopopen/shoot"
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
	Get(ctx context.Context, key string) (*KV, *shoot.Response, error)

	//shoot: Post("/set")
	Set(ctx context.Context, kv *KV) (*shoot.Response, error)

	//shoot: Get("/download")
	//shoot: headers={Content-Type:application/zip}
	Download(ctx context.Context) (*shoot.StreamResponse, error)
}
