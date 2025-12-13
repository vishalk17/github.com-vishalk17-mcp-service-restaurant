package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var logLevel string

func init() {
	logLevel = strings.ToLower(os.Getenv("LOG_LEVEL"))
	if logLevel == "" {
		logLevel = "info" // default
	}
}

// IsDebug returns true if LOG_LEVEL=debug
func IsDebug() bool { return logLevel == "debug" }

// LoggingMiddleware logs HTTP requests based on LOG_LEVEL
// LOG_LEVEL: debug, info, warn, error, none
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if logLevel == "none" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		// Log request (debug shows headers)
		if logLevel == "debug" {
			log.Printf("➡️  %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
			log.Printf("   Headers: %v", r.Header)
		} else if logLevel == "info" {
			log.Printf("➡️  %s %s", r.Method, r.URL.Path)
		}

		// Wrap response writer to capture status
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		// Call next handler
		next.ServeHTTP(wrapped, r)

		// Log response according to level
		duration := time.Since(start)
		if logLevel == "debug" {
			log.Printf("⬅️  %s %s -> %d (took %v)", r.Method, r.URL.Path, wrapped.statusCode, duration)
		} else if logLevel == "info" {
			if wrapped.statusCode >= 400 {
				log.Printf("⬅️  %s %s -> %d (took %v)", r.Method, r.URL.Path, wrapped.statusCode, duration)
			}
		} else if logLevel == "warn" && wrapped.statusCode >= 400 {
			log.Printf("⚠️  %s %s -> %d", r.Method, r.URL.Path, wrapped.statusCode)
		} else if logLevel == "error" && wrapped.statusCode >= 500 {
			log.Printf("❌ %s %s -> %d", r.Method, r.URL.Path, wrapped.statusCode)
		}
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
