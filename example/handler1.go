package main

import (
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	zk "github.com/uber/jaeger-client-go/transport/zipkin"
	"github.com/uber/jaeger-client-go/zipkin"
	"github.com/whatvn/denny/middleware/ot"

	"github.com/whatvn/denny"
	"github.com/whatvn/denny/middleware"
)

type zController struct {
	denny.Controller
}

func (z zController) Handle(ctx *denny.Context) {
	z.AddLog("receive request")
	var str = "hello"
	z.AddLog("do more thing")
	str += " denny"
	z.Infof("finished")
	ctx.Writer.Write([]byte(str))
}

func reporter() jaeger.Transport {
	transport, err := zk.NewHTTPTransport(
		"http://10.109.3.93:9411/api/v1/spans",
		zk.HTTPBatchSize(10),
		zk.HTTPLogger(jaeger.StdLogger),
	)
	if err != nil {
		panic(err)
	}
	return transport
}

func main() {
	server := denny.NewServer()
	propagator := zipkin.NewZipkinB3HTTPHeaderPropagator()
	trace, closer := jaeger.NewTracer(
		"server-1",
		jaeger.NewConstSampler(true),
		jaeger.NewRemoteReporter(reporter()),
		jaeger.TracerOptions.Injector(opentracing.HTTPHeaders, propagator),
		jaeger.TracerOptions.Extractor(opentracing.HTTPHeaders, propagator),
		jaeger.TracerOptions.ZipkinSharedRPCSpan(true),
	)
	defer closer.Close()
	opentracing.SetGlobalTracer(trace)

	server.Use(middleware.Logger()).Use(ot.RequestTracer())
	server.Controller("/", denny.HttpGet, &zController{})
	server.GraceFulStart(":8081")
}
