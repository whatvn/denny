package main

import (
	"fmt"
	"github.com/whatvn/denny/config"
	"github.com/whatvn/denny/log"
	"os"
	"path/filepath"
	"time"
)

func configFile() (*os.File, error) {
	data := []byte(`{"foo": "bar", "dennyObj": {"sister": "jenny"}}`)
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

type Denny struct {
	Age    int
	Sister string
}

var (
	dennyObj = &Denny{}
	logger   = log.New()
)

func load() {
	config.Scan(dennyObj, "dennyObj")
}

func main() {
	f, err := configFile()
	if err != nil {
		logger.Infof("error %v", err)
	}
	// read config from file
	config.New(f.Name())

	config.WithEtcd(
		config.WithEtcdAddress("10.109.3.93:7379"),
		config.WithPath("/acquiringcore/ae/config"),
	)

	load()
	logger.Println(dennyObj.Age)
	logger.Println(dennyObj.Sister)
	w, _ := config.Watch()

	for {

		_, err := w.Next()
		if err != nil {
			logger.Error(err)
		}
		load()
		logger.Println(dennyObj.Age)
		logger.Println(dennyObj.Sister)
	}
}
