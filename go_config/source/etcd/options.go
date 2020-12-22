package etcd

import (
	"context"
	"time"

	"github.com/whatvn/denny/go_config/source"
)

type addressKey struct{}
type pathKey struct{}
type basicAuthKey struct{}
type tlsAuthKey struct{}
type dialTimeoutKey struct{}

type basicAuthCreds struct {
	Username string
	Password string
}

type tlsAuthCreds struct {
	CAFile   string
	CertFile string
	KeyFile  string
}

// WithAddress sets the etcd address
func WithAddress(a ...string) source.Option {
	return func(o *source.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, addressKey{}, a)
	}
}

// WithPrefix sets the key prefix to use
func WithPath(p string) source.Option {
	return func(o *source.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, pathKey{}, p)
	}
}

// BasicAuth allows you to specify username/password
func BasicAuth(username, password string) source.Option {
	return func(o *source.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, basicAuthKey{}, &basicAuthCreds{Username: username, Password: password})
	}
}

// TLSAuth allows you to specify cafile, certfile, keyfile
func TLSAuth(caFile, certFile, keyFile string) source.Option {
	return func(o *source.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, tlsAuthKey{}, &tlsAuthCreds{CAFile: caFile, CertFile: certFile, KeyFile: keyFile})
	}
}

// WithDialTimeout set the time out for dialing to etcd
func WithDialTimeout(timeout time.Duration) source.Option {
	return func(o *source.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, dialTimeoutKey{}, timeout)
	}
}
