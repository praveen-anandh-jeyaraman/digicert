package handler

import (
    "sync"
    "time"
)

type RateLimiter struct {
    mu       sync.RWMutex
    limits   map[string]*clientLimit
    rpsLimit int
}

type clientLimit struct {
    tokens    float64
    lastCheck time.Time
}

// NewRateLimiter creates a token bucket rate limiter
func NewRateLimiter(requestsPerSecond int) *RateLimiter {
    return &RateLimiter{
        limits:   make(map[string]*clientLimit),
        rpsLimit: requestsPerSecond,
    }
}

// Allow checks if a request from clientIP should be allowed
func (rl *RateLimiter) Allow(clientIP string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    now := time.Now()
    limit, exists := rl.limits[clientIP]

    if !exists {
        rl.limits[clientIP] = &clientLimit{
            tokens:    float64(rl.rpsLimit),
            lastCheck: now,
        }
        return true
    }

    // Add tokens based on elapsed time
    elapsed := now.Sub(limit.lastCheck).Seconds()
    limit.tokens += elapsed * float64(rl.rpsLimit)

    // Cap tokens at limit
    if limit.tokens > float64(rl.rpsLimit) {
        limit.tokens = float64(rl.rpsLimit)
    }

    limit.lastCheck = now

    if limit.tokens >= 1.0 {
        limit.tokens--
        return true
    }

    return false
}

// Reset clears rate limit data (useful for testing)
func (rl *RateLimiter) Reset() {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    rl.limits = make(map[string]*clientLimit)
}