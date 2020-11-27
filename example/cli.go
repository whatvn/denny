package main

import (
	"fmt"

	pb "github.com/whatvn/denny/example/protobuf"
	"github.com/whatvn/denny/naming"
	"github.com/whatvn/discovery/redis"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {

	registry := redis.NewResolver("127.0.0.1:6379", "", "demo.brpc.svc")
	conn, err := grpc.Dial(registry.SvcName(), naming.DefaultBalancePolicy(), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := pb.NewHelloServiceClient(conn)

	response, err := client.SayHello(context.Background(), &pb.HelloRequest{Greeting: ""})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(response, err)

}
