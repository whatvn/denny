package denny

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/whatvn/denny/log"
)

type controller interface {
	Handle(*Context)
	init()
}

type Controller struct {
	binding.StructValidator
	*log.Log
}

func (c *Controller) init() {
	c.Log = log.New()
	c.StructValidator = binding.Validator
}
