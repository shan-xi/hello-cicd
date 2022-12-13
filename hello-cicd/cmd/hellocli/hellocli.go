package main

import (
	"context"
	"fmt"
	"hello-cicd/pkg/helloservice"
	"hello-cicd/pkg/hellotransport"
	"os"

	"github.com/go-kit/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	var useHTTP = false
	var useGrpc = true
	var httpAddr = "localhost:8081"
	var grpcAddr = "localhost:8082"
	var param = "spin"

	var (
		svc helloservice.Service
		err error
	)
	if useHTTP {
		svc, err = hellotransport.NewHTTPClient(httpAddr, log.NewNopLogger())
	} else if useGrpc {
		ctx := context.TODO()
		conn, err := grpc.DialContext(
			ctx,
			grpcAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v", err)
			os.Exit(1)
		}
		defer conn.Close()
		svc = hellotransport.NewGRPCClient(conn, log.NewNopLogger())
	} else {
		fmt.Fprintf(os.Stderr, "error: no remote address specified\n")
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	v, err := svc.SayHello(context.Background(), param)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%v  %v\n", param, v)
}
