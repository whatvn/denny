package main

import (
	"github.com/whatvn/denny"
	"github.com/whatvn/denny/log"
)

type xController struct {
	denny.Controller
}

func (x xController) Handle(ctx *denny.Context)  {
	x.AddLog("receive request")
	var str = "hello"
	x.AddLog("do more thing")
	str += " world"
	ctx.Writer.Write([]byte(str))
	x.Infof("finished")
}

func requestInfo() denny.HandleFunc {
	logger := log.New()
	return func(context *denny.Context) {
		clientIP := context.ClientIP()
		method := context.Request.Method
		statusCode := context.Writer.Status()
		logger.Infof("clientIp: %s method %s status %d, user agent: %s",clientIP,  method, statusCode, context.Request.UserAgent())
	}
}

func main()  {
	server := denny.NewServer()
	server.WithMiddleware(requestInfo())
	server.Controller("/", denny.HttpGet, &xController{})
	server.Start()
}
