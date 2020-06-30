package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"

	log "github.com/sirupsen/logrus"
)

func writeError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, `{"result":"","error":%q}`, message)
}

type cachedProxy struct {
	proxy       *httputil.ReverseProxy
	cache       Cacher
	proxyConfig *ProxyConfig
}

// setModifyResponse modifies the response to store it in cache
func (c *cachedProxy) setModifyResponse() *cachedProxy {
	c.proxy.ModifyResponse = c.updateResponseCache
	return c
}

func (c *cachedProxy) updateResponseCache(res *http.Response) error {
	if res.StatusCode >= 300 {
		log.Debug("status code >= 300 not caching response")
		return nil
	}
	b, _ := ioutil.ReadAll(res.Body)
	res.Body = ioutil.NopCloser(bytes.NewBuffer(b))

	requestURI := res.Request.URL.RequestURI()
	log.WithField("requestURI", requestURI).Debug("setting cache for request")
	c.cache.Set(requestURI, b)
	return nil
}

func (c *cachedProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestURI := r.URL.RequestURI()
	if c.proxyConfig.QueryParamName != "" {
		requestQuery := r.URL.Query()
		requestQuery.Add(c.proxyConfig.QueryParamName, c.proxyConfig.QueryParamValue)
		requestURI = fmt.Sprintf("%s?%s", requestURI, requestQuery.Encode())
	}
	value, ok := c.cache.Get(requestURI)
	// Serve from memory
	if ok {
		log.WithField("requestURI", requestURI).Debug("serving request from memory")
		// NOTE: This will only return an error when:
		// - If the connection was hijacked (see http.Hijacker): http.ErrHijacked
		// - If writing data to the actual connection fails.
		// This also automatically sets the headers.
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write(value)
		if err != nil {
			log.WithError(err).Error("failed to return back value")
			writeError(w, "failed to handle request")
			return
		}

		return
	}

	log.WithField("requestURI", requestURI).Debug("serving request through proxy")
	c.proxy.ServeHTTP(w, r)
}

func NewCachedProxy(pc *ProxyConfig, cacheConfig *CacheConfig) (http.Handler, error) {
	proxy, err := NewProxy(pc)
	if err != nil {
		return nil, err
	}

	cache := NewInMemoryCache(cacheConfig)

	cachedProxy := &cachedProxy{
		cache:       cache,
		proxy:       proxy,
		proxyConfig: pc,
	}

	return cachedProxy.setModifyResponse(), nil
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

	return proxy, nil
}
