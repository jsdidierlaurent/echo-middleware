package cache

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jsdidierlaurent/echo-middleware/cache/mocks"
	"github.com/stretchr/testify/assert"
	. "github.com/stretchr/testify/mock"

	"github.com/labstack/echo/v4"
)

func initEcho() (ctx echo.Context, res *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/api/v1/info", nil)
	res = httptest.NewRecorder()
	ctx = e.NewContext(req, res)

	return
}

func TestCacheHandler_RespondCachedValue(t *testing.T) {
	// Init
	ctx, res := initEcho()
	status := http.StatusOK
	header := ctx.Request().Header
	body := "😁"

	mockStore := new(mocks.Store)
	mockStore.
		On("Get", AnythingOfType("string"), AnythingOfType("*cache.ResponseCache")).
		Return(nil).
		Run(func(args Arguments) {
			arg := args.Get(1).(*ResponseCache)
			arg.Data = []byte(body)
			arg.Header = header
			arg.Status = status
		})

	// Override Default Config
	DefaultCacheMiddlewareConfig = CacheMiddlewareConfig{
		Store:     mockStore,
		KeyPrefix: DefaultCachePrefix,
		Skipper:   defaultSkipper,
		Expire:    DEFAULT,
	}

	cm := CacheMiddleware()

	handle := cm(echo.HandlerFunc(func(c echo.Context) error {
		return nil
	}))

	// Test
	if assert.NoError(t, handle(ctx)) {
		assert.Equal(t, status, res.Code)
		assert.Equal(t, header, res.Header())
		assert.Equal(t, body, res.Body.String())
		mockStore.AssertNumberOfCalls(t, "Get", 1)
		mockStore.AssertExpectations(t)
	}
}

func TestCacheHandler_NoCache(t *testing.T) {
	// Init
	ctx, res := initEcho()

	mockStore := new(mocks.Store)
	mockStore.On("Set", AnythingOfType("string"), Anything, AnythingOfType("time.Duration")).Return(nil)
	mockStore.On("Get", AnythingOfType("string"), Anything).Return(ErrCacheMiss)

	// Override Default Config
	DefaultCacheMiddlewareConfig = CacheMiddlewareConfig{
		Store:     mockStore,
		KeyPrefix: DefaultCachePrefix,
		Skipper:   defaultSkipper,
		Expire:    DEFAULT,
	}

	cm := CacheMiddleware()

	handle := cm(echo.HandlerFunc(func(c echo.Context) error {
		return c.HTML(http.StatusOK, "😁")
	}))

	// Test
	if assert.NoError(t, handle(ctx)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, "😁", res.Body.String())
		mockStore.AssertNumberOfCalls(t, "Get", 1)
		mockStore.AssertNumberOfCalls(t, "Set", 1)
		mockStore.AssertExpectations(t)
	}
}

func TestStoreHandler(t *testing.T) {
	// Init
	ctx, _ := initEcho()

	mockStore := new(mocks.Store)

	// Override Default Config
	DefaultStoreMiddlewareConfig = StoreMiddlewareConfig{
		Store:      mockStore,
		ContextKey: DefaultStoreContextKey,
	}

	cm := StoreMiddleWare()

	handle := cm(echo.HandlerFunc(func(c echo.Context) error {
		assert.NotNil(t, c.Get(DefaultStoreContextKey))
		return c.HTML(http.StatusOK, "😁")
	}))

	assert.NoError(t, handle(ctx))
}
