package main

import (
	"context"
	"fmt"
	"net/http"
	"restclientexample/myclient"
	"time"

	"github.com/loopopen/shoot"
	"github.com/loopopen/shoot/middleware"
)

func WaitMiddeleware(next http.RoundTripper) http.RoundTripper {
	return middleware.RoundTripper(func(req *http.Request) (*http.Response, error) {
		time.Sleep(time.Second)
		return next.RoundTrip(req)
	})
}

func main() {
	startServer()

	myC := shoot.NewRest[myclient.Client](
		shoot.BaseURL("http://localhost:8080"),
		shoot.Timeout("0ms"),
		shoot.EnableLogging(true),
		shoot.Use(WaitMiddeleware),
		shoot.WithRetry(shoot.RetryConfig{
			Count: 3,
		}),
	)

	ctx := context.Background()
	_, err := myC.Set(ctx, myclient.NewKV("foo", "bar"))
	if err != nil {
		panic(err)
	}

	kv, resp, err := myC.Get(
		ctx,
		"foo",
		shoot.WithHeader("X-Request-ID", fmt.Sprintf("request-%d", time.Now().UnixNano())),
		// shoot.WithTimeout(2*time.Second),
	)
	if err != nil {
		panic(string(resp.Body()))
	}
	fmt.Printf("%+v\n", *kv)
}
