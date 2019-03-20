package cache

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type cacheFactory func(*testing.T, time.Duration) Store

// Test typical cache interactions
func typicalGetSet(t *testing.T, newCache cacheFactory) {
	cache := newCache(t, time.Hour)

	set, get := "foo", ""

	assert.NoError(t, cache.Set("value", set, DEFAULT))
	assert.NoError(t, cache.Get("value", &get))
	assert.Equal(t, set, get)
}

// Test the increment-decrement cases
func incrDecr(t *testing.T, newCache cacheFactory) {
	cache := newCache(t, time.Hour)

	// Normal increment / decrement operation.
	assert.NoError(t, cache.Set("int", 10, DEFAULT))

	newValue, err := cache.Increment("int", 50)
	assert.NoError(t, err)
	assert.Equal(t, uint64(60), newValue)

	newValue, err = cache.Decrement("int", 50)
	assert.NoError(t, err)
	assert.Equal(t, uint64(10), newValue)

	// Increment wraparound
	newValue, err = cache.Increment("int", math.MaxUint64-5)
	assert.NoError(t, err)
	assert.Equal(t, uint64(4), newValue)

	// Decrement capped at 0
	newValue, err = cache.Decrement("int", 25)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), newValue)
}

func expiration(t *testing.T, newCache cacheFactory) {
	// memcached does not support expiration times less than 1 second.
	cache := newCache(t, time.Second)

	// Speed up this test
	var wg sync.WaitGroup
	wg.Add(4)

	parallel(&wg, func() {
		// Test Set w/ DEFAULT
		set, get := 10, 0
		assert.NoError(t, cache.Set("int1", set, DEFAULT))
		time.Sleep(2 * time.Second)
		assert.Equal(t, ErrCacheMiss, cache.Get("int1", &get))
	})

	parallel(&wg, func() {
		// Test Set w/ short time
		set, get := 10, 0
		assert.NoError(t, cache.Set("int2", set, time.Second))
		time.Sleep(2 * time.Second)
		assert.Equal(t, ErrCacheMiss, cache.Get("int2", &get))
	})

	parallel(&wg, func() {
		// Test Set w/ longer time.
		set, get := 10, 0
		assert.NoError(t, cache.Set("int3", set, time.Hour))
		time.Sleep(2 * time.Second)
		assert.NoError(t, cache.Get("int3", &get))
		assert.Equal(t, set, get)
	})

	parallel(&wg, func() {
		// Test Set w/ forever.
		set, get := 10, 0
		assert.NoError(t, cache.Set("int4", set, FOREVER))
		time.Sleep(2 * time.Second)
		assert.NoError(t, cache.Get("int4", &get))
		assert.Equal(t, set, get)
	})

	wg.Wait()
}

func emptyCache(t *testing.T, newCache cacheFactory) {
	var err error
	cache := newCache(t, time.Hour)

	err = cache.Get("notexist", 0)
	assert.Equal(t, ErrCacheMiss, err)

	err = cache.Delete("notexist")
	assert.Equal(t, ErrCacheMiss, err)

	_, err = cache.Increment("notexist", 1)
	assert.Equal(t, ErrCacheMiss, err)

	_, err = cache.Decrement("notexist", 1)
	assert.Equal(t, ErrCacheMiss, err)
}

func testReplace(t *testing.T, newCache cacheFactory) {
	var err error
	cache := newCache(t, time.Hour)

	// Replace in an empty cache.
	err = cache.Replace("notexist", 1, FOREVER)
	assert.Equal(t, ErrNotStored, err)

	// Set a value of 1, and replace it with 2
	err = cache.Set("int", 1, time.Second)
	assert.NoError(t, err)

	err = cache.Replace("int", 2, time.Second)
	assert.NoError(t, err)

	var i int
	err = cache.Get("int", &i)
	assert.NoError(t, err)
	assert.Equal(t, 2, i)

	// Wait for it to expire and replace with 3 (unsuccessfully).
	time.Sleep(2 * time.Second)
	err = cache.Replace("int", 3, time.Second)
	assert.Equal(t, ErrNotStored, err)
	err = cache.Get("int", &i)
	assert.Equal(t, ErrCacheMiss, err)
}

func testAdd(t *testing.T, newCache cacheFactory) {
	var err error
	cache := newCache(t, time.Hour)

	// Add to an empty cache.
	err = cache.Add("int", 1, time.Second)
	assert.NoError(t, err)

	// Try to add again. (fail)
	err = cache.Add("int", 2, time.Second)
	assert.Equal(t, ErrNotStored, err)

	// Wait for it to expire, and add again.
	time.Sleep(2 * time.Second)
	err = cache.Add("int", 3, time.Second)
	assert.NoError(t, err)

	// Get and verify the value.
	var i int
	err = cache.Get("int", &i)
	assert.NoError(t, err)
	assert.Equal(t, 3, i)
}

func parallel(wg *sync.WaitGroup, handler func()) {
	go func() {
		handler()
		wg.Done()
	}()
}
