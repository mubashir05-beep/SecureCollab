package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type failingLimiter struct{}

func (f failingLimiter) Allow(_ context.Context, _ string) (bool, error) {
	return false, context.DeadlineExceeded
}

func TestHealthz_ReturnsOK(t *testing.T) {
	router := NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}

	var payload map[string]string
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse json body: %v", err)
	}

	if payload["status"] != "ok" {
		t.Fatalf("expected status field to be ok, got %q", payload["status"])
	}
}

func TestMetrics_EndpointAvailable(t *testing.T) {
	router := NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
}

func TestProtectedRoute_WithoutToken_ReturnsUnauthorized(t *testing.T) {
	router := NewRouterWithConfig(RouterConfig{JWTSecret: "test-secret"})
	req := httptest.NewRequest(http.MethodGet, "/v1/protected/ping", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, res.Code)
	}
}

func TestProtectedRoute_WithValidToken_ReturnsOK(t *testing.T) {
	secret := "test-secret"
	router := NewRouterWithConfig(RouterConfig{JWTSecret: secret})

	req := httptest.NewRequest(http.MethodGet, "/v1/protected/ping", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, secret, "user-123"))
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
}

func TestProtectedRoute_SlidingWindowRateLimit_ReturnsTooManyRequests(t *testing.T) {
	secret := "test-secret"
	router := NewRouterWithConfig(RouterConfig{
		JWTSecret:             secret,
		SlidingWindowLimit:    2,
		SlidingWindowDuration: time.Minute,
		TokenBucketCapacity:   10,
		TokenBucketRefillRate: 10,
	})

	headers := map[string]string{"Authorization": "Bearer " + generateTestToken(t, secret, "user-sw")}
	serveProtectedRequest(router, headers)
	serveProtectedRequest(router, headers)
	code := serveProtectedRequest(router, headers)

	if code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, code)
	}
}

func TestProtectedRoute_TokenBucketRateLimit_ReturnsTooManyRequests(t *testing.T) {
	secret := "test-secret"
	router := NewRouterWithConfig(RouterConfig{
		JWTSecret:             secret,
		SlidingWindowLimit:    100,
		SlidingWindowDuration: time.Minute,
		TokenBucketCapacity:   1,
		TokenBucketRefillRate: 0.001,
	})

	headers := map[string]string{"Authorization": "Bearer " + generateTestToken(t, secret, "user-tb")}
	first := serveProtectedRequest(router, headers)
	second := serveProtectedRequest(router, headers)

	if first != http.StatusOK {
		t.Fatalf("expected first request status %d, got %d", http.StatusOK, first)
	}
	if second != http.StatusTooManyRequests {
		t.Fatalf("expected second request status %d, got %d", http.StatusTooManyRequests, second)
	}
}

func TestRateLimitMiddleware_BackendUnavailable_ReturnsServiceUnavailable(t *testing.T) {
	before := counterVecValue(t, rateLimitErrorsTotal, "sliding_window")

	router := setupUnavailableLimiterRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/v1err/ping", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, "test-secret", "user-x"))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	code := res.Code
	if code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, code)
	}

	after := counterVecValue(t, rateLimitErrorsTotal, "sliding_window")
	if after-before != 1 {
		t.Fatalf("expected gateway_rate_limiter_errors_total to increment by 1, got delta %v", after-before)
	}
}

func TestRateLimitPolicy_PerUserIsolation(t *testing.T) {
	secret := "test-secret"
	router := NewRouterWithConfig(RouterConfig{
		JWTSecret:             secret,
		SlidingWindowLimit:    1,
		SlidingWindowDuration: time.Minute,
		TokenBucketCapacity:   100,
		TokenBucketRefillRate: 100,
	})

	headersA := map[string]string{"Authorization": "Bearer " + generateTestToken(t, secret, "user-a")}
	headersB := map[string]string{"Authorization": "Bearer " + generateTestToken(t, secret, "user-b")}

	if got := serveProtectedRequest(router, headersA); got != http.StatusOK {
		t.Fatalf("expected first user-a request status %d, got %d", http.StatusOK, got)
	}
	if got := serveProtectedRequest(router, headersA); got != http.StatusTooManyRequests {
		t.Fatalf("expected second user-a request status %d, got %d", http.StatusTooManyRequests, got)
	}
	if got := serveProtectedRequest(router, headersB); got != http.StatusOK {
		t.Fatalf("expected first user-b request status %d, got %d", http.StatusOK, got)
	}
}

