package httpbp_test

import (
	"context"
	"errors"
	"net/http"
	"time"

	baseplate "github.com/reddit/baseplate.go"
	"github.com/reddit/baseplate.go/httpbp"
	"github.com/reddit/baseplate.go/log"
	"github.com/reddit/baseplate.go/secrets"
)

type body struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Handlers struct {
	secrets *secrets.Store
}

func (h Handlers) Home(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	response := httpbp.Response{
		Body: body{
			X: 1,
			Y: 2,
		},
	}
	return httpbp.WriteJSON(w, response)
}

func (h Handlers) ServerErr(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return httpbp.JSONError(httpbp.InternalServerError(), errors.New("example"))
}

func (h Handlers) Ratelimit(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return httpbp.JSONError(
		httpbp.TooManyRequests().Retryable(w, time.Minute),
		errors.New("rate-limit"),
	)
}

func (h Handlers) InvalidInput(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return httpbp.JSONError(
		httpbp.BadRequest().WithDetails(map[string]string{
			"foo": "must be >= 0",
			"bar": "must be non-nil",
		}),
		errors.New("invalid-input"),
	)
}

func (h Handlers) Endpoints() map[httpbp.Pattern]httpbp.Endpoint {
	return map[httpbp.Pattern]httpbp.Endpoint{
		"/":              {Name: "home", Handle: h.Home},
		"/err":           {Name: "err", Handle: h.ServerErr},
		"/ratelimit":     {Name: "ratelimit", Handle: h.Ratelimit},
		"/invalid-input": {Name: "invalid-input", Handle: h.InvalidInput},
	}
}

func loggingMiddleware(name string, next httpbp.HandlerFunc) httpbp.HandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		log.Infof("Request %q: %#v", name, r)
		return next(ctx, w, r)
	}
}

var (
	_ httpbp.Middleware = loggingMiddleware
)

func ExampleNewBaseplateServer() {
	ctx := context.Background()
	bp, err := baseplate.New(ctx, "example.yaml")
	if err != nil {
		panic(err)
	}
	defer bp.Close()

	handlers := Handlers{bp.Secrets()}
	server, err := httpbp.NewBaseplateServer(httpbp.ServerArgs{
		Baseplate:   bp,
		Endpoints:   handlers.Endpoints(),
		Middlewares: []httpbp.Middleware{loggingMiddleware},
	})
	if err != nil {
		panic(err)
	}
	log.Info(baseplate.Serve(ctx, server))
}
