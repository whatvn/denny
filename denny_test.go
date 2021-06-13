package denny

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	pb "github.com/whatvn/denny/example/protobuf"
	"github.com/whatvn/denny/middleware/grpc"
	"github.com/whatvn/denny/naming"
	"github.com/whatvn/denny/naming/etcd"
	"go.etcd.io/etcd/clientv3"
	grpcClient "google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type header struct {
	Key   string
	Value string
}

var mockTime = time.Date(2021, time.May, 16, 23, 19, 0, 0, time.UTC)

// grpc
type Hello struct{}

// groupPath + "/hello/" + "sayhello"
//
func (s *Hello) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	var (
		logger = GetLogger(ctx)
	)
	response := &pb.HelloResponse{
		Reply:     "hi",
		CreatedAt: timestamppb.New(mockTime),
	}

	logger.WithField("response", response)
	return response, nil
}

func (s *Hello) SayHelloAnonymous(ctx context.Context, in *empty.Empty) (*pb.HelloResponseAnonymous, error) {
	var (
		logger = GetLogger(ctx)
	)

	span, ctx := opentracing.StartSpanFromContext(ctx, "sayHello")
	defer span.Finish()
	response := &pb.HelloResponseAnonymous{
		Reply:     "hoho",
		Status:    pb.Status_STATUS_SUCCESS,
		CreatedAt: timestamppb.New(mockTime),
	}

	logger.WithField("response", response)

	return response, nil
}

func performRequest(r http.Handler, method, path string, headers ...header) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for _, h := range headers {
		req.Header.Add(h.Key, h.Value)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestSimpleRequest(t *testing.T) {
	signature := ""
	server := NewServer()
	server.Use(func(c *Context) {
		signature += "A"
		c.Next()
		signature += "B"
	})
	server.Use(func(c *Context) {
		signature += "C"
	})
	server.GET("/", func(c *Context) {
		signature += "D"
		c.String(http.StatusOK, signature)
	})
	server.NoRoute(func(c *Context) {
		signature += " X "
	})
	server.NoMethod(func(c *Context) {
		signature += " XX "
	})
	// RUN
	w := performRequest(server, "GET", "/")

	out, e := ioutil.ReadAll(w.Result().Body)

	// TEST
	assert.Equal(t, nil, e)
	assert.Equal(t, "ACD", string(out))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ACDB", signature)
}

func TestCustomJsonMarshal(t *testing.T) {
	server := NewServer(true)

	// Add your custom JSON response serializer
	AddProtoJsonResponseSerializer(
		ProtoJsonResponseSerializer(protojson.MarshalOptions{ // You can use the Proto json serializer with protojson.MarshalOptions
			Indent:          "  ",
			Multiline:       true,
			UseProtoNames:   false,
			EmitUnpopulated: true,
			UseEnumNumbers:  false,
		}))

	// setup grpc server
	grpcServer := NewGrpcServer(grpc.ValidatorInterceptor)
	pb.RegisterHelloServiceServer(grpcServer, new(Hello))
	server.WithGrpcServer(grpcServer)
	//

	//// then http
	authorized := server.NewGroup("/")
	authorized.BrpcController(&Hello{})

	// RUN
	w := performRequest(server, "GET", "/hello/say-hello-anonymous")

	out, e := ioutil.ReadAll(w.Result().Body)
	var res map[string]string
	err := json.Unmarshal(out, &res)
	if err != nil {
		assert.Errorf(t, err, "Error when Unmarshal response")
	}

	// TEST
	assert.Equal(t, nil, e)

	reply, ok := res["reply"]
	if !ok {
		assert.Errorf(t,
			fmt.Errorf("Not found reply field in response map"),
			"Not found reply field in response", res)
	}
	assert.Equal(t, "hoho", reply)

	status, ok := res["status"]
	if !ok {
		assert.Errorf(t,
			fmt.Errorf("Not found status field in response map"),
			"Not found status field in response", res)
	}
	assert.Equal(t, "STATUS_SUCCESS", status)

	createdAtField, ok := res["createdAt"] // The timestamp is now in RFC3339 format, and the fields are camelCased.
	if !ok {
		assert.Errorf(t,
			fmt.Errorf("Not found status field in response map"),
			"Not found status field in response", res)
	}
	createdAt, err := time.Parse(time.RFC3339, createdAtField)
	if err != nil {
		assert.Errorf(t, err, "Parse createAt field error", res)
	}
	assert.Equal(t, mockTime.Equal(createdAt), true)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNaming(t *testing.T) {
	server := NewServer(true)

	// setup grpc server
	grpcServer := NewGrpcServer(grpc.ValidatorInterceptor)
	pb.RegisterHelloServiceServer(grpcServer, new(Hello))
	server.WithGrpcServer(grpcServer)
	//

	//// then http
	authorized := server.NewGroup("/")
	authorized.BrpcController(&Hello{})

	// RUN
	clientCfgServer := clientv3.Config{
		Endpoints:   strings.Split("58.84.1.31:2379", ";"),
		Username:    "root",
		Password:    "phuc12345",
		DialTimeout: 15 * time.Second,
	}
	registryServer := etcd.NewWithClientConfig("bevo.profile", clientCfgServer)
	server.WithRegistry(registryServer)

	// start server in dual mode
	go server.GraceFulStart(":8081")

	// Run client
	clientCfg := clientv3.Config{
		Endpoints:   strings.Split("58.84.1.31:2379", ";"),
		Username:    "root",
		Password:    "phuc12345",
		DialTimeout: 15 * time.Second,
	}
	registry := etcd.NewResolverWithClientConfig("bevo.profile", clientCfg)
	conn, err := grpcClient.Dial(registry.SvcName(), naming.DefaultBalancePolicy(), grpcClient.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := pb.NewHelloServiceClient(conn)
	response, err := client.SayHelloAnonymous(context.Background(), &empty.Empty{})
	fmt.Println(response, err)
	assert.Equal(t, "hoho", response.Reply)
}
