package cache

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"io"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"
	emw "github.com/labstack/echo/v4/middleware"
)

const (
	DefaultStoreContextKey = "jsdidierlaurent.echo-middleware.store"
	DefaultCachePrefix     = "jsdidierlaurent.echo-middleware.cache"

	//DEFAULT Duration use Duration passed in Store constructor
	DEFAULT = time.Duration(0)
	//FOREVER Duration never remove value of the cache
	FOREVER = time.Duration(-1)
)

type (
	//StoreMiddlewareConfig Struct for Configure StoreMiddleware
	StoreMiddlewareConfig struct {
		//Store defines which type of cache you use (Default: gocache).
		Store Store

		//ContextKey defines the name you use to get the store from echo.Context (Default: cache.DefaultCacheContextKey).
		ContextKey string
	}

	//CacheMiddlewareConfig Struct for Configure CacheMiddleware
	CacheMiddlewareConfig struct {
		//Store defines which type of cache you use (Default: gocache).
		Store Store

		//KeyPrefix default cache key prefix used for stored responses (Default: cache.DefaultResponseCachePrefix).
		KeyPrefix string

		//Skipper defines a function to skip middleware. (Default: nil).
		Skipper emw.Skipper

		//Expire ttl for cache value
		Expire time.Duration
	}

	//CacheStore Interface for every Cache (GoCache, Redis, ...)
	Store interface {
		Get(key string, value interface{}) error
		Set(key string, value interface{}, expire time.Duration) error
		Add(key string, value interface{}, expire time.Duration) error
		Replace(key string, data interface{}, expire time.Duration) error
		Delete(key string) error
		Increment(key string, data uint64) (uint64, error)
		Decrement(key string, data uint64) (uint64, error)
		Flush() error
	}
)

var (
	//defaultStore used by Default config
	defaultStore = NewGoCacheStore(time.Minute*10, time.Second*30)

	//defaultSkiper skip cache if header Cache-Control=no-cache
	defaultSkipper = func(context echo.Context) bool {
		// Skip cache if Cache-Controle is set to no-cache
		return context.Request().Header.Get("Cache-Control") == "no-cache"
	}

	//DefaultConfig used by default if you don't specifies Config or value inside Config
	DefaultStoreMiddlewareConfig = StoreMiddlewareConfig{
		Store:      defaultStore,
		ContextKey: DefaultStoreContextKey,
	}

	DefaultCacheMiddlewareConfig = CacheMiddlewareConfig{
		Store:     defaultStore,
		KeyPrefix: DefaultCachePrefix,
		Skipper:   defaultSkipper,
		Expire:    DEFAULT,
	}

	ErrCacheMiss  = errors.New("cache: key not found")
	ErrNotStored  = errors.New("cache: not stored")
	ErrNotSupport = errors.New("cache: not support")
)

//StoreMiddleware for provide Store to all route using echo.Context#Set()
// Use echo.Context#Get() with cache.DefaultCacheContextKey for get store in your route
func StoreMiddleWare() echo.MiddlewareFunc {
	return StoreMiddlewareWithConfig(DefaultStoreMiddlewareConfig)
}

//StoreMiddlewareWithConfig for provide Store to all route using echo.Context#Set()
// Use echo.Context#Get() with cache.DefaultCacheContextKey for get store in your route
func StoreMiddlewareWithConfig(config StoreMiddlewareConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Store == nil {
		config.Store = DefaultStoreMiddlewareConfig.Store
	}
	if config.ContextKey == "" {
		config.ContextKey = DefaultStoreMiddlewareConfig.ContextKey
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(config.ContextKey, config.Store)
			return next(c)
		}
	}
}

//CacheMiddleware for caching response of all route and return cache if previous call is stored
// Use it in middleware definition
func CacheMiddleware() echo.MiddlewareFunc {
	return CacheMiddlewareWithConfig(DefaultCacheMiddlewareConfig)
}

//CacheMiddlewareWithConfig for caching response of all route and return cache if previous call is stored
// Use it in middleware definition
func CacheMiddlewareWithConfig(config CacheMiddlewareConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Store == nil {
		config.Store = DefaultCacheMiddlewareConfig.Store
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = DefaultCacheMiddlewareConfig.KeyPrefix
	}
	if config.Skipper == nil {
		config.Skipper = DefaultCacheMiddlewareConfig.Skipper
	}
	if config.Expire == time.Duration(0) {
		config.Expire = DefaultCacheMiddlewareConfig.Expire
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			var cache responseCache
			key := getKey(config.KeyPrefix, c.Request().RequestURI)

			if err := config.Store.Get(key, &cache); err != nil {
				// Inject Wrapped Writer
				writer := newCachedWriter(config.Store, config.Expire, c.Response().Writer, c.Response(), key)
				c.Response().Writer = writer
				return next(c)
			} else {
				for k, vals := range cache.header {
					for _, v := range vals {
						if c.Response().Header().Get(k) == "" {
							c.Response().Header().Add(k, v)
						}
					}
				}
				c.Response().WriteHeader(cache.status)
				_, _ = c.Response().Write(cache.data)
				return nil
			}
		}
	}
}

//CacheHandler for caching response of one route and return cache if previous call is stored
// Use it in route definition
func CacheHandler(handle echo.HandlerFunc) echo.HandlerFunc {
	return CacheHandlerWithConfig(DefaultCacheMiddlewareConfig, handle)
}

//CacheHandlerWithConfig for caching response of one route and return cache if previous call is stored
// Use it in route definition
func CacheHandlerWithConfig(config CacheMiddlewareConfig, handle echo.HandlerFunc) echo.HandlerFunc {
	// Defaults
	if config.Store == nil {
		config.Store = DefaultCacheMiddlewareConfig.Store
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = DefaultCacheMiddlewareConfig.KeyPrefix
	}
	if config.Skipper == nil {
		config.Skipper = DefaultCacheMiddlewareConfig.Skipper
	}
	if config.Expire == time.Duration(0) {
		config.Expire = DefaultCacheMiddlewareConfig.Expire
	}

	return func(c echo.Context) error {
		if config.Skipper(c) {
			return handle(c)
		}

		var cache responseCache
		key := getKey(config.KeyPrefix, c.Request().RequestURI)

		if err := config.Store.Get(key, &cache); err != nil {
			// Inject Wrapped Writer
			writer := newCachedWriter(config.Store, config.Expire, c.Response().Writer, c.Response(), key)
			c.Response().Writer = writer
			return handle(c)
		} else {
			for k, vals := range cache.header {
				for _, v := range vals {
					if c.Response().Header().Get(k) == "" {
						c.Response().Header().Add(k, v)
					}
				}
			}
			c.Response().WriteHeader(cache.status)
			_, _ = c.Response().Write(cache.data)
			return nil
		}
	}
}

// getKey build unique key by route with queryParams
func getKey(prefix string, u string) string {
	key := url.QueryEscape(u)

	// Hash if necessary
	if len(key) > 200 {
		h := sha1.New()
		_, _ = io.WriteString(h, u)
		key = string(h.Sum(nil))
	}

	var buffer bytes.Buffer
	buffer.WriteString(prefix)
	buffer.WriteString(":")
	buffer.WriteString(key)
	return buffer.String()
}