func TestRateLimitPolicy_PerIPFallback(t *testing.T) {
	router := gin.New()
	router.Use(rateLimitMiddleware(NewSlidingWindowLimiter(1, time.Minute), NewTokenBucketLimiter(100, 100)))
	router.GET("/ip-only", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	first := httptest.NewRequest(http.MethodGet, "/ip-only", nil)
	first.RemoteAddr = "10.1.1.9:1234"
	firstRes := httptest.NewRecorder()
	router.ServeHTTP(firstRes, first)
	if firstRes.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, firstRes.Code)
	}

	second := httptest.NewRequest(http.MethodGet, "/ip-only", nil)
	second.RemoteAddr = "10.1.1.9:4321"
	secondRes := httptest.NewRecorder()
	router.ServeHTTP(secondRes, second)
	if secondRes.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, secondRes.Code)
	}

	third := httptest.NewRequest(http.MethodGet, "/ip-only", nil)
	third.RemoteAddr = "10.1.1.10:5678"
	thirdRes := httptest.NewRecorder()
	router.ServeHTTP(thirdRes, third)
	if thirdRes.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, thirdRes.Code)
	}
}

func TestRateLimitedMetric_IncrementsOnTooManyRequests(t *testing.T) {
	secret := "test-secret"
	router := NewRouterWithConfig(RouterConfig{
		JWTSecret:             secret,
		SlidingWindowLimit:    1,
		SlidingWindowDuration: time.Minute,
		TokenBucketCapacity:   100,
		TokenBucketRefillRate: 100,
	})

	headers := map[string]string{"Authorization": "Bearer " + generateTestToken(t, secret, "metric-user")}
	path := "/v1/protected/ping"
	before := counterVecValue(t, rateLimitedTotal, http.MethodGet, path)

	serveProtectedRequest(router, headers)
	code := serveProtectedRequest(router, headers)
	if code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, code)
	}

	after := counterVecValue(t, rateLimitedTotal, http.MethodGet, path)
	if after-before != 1 {
		t.Fatalf("expected gateway_rate_limited_requests_total to increment by 1, got delta %v", after-before)
	}
}

func setupUnavailableLimiterRouter(t *testing.T) http.Handler {
	t.Helper()

	r := NewRouterWithConfig(RouterConfig{JWTSecret: "test-secret"})
	// Build a dedicated test route to assert middleware behavior on limiter error.
	v1 := r.Group("/v1err")
	v1.Use(authMiddleware("test-secret"))
	v1.Use(rateLimitMiddleware(failingLimiter{}, NewTokenBucketLimiter(1, 1)))
	v1.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	return r
}

func serveProtectedRequest(router http.Handler, headers map[string]string) int {
	req := httptest.NewRequest(http.MethodGet, "/v1/protected/ping", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res.Code
}

func generateTestToken(t *testing.T, secret, userID string) string {
	t.Helper()

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(10 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}
	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := tkn.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	return tokenString
}

func counterVecValue(t *testing.T, vec *prometheus.CounterVec, labels ...string) float64 {
	t.Helper()

	metric, err := vec.GetMetricWithLabelValues(labels...)
	if err != nil {
		t.Fatalf("get metric with labels %v: %v", labels, err)
	}

	pb := &dto.Metric{}
	if err := metric.Write(pb); err != nil {
		t.Fatalf("write metric protobuf: %v", err)
	}

	return pb.GetCounter().GetValue()
}
