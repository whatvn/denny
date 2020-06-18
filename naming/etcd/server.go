package etcd

import (
	"context"
	"errors"
	"github.com/whatvn/denny/log"
	"github.com/whatvn/denny/naming"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/resolver"
	"strings"
	"time"
)

type etcd struct {
	cli         *clientv3.Client
	log         *log.Log
	addrs       string
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
		log:         log.New(),
		addrs:       etcdAddrs,
		serviceName: serviceName,
	}
	registry.log.WithField("etcd", etcdAddrs)
	return registry
}

// Register starts etcd client and check for registered key, if it's not available
// in etcd storage, it will write service host:port to etcd and start watching to keep writing
// if its data is not available again
func (r *etcd) Register(addr string, ttl int) error {

	var (
		ticker  = time.NewTicker(time.Second * time.Duration(ttl))
		err     error
		svcPath = "/" + naming.Prefix + "/" + r.serviceName + "/" + addr
	)

	r.log.Infof("register %s with registy", svcPath)
	err = r.register(r.serviceName, addr, ttl)
	if err != nil {
		r.log.Errorf("error %v", err)
	}

	go func() {
		for {
			select {
			case _ = <-ticker.C:
				resp, err := r.cli.Get(context.Background(), svcPath)
				if err != nil {
					r.log.Errorf("error %v", err)
				} else if resp.Count == 0 {
					err = r.register(r.serviceName, addr, ttl)
					if err != nil {
						r.log.Errorf("error %v", err)
					}
				}
			case _ = <-r.shutdown:
				// receive message from shutdown channel
				// will stop current thread and stop ticker to prevent thread leak
				ticker.Stop()
				return
			}
		}
	}()

	return nil
}

func (r *etcd) register(name string, addr string, ttl int) error {
	leaseResp, err := r.cli.Grant(context.Background(), int64(ttl))
	if err != nil {
		return err
	}

	_, err = r.cli.Put(context.Background(), "/"+naming.Prefix+"/"+name+"/"+addr, addr, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	_, err = r.cli.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		return err
	}
	return nil
}

// UnRegister deletes itself in etcd storage
// also stop watch goroutine and ticker inside it
func (r *etcd) UnRegister(name string, addr string) error {
	_, _ = r.cli.Delete(context.Background(), "/"+naming.Prefix+"/"+name+"/"+addr)
	r.shutdown <- "stop"
	return nil
}
