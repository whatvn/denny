package ot

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

var spanKey = "span"

type options struct {
	beforeHook      func(opentracing.Span, *gin.Context)
	afterHook       func(opentracing.Span, *gin.Context)
	operationNameFn func(*gin.Context) string
	errorFn         func(*gin.Context) bool
	resourceNameFn  func(*gin.Context) string
}

func RequestTracer(opts ...OptionFunc) gin.HandlerFunc {
	mwOptions := &options{}
	for _, opt := range opts {
		opt(mwOptions)
	}

	mwOptions.handleDefaultOptions()

	return func(c *gin.Context) {
		var (
			span    opentracing.Span = opentracing.StartSpan(c.FullPath())
			tracer                   = opentracing.GlobalTracer()
			carrier                  = opentracing.HTTPHeadersCarrier(c.Request.Header)
		)
		ctx, err := tracer.Extract(opentracing.HTTPHeaders, carrier)

		if err == nil {
			span = opentracing.StartSpan(c.FullPath(), opentracing.ChildOf(ctx))
		}

		ext.HTTPMethod.Set(span, c.Request.Method)
		ext.HTTPUrl.Set(span, c.Request.URL.String())
		span.SetTag("resource.name", mwOptions.resourceNameFn(c))
		c.Request = c.Request.WithContext(opentracing.ContextWithSpan(c, span))
		mwOptions.beforeHook(span, c)
		c.Next()
		mwOptions.afterHook(span, c)
		ext.Error.Set(span, mwOptions.errorFn(c))
		ext.HTTPStatusCode.Set(span, uint16(c.Writer.Status()))
		span.Finish()
	}
}

func GetSpan(c context.Context) (opentracing.Span, bool) {
	var span = opentracing.SpanFromContext(c)
	if span != nil {
		return span, true
	}
	return nil, false
}
