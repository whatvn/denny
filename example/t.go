package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/whatvn/denny"
)

type HandlerF gin.HandlerFunc

func tHandler(ctx *denny.Context)  {
	fmt.Println("ginHandler")
}

func convertHandlerFunc(h HandlerF) gin.HandlerFunc {
	var temp interface{} = h
	return temp.(gin.HandlerFunc)
}

func main()  {
	h := convertHandlerFunc(tHandler)
	ctx := gin.Context{}
	h(&ctx)
}