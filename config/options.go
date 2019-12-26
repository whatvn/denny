package config

import (
	"github.com/whatvn/denny/go_config/source"
	"github.com/whatvn/denny/go_config/source/etcd"
)

func WithEtcdAddress(addr ...string) source.Option {
	return etcd.WithAddress(addr...)
}

func WithEtdAuth(user, pass string) source.Option {
	return etcd.Auth(user, pass)
}

func WithPath(path string) source.Option {
	return etcd.WithPath(path)
}
