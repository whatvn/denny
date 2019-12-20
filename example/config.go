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

type Denny struct {
	Age    int
	Sister string
}

var (
	denny = &Denny{}
)

func load() {
	config.Scan(denny, "denny")
}

func main() {
	f, err := configFile()
	if err != nil {
		fmt.Println(err)
	}
	// read config from file
	config.New(f.Name())

	config.WithEtcd(
		config.WithEtcdAddress("http://127.0.0.1:2379"),
		config.WithPath("/acquiringcore/ae/config"),
	)

	load()
	fmt.Println(denny.Age)
	fmt.Println(denny.Sister)
	w, _ := config.Watch()

	for {

		_, err := w.Next()
		if err != nil {
			// do something
		}
		load()
		fmt.Println(denny.Age)
		fmt.Println(denny.Sister)
	}
}
