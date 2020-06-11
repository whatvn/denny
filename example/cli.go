package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/whatvn/denny/example/protobuf"
	"google.golang.org/grpc"
)

func main() {

	conn, _ := grpc.Dial("127.0.0.1:8080", grpc.WithInsecure())
	client := pb.NewHelloServiceClient(conn)
	response, err := client.SayHelloAnonymous(context.Background(), &empty.Empty{})
	fmt.Println(response, err)

	response, err = client.SayHelloAnonymous(context.Background(), &empty.Empty{})
	fmt.Println(response, err)

	response, err = client.SayHelloAnonymous(context.Background(), &empty.Empty{})
	fmt.Println(response, err)
}
