module github.com/whatvn/denny

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/bitly/go-simplejson v0.5.0
	github.com/coreos/etcd v0.0.0-00010101000000-000000000000
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-00010101000000-000000000000 // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/fsnotify/fsnotify v1.4.7
	github.com/ghodss/yaml v1.0.0
	github.com/gin-gonic/gin v1.5.0
	github.com/go-redis/redis v6.15.8+incompatible
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.4.1
	github.com/google/uuid v1.1.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/hashicorp/hcl v1.0.0
	github.com/imdario/mergo v0.3.8
	github.com/opentracing/opentracing-go v1.1.0
	github.com/sirupsen/logrus v1.4.2
	github.com/soheilhy/cmux v0.1.4
	github.com/stretchr/testify v1.4.0
	github.com/uber/jaeger-client-go v2.20.1+incompatible // indirect
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	go.etcd.io/etcd v3.3.22+incompatible
	go.uber.org/zap v1.13.0 // indirect
	google.golang.org/grpc v1.27.0
	google.golang.org/protobuf v1.24.0
)

replace (
	github.com/coreos/etcd => github.com/ozonru/etcd v3.3.20-grpc1.27-origmodule+incompatible
	github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0
)
