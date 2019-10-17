package denny

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/whatvn/denny/log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Context = gin.Context

type HandleFunc = gin.HandlerFunc

type methodHandlerMap struct {
	method  HttpMethod
	handler HandleFunc
}

type Denny struct {
	sync.Mutex
	*log.Log
	handlerMap map[string]*methodHandlerMap
	*gin.Engine
	initialised bool
}

func NewServer() *Denny {
	return &Denny{
		handlerMap:  make(map[string]*methodHandlerMap),
		Engine:      gin.New(),
		Log:         log.New(),
		initialised: false,
	}
}

func (r *Denny) Controller(path string, method HttpMethod, ctl controller) {
	r.Lock()
	defer r.Unlock()
	m := &methodHandlerMap{
		method: method,
		handler: func(ctx *Context) {
			ctl.init()
			ctl.Handle(ctx)
		},
	}
	r.handlerMap[path] = m
}

func (r *Denny) initRoute() {
	if r.initialised {
		return
	}
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
	gin.SetMode(gin.ReleaseMode)
	r.initialised = true
}

// ServeHTTP conforms to the http.Handler interface.
func (r *Denny) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.initRoute()
	r.Engine.ServeHTTP(w, req)
}

func (r *Denny) WithMiddleware(middleware ...HandleFunc) {
	r.Use(middleware...)
}

func (r *Denny) Start(addrs ...string) error {
	r.initRoute()
	return r.Run(addrs...)
}

// gracefulStart uses net http standard server
// and register channel listen to os signals to make it graceful stop
func (r *Denny) GraceFulStart(addrs ...string) error {
	var (
		address = r.resolveAddress(addrs)
	)
	r.initRoute()
	srv := &http.Server{
		Addr:    address,
		Handler: r,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			r.Fatalf("listen: %v\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	r.Infof("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		r.Fatal("Server Shutdown: ", err)
		return err
	}
	return nil
}

func (r *Denny) resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		if port := os.Getenv("PORT"); port != "" {
			r.Debugf("environment variable PORT=\"%s\"", port)
			return ":" + port
		}
		r.Debug("environment variable PORT is undefined. Using port :8080 by default")
		return ":8080"
	case 1:
		return addr[0]
	default:
		panic("too many parameters")
	}
}
