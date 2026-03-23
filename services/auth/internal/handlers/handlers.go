package handlers

import (
	"errors"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"securecollab/services/auth/internal/auth"
	"securecollab/services/auth/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/bcrypt"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "auth_http_requests_total", Help: "Total HTTP requests."},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "auth_http_request_duration_seconds", Help: "Request duration.", Buckets: prometheus.DefBuckets},
		[]string{"method", "path", "status"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
}

func metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		status := http.StatusText(c.Writer.Status())
		httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath(), status).Observe(time.Since(start).Seconds())
	}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
}

func NewRouter(userStore store.UserStore) *gin.Engine {
	if userStore == nil {
		userStore = store.NewInMemoryUserStore()
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(metricsMiddleware())

	router.POST("/register", handleRegister(userStore))
	router.POST("/login", handleLogin(userStore))
	router.POST("/refresh", handleRefresh())
	router.GET("/healthz", handleHealth)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return router
}

func corsMiddleware() gin.HandlerFunc {
	allowedOrigin := corsOriginFromEnv()
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if allowedOrigin == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if origin != "" && originAllowed(origin, allowedOrigin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func corsOriginFromEnv() string {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "prod") {
		if origins := strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS")); origins != "" {
			return origins
		}
	}
	return "*"
}

func originAllowed(origin, allowed string) bool {
	for _, o := range strings.Split(allowed, ",") {
		if strings.TrimSpace(o) == origin {
			return true
		}
	}
	return false
}

func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func validateRegistration(req RegisterRequest) string {
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return "username must be between 3 and 50 characters"
	}
	if !emailRegex.MatchString(req.Email) {
		return "invalid email format"
	}
	if len(req.Password) < 8 {
		return "password must be at least 8 characters"
	}
	return ""
}

func handleRegister(userStore store.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		if msg := validateRegistration(req); msg != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to secure password"})
			return
		}

		user, err := userStore.CreateUser(c.Request.Context(), req.Username, req.Email, string(hashedPassword))
		if err != nil {
			if errors.Is(err, store.ErrUserExists) {
				c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
			return
		}

		token, err := auth.GenerateAccessToken(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		c.JSON(http.StatusCreated, TokenResponse{AccessToken: token, UserID: user.ID, Username: user.Username})
	}
}

func handleLogin(userStore store.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		user, err := userStore.GetUserByUsername(c.Request.Context(), req.Username)
		if err != nil {
			if errors.Is(err, store.ErrUserNotFound) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		token, err := auth.GenerateAccessToken(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, TokenResponse{AccessToken: token, UserID: user.ID, Username: user.Username})
	}
}

func handleRefresh() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(authorization, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		incomingToken := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
		claims, err := auth.ValidateToken(incomingToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		freshToken, err := auth.GenerateAccessToken(claims.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, TokenResponse{AccessToken: freshToken})
	}
}
