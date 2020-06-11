package denny

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/soheilhy/cmux"
	"github.com/whatvn/denny/log"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
)

type Context = gin.Context

var invalidMethodType = errors.New("invalid method signature")

type HandleFunc = gin.HandlerFunc

type methodHandlerMap struct {
	method  HttpMethod
	handler HandleFunc
}

type group struct {
	path        string
	routerGroup *gin.RouterGroup
	handlerMap  map[string]*methodHandlerMap
	engine      *Denny
}

func newGroup(path string, routerGroup *gin.RouterGroup) *group {
	return &group{path: path, routerGroup: routerGroup}
}

type Denny struct {
	sync.Mutex
	*log.Log
	handlerMap map[string]*methodHandlerMap
	groups     []*group
	*gin.Engine
	initialised bool
	validator   binding.StructValidator
	grpcServer  *grpc.Server
}

func NewServer(debug ...bool) *Denny {
	if len(debug) == 0 || !debug[0] {
		gin.SetMode(gin.ReleaseMode)
	}
	return &Denny{
		handlerMap:  make(map[string]*methodHandlerMap),
		groups:      []*group{},
		Engine:      gin.New(),
		Log:         log.New(),
		initialised: false,
	}
}

func (r *Denny) WithGrpcServer(server *grpc.Server) {
	if server == nil {
		panic("server is not initialised")
	}
	r.grpcServer = server
}

func (r *Denny) Controller(path string, method HttpMethod, ctl controller) *Denny {
	r.Lock()
	defer r.Unlock()
	if r.validator != nil {
		ctl.SetValidator(r.validator)
	}
	m := &methodHandlerMap{
		method: method,
		handler: func(ctx *Context) {
			ctl.init()
			ctl.Handle(ctx)
		},
	}

	r.handlerMap[path] = m
	return r
}

func (r *Denny) NewGroup(path string) *group {
	r.Lock()
	defer r.Unlock()
	routerGroup := r.Group(path)
	ng := newGroup(path, routerGroup)
	ng.engine = r
	r.groups = append(r.groups, ng)
	return ng
}

func (g *group) Use(handleFunc HandleFunc) *group {
	g.routerGroup.Use(handleFunc)
	return g
}

func (g *group) Controller(path string, method HttpMethod, ctl controller) *group {
	if g.engine.validator != nil {
		ctl.SetValidator(g.engine.validator)
	}
	m := &methodHandlerMap{
		method: method,
		handler: func(ctx *Context) {
			ctl.init()
			ctl.Handle(ctx)
		},
	}
	if g.handlerMap == nil {
		g.handlerMap = make(map[string]*methodHandlerMap)
	}

	g.handlerMap[path] = m
	return g
}

func httpMethod(method reflect.Method) HttpMethod {
	in := method.Type.In(2)
	if in == reflect.TypeOf(&empty.Empty{}) {
		return HttpGet
	}
	return HttpPost
}

func httpRouterPath(controllerName string, method reflect.Method) string {
	return strings.ToLower(controllerName + "/" + method.Name)
}

var underlyContextType = reflect.TypeOf(new(context.Context)).Elem()
var underlyErrorType = reflect.TypeOf(new(error)).Elem()

func (g *group) BrpcController(controllerGroup interface{}) {
	g.registerHttpController(controllerGroup)
}

func (g *group) registerHttpController(controllerGroup interface{}) {
	var (
		controllerReferenceType              = reflect.TypeOf(controllerGroup)
		controllerReferenceValue             = reflect.ValueOf(controllerGroup)
		indirectControllerReferenceValueType = reflect.Indirect(controllerReferenceValue).Type()
		controllerName                       = indirectControllerReferenceValueType.Name()
	)
	// Install the methods
	for m := 0; m < controllerReferenceType.NumMethod(); m++ {

		method := controllerReferenceType.Method(m)

		if method.Type.NumIn() != 3 {
			panic(invalidMethodType)
		}

		groupType, contextType, requestType := method.Type.In(0), method.Type.In(1), method.Type.In(2)

		if groupType.Kind() != reflect.Ptr {
			panic(invalidMethodType)
		}

		if contextType.Kind() != reflect.Ptr && contextType.Kind() != reflect.Interface {
			panic(invalidMethodType)
		}

		if !contextType.Implements(underlyContextType) {
			panic(invalidMethodType)
		}

		if requestType.Kind() != reflect.Ptr {
			panic(invalidMethodType)
		}

		if method.Type.NumOut() != 2 {
			panic(invalidMethodType)
		}

		outErrorType, outResponseType := method.Type.Out(1), method.Type.Out(0)
		if outErrorType != underlyErrorType {
			panic(invalidMethodType)
		}

		if outResponseType.Kind() != reflect.Ptr {
			panic(invalidMethodType)
		}

		routerPath, httpMethod := httpRouterPath(controllerName, method), httpMethod(method)
		g.registerHandler(controllerReferenceValue, method, routerPath, httpMethod)

	}
}

