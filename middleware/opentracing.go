package middleware

import (
	"github.com/opentracing-contrib/go-gin/ginhttp"
	"github.com/opentracing/opentracing-go"
	"github.com/whatvn/denny"
)



// setting up open tracing middleware
// example usage
// https://github.com/opentracing-contrib/go-gin/blob/master/examples/example_test.go
func OpenTracing(tr opentracing.Tracer, options ...ginhttp.MWOption) denny.HandleFunc {
	return ginhttp.Middleware(tr, options...)
}