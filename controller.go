package denny

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/whatvn/denny/log"
)

type controller interface {
	Handle(*Context)
	init()
	SetValidator(validator binding.StructValidator)
}

type Controller struct {
	binding.StructValidator
	*log.Log
}

func (c *Controller) init() {
	c.Log = log.New()
	c.StructValidator = binding.Validator
}

func (c *Controller) SetValidator(v binding.StructValidator) {
	c.StructValidator = v
}
