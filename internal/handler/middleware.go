package handler

import (
    "context"
    "log"
    "net/http"
    "time"

    "github.com/google/uuid"
)
type ContextKey string

const RequestIDKey ContextKey = "request-id"
// RequestIDMiddleware adds unique request ID to all requests
func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        w.Header().Set("X-Request-ID", requestID)
        ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// LoggingMiddleware logs HTTP requests with timing and request ID
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        next.ServeHTTP(wrapped, r)

        duration := time.Since(start)
        requestID := GetRequestID(r.Context())

        log.Printf("[%s] %s %s %s - %d (%dms)",
            requestID, r.Method, r.RequestURI, r.RemoteAddr, wrapped.statusCode, duration.Milliseconds())
    })
}

// RecoveryMiddleware handles panics gracefully
func RecoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                requestID := GetRequestID(r.Context())
                log.Printf("[%s] [PANIC] %v", requestID, err)
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// RateLimitMiddleware implements simple rate limiting per IP
func RateLimitMiddleware(requestsPerSecond int) func(http.Handler) http.Handler {
    limiter := NewRateLimiter(requestsPerSecond)

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            clientIP := r.RemoteAddr
            if !limiter.Allow(clientIP) {
                requestID := GetRequestID(r.Context())
                log.Printf("[%s] Rate limit exceeded for IP: %s", requestID, clientIP)
                http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) string {
    id, ok := ctx.Value(RequestIDKey).(string)
    if !ok {
        return "unknown"
    }
    return id
}