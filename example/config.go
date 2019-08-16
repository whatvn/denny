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

func main() {
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
