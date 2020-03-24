package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type responder struct {
	value []byte
}

func (r *responder) SetValue(value []byte) {
	r.value = value
}

func (r *responder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Write(r.value)
}

func TestProxyRoundtrip(t *testing.T) {
	firstValue := []byte(fmt.Sprintf("{%q:%q}", "message", "hello"))
	secondValue := []byte(fmt.Sprintf("{%q:%q}", "message", "world"))

	backend := &responder{value: firstValue}
	ts := httptest.NewServer(backend)
	defer ts.Close()

	cachedProxy, err := NewCachedProxy(&ProxyConfig{
		TargetURL: ts.URL,
	}, &CacheConfig{
		CacheExpiration: 2 * time.Minute,
	})
	require.NoError(t, err)

	ps := httptest.NewServer(cachedProxy)

	proxyClient := ps.Client()

	resp, err := proxyClient.Get(ps.URL)
	require.NoError(t, err)

	firstResponseBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, firstResponseBody, firstValue)

	// We set on the backend a new value and check whether or not the proxy
	// responds with the cached request.
	backend.SetValue(secondValue)

	resp2, err := proxyClient.Get(ps.URL)
	require.NoError(t, err)

	secondResponseBody, err := ioutil.ReadAll(resp2.Body)
	require.NoError(t, err)
	require.Equal(t, secondResponseBody, firstValue)
}
