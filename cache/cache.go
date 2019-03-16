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
	DefaultCacheContextKey     = "jsdidierlaurent.echo-middleware.cache"
	DefaultResponseCachePrefix = "jsdidierlaurent.echo-middleware.response.cache"

	//DEFAULT Duration use Duration passed in Store constructor
	DEFAULT = time.Duration(0)
	//FOREVER Duration never remove value of the cache
	FOREVER = time.Duration(-1)
)

type (
	//ProvidedCacheConfig Struct for Configure ProvidedCache
	ProvidedCacheConfig struct {
		//Store defines which type of cache you use (Default: inmemory).
		Store Store

		//ContextKey defines the name you use to get the store from echo.Context (Default: cache.DefaultCacheContextKey).
		ContextKey string
	}

	//ResponseCacheConfig Struct for Configure ProvidedCache
	ResponseCacheConfig struct {
		//Store defines which type of cache you use (Default: inmemory).
		Store Store

		//KeyPrefix default cache key prefix used for stored responses (Default: cache.DefaultResponseCachePrefix).
		KeyPrefix string

		//Skipper defines a function to skip middleware. (Default: nil).
		Skipper emw.Skipper

		//Expire ttl for cache value
		Expire time.Duration
	}

	//CacheStore Interface for every Cache (InMemory, GCache, ...)
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
	defaultStore = NewInMemoryStore(time.Minute*10, time.Second*30)

	//DefaultConfig used by default if you don't specifies Config or value inside Config
	DefaultProvidedCacheConfig = ProvidedCacheConfig{
		Store:      defaultStore,
		ContextKey: DefaultCacheContextKey,
	}

	DefaultResponseCacheConfig = ResponseCacheConfig{
		Store:     defaultStore,
		KeyPrefix: DefaultResponseCachePrefix,
		Skipper:   emw.DefaultSkipper,
		Expire:    DEFAULT,
	}

	ErrCacheMiss  = errors.New("cache: key not found")
	ErrNotStored  = errors.New("cache: not stored")
	ErrNotSupport = errors.New("cache: not support")
)

//ProvidedCache Middleware for provide Store to all route using echo.Context#Set()
// Use echo.Context#Get() with cache.DefaultCacheContextKey for get store in your route
func ProvidedCache() echo.MiddlewareFunc {
	return ProvidedCacheWithConfig(DefaultProvidedCacheConfig)
}

//ProvidedCacheWithConfig Middleware for provide Store to all route using echo.Context#Set()
// Use echo.Context#Get() with cache.DefaultCacheContextKey for get store in your route
func ProvidedCacheWithConfig(config ProvidedCacheConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Store == nil {
		config.Store = DefaultProvidedCacheConfig.Store
	}
	if config.ContextKey == "" {
		config.ContextKey = DefaultProvidedCacheConfig.ContextKey
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(config.ContextKey, config.Store)
			return next(c)
		}
	}
}

//ResponseCache Middleware for caching response of all route and return cache if previous call is stored
// Use it in middleware definition
func ResponseCache() echo.MiddlewareFunc {
	return ResponseCacheWithConfig(DefaultResponseCacheConfig)
}

//ResponseCacheWithConfig Middleware for caching response of all route and return cache if previous call is stored
// Use it in middleware definition
func ResponseCacheWithConfig(config ResponseCacheConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Store == nil {
		config.Store = DefaultResponseCacheConfig.Store
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = DefaultResponseCacheConfig.KeyPrefix
	}
	if config.Skipper == nil {
		config.Skipper = DefaultResponseCacheConfig.Skipper
	}
	if config.Expire == time.Duration(0) {
		config.Expire = DefaultResponseCacheConfig.Expire
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
				c.Response().WriteHeader(cache.status)
				for k, vals := range cache.header {
					for _, v := range vals {
						c.Response().Header().Add(k, v)
					}
				}
				_, _ = c.Response().Write(cache.data)
				return nil
			}
		}
	}
}

//ResponseCacheHandler Handle for caching response of one route and return cache if previous call is stored
// Use it in route definition
func ResponseCacheHandler(handle echo.HandlerFunc) echo.HandlerFunc {
	return ResponseCacheHandlerWithConfig(DefaultResponseCacheConfig, handle)
}

//ResponseCacheHandlerWithConfig Handle for caching response of one route and return cache if previous call is stored
// Use it in route definition
func ResponseCacheHandlerWithConfig(config ResponseCacheConfig, handle echo.HandlerFunc) echo.HandlerFunc {
	// Defaults
	if config.Store == nil {
		config.Store = DefaultResponseCacheConfig.Store
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = DefaultResponseCacheConfig.KeyPrefix
	}
	if config.Skipper == nil {
		config.Skipper = DefaultResponseCacheConfig.Skipper
	}
	if config.Expire == time.Duration(0) {
		config.Expire = DefaultResponseCacheConfig.Expire
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
			c.Response().WriteHeader(cache.status)
			for k, vals := range cache.header {
				for _, v := range vals {
					c.Response().Header().Add(k, v)
				}
			}
			_, _ = c.Response().Write(cache.data)
			return nil
		}
	}
}

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
