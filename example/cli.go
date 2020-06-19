package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/whatvn/denny/example/protobuf"
	"github.com/whatvn/denny/naming"
	"github.com/whatvn/denny/naming/redis"
	"google.golang.org/grpc"
)

func main() {

	registry := redis.NewResolver("127.0.0.1:6379", "", "demo.brpc.svc")
	conn, err := grpc.Dial(registry.SvcName(), naming.DefaultBalancePolicy(), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := pb.NewHelloServiceClient(conn)
	response, err := client.SayHelloAnonymous(context.Background(), &empty.Empty{})
	fmt.Println(response, err)

}
