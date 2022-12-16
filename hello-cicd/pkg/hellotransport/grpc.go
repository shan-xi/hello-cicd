package hellotransport

import (
	"context"
	"errors"
	"hello-cicd/pb"
	"hello-cicd/pkg/helloendpoint"
	"hello-cicd/pkg/helloservice"
	"time"

	"google.golang.org/grpc"

	// stdopentracing "github.com/opentracing/opentracing-go"
	// stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/ratelimit"

	// "github.com/go-kit/kit/tracing/opentracing"
	// "github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/kit/transport"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/go-kit/log"
)

type grpcServer struct {
	sayHello grpctransport.Handler
}

// func NewGRPCServer(endpoints helloendpoint.Set, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) pb.HelloServer {

// NewGRPCServer makes a set of endpoints available as a gRPC AddServer.
func NewGRPCServer(endpoints helloendpoint.Set, logger log.Logger) pb.HelloServer {
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}

	// if zipkinTracer != nil {
	// 	options = append(options, zipkin.GRPCServerTrace(zipkinTracer))
	// }

	return &grpcServer{
		sayHello: grpctransport.NewServer(
			endpoints.SayHelloEndpoint,
			decodeGRPCSayHelloRequest,
			encodeGRPCSayHelloResponse,
			options...,
		),
	}
}

func (s *grpcServer) SayHello(ctx context.Context, req *pb.SayHelloRequest) (*pb.SayHelloReply, error) {
	_, rep, err := s.sayHello.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.SayHelloReply), nil
}

// NewGRPCClient returns an HelloService backed by a gRPC server at the other end
// of the conn. The caller is responsible for constructing the conn, and
// eventually closing the underlying transport. We bake-in certain middlewares,
// implementing the client library pattern.
func NewGRPCClient(conn *grpc.ClientConn, logger log.Logger) helloservice.Service {
	// We construct a single ratelimiter middleware, to limit the total outgoing
	// QPS from this client to all methods on the remote instance. We also
	// construct per-endpoint circuitbreaker middlewares to demonstrate how
	// that's done, although they could easily be combined into a single breaker
	// for the entire remote instance, too.
	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 100))

	// Each individual endpoint is an grpc/transport.Client (which implements
	// endpoint.Endpoint) that gets wrapped with various middlewares. If you
	// made your own client library, you'd do this work there, so your server
	// could rely on a consistent set of client behavior.
	var sayHelloEndpoint endpoint.Endpoint
	{
		sayHelloEndpoint = grpctransport.NewClient(
			conn,
			"pb.Hello",
			"SayHello",
			encodeGRPCSayHelloRequest,
			decodeGRPCSayHelloResponse,
			pb.SayHelloReply{},
		).Endpoint()
		sayHelloEndpoint = limiter(sayHelloEndpoint)
		sayHelloEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "SayHello",
			Timeout: 30 * time.Second,
		}))(sayHelloEndpoint)
	}

	// Returning the endpoint.Set as a service.Service relies on the
	// endpoint.Set implementing the Service methods. That's just a simple bit
	// of glue code.
	return helloendpoint.Set{
		SayHelloEndpoint: sayHelloEndpoint,
	}
}

// decodeGRPCSayHelloRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC sayHello request to a user-domain sayHello request. Primarily useful in a server.
func decodeGRPCSayHelloRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.SayHelloRequest)
	return helloendpoint.SayHelloRequest{A: req.A}, nil
}

// decodeGRPCSayHelloResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC sayHello reply to a user-domain sayHello response. Primarily useful in a client.
func decodeGRPCSayHelloResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.SayHelloReply)
	return helloendpoint.SayHelloResponse{V: reply.V, Err: str2err(reply.Err)}, nil
}

// encodeGRPCSayHelloResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain sayHello response to a gRPC sayHello reply. Primarily useful in a server.
func encodeGRPCSayHelloResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(helloendpoint.SayHelloResponse)
	return &pb.SayHelloReply{V: resp.V, Err: err2str(resp.Err)}, nil
}

// encodeGRPCSayHelloRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain sayHello request to a gRPC sayHello request. Primarily useful in a client.
func encodeGRPCSayHelloRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(helloendpoint.SayHelloRequest)
	return &pb.SayHelloRequest{A: req.A}, nil
}

// These annoying helper functions are required to translate Go error types to
// and from strings, which is the type we use in our IDLs to represent errors.
// There is special casing to treat empty strings as nil errors.

func str2err(s string) error {
	if s == "" {
		return nil
	}
	return errors.New(s)
}

func err2str(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
