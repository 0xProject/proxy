package main

import (
	"sync"
	"time"
)

type Cacher interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte)
}

type Item struct {
	value    []byte
	expireAt time.Time
}

func (i *Item) Bytes() []byte {
	return i.value
}

func (i *Item) Expired() bool {
	return i.expireAt.Before(time.Now())
}

type InMemoryCache struct {
	mu sync.RWMutex

	storage     map[string]*Item
	cacheConfig *CacheConfig
}

func (cache *InMemoryCache) Get(key string) ([]byte, bool) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	value, ok := cache.storage[key]
	if !ok || value.Expired() {
		return nil, false
	}

	return value.Bytes(), true
}

func (cache *InMemoryCache) Set(key string, value []byte) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.storage[key] = &Item{
		value:    value,
		expireAt: time.Now().Add(cache.cacheConfig.CacheExpiration),
	}
}

func NewInMemoryCache(cacheConfig *CacheConfig) Cacher {
	return &InMemoryCache{
		storage:     map[string]*Item{},
		cacheConfig: cacheConfig,
	}
}
