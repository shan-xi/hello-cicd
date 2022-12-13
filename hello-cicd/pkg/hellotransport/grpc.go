package hellotransport

import (
	"context"
	"errors"
	"hello-cicd/pb"
	"hello-cicd/pkg/helloendpoint"
	"hello-cicd/pkg/helloservice"
	"time"

	"google.golang.org/grpc"

	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/kit/transport"
	grpctransport "github.com/go-kit/kit/transport/grpc"
)

type grpcServer struct {
	sayHello grpctransport.Handler
}

func NewGRPCServer(endpoints helloendpoint.Set, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) pb.HelloServer {
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}

	if zipkinTracer != nil {
		options = append(options, zipkin.GRPCServerTrace(zipkinTracer))
	}

	return &grpcServer{
		sayHello: grpctransport.NewServer(
			endpoints.SayHelloEndpoint,
			decodeGRPCSayHelloRequest,
			encodeGRPCSayHelloResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(otTracer, "SayHello", logger)))...,
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

func NewGRPCClient(conn *grpc.ClientConn, logger log.Logger) helloservice.Service {
	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 100))
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

	return helloendpoint.Set{
		SayHelloEndpoint: sayHelloEndpoint,
	}
}

func decodeGRPCSayHelloRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.SayHelloRequest)
	return helloendpoint.SayHelloRequest{A: req.A}, nil
}

func decodeGRPCSayHelloResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.SayHelloReply)
	return helloendpoint.SayHelloResponse{V: reply.V, Err: str2err(reply.Err)}, nil
}

func encodeGRPCSayHelloResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(helloendpoint.SayHelloResponse)
	return &pb.SayHelloReply{V: resp.V, Err: err2str(resp.Err)}, nil
}

func encodeGRPCSayHelloRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(helloendpoint.SayHelloRequest)
	return &pb.SayHelloRequest{A: req.A}, nil
}

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
