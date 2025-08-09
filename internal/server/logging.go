package server

import (
	"log"
	"net/http"
	"time"
)

// LoggingResponseWriter wraps http.ResponseWriter to capture status code
type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{
		ResponseWriter: w,
		statusCode:     200, // Default status
	}
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *LoggingResponseWriter) Write(data []byte) (int, error) {
	n, err := lrw.ResponseWriter.Write(data)
	lrw.size += n
	return n, err
}

// LoggingMiddleware creates HTTP access logging middleware
func LoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Wrap response writer to capture status
			lrw := NewLoggingResponseWriter(w)
			
			// Process request
			next.ServeHTTP(lrw, r)
			
			// Log in Common Log Format with additional info
			duration := time.Since(start)
			userAgent := r.Header.Get("User-Agent")
			if userAgent == "" {
				userAgent = "-"
			}
			
			// Get client IP, handling forwarded headers
			clientIP := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				clientIP = forwarded
			} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
				clientIP = realIP
			}
			
			// Get authentication info for logging (without exposing the key)
			authMethod := "-"
			if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
				authMethod = "api_key"
			} else if auth := r.Header.Get("Authorization"); auth != "" {
				if auth == "Bearer " {
					authMethod = "bearer"
				} else if auth == "ApiKey " {
					authMethod = "api_key"
				}
			}
			
			// Log format: IP - - [timestamp] "METHOD /path HTTP/1.1" status size "User-Agent" duration auth_method
			log.Printf("%s - - [%s] \"%s %s %s\" %d %d \"%s\" %dms %s",
				clientIP,
				start.Format("02/Jan/2006:15:04:05 -0700"),
				r.Method,
				r.RequestURI,
				r.Proto,
				lrw.statusCode,
				lrw.size,
				userAgent,
				duration.Milliseconds(),
				authMethod,
			)
		})
	}
}