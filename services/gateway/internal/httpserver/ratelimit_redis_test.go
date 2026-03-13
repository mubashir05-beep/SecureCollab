package httpserver

import "testing"

func TestNewRateLimiters_WithNoRedis_UsesInMemoryLimiters(t *testing.T) {
	sliding, bucket := newRateLimiters(RouterConfig{RedisAddr: ""})

	if _, ok := sliding.(*SlidingWindowLimiter); !ok {
		t.Fatalf("expected in-memory sliding limiter, got %T", sliding)
	}
	if _, ok := bucket.(*TokenBucketLimiter); !ok {
		t.Fatalf("expected in-memory token bucket limiter, got %T", bucket)
	}
}

func TestNewRateLimiters_WithUnavailableRedis_FallsBackToInMemory(t *testing.T) {
	cfg := RouterConfig{
		RedisAddr:             "127.0.0.1:1",
		SlidingWindowLimit:    3,
		TokenBucketCapacity:   3,
		TokenBucketRefillRate: 1,
	}

	sliding, bucket := newRateLimiters(cfg)

	if _, ok := sliding.(*SlidingWindowLimiter); !ok {
		t.Fatalf("expected in-memory sliding limiter fallback, got %T", sliding)
	}
	if _, ok := bucket.(*TokenBucketLimiter); !ok {
		t.Fatalf("expected in-memory token bucket fallback, got %T", bucket)
	}
}
