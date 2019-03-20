package cache

import (
	"net"
	"testing"
	"time"
)

// These tests require memcached running on localhost:11211 (the default)
const memcacheTestServer = "localhost:11211"

var newMemcachedStore = func(t *testing.T, defaultExpiration time.Duration) Store {
	c, err := net.Dial("tcp", memcacheTestServer)
	if err == nil {
		c.Write([]byte("flush_all\r\n"))
		c.Close()
		return NewMemcachedStore([]string{memcacheTestServer}, defaultExpiration)
	}
	t.Errorf("couldn't connect to memcached on %s", memcacheTestServer)
	t.FailNow()
	panic("")
}

func TestMemcachedCache_TypicalGetSet(t *testing.T) {
	typicalGetSet(t, newMemcachedStore)
}

func TestMemcachedCache_IncrDecr(t *testing.T) {
	incrDecr(t, newMemcachedStore)
}

func TestMemcachedCache_Expiration(t *testing.T) {
	expiration(t, newMemcachedStore)
}

func TestMemcachedCache_EmptyCache(t *testing.T) {
	emptyCache(t, newMemcachedStore)
}

func TestMemcachedCache_Replace(t *testing.T) {
	testReplace(t, newMemcachedStore)
}

func TestMemcachedCache_Add(t *testing.T) {
	testAdd(t, newMemcachedStore)
}
