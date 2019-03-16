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

	config := cache.ResponseCacheConfig{
		Store: cache.NewInMemoryStore(time.Second*5, time.Second),
	}

	r.Use(cache.ResponseCacheWithConfig(config))

	r.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	r.Logger.Fatal(r.Start(":1323"))
}
