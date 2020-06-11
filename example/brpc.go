package main

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/whatvn/denny"
	pb "github.com/whatvn/denny/example/protobuf"
	"github.com/whatvn/denny/middleware"
	"google.golang.org/grpc"
)

// grpc
type Hello struct{}

func (s *Hello) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {

	response := &pb.HelloResponse{
		Reply: "hi",
	}

	return response, nil
}

func (s *Hello) SayHelloAnonymous(ctx context.Context, in *empty.Empty) (*pb.HelloResponse, error) {

	response := &pb.HelloResponse{
		Reply: "hi",
	}

	return response, nil
}
func main() {
	server := denny.NewServer(true)
	server.Use(middleware.Logger())

	// setup grpc server
	grpcServer := grpc.NewServer()
	pb.RegisterHelloServiceServer(grpcServer, new(Hello))
	//l, _ := net.Listen("tcp", ":8080")
	//grpcServer.Serve(l)

	server.WithGrpcServer(grpcServer)
	//

	//// then http
	authorized := server.NewGroup("/")
	authorized.BrpcController(&Hello{})

	// start server in dual mode
	server.GraceFulStart()
}
