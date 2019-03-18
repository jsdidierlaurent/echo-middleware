package main

import (
	"net/http"
	"time"

	"github.com/jsdidierlaurent/echo-middleware/cache"
	"github.com/labstack/echo/v4"
	emw "github.com/labstack/echo/v4/middleware"
)

func main() {
	r := echo.New()
	r.Use(emw.Logger())

	config := cache.StoreMiddlewareConfig{
		Store: cache.NewInMemoryStore(time.Second*5, time.Second),
	}

	r.Use(cache.StoreMiddlewareWithConfig(config))

	r.GET("/ping", func(c echo.Context) error {
		store := c.Get(cache.DefaultStoreContextKey).(*cache.InMemoryStore)
		key := c.QueryString()

		var cachedValue string
		if err := store.Get(key, &cachedValue); err == nil && cachedValue != "" {
			c.Response().Header().Set("From-Cache", "true")
			c.Response().Header().Set("Cache-Control", "max-age=5")
			return c.String(http.StatusOK, cachedValue)
		} else {
			c.Logger().Errorf("Enable to get value in cache %s\n", err)
		}

		// Awesome value
		value := "pong"

		err := store.Add(key, value, cache.DEFAULT)
		if err != nil {
			c.Logger().Errorf("Enable to store value in cache %s\n", err)
		}

		return c.String(http.StatusOK, value)
	})

	r.Logger.Fatal(r.Start(":1323"))
}
