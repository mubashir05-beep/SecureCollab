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
	router := setupUnavailableLimiterRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/v1err/ping", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestToken(t, "test-secret", "user-x"))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	code := res.Code
	if code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, code)
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
