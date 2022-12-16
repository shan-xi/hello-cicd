package helloservice

import (
	"context"

	"github.com/go-kit/log"
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(Service) Service

// LoggingMiddleware takes a logger as a dependency
// and returns a service Middleware.
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return loggingMiddleware{logger, next}
	}
}

type loggingMiddleware struct {
	logger log.Logger
	next   Service
}

func (mw loggingMiddleware) SayHello(ctx context.Context, a string) (v string, err error) {
	defer func() {
		mw.logger.Log("method", "SayHello", "a", a, "v", v, "err", err)
	}()
	return mw.next.SayHello(ctx, a)
}
