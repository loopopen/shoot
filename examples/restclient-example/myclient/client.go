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
	//@headers={Authorization:Basic dXNlcm5hbWU6cGFzc3dvcmQ=}
	shoot.RestClient[Client]

	//@Get("/get")
	Get(ctx context.Context, key string, opts ...shoot.RequestOption) (*KV, *shoot.Response, error)

	//@Post("/set")
	Set(ctx context.Context, kv *KV, opts ...shoot.RequestOption) (*shoot.Response, error)

	//@Get("/download")
	//@headers={Content-Type:application/zip}
	Download(ctx context.Context) (*shoot.StreamResponse, error)
}
