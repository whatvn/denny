package main

import (
	"github.com/whatvn/denny"
	"github.com/whatvn/denny/middleware"
)

type xController struct {
	denny.Controller
}

func (x xController) Handle(ctx *denny.Context) {
	x.AddLog("receive request")
	var str = "hello"
	x.AddLog("do more thing")
	str += " world"
	ctx.Writer.Write([]byte(str))
	x.Infof("finished")
}

type yController struct {
	denny.Controller
}

func (y yController) Handle(ctx *denny.Context) {
	y.AddLog("receive request")
	var str = "hello"
	y.AddLog("do more thing")
	str += " denny"
	ctx.Writer.Write([]byte(str))
	y.Infof("finished")
}

func main() {
	server := denny.NewServer()
	server.WithMiddleware(middleware.Logger())
	server.Controller("/", denny.HttpGet, &xController{})
	server.Controller("/denny", denny.HttpGet, &yController{})
	server.Start()
}
