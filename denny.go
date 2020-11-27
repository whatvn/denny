package denny

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/soheilhy/cmux"
	"github.com/whatvn/denny/log"
	"github.com/whatvn/denny/middleware"
	"github.com/whatvn/denny/naming"
	"google.golang.org/grpc"
)

type (
	Context          = gin.Context
	HandleFunc       = gin.HandlerFunc
	methodHandlerMap struct {
		method  HttpMethod
		handler HandleFunc
	}
	group struct {
		path        string
		cors        bool
		routerGroup *gin.RouterGroup
		handlerMap  map[string]*methodHandlerMap
		engine      *Denny
	}

	Denny struct {
		sync.Mutex
		*log.Log
		handlerMap map[string]*methodHandlerMap
		groups     []*group
		*gin.Engine
		initialised     bool
		validator       binding.StructValidator
		notFoundHandler HandleFunc
		noMethodHandler HandleFunc
		grpcServer      *grpc.Server
		// for naming registry/dicovery
		registry naming.Registry
	}
)

var (
	invalidMethodType   = errors.New("invalid method signature")
	notFoundHandlerFunc = func(ctx *Context) {
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"path":   ctx.Request.RequestURI,
				"status": http.StatusNotFound,
				"method": ctx.Request.Method,
			},
		)
	}

	underlyContextType = reflect.TypeOf(new(context.Context)).Elem()
	underlyErrorType   = reflect.TypeOf(new(error)).Elem()
)

const (
	SelectCapital  = "([a-z])([A-Z])"
	ReplaceCapital = "$1 $2"
)

func newGroup(path string, routerGroup *gin.RouterGroup) *group {
	return &group{path: path, routerGroup: routerGroup}
}

// NewServer init denny with default parameter
func NewServer(debug ...bool) *Denny {
	if len(debug) == 0 || !debug[0] {
		gin.SetMode(gin.ReleaseMode)
	}
	return &Denny{
		handlerMap:      make(map[string]*methodHandlerMap),
		groups:          []*group{},
		Engine:          gin.New(),
		Log:             log.New(),
		initialised:     false,
		notFoundHandler: notFoundHandlerFunc,
		noMethodHandler: notFoundHandlerFunc,
	}
}

// WithRegistry makes Denny discoverable via naming registry
func (r *Denny) WithRegistry(registry naming.Registry) *Denny {
	r.registry = registry
	return r
}

// WithGrpcServer turns Denny into grpc server
func (r *Denny) WithGrpcServer(server *grpc.Server) *Denny {
	if server == nil {
		panic("server is not initialised")
	}
	r.grpcServer = server
	return r
}

// Controller register a controller with given path, method to http routes
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

// NewGroup adds new group into server routes
func (r *Denny) NewGroup(path string) *group {
	r.Lock()
	defer r.Unlock()
	routerGroup := r.Group(path)
	ng := newGroup(path, routerGroup)
	ng.engine = r
	r.groups = append(r.groups, ng)
	return ng
}

// same with WithMiddleware
func (g *group) Use(handleFunc HandleFunc) *group {
	g.routerGroup.Use(handleFunc)
	return g
}

// Controller is the same with router Controller, but register a controller with given path within group
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

func toKebabCase(input string, rule ...string) string {

	re := regexp.MustCompile(SelectCapital)
	input = re.ReplaceAllString(input, ReplaceCapital)

	input = strings.Join(strings.Fields(strings.TrimSpace(input)), " ")

	rule = append(rule, ".", " ", "_", " ", "-", " ")

	replacer := strings.NewReplacer(rule...)
	input = replacer.Replace(input)
	return strings.ToLower(strings.Join(strings.Fields(input), "-"))
}

func httpMethod(method reflect.Method) HttpMethod {
	in := method.Type.In(2)
	if in == reflect.TypeOf(&empty.Empty{}) {
		return HttpGet
	}
	return HttpPost
}

func httpRouterPath(controllerName string, method reflect.Method) string {
	return toKebabCase(controllerName) + "/" + toKebabCase(method.Name)
}

// BrpcController register a grpc service implements as multiple http enpoints
// endpoint usually start with service name (class name), and end with method name
// so if your have your service name: Greeting and have method hi, when registered with server
// under group v1, your http endpoint will be /v1/greeting/hi
func (g *group) BrpcController(controllerGroup interface{}) {
	g.registerHttpController(controllerGroup)
}

