// Package memcache provides a constructor for the go-cache
package memcache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

// NewMemoryCache constructs a new instance of go-cache's memory cache
func NewMemoryCache() *cache.Cache {
	return cache.New(
		time.Minute*5,
		time.Minute*10,
	)
}
