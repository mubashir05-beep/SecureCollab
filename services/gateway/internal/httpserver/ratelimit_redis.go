package httpserver

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const tokenBucketLua = `
local key = KEYS[1]
local now_ms = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local refill_per_sec = tonumber(ARGV[3])
local ttl_ms = tonumber(ARGV[4])

local values = redis.call('HMGET', key, 'tokens', 'last_ms')
local tokens = tonumber(values[1])
local last_ms = tonumber(values[2])
if tokens == nil then tokens = capacity end
if last_ms == nil then last_ms = now_ms end

local elapsed = (now_ms - last_ms) / 1000.0
if elapsed < 0 then elapsed = 0 end

tokens = math.min(capacity, tokens + elapsed * refill_per_sec)
local allowed = 0
if tokens >= 1 then
  tokens = tokens - 1
  allowed = 1
end

redis.call('HSET', key, 'tokens', tokens, 'last_ms', now_ms)
redis.call('PEXPIRE', key, ttl_ms)
return allowed
`

type RedisSlidingWindowLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

func NewRedisSlidingWindowLimiter(client *redis.Client, limit int, window time.Duration) *RedisSlidingWindowLimiter {
	if limit <= 0 {
		limit = 60
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RedisSlidingWindowLimiter{client: client, limit: limit, window: window}
}

func (l *RedisSlidingWindowLimiter) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now()
	zkey := "gateway:ratelimit:sw:" + key
	cutoff := strconv.FormatInt(now.Add(-l.window).UnixMilli(), 10)
	nowMs := strconv.FormatInt(now.UnixMilli(), 10)
	member := fmt.Sprintf("%s-%d", nowMs, now.UnixNano())

	pipe := l.client.TxPipeline()
	pipe.ZRemRangeByScore(ctx, zkey, "-inf", cutoff)
	countCmd := pipe.ZCard(ctx, zkey)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}
	if countCmd.Val() >= int64(l.limit) {
		return false, nil
	}

	pipe = l.client.TxPipeline()
	pipe.ZAdd(ctx, zkey, redis.Z{Score: float64(now.UnixMilli()), Member: member})
	pipe.Expire(ctx, zkey, l.window)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}

type RedisTokenBucketLimiter struct {
	client       *redis.Client
	capacity     int
	refillPerSec float64
	window       time.Duration
}

func NewRedisTokenBucketLimiter(client *redis.Client, capacity int, refillPerSec float64, window time.Duration) *RedisTokenBucketLimiter {
	if capacity <= 0 {
		capacity = 30
	}
	if refillPerSec <= 0 {
		refillPerSec = 10
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RedisTokenBucketLimiter{client: client, capacity: capacity, refillPerSec: refillPerSec, window: window}
}

func (l *RedisTokenBucketLimiter) Allow(ctx context.Context, key string) (bool, error) {
	hkey := "gateway:ratelimit:tb:" + key
	nowMs := time.Now().UnixMilli()
	ttlMs := l.window.Milliseconds()
	if ttlMs <= 0 {
		ttlMs = int64((time.Minute).Milliseconds())
	}

	res, err := l.client.Eval(ctx, tokenBucketLua, []string{hkey}, nowMs, l.capacity, l.refillPerSec, ttlMs).Int()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}

func newRateLimiters(cfg RouterConfig) (RateLimiter, RateLimiter) {
	sliding := NewSlidingWindowLimiter(cfg.SlidingWindowLimit, cfg.SlidingWindowDuration)
	bucket := NewTokenBucketLimiter(cfg.TokenBucketCapacity, cfg.TokenBucketRefillRate)

	if cfg.RedisAddr == "" {
		return sliding, bucket
	}

	client := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return sliding, bucket
	}

	redisSliding := NewRedisSlidingWindowLimiter(client, cfg.SlidingWindowLimit, cfg.SlidingWindowDuration)
	redisBucket := NewRedisTokenBucketLimiter(client, cfg.TokenBucketCapacity, cfg.TokenBucketRefillRate, cfg.SlidingWindowDuration)
	return redisSliding, redisBucket
}
