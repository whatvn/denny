package middleware

import (
	"github.com/whatvn/denny"
	"github.com/whatvn/denny/log"
	"time"
)

func Logger() denny.HandleFunc {
	logger := log.New()
	return func(ctx *denny.Context) {
		var (
			clientIP   = ctx.ClientIP()
			method     = ctx.Request.Method
			statusCode = ctx.Writer.Status()
			userAgent  = ctx.Request.UserAgent()
			uri        = ctx.Request.RequestURI
			errs       string
		)

		if ctx.Errors != nil {
			bs, err := ctx.Errors.MarshalJSON()
			if err == nil {
				errs = string(bs)
			}
		}
		logger.WithFields(map[string]interface{}{
			"ClientIP":      clientIP,
			"RequestMethod": method,
			"Status":        statusCode,
			"UserAgent":     userAgent,
			"Uri":           uri,
		})
		if len(errs) > 0 {
			logger.WithField("Errors", errs)
		}
		logger.Infof(time.Now().Format(time.RFC3339))
	}
}
