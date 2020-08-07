package denny

import (
	"context"
	"github.com/whatvn/denny/log"
)

func GetLogger(ctx context.Context) *log.Log {
	var (
		logger interface{}
	)
	if ctx, ok := ctx.(*Context); ok {
		logger, ok := ctx.Get(log.LogKey)
		if !ok {
			logger := log.New()
			ctx.Set(log.LogKey, logger)
			return logger
		}
		return logger.(*log.Log)
	}
	logger, ok := ctx.Value(log.LogKey).(*log.Log)
	if !ok {
		logger := log.New()
		ctx = context.WithValue(
			ctx,
			log.LogKey, logger)
		return logger
	}
	return logger.(*log.Log)
}
