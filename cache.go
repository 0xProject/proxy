package main

import (
	"bufio"
	"bytes"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func cacheKey(r *http.Request) string {
	return r.URL.String()
}

func cachedResponse(b []byte, r *http.Request) (*http.Response, error) {
	buf := bytes.NewBuffer(b)
	return http.ReadResponse(bufio.NewReader(buf), r)
}

type Cacher interface {
	Get(r *http.Request) ([]byte, bool)
	Set(r *http.Request, value []byte)
	RoundTrip(r *http.Request) (*http.Response, error)
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

	storage           map[string]*Item
	cacheConfig       *CacheConfig
	originalTransport http.RoundTripper
}

func (cache *InMemoryCache) Get(r *http.Request) ([]byte, bool) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	key := cacheKey(r)
	value, ok := cache.storage[key]
	if !ok || value.Expired() {
		return nil, false
	}

	return value.Bytes(), true
}

func (cache *InMemoryCache) Set(r *http.Request, value []byte) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	key := cacheKey(r)
	cache.storage[key] = &Item{
		value:    value,
		expireAt: time.Now().Add(cache.cacheConfig.CacheExpiration),
	}
}

func (cache *InMemoryCache) RoundTrip(r *http.Request) (*http.Response, error) {
	// Return the cached response if present and not expired.
	if val, ok := cache.Get(r); ok {
		log.WithField("requestURI", r.URL.String()).Debug("Fetching the response from cache")
		return cachedResponse(val, r)
	}

	log.WithField("requestURI", r.URL.String()).Debug("Serving request using default transport")
	resp, err := cache.originalTransport.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 300 {
		// Store the response in cache.
		buf, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, err
		}
		cache.Set(r, buf)
	}
	return resp, nil
}

func NewInMemoryCache(cacheConfig *CacheConfig) Cacher {
	return &InMemoryCache{
		storage:           map[string]*Item{},
		cacheConfig:       cacheConfig,
		originalTransport: http.DefaultTransport,
	}
}