func unmarshal(ctx *Context, in interface{}) error {
	return ctx.ShouldBind(in)
}

// getCaller extract grpc service implementation into gin http handlerFunc
func getCaller(fn, obj reflect.Value) (func(*gin.Context), error) {
	var (
		funcType    = fn.Type()
		requestType = funcType.In(2)
	)

	handlerFunc := func(c *gin.Context) interface{} { return c }

	reqIsValue := true
	if requestType.Kind() == reflect.Ptr {
		reqIsValue = false
	}
	return func(c *Context) {
		req := reflect.New(requestType)
		if !reqIsValue {
			req = reflect.New(requestType.Elem())
		}
		if err := unmarshal(c, req.Interface()); err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		if reqIsValue {
			req = req.Elem()
		}

		var vals []reflect.Value
		// call grpc service with provided method
		// obj is service which implements grpc service interface
		vals = fn.Call([]reflect.Value{obj, reflect.ValueOf(handlerFunc(c)), req})

		if vals != nil {
			response, err := vals[0].Interface(), vals[1].Interface()
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err.(error))
				return
			}
			c.JSON(http.StatusOK, response)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{})
	}, nil
}

func handlerFuncObj(function, obj reflect.Value) gin.HandlerFunc {
	call, err := getCaller(function, obj)
	if err != nil {
		panic(err)
	}

	return call
}

func (g *group) registerHandler(
	controllerReferenceValue reflect.Value,
	method reflect.Method, path string, httpMethod HttpMethod) {
	handlerFunc := handlerFuncObj(method.Func, controllerReferenceValue)
	switch httpMethod {
	case HttpPost:
		g.engine.POST(path, handlerFunc)
		break
	case HttpGet:
		g.engine.GET(path, handlerFunc)
		break
	default:
		panic("not implemetation")
	}
}

func (r *Denny) initRoute() {
	if r.initialised {
		return
	}
	for p, m := range r.handlerMap {
		setupHandler(m, r, p)
	}

	for _, g := range r.groups {
		for p, m := range g.handlerMap {
			setupHandler(m, g.routerGroup, p)
		}
	}
	r.initialised = true
}

func setupHandler(m *methodHandlerMap, router gin.IRouter, p string) {
	switch m.method {
	case HttpGet:
		router.GET(p, m.handler)
	case HttpPost:
		router.POST(p, m.handler)
	case HttpDelete:
		router.DELETE(p, m.handler)
	case HttpOption:
		router.OPTIONS(p, m.handler)
	case HttpPatch:
		router.PATCH(p, m.handler)
	}
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

func (r *Denny) SetValidator(v binding.StructValidator) *Denny {
	r.validator = v
	return r
}

// gracefulStart uses net http standard server
// and register channel listen to os signals to make it graceful stop
func (r *Denny) GraceFulStart(addrs ...string) error {
	var (
		httpSrv = &http.Server{
			Handler: r,
		}
		enableBrpc   = r.grpcServer != nil
		listener     net.Listener
		grpcListener net.Listener
		httpListener net.Listener
		muxer        cmux.CMux
		err          error
	)

	r.initRoute()

	if enableBrpc {
		listener, err = net.Listen("tcp", "127.0.0.1:8080")
		if err != nil {
			return err
		}
		// Create a cmux.
		muxer = cmux.New(listener)
		// Match connections in order:
		// First grpc, then HTTP, and otherwise Go RPC/TCP.
		grpcListener = muxer.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
		httpListener = muxer.Match(cmux.HTTP1Fast())

		go func() {
			r.Info("start grpc server...")
			if err := r.grpcServer.Serve(grpcListener); err != nil {
				r.Fatalf("listen: %v\n", err)
			}
		}()

		go func() {
			r.Info("start http server...")
			if err := httpSrv.Serve(httpListener); err != nil && err != http.ErrServerClosed {
				r.Fatalf("listen: %v\n", err)
			}
		}()

		fmt.Println("what the fuck")
		if err = muxer.Serve(); err != nil {
			panic(err)
		}

	} else {
		go func() {
			// service connections
			r.Info("start http server...")
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				r.Fatalf("listen: %v\n", err)
			}
		}()
	}

	//// Wait for interrupt signal to gracefully shutdown the server with
	//// a timeout of 5 seconds.
	//quit := make(chan os.Signal, 1)
	//// kill (no param) default send syscanll.SIGTERM
	//// kill -2 is syscall.SIGINT
	//// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	//signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	//<-quit
	//r.Infof("Shutdown Server ...")
	//
	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()
	//if err := httpSrv.Shutdown(ctx); err != nil {
	//	r.Fatal("Server Shutdown: ", err)
	//	return err
	//}
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
