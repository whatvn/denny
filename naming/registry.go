package naming

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

const (
	// PREFIX uses here to differentiate between denny etcd prefix and other service prefix
	// in etcd directory/files
	Prefix = "_DENNY_"
)

// Registry is based interface, which is composed of grpc resolver.Builder, resolver.Resolver and also
// contains method to register and unregister from naming storage
type Registry interface {
	Register(addr string, ttl int) error
	UnRegister(addr string) error
	Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error)
	Scheme() string
	SvcName() string
}

const defaultBalancingPolicy = `{"loadBalancingPolicy":"round_robin"}`

// DefaultBalancePolicy returns default grpc service config
// which required by new grpc API so client does not have to supply
// json config everytime
func DefaultBalancePolicy() grpc.DialOption {
	return grpc.WithDefaultServiceConfig(defaultBalancingPolicy)
}

func Exist(l []resolver.Address, addr string) bool {
	for i := range l {
		if l[i].Addr == addr {
			return true
		}
	}
	return false
}

func Remove(s []resolver.Address, addr string) ([]resolver.Address, bool) {
	for i := range s {
		if s[i].Addr == addr {
			s[i] = s[len(s)-1]
			return s[:len(s)-1], true
		}
	}
	return nil, false
}
