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
type JSONCache struct {
	mu           sync.RWMutex
	cachedJSON   []byte
	lastModified time.Time
	isValid      bool
}

func NewJSONCache() *JSONCache {
	return &JSONCache{}
}

func (c *JSONCache) Get() []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.isValid {
		return c.cachedJSON
	}
	return nil
}

func (c *JSONCache) Set(data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cachedJSON = data
	c.lastModified = time.Now()
	c.isValid = true
}

func (c *JSONCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.isValid = false
}

func (c *JSONCache) GetLastModified() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastModified
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

			data := cache.Get()
			if data != nil {
				w.Header().Set("content-type", "application/json")
				w.Header().Set("cache-control", fmt.Sprintf("public, max-age=%d", int(cfg.MaxAge.Seconds())))
				lastModified := cache.GetLastModified()
				w.Header().Set("last-modified", lastModified.UTC().Format(http.TimeFormat))

				if ifModSince := r.Header.Get("if-modified-since"); ifModSince != "" {
					if t, err := time.Parse(http.TimeFormat, ifModSince); err == nil {
						if !lastModified.After(t) {
							w.WriteHeader(http.StatusNotModified)
							return
						}
					}
				}

				w.WriteHeader(http.StatusOK)

				// Write cached JSON data directly
				_, err := w.Write(data)
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
