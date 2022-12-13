package main

import (
	"context"
	"fmt"
	"hello-cicd/pkg/helloservice"
	"hello-cicd/pkg/hellotransport"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"google.golang.org/grpc"
)

func main() {
	var useHttp = true
	var useGrpc = false
	var httpAddr = "35.247.65.245:8081"
	var grpcAddr = "35.247.65.245:8082"
	var param = "spin"

	var (
		svc helloservice.Service
		err error
	)
	if useHttp {
		svc, err = hellotransport.NewHTTPClient(httpAddr, log.NewNopLogger())
	} else if useGrpc {
		conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure(), grpc.WithTimeout(time.Second))
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
	fmt.Fprintf(os.Stdout, "%d  %d\n", param, v)
}
