package denny

import (
	"github.com/gin-gonic/gin"
	"github.com/whatvn/denny/log"
	"sync"
)


type Handler struct {
	gin.Context

}

type methodMap struct {
	method  HttpMethod
	handler gin.HandlerFunc
	*log.Log
}

type route struct {
	sync.Mutex
	handlerMap map[string]*methodMap
	*gin.Engine
}

func NewRouter() *route {
	return &route{
		handlerMap: make(map[string]*methodMap),
		Engine:     gin.New(),
	}
}

func (r *route) Add(path string, method HttpMethod, handler gin.HandlerFunc) {
	r.Lock()
	defer r.Unlock()
	m := &methodMap{
		method:  method,
		handler: handler,
		Log:     log.New(path),
	}
	m.Infof("setting up handler for path %s, method %V", path, method)
	r.handlerMap[path] = m
}


func (r *route) initRoute() {
	for p, m := range r.handlerMap {
		switch m.method {
		case HttpGet:
			r.GET(p, m.handler)
		case HttpPost:
			r.POST(p, m.handler)
		case HttpDelete:
			r.DELETE(p, m.handler)
		case HttpOption:
			r.OPTIONS(p, m.handler)
		case HttpPatch:
			r.PATCH(p, m.handler)
		}
	}
}

func (r *route) WithMiddleware(middleware ...gin.HandlerFunc) {
	r.Use(middleware...)
}

func (r *route) Start(addrs ...string) error {
	r.initRoute()
	return r.Run(addrs...)
}
