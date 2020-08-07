package main

import (
	"github.com/whatvn/denny"
	"golang.org/x/net/context"
)

func logfunc(ctx context.Context) {
	logger := denny.GetLogger(ctx)

	logger.AddLog("logfunc")
}

func main() {

	ctx := context.Background()
	logger := denny.GetLogger(ctx)
	logger.AddLog("mainfunc")
	logfunc(ctx)

	logger.Infof("finish")

	//registry := redis.NewResolver("127.0.0.1:6379", "", "demo.brpc.svc")
	//conn, err := grpc.Dial(registry.SvcName(), naming.DefaultBalancePolicy(), grpc.WithInsecure())
	//if err != nil {
	//	panic(err)
	//}
	//client := pb.NewHelloServiceClient(conn)
	//response, err := client.SayHelloAnonymous(context.Background(), &empty.Empty{})
	//fmt.Println(response, err)

}
