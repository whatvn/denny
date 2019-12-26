package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

var (
	cli, _ = clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://10.109.3.26:6379"},
		DialTimeout: 100 * time.Second,
	})
)

func main() {
	//ctx := context.Background()
	//txn := cli.Txn(ctx)
	//
	//then := txn.If(clientv3.Compare(clientv3.Value("/acquiringcore/ae/config"), "!=", "df")).Then(
	//	clientv3.OpGet("/acquiringcore/", clientv3.WithPrefix()))
	//
	//txnResponse, e := then.Commit()
	//
	//if e != nil {
	//	panic(e)
	//}
	//
	//for _, kv := range txnResponse.Responses {
	//	fmt.Println(kv)
	//}
	//cli.Put(context.Background(), "/acquiringcore/ae/config", `{"ae": "def"}`)
	response, e := cli.Get(context.Background(), "/", clientv3.WithPrefix())
	if e == nil {
		values := response.Kvs
		for _, v := range values {
			fmt.Println(string(v.Value))
		}
	} else {
		fmt.Println(e)
	}

}
