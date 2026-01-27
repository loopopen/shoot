package main

import (
	"context"
	"fmt"
	"net/http"
	"restclientexample/myclient"
	"time"

	"github.com/lopolopen/shoot"
	"github.com/lopolopen/shoot/middleware"
)

func WaitMiddeleware(next http.RoundTripper) http.RoundTripper {
	return middleware.RoundTripper(func(req *http.Request) (*http.Response, error) {
		time.Sleep(time.Second)
		return next.RoundTrip(req)
	})
}

func main() {
	myC := shoot.NewRest[myclient.Client](
		shoot.BaseURL("http://localhost:8080"),
		shoot.Timeout(3000*time.Millisecond),
		shoot.EnableLogging(true),
		shoot.Use(WaitMiddeleware),
	)

	ctx := context.Background()
	_, err := myC.Set(ctx, myclient.NewKV("foo", "bar"))
	if err != nil {
		panic(err)
	}

	kv, _, err := myC.Get(ctx, "foo")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", *kv)
}
