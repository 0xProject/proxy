package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStateDiskStorage(t *testing.T) {

	cache := NewInMemoryCache(&CacheConfig{
		CacheExpiration: 2 * time.Minute,
	})

	testStoreKey := "testkey"
	testStoreValue := []byte("hey")
	cache.Set(testStoreKey, testStoreValue)

	value, ok := cache.Get(testStoreKey)
	require.Equal(t, ok, true)
	require.Equal(t, value, testStoreValue)
}
