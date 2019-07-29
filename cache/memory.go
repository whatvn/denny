package cache

import (
	"sync"
	"time"
)

type memory struct {
	count int64
	sync.RWMutex
	storage *sync.Map
}

type item struct {
	expire      time.Duration
	createdTime time.Time
	value       interface{}
}

func (i *item) life() int64 {
	if i.expire == 0 {
		return 0
	}
	return int64(time.Now().Sub(i.createdTime).Seconds())
}

func (i *item) isExpire() bool {
	return time.Now().Sub(i.createdTime) > i.expire
}

// Get return value if key exist or nil if it does not
func (c *memory) Get(key string) interface{} {
	v, ok := c.storage.Load(key)
	if ok {
		return v
	}
	return nil
}

// Set store key in sync map
func (c *memory) Set(key string, val interface{}, expire int64) {
	c.storage.Store(key, &item{
		value:       val,
		createdTime: time.Now(),
		expire:      time.Duration(expire) * time.Second,
	})
}

// Get multi will load all values with given keys
// caller has to check request return value against nil befors using
// as this call will not check key existence
func (c *memory) GetMulti(keys []string) []interface{} {
	var values []interface{}
	for _, k := range keys {
		el, _ := c.storage.Load(k)
		v, ok := el.(*item)
		if ok {
			values = append(values, v)
		} else {
			values = append(values, nil)
		}
	}
	return values
}

// Delete delete key in map if it exists
func (c *memory) Delete(key string) {
	c.storage.Delete(key)
}

// Incr incr key in map if it exists
func (c *memory) Incr(key string) error {
	c.Lock()
	defer c.Unlock()
	el, ok := c.storage.Load(key)
	if !ok {
		return ValueNotExistError
	}

	v, ok := el.(*item)
	if !ok {
		return InvalidValueTypeError
	}

	i, ok := v.value.(int64)
	if !ok {
		return InvalidValueTypeError
	}

	c.Set(key, i, v.life())
	return nil
}

// Incr incr key in map if it exists
func (c *memory) Decr(key string) error {
	c.Lock()
	defer c.Unlock()
	el, ok := c.storage.Load(key)
	if !ok {
		return ValueNotExistError
	}

	v, ok := el.(*item)
	if !ok {
		return InvalidValueTypeError
	}

	i, ok := v.value.(int64)
	if !ok {
		return InvalidValueTypeError
	}

	c.Set(key, i, v.life())
	return nil
}

func (c *memory) IsExist(key string) bool {
	_, ok := c.storage.Load(key)
	return ok
}

func (c *memory) ClearAll() {
	c.storage.Range(func(key, value interface{}) bool {
		c.storage.Delete(key)
		return true
	})
}

func (c *memory) runGc(config Config) {
	for {
		<-time.After(time.Duration(config.GcDuration) * time.Second)
		for _, k := range c.expires() {
			c.Delete(k)
		}
	}
}

func (c *memory) expires() []string {
	var keys []string
	c.storage.Range(func(k, el interface{}) bool {
		v := el.(*item)
		if v.expire != 0 {
			if v.isExpire() {
				keys = append(keys, k.(string))
			}
		}
		return true
	})
	return keys
}

func New(cfg Config) Cache {
	c := &memory{}
	go c.runGc(cfg)
	return c
}
