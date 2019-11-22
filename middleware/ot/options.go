package ot

import (
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
)

type OptionFunc func(*options)

func SetOperationNameFn(fn func(*gin.Context) string) OptionFunc {
	return func(opts *options) {
		opts.operationNameFn = fn
	}
}

func SetErrorFn(fn func(*gin.Context) bool) OptionFunc {
	return func(opts *options) {
		opts.errorFn = fn
	}
}

func SetResourceNameFn(fn func(*gin.Context) string) OptionFunc {
	return func(opts *options) {
		opts.resourceNameFn = fn
	}
}

func SetBeforeHook(fn func(opentracing.Span, *gin.Context)) OptionFunc {
	return func(opts *options) {
		opts.beforeHook = fn
	}
}

func SetAfterHook(fn func(opentracing.Span, *gin.Context)) OptionFunc {
	return func(opts *options) {
		opts.afterHook = fn
	}
}

func (opts *options) handleDefaultOptions() {
	if opts.operationNameFn == nil {
		opts.operationNameFn = func(ctx *gin.Context) string {
			return "gin.request"
		}
	}

	if opts.errorFn == nil {
		opts.errorFn = func(ctx *gin.Context) bool {
			return ctx.Writer.Status() >= 400 || len(ctx.Errors) > 0
		}
	}

	if opts.resourceNameFn == nil {
		opts.resourceNameFn = func(ctx *gin.Context) string {
			return ctx.HandlerName()
		}
	}

	if opts.beforeHook == nil {
		opts.beforeHook = func(span opentracing.Span, ctx *gin.Context) {
			return
		}
	}

	if opts.afterHook == nil {
		opts.afterHook = func(span opentracing.Span, ctx *gin.Context) {
			return
		}
	}
}
