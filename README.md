# denny

common http server which simplify request handling and logging by combining libraries, framework to be able to 
- support both http and grpc in one controller, write once, support both protocol. See [example](https://github.com/whatvn/denny/blob/master/example/brpc.go)
- use class base request controller, one controller for one handler 
- make cache usage simpler
- use open tracing  
- make config reader simpler
- make logger attached in controller, log should be showed as steps and in only one line for every request 


`denny` is not a http server from scratch, by now it's based on [gin framework](https://github.com/gin-gonic/gin) (currently) but aim to allow user to switch framework by configuration. 
It also borrow many component from well known libraries (go-config, beego, logrus...).  


## usage example

### setting up request handler 
```go

package main

import (
	"github.com/whatvn/denny"
	"github.com/whatvn/denny/middleware"
)

type xController struct {
	denny.Controller
}

// define handle function for controller  
func (x xController) Handle(ctx *denny.Context) {
	x.AddLog("receive request")
	var str = "hello"
	x.AddLog("do more thing") // logger middleware will log automatically when request finished
	str += " world"
	ctx.Writer.Write([]byte(str))
}

func main() {
	server := denny.NewServer()
	server.WithMiddleware(middleware.Logger())
	server.Controller("/", denny.HttpGet, &xController{})
	server.Start()
}



```

### Reading config 

```go

package main

import (
	"fmt"
	"github.com/whatvn/denny/config"
	"os"
	"path/filepath"
	"time"
)

func configFile() (*os.File, error) {
	data := []byte(`{"foo": "bar", "denny": {"sister": "jenny"}}`)
	path := filepath.Join(os.TempDir(), fmt.Sprintf("file.%d", time.Now().UnixNano()))
	fh, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	_, err = fh.Write(data)
	if err != nil {
		return nil, err
	}

	return fh, nil
}

func main()  {
	f, err := configFile()
	if err != nil {
		fmt.Println(err)
	}
	// read config from file
	config.New(f.Name())
	fmt.Println(config.GetString("foo"))
	fmt.Println(config.GetString("denny", "sister"))

	// config from evn takes higher priority
	os.Setenv("foo", "barbar")
	os.Setenv("denny_sister", "Jenny")
	config.Reload()
	fmt.Println(config.GetString("foo"))
	fmt.Println(config.GetString("denny", "sister"))
}
```