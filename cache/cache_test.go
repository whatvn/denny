package cache

import (
	"testing"
	"time"
)

func TestMemoryCache(t *testing.T) {

	var (
		cache = NewMemoryCache(Config{
			GcDuration: 2,
		})
		key   = "name"
		value = "denny"
	)

	cache.Set(key, value, 2)
	v := cache.Get(key)
	if v.(string) != value {
		t.Error("wrong value return")
	}

	time.Sleep(time.Duration(3) * time.Second)
	v = cache.Get(key)
	if v != nil {
		t.Error("gc cannot run")
	}
	cache.Set(key, value, 0)
	cache.Delete(key)
	v = cache.Get(key)
	if v != nil {
		t.Error("cannot delete key")
	}

}
