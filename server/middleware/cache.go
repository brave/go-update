package middleware

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/brave/go-update/logger"
)

// Package middleware provides HTTP middleware functions for caching and other common operations.
//
// Example usage:
//
//	cache := middleware.NewJSONCache()
//	r.With(middleware.JSONCacheMiddleware(cache)).Get("/api/data", handler)

// CacheEntry represents a cached response
type CacheEntry struct {
	Data         []byte
	LastModified time.Time
}

// JSONCache uses atomic.Value for safe lock-free reads, optimized for single-entry caching
// with high read concurrency
type JSONCache struct {
	entry atomic.Value // stores *CacheEntry or nil
}

func NewJSONCache() *JSONCache {
	cache := &JSONCache{}
	// Start with nil (no cached data)
	cache.entry.Store((*CacheEntry)(nil))
	return cache
}

func (c *JSONCache) GetEntry() *CacheEntry {
	if entry := c.entry.Load().(*CacheEntry); entry != nil {
		return entry
	}
	return nil
}

func (c *JSONCache) Get() []byte {
	if entry := c.GetEntry(); entry != nil {
		return entry.Data
	}
	return nil
}

func (c *JSONCache) Set(data []byte) {
	newEntry := &CacheEntry{
		Data:         data,
		LastModified: time.Now(),
	}
	c.entry.Store(newEntry)
}

func (c *JSONCache) Invalidate() {
	c.entry.Store((*CacheEntry)(nil))
}

func (c *JSONCache) GetLastModified() time.Time {
	if entry := c.GetEntry(); entry != nil {
		return entry.LastModified
	}
	return time.Time{}
}

type JSONCacheConfig struct {
	MaxAge time.Duration
}

func DefaultJSONCacheConfig() JSONCacheConfig {
	return JSONCacheConfig{
		MaxAge: 10 * time.Minute,
	}
}

func JSONCacheMiddleware(cache *JSONCache, config ...JSONCacheConfig) func(next http.Handler) http.Handler {
	cfg := DefaultJSONCacheConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logger.FromContext(r.Context())

			// Check cache with atomic load - lock-free operation
			if entry := cache.GetEntry(); entry != nil {
				// Set response headers
				w.Header().Set("content-type", "application/json")
				w.Header().Set("cache-control", fmt.Sprintf("public, max-age=%d", int(cfg.MaxAge.Seconds())))
				w.Header().Set("last-modified", entry.LastModified.UTC().Format(http.TimeFormat))

				// Handle conditional requests
				if ifModSince := r.Header.Get("if-modified-since"); ifModSince != "" {
					if t, err := time.Parse(http.TimeFormat, ifModSince); err == nil {
						if !entry.LastModified.After(t) {
							w.WriteHeader(http.StatusNotModified)
							return
						}
					}
				}

				w.WriteHeader(http.StatusOK)

				// nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter
				if _, err := w.Write(entry.Data); err != nil {
					logger.Error("Error writing cached response", "error", err)
				}
				return
			}

			// Cache miss, continue to handler
			next.ServeHTTP(w, r)
		})
	}
}
