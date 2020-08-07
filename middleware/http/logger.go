package http

import (
	"context"
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
			bs, err := ctx.Errors.MarshalJSON()
			if err == nil {
				errs = string(bs)
			}
		}
		if len(errs) > 0 {
			logger.WithField("Errors", errs)
		}
		end := time.Now()
		logger.WithField("end", end)
		logger.Infof("latency: %v", end.Sub(start))
	}
}
