package tagcache

import (
	"fmt"
)

const _VERSION = "0.1.0"

func Version() string {
	return _VERSION
}

var _ Cache = new(Engine)

// Cache is the interface that operates the cache data.
type CacheStore interface {
	// Set Sets value into cache with key and expire time.
	Set(key, val string, timeout int64) error
	MSet(items map[string]string, timeout int64) error
	// Get gets cached value by given key.
	Get(key string) string
	// Get Multi keys
	MGet(keys []string) []string
	// Delete deletes cached value by given key.
	Delete(key string) error
	// Incr increases cached int-type value by given key as a counter.
	Incr(key string) (int64, error)
	// Decr decreases cached int-type value by given key as a counter.
	Decr(key string) (int64, error)
	// Flush deletes all cached data.
	Flush() error
	// StartAndGC starts GC routine based on config string settings.
	StartAndGC(opt Options) error
	// update expire time
	Touch(key string, expire int64) error
}

type Cache interface {
	CacheStore
	Tags(tags []string) Cache
}

type Options struct {
	// Name of adapter. Default is "memory".
	Adapter string
	// Adapter configuration, it's corresponding to adapter.
	AdapterConfig string
	// key prefix Default is ""
	Section string
}

func New(opt Options) (Cache, error) {

	adapter, ok := adapters[opt.Adapter]
	if !ok {
		return nil, fmt.Errorf("cache: unknown adapter '%s'(forgot to import?)", opt.Adapter)
	}

	engine := &Engine{}
	engine.Opt = opt
	engine.store = adapter

	return engine, adapter.StartAndGC(opt)
}

type Engine struct {
	Opt   Options
	store CacheStore
}

func (this *Engine) Set(key, val string, timeout int64) error {
	return this.store.Set(key, val, timeout)
}

func (this *Engine) MSet(items map[string]string, timeout int64) error {
	return this.store.MSet(items, timeout)
}

func (this *Engine) Get(key string) string {
	return this.store.Get(key)
}

func (this *Engine) MGet(keys []string) []string {
	return this.store.MGet(keys)
}

func (this *Engine) Delete(key string) error {
	return this.store.Delete(key)
}

func (this *Engine) Incr(key string) (int64, error) {
	return this.store.Incr(key)
}

func (this *Engine) Decr(key string) (int64, error) {
	return this.store.Decr(key)
}

func (this *Engine) Flush() error {
	return this.store.Flush()
}

func (this *Engine) StartAndGC(opt Options) error {
	return this.store.StartAndGC(opt)
}

func (this *Engine) Touch(key string, expire int64) error {
	return this.store.Touch(key, expire)
}

func (this *Engine) Tags(tags []string) Cache {
	return NewTagCache(this.store, tags...)
}

var adapters = make(map[string]CacheStore)

// Register registers a adapter.
func Register(name string, adapter CacheStore) {
	if adapter == nil {
		panic("cache: cannot register adapter with nil value")
	}
	if _, dup := adapters[name]; dup {
		panic(fmt.Errorf("cache: cannot register adapter '%s' twice", name))
	}
	adapters[name] = adapter
}
