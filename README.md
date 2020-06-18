# denny

common http server which simplify request handling and logging by combining libraries, framework to be able to 
- support both http and grpc in one controller, write once, support both protocol. See [example](https://github.com/whatvn/denny/blob/master/example/brpc.go)
- support class base request controller, one controller for one handler, **or** you can describe your service in grpc and impelement grpc service, denny will then support HTTP/gRPC when you start it in brpc mode, see [example](https://github.com/whatvn/denny/blob/master/example/brpc.go)
- make cache usage simpler
- use open tracing  
- make config reader simpler
- make logger attached in request context, log should be showed as steps and in only one line for every request


`denny` is not a http/grpc server from scratch, by now it's based on [gin framework](https://github.com/gin-gonic/gin) and google grpc server, with help of Golang reflection to support both protocol while requires user just minimum about of work. Eq: you just need to write your service in grpc, `denny` will also support HTTP for your implementation. 

`denny` is different from [grpc gateway](https://github.com/grpc-ecosystem/grpc-gateway), grpc gateway uses code generation to generate http proxy call, a request to http enpoint of grpc gateway will also trigger another grpc call to your service. it's http proxy to grpc, with `denny`, a call to http will only invoke the code you wrote, does not trigger grpc call. It applies also for grpc call. Because of that, using grpc your service has to start with 2 services port, 1 for http and 1 for grpc, `denny` need only one for both protocol.

It also borrow many component from well known libraries (go-config, beego, logrus...).  


## usage example

### Register discoverable grpc service 

```go

func main() {

	server := denny.NewServer(true)
	// setup grpc server

	grpcServer := denny.NewGrpcServer()
	pb.RegisterHelloServiceServer(grpcServer, new(Hello))
	server.WithGrpcServer(grpcServer)
	
	registry := etcd.New("127.0.0.1:7379", "demo.brpc.svc")
	server.WithRegistry(registry)

	// start server in dual mode
	server.GraceFulStart(":8081")
}
``` 


### Connect to grpc server using grpc resolver 

```go

package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/whatvn/denny/example/protobuf"
	"github.com/whatvn/denny/naming/etcd"
	"google.golang.org/grpc"
)

func main() {

	registry := etcd.NewResolver("127.0.0.1:7379", "demo.brpc.svc")
	conn, err := grpc.Dial(registry.SvcName(), etcd.DefaultBalancePolicy(), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := pb.NewHelloServiceClient(conn)
	response, err := client.SayHelloAnonymous(context.Background(), &empty.Empty{})
	fmt.Println(response, err)
}
```

### Write grpc code but support both http/grpc
```go
package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/zipkin"
	"github.com/whatvn/denny"
	pb "github.com/whatvn/denny/example/protobuf"
	"github.com/whatvn/denny/middleware/http"
	"github.com/whatvn/denny/naming/etcd"
	"io"
)

// grpc
type Hello struct{}

// groupPath + "/hello/" + "sayhello"
//
func (s *Hello) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	var (
		logger = denny.GetLogger(ctx)
	)
	response := &pb.HelloResponse{
		Reply: "hi",
	}

	logger.WithField("response", response)
	return response, nil
}

// http get request
// when define grpc method with input is empty.Empty object, denny will consider request as get request
// router will be:
// groupPath + "/hello/" + "sayhelloanonymous"
// rule is rootRoute + "/" lowerCase(serviceName) + "/" lowercase(methodName)

func (s *Hello) SayHelloAnonymous(ctx context.Context, in *empty.Empty) (*pb.HelloResponse, error) {

	var (
		logger = denny.GetLogger(ctx)
	)

	span, ctx := opentracing.StartSpanFromContext(ctx, "sayHello")
	defer span.Finish()
	response := &pb.HelloResponse{
		Reply: "ha",
	}

	logger.WithField("response", response)

	return response, nil
}

type TestController struct {
	denny.Controller
}

func (t *TestController) Handle(ctx *denny.Context) {
	ctx.JSON(200, &pb.HelloResponse{
		Reply: "ha",
	})
}

func newReporterUDP(jaegerAddr string, port int, packetLength int) jaeger.Transport {
	hostString := fmt.Sprintf("%s:%d", jaegerAddr, port)
	transport, err := jaeger.NewUDPTransport(hostString, packetLength)
	if err != nil {
		panic(err)
	}
	return transport
}
func initTracerUDP(jaegerAddr string, port int, packetLength int, serviceName string) (opentracing.Tracer, io.Closer) {
	var (
		propagator = zipkin.NewZipkinB3HTTPHeaderPropagator()
	)

	return jaeger.NewTracer(
		serviceName,
		jaeger.NewConstSampler(true),
		jaeger.NewRemoteReporter(newReporterUDP(jaegerAddr, port, packetLength)),
		jaeger.TracerOptions.Injector(opentracing.HTTPHeaders, propagator),
		jaeger.TracerOptions.Extractor(opentracing.HTTPHeaders, propagator),
		jaeger.TracerOptions.ZipkinSharedRPCSpan(true),
	)
}

func main() {

	// open tracing
	tracer, _ := initTracerUDP(
		"127.0.0.1",
		6831,
		65000,
		"brpc.server.demo",
	)
	opentracing.SetGlobalTracer(tracer)



	server := denny.NewServer(true)
	server.Use(http.Logger())
	group := server.NewGroup("/hi")
	group.Controller("/hi", denny.HttpPost, new(TestController))

	// setup grpc server

	grpcServer := denny.NewGrpcServer()
	pb.RegisterHelloServiceServer(grpcServer, new(Hello))
	server.WithGrpcServer(grpcServer)
	//

	//// then http
	authorized := server.NewGroup("/")
	// http://127.0.0.1:8080/hello/sayhello  (POST)
	// http://127.0.0.1:8080/hello/sayhelloanonymous  (GET)
	authorized.BrpcController(&Hello{})


	// naming registry
	registry := etcd.New("127.0.0.1:7379", "demo.brpc.svc")
	server.WithRegistry(registry)

	// start server in dual mode
	server.GraceFulStart(":8081")
}

```


### setting up simple http request handler 
```go

package main

import (
	"github.com/whatvn/denny"
	"github.com/whatvn/denny/middleware/http"
)

type xController struct {
	denny.Controller
}

// define handle function for controller  
func (x xController) Handle(ctx *denny.Context) {
	var (
		logger = denny.GetLogger(ctx)
	)
	logger.AddLog("receive request")
	var str = "hello"
	logger.AddLog("do more thing") // logger middleware will log automatically when request finished
	str += " world"
	ctx.Writer.Write([]byte(str))
}

func main() {
	server := denny.NewServer()
	server.WithMiddleware(http.Logger())
	server.Controller("/", denny.HttpGet, &xController{})
	server.Start()
}



```

### Reading config 

```go

package main

import (
	"fmt"
	"github.com/whatvn/denny/config"
	"os"
	"path/filepath"
	"time"
)

func configFile() (*os.File, error) {
	data := []byte(`{"foo": "bar", "denny": {"sister": "jenny"}}`)
	path := filepath.Join(os.TempDir(), fmt.Sprintf("file.%d", time.Now().UnixNano()))
	fh, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	_, err = fh.Write(data)
	if err != nil {
		return nil, err
	}

	return fh, nil
}

func main()  {
	f, err := configFile()
	if err != nil {
		fmt.Println(err)
	}
	// read config from file
	config.New(f.Name())
	fmt.Println(config.GetString("foo"))
	fmt.Println(config.GetString("denny", "sister"))

	// config from evn takes higher priority
	os.Setenv("foo", "barbar")
	os.Setenv("denny_sister", "Jenny")
	config.Reload()
	fmt.Println(config.GetString("foo"))
	fmt.Println(config.GetString("denny", "sister"))
}
```


# limit 

Denny uses etcd as naming registry, but etcd packaging is somewhat complicated, new version links to old version, old version links to older version which is very difficult to optimize import, so currently it use a fork version of etcd [here](https://github.com/ozonru/etcd/releases/tag/v3.3.20-grpc1.27-origmodule). Look at this [issue](https://github.com/etcd-io/etcd/issues/11721) to track 