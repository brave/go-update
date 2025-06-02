package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/brave/go-update/logger"
)

// Package middleware provides HTTP middleware functions for caching and other common operations.
//
// Example usage:
//
//	cache := middleware.NewJSONCache()
//	r.With(middleware.JSONCacheMiddleware(cache)).Get("/api/data", handler)

type CacheEntry struct {
	Data         []byte
	LastModified time.Time
	IsValid      bool
}

type JSONCache struct {
	mu    sync.RWMutex
	entry *CacheEntry
}

func NewJSONCache() *JSONCache {
	return &JSONCache{
		entry: &CacheEntry{},
	}
}

func (c *JSONCache) GetEntry() *CacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.entry.IsValid {
		// Return a copy to avoid race conditions
		return &CacheEntry{
			Data:         c.entry.Data,
			LastModified: c.entry.LastModified,
			IsValid:      c.entry.IsValid,
		}
	}
	return nil
}

func (c *JSONCache) Get() []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.entry.IsValid {
		return c.entry.Data
	}
	return nil
}

func (c *JSONCache) Set(data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entry.Data = data
	c.entry.LastModified = time.Now()
	c.entry.IsValid = true
}

func (c *JSONCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entry.IsValid = false
}

func (c *JSONCache) GetLastModified() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.entry.LastModified
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

			// Single lock operation to get cache entry with metadata
			entry := cache.GetEntry()
			if entry != nil {
				w.Header().Set("content-type", "application/json")
				w.Header().Set("cache-control", fmt.Sprintf("public, max-age=%d", int(cfg.MaxAge.Seconds())))
				w.Header().Set("last-modified", entry.LastModified.UTC().Format(http.TimeFormat))

				if ifModSince := r.Header.Get("if-modified-since"); ifModSince != "" {
					if t, err := time.Parse(http.TimeFormat, ifModSince); err == nil {
						if !entry.LastModified.After(t) {
							w.WriteHeader(http.StatusNotModified)
							return
						}
					}
				}

				w.WriteHeader(http.StatusOK)

				// Write cached JSON data directly
				_, err := w.Write(entry.Data) // nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter
				if err != nil {
					logger.Error("Error writing cached response", "error", err)
				}
				return
			}

			// Cache miss, continue to handler
			next.ServeHTTP(w, r)
		})
	}
}
