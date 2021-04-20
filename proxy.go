package main

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"time"
)

var (
	swapQuotePath = regexp.MustCompile(`(.*)/swap/v1/quote(.*)`)
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
		newQueryValues := req.URL.Query()
		if swapQuotePath.MatchString(req.URL.Path) {
			newQueryValues.Add("skipValidation", "false")
		}
		if pc.QueryParamName != "" {
			// Add new query params
			newQueryValues.Add(pc.QueryParamName, pc.QueryParamValue)
		}
		if pc.HeaderName != "" {
			req.Header.Add(pc.HeaderName, pc.HeaderValue)
		}

		req.URL.RawQuery = newQueryValues.Encode()
	}

	proxy.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: pc.ResponseTimeout,
	}

	return proxy, nil
}
