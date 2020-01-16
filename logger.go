package denny

import (
	"fmt"
	"github.com/whatvn/denny/log"
)

const (
	logKey = "dennyLogger"
)

func GetLogger(ctx *Context) *log.Log {
	logger, ok := ctx.Get(logKey)
	if !ok {
		return log.New()
	}
	if l, ok := logger.(*log.Log); ok {
		return l
	}
	panic(fmt.Errorf("%v is not logger", logger))
}
