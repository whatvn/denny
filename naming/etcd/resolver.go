package etcd

import (
	"context"
	"errors"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/whatvn/denny/naming"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/resolver"
	"strings"
)

// NewResolver is alias to New(), and also register resolver automatically
// so client does not have to call register resolver everytime
func NewResolver(etcdAddrs, serviceName string) naming.Registry {
	registry := New(etcdAddrs, serviceName)
	resolver.Register(registry)
	return registry
}

// Build implements grpc Builder.Build method so grpc client know how to construct resolver Builder
func (r *etcd) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	if r.cli == nil {
		return nil, errors.New("etcd client was not initialised")
	}
	r.cc = cc
	r.WithFields(map[string]interface{}{
		"scheme":   target.Scheme,
		"endpoint": target.Endpoint,
	})
	keyPrefix := "/" + target.Scheme + "/" + target.Endpoint + "/"
	go r.watch(keyPrefix)
	return r, nil
}

// Scheme implements Builder.Scheme method to get prefix hint for grpc resolver
func (r etcd) Scheme() string {
	return naming.Prefix
}

// SvcName is shortcut for client's user, it return full service url
// so clients does not have to construct service url themself
func (r etcd) SvcName() string {
	return r.Scheme() + ":///" + r.serviceName
}

// ResolveNow force grpc clients to resolve service address immediately
// it's TODO implementation
func (r etcd) ResolveNow(rn resolver.ResolveNowOptions) {
	// will force to update address list immediately
}

// Close closes the resolver.
func (r etcd) Close() {
	_ = r.cli.Close()
}

func (r *etcd) watch(keyPrefix string) {
	var addrList []resolver.Address

	resp, err := r.cli.Get(context.Background(), keyPrefix, clientv3.WithPrefix())
	if err != nil {
		r.Errorf("error %v", err)
	} else {
		for i := range resp.Kvs {
			addrList = append(addrList, resolver.Address{Addr: strings.TrimPrefix(string(resp.Kvs[i].Key), keyPrefix)})
		}
	}

	r.cc.UpdateState(resolver.State{
		Addresses: addrList,
	})

	rch := r.cli.Watch(context.Background(), keyPrefix, clientv3.WithPrefix())
	for n := range rch {
		for _, ev := range n.Events {
			addr := strings.TrimPrefix(string(ev.Kv.Key), keyPrefix)
			switch ev.Type {
			case mvccpb.PUT:
				if !naming.Exist(addrList, addr) {
					addrList = append(addrList, resolver.Address{Addr: addr})
					r.cc.UpdateState(resolver.State{Addresses: addrList})
				}
			case mvccpb.DELETE:
				if s, ok := naming.Remove(addrList, addr); ok {
					addrList = s
					r.cc.UpdateState(resolver.State{Addresses: addrList})
				}
			}
		}
	}
}
