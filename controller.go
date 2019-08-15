package denny

import (
	"github.com/whatvn/denny/log"
)

type controller interface {
	Handle(*Context)
}

type Controller struct {
	log.Log
}
