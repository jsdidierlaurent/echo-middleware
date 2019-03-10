package cache

import (
	"errors"
	"time"

	"github.com/labstack/echo/v4"
	emw "github.com/labstack/echo/v4/middleware"
)

const (
	DefaultCacheContextKey = "jsdidierlaurent/echo-middleware/default.store.name"

	//DEFAULT Duration use Duration passed in Store constructor
	DEFAULT = time.Duration(0)
	//FOREVER Duration never remove value of the cache
	FOREVER = time.Duration(-1)
)

type (
	//CacheConfig Struct for Configure cache
	Config struct {
		//Skipper defines a function to skip middleware.
		Skipper emw.Skipper

		//Store defines which type of cache you use (Default: inmemory).
		// Required.
		Store Store

		//StoreName defines the name you use to get the store from echo.Context
		ContextKey string
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
	//DefaultConfig used by default if you don't specifies Config or value inside Config
	DefaultConfig = Config{
		Skipper:    emw.DefaultSkipper,
		Store:      NewInMemoryStore(time.Minute*10, time.Second*30),
		ContextKey: DefaultCacheContextKey,
	}

	ErrCacheMiss  = errors.New("cache: key not found")
	ErrNotStored  = errors.New("cache: not stored")
	ErrNotSupport = errors.New("cache: not support")
)

//ManualCache Middleware for manage cache manually with echo.Contect#Get()
func ManualCache() echo.MiddlewareFunc {
	return ManualCacheWithConfig(DefaultConfig)
}

//ManualCacheWithConfig Middleware for manage cache manually with echo.Contect#Get()
func ManualCacheWithConfig(config Config) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultConfig.Skipper
	}
	if config.Store == nil {
		config.Store = DefaultConfig.Store
	}
	if config.ContextKey == "" {
		config.ContextKey = DefaultConfig.ContextKey
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			c.Set(config.ContextKey, config.Store)
			return next(c)
		}
	}
}
