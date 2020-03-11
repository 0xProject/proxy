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
		req.URL.Query().Add(pc.QueryParamName, pc.QueryParamValue)
	}
	return proxy, nil
}
