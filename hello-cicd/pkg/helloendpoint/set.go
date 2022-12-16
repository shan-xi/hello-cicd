package helloendpoint

import (
	"context"
	"hello-cicd/pkg/helloservice"
	"time"

	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/log"
)

// Set collects all of the endpoints that compose an hello service. It's meant to
// be used as a helper struct, to collect all of the endpoints into a single
// parameter.
type Set struct {
	SayHelloEndpoint endpoint.Endpoint
}

// New returns a Set that wraps the provided server, and wires in all of the
// expected endpoint middlewares via the various parameters.
func New(svc helloservice.Service, logger log.Logger, duration metrics.Histogram) Set {
	var sayHelloEndpoint endpoint.Endpoint
	{
		sayHelloEndpoint = MakeSayHelloEndpoint(svc)
		// SayHello is limited to 1 request per second with burst of 1 request.
		// Note, rate is defined as a time interval between requests.
		sayHelloEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(sayHelloEndpoint)
		sayHelloEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(sayHelloEndpoint)
		sayHelloEndpoint = LoggingMiddleware(log.With(logger, "method", "SayHello"))(sayHelloEndpoint)
		sayHelloEndpoint = InstrumentingMiddleware(duration.With("method", "SayHello"))(sayHelloEndpoint)
	}
	return Set{
		SayHelloEndpoint: sayHelloEndpoint,
	}
}

// SayHello implements the service interface, so Set may be used as a service.
// This is primarily useful in the context of a client library.
func (s Set) SayHello(ctx context.Context, a string) (string, error) {
	resp, err := s.SayHelloEndpoint(ctx, SayHelloRequest{A: a})
	if err != nil {
		return "", err
	}
	response := resp.(SayHelloResponse)
	return response.V, response.Err
}

// MakeSayHelloEndpoint constructs a SayHello endpoint wrapping the service.
func MakeSayHelloEndpoint(s helloservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(SayHelloRequest)
		v, err := s.SayHello(ctx, req.A)
		return SayHelloResponse{V: v, Err: err}, nil
	}
}

// compile time assertions for our response types implementing endpoint.Failer
var (
	_ endpoint.Failer = SayHelloResponse{}
)

// SayHelloRequest collects the request parameters for the SayHello method.
type SayHelloRequest struct {
	A string
}

// SayHelloResponse collects the response values for the SayHello method.
type SayHelloResponse struct {
	V   string `json:"v"`
	Err error  `json:"-"` // should be intercepted by Failed/errorEncoder
}

// Failed implements Failer.
func (r SayHelloResponse) Failed() error { return r.Err }
