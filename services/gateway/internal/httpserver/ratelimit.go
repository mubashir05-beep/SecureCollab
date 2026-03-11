package httpserver

import (
	"context"
	"math"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

type SlidingWindowLimiter struct {
	mu      sync.Mutex
	window  time.Duration
	limit   int
	events  map[string][]time.Time
	clockFn func() time.Time
}

func NewSlidingWindowLimiter(limit int, window time.Duration) *SlidingWindowLimiter {
	if limit <= 0 {
		limit = 60
	}
	if window <= 0 {
		window = time.Minute
	}

	return &SlidingWindowLimiter{
		window:  window,
		limit:   limit,
		events:  make(map[string][]time.Time),
		clockFn: time.Now,
	}
}

func (l *SlidingWindowLimiter) Allow(_ context.Context, key string) (bool, error) {
	now := l.clockFn()
	cutoff := now.Add(-l.window)

	l.mu.Lock()
	defer l.mu.Unlock()

	events := l.events[key]
	idx := 0
	for idx < len(events) && events[idx].Before(cutoff) {
		idx++
	}
	if idx > 0 {
		events = events[idx:]
	}

	if len(events) >= l.limit {
		l.events[key] = events
		return false, nil
	}

	events = append(events, now)
	l.events[key] = events
	return true, nil
}

type tokenBucketState struct {
	tokens     float64
	lastRefill time.Time
}

type TokenBucketLimiter struct {
	mu           sync.Mutex
	capacity     float64
	refillPerSec float64
	buckets      map[string]tokenBucketState
	clockFn      func() time.Time
}

func NewTokenBucketLimiter(capacity int, refillPerSecond float64) *TokenBucketLimiter {
	if capacity <= 0 {
		capacity = 30
	}
	if refillPerSecond <= 0 {
		refillPerSecond = 10
	}

	return &TokenBucketLimiter{
		capacity:     float64(capacity),
		refillPerSec: refillPerSecond,
		buckets:      make(map[string]tokenBucketState),
		clockFn:      time.Now,
	}
}

func (l *TokenBucketLimiter) Allow(_ context.Context, key string) (bool, error) {
	now := l.clockFn()

	l.mu.Lock()
	defer l.mu.Unlock()

	state, ok := l.buckets[key]
	if !ok {
		state = tokenBucketState{tokens: l.capacity, lastRefill: now}
	}

	elapsed := now.Sub(state.lastRefill).Seconds()
	state.tokens = math.Min(l.capacity, state.tokens+elapsed*l.refillPerSec)
	state.lastRefill = now

	if state.tokens < 1 {
		l.buckets[key] = state
		return false, nil
	}

	state.tokens -= 1
	l.buckets[key] = state
	return true, nil
}

func rateLimitMiddleware(sliding RateLimiter, bucket RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := rateLimitKey(c)
		allowedSliding, err := sliding.Allow(c.Request.Context(), key)
		if err != nil {
			rateLimitErrorsTotal.WithLabelValues("sliding_window").Inc()
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "rate limiter unavailable"})
			return
		}

		allowedBucket, err := bucket.Allow(c.Request.Context(), key)
		if err != nil {
			rateLimitErrorsTotal.WithLabelValues("token_bucket").Inc()
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "rate limiter unavailable"})
			return
		}

		if !allowedSliding || !allowedBucket {
			rateLimitedTotal.WithLabelValues(c.Request.Method, c.FullPath()).Inc()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

func rateLimitKey(c *gin.Context) string {
	if userID, ok := c.Get("user_id"); ok {
		if value, ok := userID.(string); ok && value != "" {
			return "user:" + value
		}
	}

	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err == nil && host != "" {
		return "ip:" + host
	}
	if c.ClientIP() != "" {
		return "ip:" + c.ClientIP()
	}

	return "ip:unknown"
}
