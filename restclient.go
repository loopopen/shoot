//go:generate go run github.com/lopolopen/shoot/cmd/shoot new -getset -opt -short -type=RestConf -ver=v0.0.0
package shoot

import (
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/lopolopen/shoot/middleware"
)

// RestConf holds configuration parameters for initializing a RestClient.
type RestConf struct {
	baseURL       string
	timeout       time.Duration
	enableLogging bool
	//todo: enableTrace    bool
	defaultHeaders map[string]string
	_middlewares   []middleware.Middleware
}

// BuildMiddleware constructs the middleware chain by wrapping the default HTTP transport.
func (r *RestConf) BuildMiddleware() http.RoundTripper {
	t := http.DefaultTransport
	for i := len(r._middlewares) - 1; i >= 0; i-- {
		t = r._middlewares[i](t)
	}
	if r.enableLogging {
		t = middleware.LoggingMiddleware(t)
	}
	return t
}

// Use returns an Option function that applies the given middleware to a RestConf instance.
func Use(middleware middleware.Middleware) Option[RestConf, *RestConf] {
	return func(r *RestConf) {
		r._middlewares = append(r._middlewares, middleware)
	}
}

// RestClient defines the interface for REST-capable clients.
type RestClient[T any] interface {
	ConfigHTTPClient(config func(client *http.Client)) T
	ShootRest()
}

// Register associates a constructor function with a specific RestClient implementation.
// The constructor must accept a RestConf and return a concrete type that satisfies RestClient.
func Register[T RestClient[T]](ctor func(RestConf) T) {
	typ := reflect.TypeOf((*T)(nil))
	_, ok := ctorRegistry[typ]
	if ok {
		panic(fmt.Errorf("ctor of interface %s should not be registered multiple times", typ.Elem()))
	}
	ctorRegistry[typ] = ctor
}

// NewRest creates a new instance of type T that implements the RestClient interface.
func NewRest[T RestClient[T]](opts ...Option[RestConf, *RestConf]) T {
	typ := reflect.TypeOf((*T)(nil))
	conf := NewWith(opts...)
	ctor, ok := ctorRegistry[typ]
	if !ok {
		panic(fmt.Errorf("ctor of interface %s is not regstered", typ.Elem()))
	}
	typedCtor, ok := ctor.(func(RestConf) T)
	if !ok {
		panic(fmt.Errorf("registered ctor of interface %s has invalid type %T", typ.Elem(), ctor))
	}
	t := typedCtor(*conf)
	return t
}

var ctorRegistry = map[reflect.Type]any{}
