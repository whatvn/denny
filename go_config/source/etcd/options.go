package etcd

import (
	"context"
	"time"

	"github.com/whatvn/denny/go_config/source"
)

type addressKey struct{}
type pathKey struct{}
type authKey struct{}
type dialTimeoutKey struct{}

type authCreds struct {
	Username string
	Password string
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

// Auth allows you to specify username/password
func Auth(username, password string) source.Option {
	return func(o *source.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, authKey{}, &authCreds{Username: username, Password: password})
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
