package cache

import (
	"testing"
	"time"
)

var newGoCacheStore = func(_ *testing.T, defaultExpiration time.Duration) Store {
	return NewGoCacheStore(defaultExpiration, time.Second)
}

func TestGoCacheCache_TypicalGetSet(t *testing.T) {
	typicalGetSet(t, newGoCacheStore)
}

func TestGoCacheCache_IncrDecr(t *testing.T) {
	incrDecr(t, newGoCacheStore)
}

func TestGoCacheCache_Expiration(t *testing.T) {
	expiration(t, newGoCacheStore)
}

func TestGoCacheCache_EmptyCache(t *testing.T) {
	emptyCache(t, newGoCacheStore)
}

func TestGoCacheCache_Replace(t *testing.T) {
	testReplace(t, newGoCacheStore)
}

func TestGoCacheCache_Add(t *testing.T) {
	testAdd(t, newGoCacheStore)
}
