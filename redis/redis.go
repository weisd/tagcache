package redis

import (
	"encoding/json"
	"fmt"
	"time"

	redigo "github.com/garyburd/redigo/redis"
	"github.com/weisd/tagcache"
)

var _ tagcache.CacheStore = new(RedisCache)

var DefualtPrefix = "tc"
var DefualtExpire int64 = 3600 // 默认一个小时

func init() {
	tagcache.Register("redis", &RedisCache{})
}

type RedisConfig struct {
	Addr        string
	Passwd      string
	SelectDB    int
	MaxIdle     int
	MaxActive   int
	IdleTimeout int
	Wait        bool
}

func prepareConfig(conf RedisConfig) RedisConfig {
	if conf.MaxIdle == 0 {
		conf.MaxIdle = 10
	}
	if conf.MaxActive == 0 {
		conf.MaxActive = 10
	}

	if conf.IdleTimeout == 0 {
		conf.IdleTimeout = 60
	}

	return conf
}

type RedisCache struct {
	pool   *redigo.Pool
	prefix string
}

func (r *RedisCache) key(key string) string {
	if len(r.prefix) > 0 {
		return r.prefix + ":" + key
	}

	return key
}

func (r *RedisCache) do(commandName string, args ...interface{}) (reply interface{}, err error) {
	c := r.pool.Get()
	defer c.Close()
	return c.Do(commandName, args...)
}

func (r *RedisCache) Set(key, val string, timeout int64) (err error) {
	if timeout == 0 {
		_, err = r.do("SET", r.key(key), val)
		return
	}
	_, err = r.do("SETEX", r.key(key), timeout, val)
	return
}

// func (r *RedisCache) MSet(items map[string]string, timeout int64) (err error) {
// 	if timeout == 0 {
// 		args := make([]interface{}, 0, len(items)*2)
// 		for k, _ := range items {
// 			args = append(args, r.key(k), items[k])
// 		}
// 		_, err = r.do("MSET", args...)
// 		return
// 	}

// 	c := r.pool.Get()
// 	defer c.Close()

// 	for k, _ := range items {
// 		c.Send("SETEX", r.key(k), timeout, items[k])
// 	}

// 	return c.Flush()
// }

// func (r *RedisCache) Forever(key, val string) (err error) {
// 	_, err = r.do("SET", r.key(key), val)
// 	return
// }

func (r *RedisCache) Get(key string) string {

	c := r.pool.Get()
	defer c.Close()

	c.Send("GET", r.key(key))
	c.Send("TTL", r.key(key))

	err := c.Flush()
	if err != nil {
		return ""
	}

	v, _ := redigo.String(c.Receive())
	ttl, _ := redigo.Int64(c.Receive())

	if ttl > 0 && ttl < DefualtExpire {
		r.Touch(key, DefualtExpire+ttl)
	}

	return v

	// v, _ := redigo.String(r.do("GET", r.key(key)))
	// // 有值，自动添加过期时间
	// if len(v) > 0 {

	// }
	// return v
}

// func (r *RedisCache) MGet(keys []string) []string {
// 	args := make([]interface{}, len(keys))

// 	for i, _ := range keys {
// 		args[i] = r.key(keys[i])
// 	}

// 	v, _ := redigo.Strings(r.do("MGET", args...))
// 	// 有值，自动添加过期时间
// 	if len(v) > 0 {

// 	}
// 	return v
// }

// Delete deletes cached value by given key.
func (r *RedisCache) Delete(key string) (err error) {
	_, err = r.do("DEL", r.key(key))
	if err != redigo.ErrNil {
		return
	}

	return nil
}

// Incr increases cached int-type value by given key as a counter.
func (r *RedisCache) Incr(key string) (int64, error) {
	return redigo.Int64(r.do("INCR", r.key(key)))
}

// Decr decreases cached int-type value by given key as a counter.
func (r *RedisCache) Decr(key string) (int64, error) {
	return redigo.Int64(r.do("DECR", r.key(key)))
}

// Flush deletes all cached data.
func (r *RedisCache) Flush() (err error) {

	keys, err := redigo.MultiBulk(r.do("KEYS", r.key("*")))
	if err != nil {
		return
	}

	conn := r.pool.Get()
	defer conn.Close()

	_, err = conn.Do("DEL", keys...)

	return
}

func (r *RedisCache) startGC() {

}

// StartAndGC starts GC routine based on config string settings.
func (r *RedisCache) StartAndGC(opt tagcache.Options) error {
	var conf RedisConfig
	err := json.Unmarshal([]byte(opt.AdapterConfig), &conf)
	if err != nil {
		return fmt.Errorf("RedisConfig parse err %v", err)
	}

	conf = prepareConfig(conf)

	r.prefix = opt.Section
	if len(r.prefix) == 0 {
		r.prefix = DefualtPrefix
	}

	r.pool = newRedisPool(conf)

	conn := r.pool.Get()

	_, err = conn.Do("PING")
	if err != nil {
		return fmt.Errorf("redis conn err %v", err)
	}
	conn.Close()

	go r.startGC()

	return nil
}

func newRedisPool(conf RedisConfig) *redigo.Pool {
	return &redigo.Pool{
		MaxIdle:     conf.MaxIdle,
		IdleTimeout: time.Duration(conf.IdleTimeout) * time.Second,
		Dial: func() (redigo.Conn, error) {
			c, err := redigo.Dial("tcp", conf.Addr)
			if err != nil {
				return nil, err
			}
			if len(conf.Passwd) > 0 {
				if _, err := c.Do("AUTH", conf.Passwd); err != nil {
					c.Close()
					return nil, err
				}
			}
			_, err = c.Do("SELECT", conf.SelectDB)
			if err != nil {
				c.Close()
				return nil, err
			}

			return c, err
		},
		TestOnBorrow: func(c redigo.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// update expire time
func (r *RedisCache) Touch(key string, expire int64) (err error) {
	if _, err = r.do("EXPIRE", r.key(key), expire); err != redigo.ErrNil {
		return
	}
	return nil
}
