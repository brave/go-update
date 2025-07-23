package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/klauspost/compress/gzhttp"

	"github.com/brave/go-update/logger"
)

// OptimizedCompress is a middleware that provides HTTP compression using klauspost's optimized
// gzip compression library. If that fails, it falls back to Chi's standard compression.
//
// The level parameter specifies the compression level (1-9, with 1 being fastest)
// The minSize parameter specifies the minimum size (in bytes) for responses to be compressed
// The types parameter specifies which content types should be compressed.
func OptimizedCompress(level, minSize int, types ...string) func(next http.Handler) http.Handler {
	gzWrapper, err := gzhttp.NewWrapper(
		gzhttp.CompressionLevel(level),
		gzhttp.MinSize(minSize),
		gzhttp.ContentTypes(types),
	)

	if err == nil {
		return func(next http.Handler) http.Handler {
			return gzWrapper(next)
		}
	}

	logger.New().Warn("Failed to use optimized compression, falling back to standard implementation", "error", err)

	compressor := middleware.NewCompressor(level, types...)
	return compressor.Handler
}
