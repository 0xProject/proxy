package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

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
		// Add new query params
		newQueryValues := req.URL.Query()
		newQueryValues.Add(pc.QueryParamName, pc.QueryParamValue)
		req.URL.RawQuery = newQueryValues.Encode()
	}
	return proxy, nil
}
