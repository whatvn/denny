package http

import (
	"context"
	"fmt"
	"github.com/whatvn/denny"
	"github.com/whatvn/denny/log"
	"time"
)

func Logger() denny.HandleFunc {
	return func(ctx *denny.Context) {
		logger := log.New(&log.JSONFormatter{})
		var (
			clientIP = ctx.ClientIP()
			method   = ctx.Request.Method

			userAgent = ctx.Request.UserAgent()
			uri       = ctx.Request.URL
			errs      string
			start     = time.Now()
			isError bool
		)

		logger.WithFields(map[string]interface{}{
			"client_ip":      clientIP,
			"request_method": method,
			"user_agent":     userAgent,
			"uri":            uri,
		})
		ctx.Request = ctx.Request.WithContext(context.WithValue(ctx, log.LogKey, logger))
		ctx.Set(log.LogKey, logger)
		ctx.Next()
		var (
			statusCode = ctx.Writer.Status()
		)
		logger.WithField("Status", statusCode)
		if ctx.Errors != nil {
			isError = true
			bs, err := ctx.Errors.MarshalJSON()
			if err == nil {
				errs = string(bs)
			} else {
				errs = ctx.Errors.String()
			}
		}
		if len(errs) > 0 {
			logger.WithField("Errors", errs)
		}
		end := time.Now()
		logger.WithField("end", end)
		msg := fmt.Sprintf("latency: %v", end.Sub(start))
		if isError {
			logger.Error(msg)
		} else {
			logger.Info(msg)
		}
	}
}
