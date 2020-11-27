module github.com/whatvn/denny

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/HdrHistogram/hdrhistogram-go v1.0.1 // indirect
	github.com/bitly/go-simplejson v0.5.0
	github.com/coreos/etcd v3.3.13+incompatible
	github.com/envoyproxy/protoc-gen-validate v0.1.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/ghodss/yaml v1.0.0
	github.com/gin-gonic/gin v1.6.3
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.1.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/hashicorp/hcl v1.0.0
	github.com/imdario/mergo v0.3.8
	github.com/kr/pretty v0.2.1 // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/sirupsen/logrus v1.6.0
	github.com/soheilhy/cmux v0.1.4
	github.com/stretchr/testify v1.6.1
	github.com/uber/jaeger-client-go v2.20.1+incompatible
	github.com/uber/jaeger-lib v2.4.0+incompatible // indirect
	github.com/whatvn/discovery v1.0.0
	go.etcd.io/etcd v3.3.22+incompatible
	golang.org/x/net v0.0.0-20201006153459-a7d1128ccaa0
	google.golang.org/grpc v1.33.1
	google.golang.org/protobuf v1.25.0
)

replace (
	github.com/codahale/hdrhistogram => github.com/HdrHistogram/hdrhistogram-go v1.0.0
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5
	github.com/coreos/etcd => github.com/ozonru/etcd v3.3.20-grpc1.27-origmodule+incompatible
	github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0
	google.golang.org/grpc => google.golang.org/grpc v1.27.0
)
