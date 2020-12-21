package etcd

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/whatvn/denny/go_config/source"
	cetcd "go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/pkg/transport"
)

// Currently a single etcd reader
type etcd struct {
	path   string
	opts   source.Options
	client *cetcd.Client
	cerr   error
}

func (c *etcd) Read() (*source.ChangeSet, error) {
	if c.cerr != nil {
		return nil, c.cerr
	}

	rsp, err := c.client.Get(context.Background(), c.path)
	if err != nil {
		return nil, err
	}

	if rsp == nil || len(rsp.Kvs) == 0 {
		return nil, fmt.Errorf("source not found: %s", c.path)
	}

	b := rsp.Kvs[0].Value

	cs := &source.ChangeSet{
		Timestamp: time.Now(),
		Source:    c.String(),
		Data:      b,
		Format:    c.opts.Encoder.String(),
	}
	cs.Checksum = cs.Sum()

	return cs, nil
}

func (c *etcd) String() string {
	return "etcd"
}

func (c *etcd) Watch() (source.Watcher, error) {
	if c.cerr != nil {
		return nil, c.cerr
	}
	cs, err := c.Read()
	if err != nil {
		return nil, err
	}
	return newWatcher(c.path, c.client.Watcher, cs, c.opts)
}

func NewSource(opts ...source.Option) source.Source {
	options := source.NewOptions(opts...)

	var endpoints []string

	// check if there are any addrs
	addrs, ok := options.Context.Value(addressKey{}).([]string)
	if ok {
		for _, a := range addrs {
			addr, port, err := net.SplitHostPort(a)
			if ae, ok := err.(*net.AddrError); ok && ae.Err == "missing port in address" {
				port = "2379"
				addr = a
				endpoints = append(endpoints, fmt.Sprintf("%s:%s", addr, port))
			} else if err == nil {
				endpoints = append(endpoints, fmt.Sprintf("%s:%s", addr, port))
			}
		}
	}

	if len(endpoints) == 0 {
		endpoints = []string{"localhost:2379"}
	}

	// check dial timeout option
	dialTimeout, ok := options.Context.Value(dialTimeoutKey{}).(time.Duration)
	if !ok {
		dialTimeout = 3 * time.Second // default dial timeout
	}

	config := cetcd.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	}

	u, ok := options.Context.Value(basicAuthKey{}).(*basicAuthCreds)
	if ok {
		config.Username = u.Username
		config.Password = u.Password
	}

	tls, ok := options.Context.Value(tlsAuthKey{}).(*tlsAuthCreds)
	if ok {
		var (
			cfgTLS *transport.TLSInfo
			err    error
		)
		cfgTLS = &transport.TLSInfo{
			CertFile: tls.CertFile,
			KeyFile:  tls.KeyFile,
			CAFile:   tls.CAFile,
		}
		config.TLS, err = cfgTLS.ClientConfig()
		if err != nil {
			panic(err)
		}
	}

	// use default config
	client, err := cetcd.New(config)

	path, ok := options.Context.Value(pathKey{}).(string)

	if !ok {
		panic("cannot setup etcd source with empty path")
	}

	if strings.HasSuffix(path, "/") {
		panic("etcd path cannot be directory")
	}

	return &etcd{
		path:   path,
		opts:   options,
		client: client,
		cerr:   err,
	}
}