func (g *group) WithCors() {
	g.cors = true
}

func cors() HandleFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Origin, Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
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

		// check if request has validate function enabled
		if v, ok := req.Interface().(middleware.IValidator); ok {
			if err := v.Validate(); err != nil {
				c.JSON(http.StatusBadRequest, err)
				return
			}
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
	if g.cors {
		g.routerGroup.OPTIONS(path, cors())
	}
	switch httpMethod {
	case HttpPost:
		g.routerGroup.POST(path, cors(), handlerFunc)
		break
	case HttpGet:
		g.routerGroup.GET(path, cors(), handlerFunc)
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
	r.RemoveExtraSlash = true
	r.NoRoute(r.notFoundHandler)
	r.NoMethod(r.noMethodHandler)
	r.WithMiddleware(gin.Recovery())
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

// WithMiddleware registers middleware to http server
// only gin style middleware is supported
func (r *Denny) WithMiddleware(middleware ...HandleFunc) {
	r.Use(middleware...)
}

// WithNotFoundHandler registers middleware to http server to serve not found route
// only gin style middleware is supported
func (r *Denny) WithNotFoundHandler(handler HandleFunc) {
	r.notFoundHandler = handler
}

// Start http server with given address
// Deprecated: use GraceFulStart(addrs ...string) instead.
func (r *Denny) Start(addrs ...string) error {
	r.initRoute()
	return r.Run(addrs...)
}

// SetValidator overwrites default gin validate with provides validator
// we're using v10 validator
func (r *Denny) SetValidator(v binding.StructValidator) *Denny {
	r.validator = v
	return r
}

// GraceFulStart uses net http standard server
// it also detect if grpc server and discovery registry are available
// to start Denny in brpc mode, in this mode, server will support both protocol using same port
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
		addr         = r.resolveAddress(addrs)
		ip           string
	)

	r.initRoute()
	enableHttp := len(r.Handlers) > 0 || len(r.groups) > 0

	if enableBrpc {
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			r.Fatalf("listen: %v\n", err)
		}

		// register service into registered registry
		if r.registry != nil {
			ip, err = localIp()
			if err != nil {
				panic(err)
			}
			if err = r.registry.Register(ip+addr, 5); err != nil {
				panic(err)
			}
		}

		// Create a cmux.
		muxer = cmux.New(listener)
		// Match connections in order:
		// First grpc, then HTTP, and otherwise Go RPC/TCP.
		grpcListener = muxer.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
		httpListener = muxer.Match(cmux.HTTP1Fast())

		wg := sync.WaitGroup{}
		if enableHttp {
			wg.Add(1)
		}
		if enableBrpc {
			wg.Add(1)
		}

		go func() {
			r.Info("start grpc server ", addr)
			wg.Done()
			if err := r.grpcServer.Serve(grpcListener); err != nil {
				r.Fatalf("listen: %v\n", err)
			}
		}()

		if enableHttp {
			go func() {
				r.Info("start http server ", addr)
				wg.Done()
				if err := httpSrv.Serve(httpListener); err != nil && err != http.ErrServerClosed {
					r.Fatalf("listen: %v\n", err)
				}
			}()
		}

		go func() {
			if err = muxer.Serve(); err != nil {
				r.Fatalf("listen: %v\n", err)
			}
		}()

	} else {
		go func() {
			// service connections
			httpSrv.Addr = addr
			r.Info("start http server ", addr)
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				r.Fatalf("listen: %v\n", err)
			}
		}()
	}

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	<-quit

	r.Infof("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if r.registry != nil {
		r.Infof("unregister from registry")
		_ = r.registry.UnRegister(ip + addr)
	}
	if r.grpcServer != nil {
		r.Infof("stop grpc server")
		r.grpcServer.GracefulStop()
	}

	if enableHttp {
		r.Infof("stop http server")
		_ = httpSrv.Shutdown(ctx)
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

func localIp() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok &&
			!ipnet.IP.IsLoopback() {
			ipv4 := ipnet.IP.To4()
			if ipv4 != nil && strings.Index(ipv4.String(), "127") != 0 {
				return ipv4.String(), nil
			}
		}
	}
	return "", errors.New("cannot lookup local ip address")
}
