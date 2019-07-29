package denny

import (
	"github.com/whatvn/denny/log"
	"time"
)

func Logger() HandleFunc {
	logger := log.New()
	return func(ctx *Context) {
		var (
			clientIP   = ctx.ClientIP()
			method     = ctx.Request.Method
			statusCode = ctx.Writer.Status()
			userAgent  = ctx.Request.UserAgent()
		)

		logger.WithField("ClientIP", clientIP)
		logger.WithField("RequestMethod", method)
		logger.WithField("Status", statusCode)
		logger.WithField("UserAgent", userAgent)
		logger.Infof(time.Now().Format(time.RFC3339))
		ctx.Error()
	}
}
