package cache

import (
	"errors"
	"time"
)

var (
	ValueNotExistError    = errors.New("value not exist")
	InvalidValueTypeError = errors.New("invalid value type")
)

type Cache interface {
	// get cached value by key.
	Get(key string) interface{}
	// GetOrElse return value if it exists, else warmup using warmup function
	GetOrElse(key string, warmUpFunc func(key string) interface{}, expire ...int64) interface{}
	// GetMulti is a batch version of Get.
	GetMulti(keys []string) []interface{}
	// set cached value with key and expire time.
	Set(key string, val interface{}, expire int64)
	// delete cached value by key.
	Delete(key string)
	// increase cached int value by key, as a counter.
	Incr(key string) error
	// decrease cached int value by key, as a counter.
	Decr(key string) error
	// check if cached value exists or not.
	IsExist(key string) bool
	// clear all cache.
	ClearAll()
	// start gc routine based on config string settings.
	runGc(config Config)
}

type Config struct {
	GcDuration time.Duration
	GcEvery    int //second
}
