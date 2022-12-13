package helloendpoint

import (
	"context"
	"hello-cicd/pkg/helloservice"
	"time"

	"golang.org/x/time/rate"

	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/sony/gobreaker"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
)

type Set struct {
	SayHelloEndpoint endpoint.Endpoint
}

func New(svc helloservice.Service, logger log.Logger, duration metrics.Histogram, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) Set {
	var sayHelloEndpoint endpoint.Endpoint
	{
		sayHelloEndpoint = MakeSayHelloEndpoint(svc)
		sayHelloEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(sayHelloEndpoint)
		sayHelloEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(sayHelloEndpoint)
		sayHelloEndpoint = opentracing.TraceServer(otTracer, "SayHello")(sayHelloEndpoint)
		if zipkinTracer != nil {
			sayHelloEndpoint = zipkin.TraceEndpoint(zipkinTracer, "SayHello")(sayHelloEndpoint)
		}
		sayHelloEndpoint = LoggingMiddleware(log.With(logger, "method", "SayHello"))(sayHelloEndpoint)
		sayHelloEndpoint = InstrumentingMiddleware(duration.With("method", "SayHello"))(sayHelloEndpoint)
	}
	return Set{
		SayHelloEndpoint: sayHelloEndpoint,
	}
}

func (s Set) SayHello(ctx context.Context, a string) (string, error) {
	resp, err := s.SayHelloEndpoint(ctx, SayHelloRequest{A: a})
	if err != nil {
		return "", err
	}
	response := resp.(SayHelloResponse)
	return response.V, response.Err
}

func MakeSayHelloEndpoint(s helloservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(SayHelloRequest)
		v, err := s.SayHello(ctx, req.A)
		return SayHelloResponse{V: v, Err: err}, nil
	}
}

var (
	_ endpoint.Failer = SayHelloResponse{}
)

type SayHelloRequest struct {
	A string
}

type SayHelloResponse struct {
	V   string `json:"v"`
	Err error  `json:"-"` // should be intercepted by Failed/errorEncoder
}

func (r SayHelloResponse) Failed() error { return r.Err }
