package config

import (
	"github.com/whatvn/denny/go_config/source"
	"github.com/whatvn/denny/go_config/source/etcd"
)

func WithEtcdAddress(addr ...string) source.Option {
	return etcd.WithAddress(addr...)
}

func WithEtcdAuth(user, pass string) source.Option {
	return etcd.BasicAuth(user, pass)
}

func WithEtcdTLSAuth(certFile, keyFile, caFile string) source.Option {
	return etcd.TLSAuth(caFile, certFile, keyFile)
}

func WithPath(path string) source.Option {
	return etcd.WithPath(path)
}
