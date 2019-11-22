package main

import (
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	zk "github.com/uber/jaeger-client-go/transport/zipkin"
	"github.com/uber/jaeger-client-go/zipkin"
	"github.com/whatvn/denny/middleware/ot"
	"io/ioutil"
	"net/http"

	"github.com/whatvn/denny"
	"github.com/whatvn/denny/middleware"
)

type xController struct {
	denny.Controller
}

func (x xController) Handle(ctx *denny.Context) {
	x.Infof("log something to test log init")
	x.WithField("x", "y")
	x.AddLog("receive request")
	var str = "hello"
	x.AddLog("do more thing")
	str += " world"
	x.Infof("finished")
	ctx.Writer.Write([]byte(str))
}

type yController struct {
	denny.Controller
}

func (y yController) Handle(ctx *denny.Context) {
	var span = ot.GetSpan(ctx)
	y.AddLog("receive request")
	var str = "hello"
	y.AddLog("do more thing")
	str += " denny"
	y.Infof("finished")
	cli := http.Client{}
	req, _ := http.NewRequest("GET", "http://127.0.0.1:8081/", nil)
	opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header))
	y.Infof("headers %v", req.Header)
	response, _ := cli.Do(req.WithContext(ctx))
	bytes, _ := ioutil.ReadAll(response.Body)
	ctx.Writer.Write([]byte(str + " remote " + string(bytes)))
}

func newReporter() jaeger.Transport {
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
		"api_gateway",
		jaeger.NewConstSampler(true),
		jaeger.NewRemoteReporter(newReporter()),
		jaeger.TracerOptions.Injector(opentracing.HTTPHeaders, propagator),
		jaeger.TracerOptions.Extractor(opentracing.HTTPHeaders, propagator),
		jaeger.TracerOptions.ZipkinSharedRPCSpan(true),
	)
	defer closer.Close()
	opentracing.SetGlobalTracer(trace)

	server.Use(middleware.Logger()).Use(ot.RequestTracer())
	server.Controller("/", denny.HttpGet, &xController{})
	server.Controller("/denny", denny.HttpGet, &yController{})
	server.GraceFulStart()
}
