package httpserver

import (
	"net/http"
	"os"
	"time"

	"securecollab/services/gateway/internal/observability"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_http_requests_total",
			Help: "Total HTTP requests handled by gateway.",
		},
		[]string{"method", "path", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_http_request_duration_seconds",
			Help:    "Duration of HTTP requests handled by gateway.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	rateLimitedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_rate_limited_requests_total",
			Help: "Total rate-limited requests in gateway.",
		},
		[]string{"method", "path"},
	)

	rateLimitErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_rate_limiter_errors_total",
			Help: "Total errors from configured rate limiter backends.",
		},
		[]string{"algorithm"},
	)
)

func init() {
	prometheus.MustRegister(requestTotal, requestDuration, rateLimitedTotal, rateLimitErrorsTotal)
}

type RouterConfig struct {
	JWTSecret             string
	SlidingWindowLimit    int
	SlidingWindowDuration time.Duration
	TokenBucketCapacity   int
	TokenBucketRefillRate float64
	RedisAddr             string
}

func NewRouter() *gin.Engine {
	return NewRouterWithConfig(RouterConfig{})
}

func NewRouterWithConfig(cfg RouterConfig) *gin.Engine {
	logger := observability.NewLogger()
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestMetrics())
	router.Use(requestLogging(logger))

	secret := cfg.JWTSecret
	if secret == "" {
		secret = jwtSecretFromEnv()
	}
	if cfg.RedisAddr == "" {
		cfg.RedisAddr = os.Getenv("REDIS_ADDR")
	}

	sliding, bucket := newRateLimiters(cfg)

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := router.Group("/v1")
	v1.Use(authMiddleware(secret))
	v1.Use(rateLimitMiddleware(sliding, bucket))
	v1.GET("/protected/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "user_id": c.GetString("user_id")})
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	return router
}

func requestMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := http.StatusText(c.Writer.Status())
		requestTotal.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
		requestDuration.WithLabelValues(c.Request.Method, c.FullPath(), status).Observe(time.Since(start).Seconds())
	}
}
