package etcd

import (
	"errors"
	"github.com/whatvn/denny/log"
	"github.com/whatvn/denny/naming"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/resolver"
	"strings"
	"time"
)

type etcd struct {
	cli *clientv3.Client
	*log.Log
	shutdown    chan interface{}
	cc          resolver.ClientConn
	serviceName string
}

// etcd
// implement github.com/whatvn/denny/naming#Registry
// with 2 methods: Register and UnRegister
func New(etcdAddrs, serviceName string) naming.Registry {

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(etcdAddrs, ";"),
		DialTimeout: 15 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	if len(serviceName) == 0 {
		panic(errors.New("invalid service name"))
	}
	registry := &etcd{
		cli:         cli,
		Log:         log.New(),
		serviceName: serviceName,
		shutdown:    make(chan interface{}, 1),
	}
	registry.WithField("etcd", etcdAddrs)
	return registry
}

var _ naming.Registry = new(etcd)
