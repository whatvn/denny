package redis

import (
	redisCli "github.com/go-redis/redis"
	"github.com/whatvn/denny/log"
	"github.com/whatvn/denny/naming"
	"google.golang.org/grpc/resolver"
)

type redis struct {
	cli *redisCli.Client
	*log.Log
	shutdown    chan interface{}
	cc          resolver.ClientConn
	serviceName string
}

func New(redisAddr, redisPassword, serviceName string) naming.Registry {
	client := redisCli.NewClient(&redisCli.Options{
		Addr:     redisAddr,
		Password: redisPassword,
	})

	registry := &redis{
		cli:         client,
		Log:         log.New(),
		serviceName: serviceName,
	}
	registry.WithField("redis", redisAddr)
	if len(redisPassword) > 0 {
		registry.WithField("redisPassword", redisPassword)
	}
	return registry
}
