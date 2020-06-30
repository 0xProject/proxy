package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	log "github.com/sirupsen/logrus"
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

func TestEthGasStationRoundtrip(t *testing.T) {
	requestTarget := "https://ethgasstation.info/api/ethgasAPI.json?api-key=x"
	query := "/api/ethgasAPI.json?hey=whats-the-rate-limit"
	log.SetLevel(6)
	cachedProxy, err := NewCachedProxy(&ProxyConfig{
		TargetURL: requestTarget,
	}, &CacheConfig{
		CacheExpiration: 5 * time.Second,
	})
	require.NoError(t, err)

	ps := httptest.NewServer(cachedProxy)

	proxyClient := ps.Client()

	_, err = proxyClient.Get(ps.URL + query)
	require.NoError(t, err)

	for i := 0; i < 8; i++ {
		resp2, err := proxyClient.Get(ps.URL + query)
		require.NoError(t, err)
		_, err = ioutil.ReadAll(resp2.Body)
		require.NoError(t, err)
		time.Sleep(1 * time.Second)
	}

}
