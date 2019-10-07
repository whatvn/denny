package cache

import (
	redisCli "github.com/go-redis/redis"
	"time"
)

type redis struct {
	cli *redisCli.Client
}

// Get return value if key exist or nil if it does not
func (c *redis) Get(key string) interface{} {
	cmd := c.cli.Get(key)
	s, err := cmd.Result()
	if err != nil {
		return nil
	}
	return s
}

// Set store key in sync map
func (c *redis) Set(key string, val interface{}, expire int64) {
	c.cli.Set(key, val, time.Duration(expire)*time.Second)
}

// Get multi will load all values with given keys
// caller has to check request return value against nil befors using
// as this call will not check key existence
func (c *redis) GetMulti(keys []string) []interface{} {
	cmd := c.cli.MGet(keys...)
	result, err := cmd.Result()
	if err != nil {
		return nil
	}
	return result
}

// Delete delete key in map if it exists
func (c *redis) Delete(key string) {
	c.cli.Del(key)
}

// Incr incr key in map if it exists
func (c *redis) Incr(key string) error {
	cmd := c.cli.Incr(key)
	if _, err := cmd.Result(); err != nil {
		return err
	}
	return nil
}

// Incr incr key in map if it exists
func (c *redis) Decr(key string) error {
	cmd := c.cli.Decr(key)
	if _, err := cmd.Result(); err != nil {
		return err
	}
	return nil
}

func (c *redis) IsExist(key string) bool {
	cmd := c.cli.Exists(key)
	val, err := cmd.Result()
	if err != nil {
		return false
	}
	return val > 0
}

func (c *redis) ClearAll() {
	c.cli.FlushAll()
}

func (c *redis) runGc(config Config) {
}

func (c *redis) expires() []string {
	return nil
}

func NewRedis(address, password string) Cache {
	c := &redis{
		cli: redisCli.NewClient(&redisCli.Options{
			Addr:     address,
			Password: password,
		}),
	}
	return c
}
