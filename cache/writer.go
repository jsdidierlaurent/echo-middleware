package cache

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type (
	responseCache struct {
		status int
		header http.Header
		data   []byte
	}

	cachedWriter struct {
		writer   http.ResponseWriter
		response *echo.Response

		status  int
		written bool

		store  Store
		expire time.Duration
		key    string
	}
)

func newCachedWriter(store Store, expire time.Duration, writer http.ResponseWriter, response *echo.Response, key string) *cachedWriter {
	return &cachedWriter{writer, response, 0, false, store, expire, key}
}

func (w *cachedWriter) Header() http.Header {
	return w.writer.Header()
}

func (w *cachedWriter) WriteHeader(code int) {
	w.status = code
	w.written = true
	w.writer.WriteHeader(code)
}

func (w *cachedWriter) Write(data []byte) (int, error) {
	ret, err := w.writer.Write(data)
	if err == nil {
		header := w.response.Header()

		// Add cache Header
		currentTime := time.Now()
		header.Add("Last-Modified", currentTime.Format(time.RFC1123))

		val := responseCache{
			w.response.Status,
			header,
			data,
		}
		_ = w.store.Set(w.key, val, w.expire)
	}
	return ret, err
}
