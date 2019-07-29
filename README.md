# denny
http server which simplify request handling and logging

- use class base request controller, one controller for one handler 
- simple cache usage
- open tracing attached 
- simpler config reader
- logger is attached in controller, log should be showed as steps and in only one line for every request 

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

func (x xController) Handle(ctx *denny.Context) {
	x.AddLog("receive request")
	var str = "hello"
	x.AddLog("do more thing")
	str += " world"
	ctx.Writer.Write([]byte(str))
	x.Infof("finished")
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
	config.New(f.Name())
	fmt.Println(config.GetString("foo"))
	fmt.Println(config.GetString("denny", "sister"))
}
```