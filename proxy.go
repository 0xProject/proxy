package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type cachedProxy struct {
	proxy       *httputil.ReverseProxy
	cache       Cacher
	proxyConfig *ProxyConfig
}

func (c *cachedProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.proxy.ServeHTTP(w, r)
}

func NewCachedProxy(pc *ProxyConfig, cacheConfig *CacheConfig) (http.Handler, error) {
	proxy, err := NewProxy(pc)
	if err != nil {
		return nil, err
	}

	cache := NewInMemoryCache(cacheConfig)
	proxy.Transport = cache

	cachedProxy := &cachedProxy{
		cache:       cache,
		proxy:       proxy,
		proxyConfig: pc,
	}

	return cachedProxy, nil
}

func NewProxy(pc *ProxyConfig) (*httputil.ReverseProxy, error) {
	targetURL, err := url.Parse(pc.TargetURL)
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.Host = targetURL.Host
		if pc.QueryParamName != "" {
			// Add new query params
			newQueryValues := req.URL.Query()
			newQueryValues.Add(pc.QueryParamName, pc.QueryParamValue)
			req.URL.RawQuery = newQueryValues.Encode()
		}
		if pc.HeaderName != "" {
			req.Header.Add(pc.HeaderName, pc.HeaderValue)
		}
	}

	if pc.ResponseTimeout != 0 {
		proxy.Transport = &http.Transport{
			ResponseHeaderTimeout: time.Duration(pc.ResponseTimeout) * time.Second,
		}
	}

	return proxy, nil
}
