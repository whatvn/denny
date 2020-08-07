package main

import (
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	zk "github.com/uber/jaeger-client-go/transport/zipkin"
	"github.com/uber/jaeger-client-go/zipkin"
	"github.com/whatvn/denny/middleware/http"
	"github.com/whatvn/denny/middleware/http/ot"
	"io/ioutil"
	http_lib "net/http"

	"github.com/whatvn/denny"
	"github.com/whatvn/denny/middleware"
)

type xController struct {
	denny.Controller
}

func (x xController) Handle(ctx *denny.Context) {
	var logger = denny.GetLogger(ctx)

	logger.AddLog("log something to test log init")
	logger.WithField("x", "y")
	logger.AddLog("receive request")
	var str = "hello"
	logger.AddLog("do more thing")
	str += " world"
	logger.AddLog("finished")
	ctx.Writer.Write([]byte(str))
}

type yController struct {
	denny.Controller
}

func (y yController) Handle(ctx *denny.Context) {
	y.AddLog("receive request")
	var str = "hello"
	var logger = denny.GetLogger(ctx)
	logger.AddLog("do more thing")
	str += " denny"
	y.Infof("finished")
	cli := http_lib.Client{}
	req, _ := http_lib.NewRequest("GET", "http://127.0.0.1:8081/", nil)
	var span, ok = ot.GetSpan(ctx)
	if ok {
		opentracing.GlobalTracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(req.Header))
		defer func() {
			span.Finish()
		}()
	}

	logger.AddLog("headers", req.Header)
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

	authorized := server.NewGroup("/")
	// per group middleware! in this case we use the custom created
	// AuthRequired() middleware just in the "authorized" group.
	authorized.Use(http.Logger())
	{
		authorized.Controller("/login", denny.HttpGet, &xController{})
		authorized.Controller("/logout", denny.HttpGet, &xController{})
		authorized.Controller("/profile", denny.HttpGet, &xController{})
		// nested group
	}
	opentracing.SetGlobalTracer(trace)

	server.Use(http.Logger()).Use(ot.RequestTracer())
	server.Controller("/", denny.HttpGet, &xController{})
	server.Controller("/denny", denny.HttpGet, &yController{})
	server.GraceFulStart()
}
