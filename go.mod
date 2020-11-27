module github.com/whatvn/denny

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/bitly/go-simplejson v0.5.0
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/coreos/etcd v3.3.13+incompatible
	github.com/envoyproxy/protoc-gen-validate v0.1.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/ghodss/yaml v1.0.0
	github.com/gin-gonic/gin v1.6.3
	github.com/go-redis/redis v6.15.8+incompatible
	github.com/golang/protobuf v1.4.2
	github.com/google/uuid v1.1.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/hashicorp/hcl v1.0.0
	github.com/imdario/mergo v0.3.8
	github.com/opentracing/opentracing-go v1.1.0
	github.com/sirupsen/logrus v1.6.0
	github.com/soheilhy/cmux v0.1.4
	github.com/stretchr/testify v1.6.1
	github.com/uber/jaeger-client-go v2.20.1+incompatible
	github.com/whatvn/discovery v0.0.0-20200624103305-206667ab8840
	go.etcd.io/etcd v3.3.22+incompatible
	golang.org/x/net v0.0.0-20200625001655-4c5254603344
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.24.0
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace (
	github.com/codahale/hdrhistogram => github.com/HdrHistogram/hdrhistogram-go v1.0.0
	github.com/coreos/etcd => github.com/ozonru/etcd v3.3.20-grpc1.27-origmodule+incompatible
	github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0
)
