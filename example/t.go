package main

import (
	"github.com/whatvn/denny"
	"github.com/whatvn/denny/log"
)

type xController struct {
	denny.Controller
}

func (x xController) Handle(ctx *denny.Context)  {
	x.AddLine("receive request")
	var str = "hello"
	x.AddLine("do more thing")
	str += " world"
	ctx.Writer.Write([]byte("Hello word"))
	x.Infof("finished")
}

func requestInfo() denny.HandleFunc {
	log := log.New("request info")
	return func(context *denny.Context) {
		clientIP := context.ClientIP()
		method := context.Request.Method
		statusCode := context.Writer.Status()
		log.Infof("clientIp ", clientIP, "method ", method, "status ", statusCode)
	}
}

func main()  {
	server := denny.NewServer()
	server.WithMiddleware(requestInfo())
	server.Controller("/", denny.HttpGet, &xController{})
	server.Start()
}
