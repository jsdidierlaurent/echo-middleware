package cache

import (
	"reflect"
	"time"

	"github.com/robfig/go-cache"
)

type GoCacheStore struct {
	cache.Cache
}

func NewGoCacheStore(defaultExpiration time.Duration, cleanupInterval time.Duration) *GoCacheStore {
	return &GoCacheStore{*cache.New(defaultExpiration, cleanupInterval)}
}

func (c *GoCacheStore) Get(key string, value interface{}) error {
	val, found := c.Cache.Get(key)
	if !found {
		return ErrCacheMiss
	}

	v := reflect.ValueOf(value)
	if v.Type().Kind() == reflect.Ptr && v.Elem().CanSet() {
		v.Elem().Set(reflect.ValueOf(val))
		return nil
	}
	return ErrNotStored
}

func (c *GoCacheStore) Set(key string, value interface{}, expires time.Duration) error {
	// NOTE: go-cache understands the values of DEFAULT and FOREVER
	c.Cache.Set(key, value, expires)
	return nil
}

func (c *GoCacheStore) Add(key string, value interface{}, expires time.Duration) error {
	err := c.Cache.Add(key, value, expires)
	if err == cache.ErrKeyExists {
		return ErrNotStored
	}
	return err
}

func (c *GoCacheStore) Replace(key string, value interface{}, expires time.Duration) error {
	if err := c.Cache.Replace(key, value, expires); err != nil {
		return ErrNotStored
	}
	return nil
}

func (c *GoCacheStore) Delete(key string) error {
	if found := c.Cache.Delete(key); !found {
		return ErrCacheMiss
	}
	return nil
}

func (c *GoCacheStore) Increment(key string, n uint64) (uint64, error) {
	newValue, err := c.Cache.Increment(key, n)
	if err == cache.ErrCacheMiss {
		return 0, ErrCacheMiss
	}
	return newValue, err
}

func (c *GoCacheStore) Decrement(key string, n uint64) (uint64, error) {
	newValue, err := c.Cache.Decrement(key, n)
	if err == cache.ErrCacheMiss {
		return 0, ErrCacheMiss
	}
	return newValue, err
}

func (c *GoCacheStore) Flush() error {
	c.Cache.Flush()
	return nil
}
