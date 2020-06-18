package naming

import "google.golang.org/grpc/resolver"

const (
	// PREFIX uses here to differentiate between denny etcd prefix and other service prefix
	// in etcd directory/files
	Prefix = "_DENNY_"
)

type Registry interface {
	Register(addr string, ttl int) error
	UnRegister(serviceName, addr string) error
	Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error)
	Scheme() string
	SvcName() string
}
