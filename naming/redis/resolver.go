package redis

import (
	"errors"
	"github.com/whatvn/denny/naming"
	"google.golang.org/grpc/resolver"
	"strings"
	"time"
)

// NewResolver is alias to New(), and also register resolver automatically
// so client does not have to call register resolver everytime
func NewResolver(redisAddr, redisPwd, serviceName string) naming.Registry {
	registry := New(redisAddr, redisPwd, serviceName)
	resolver.Register(registry)
	return registry
}

func (r *redis) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
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

func (r *redis) watch(keyPrefix string) {
	var (
		addrList []resolver.Address
		ticker   = time.NewTicker(time.Duration(5) * time.Second)
	)

	list, err := r.addressList(keyPrefix, addrList)
	if err != nil {
		r.Errorf("cannot get address list: %v", err)
	} else {
		addrList = list[:]
	}

	r.cc.UpdateState(resolver.State{
		Addresses: addrList,
	})

	for {
		select {
		case _ = <-ticker.C:
			updatedList, err := r.addressList(keyPrefix, addrList)
			if err != nil {
				r.Errorf("cannot get address list: %v", err)
			} else {
				needUpdate := false
				// append to state list if it's not exist
				for _, addr := range updatedList {
					if !naming.Exist(addrList, addr.Addr) {
						needUpdate = true
						addrList = append(addrList, addr)
					}
				}

				// remove dead peer
				for _, addr := range addrList {
					if !naming.Exist(updatedList, addr.Addr) {
						needUpdate = true
						if s, ok := naming.Remove(addrList, addr.Addr); ok {
							addrList = s
						}
					}
				}

				if needUpdate {
					r.cc.UpdateState(resolver.State{Addresses: addrList})
				}
			}

		}

	}
}

func (r *redis) addressList(keyPrefix string, addrList []resolver.Address) ([]resolver.Address, error) {
	resp, err := r.cli.Keys(keyPrefix + "*").Result()
	if err != nil {
		return nil, err
	}
	for _, key := range resp {
		addrList = append(addrList, resolver.Address{Addr: strings.TrimPrefix(key, keyPrefix)})
	}
	return addrList, nil
}

func (r redis) Scheme() string {
	return naming.Prefix
}

func (r redis) SvcName() string {
	return r.Scheme() + ":///" + r.serviceName
}

func (r *redis) ResolveNow(resolver.ResolveNowOptions) {
	// TODO
}

func (r *redis) Close() {
	_ = r.cli.Close()
}
