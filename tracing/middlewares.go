package tracing

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// InjectHTTPServerSpan returns a go-kit endpoint.Middleware that injects a server
// span into the `next` context.
//
// Note, this depends on the edge context headers already being set on the
// context object.  This can be done by adding httpbp.PopulateRequestContext as
// a ServerBefore option when setting up the request handler for an endpoint.
func InjectHTTPServerSpan(name string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			ctx, _ = StartSpanFromHTTPContext(ctx, name)
			return next(ctx, request)
		}
	}
}

// InjectHTTPServerSpanWithTracer is the same as InjectHTTPServerSpan except it
// uses StartSpanFromHTTPContextWithTracer to initialize the server span rather
// than StartSpanFromHTTPContext.
func InjectHTTPServerSpanWithTracer(tracer *Tracer, name string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			ctx, _ = StartSpanFromHTTPContextWithTracer(ctx, name, tracer)
			return next(ctx, request)
		}
	}
}