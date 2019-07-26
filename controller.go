package denny

import "github.com/whatvn/denny/log"

type controller interface {
	Handle(*Context)
	init()
}

type Controller struct {
	*log.Log
}

func (c *Controller) init() {
	c.Log = log.New("xyz")
}


