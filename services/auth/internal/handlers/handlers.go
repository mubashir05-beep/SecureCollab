package handlers

import (
	"errors"
	"net/http"
	"strings"

	"securecollab/services/auth/internal/auth"
	"securecollab/services/auth/internal/store"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

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
}

func NewRouter(userStore store.UserStore) *gin.Engine {
	if userStore == nil {
		userStore = store.NewInMemoryUserStore()
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.POST("/register", handleRegister(userStore))
	router.POST("/login", handleLogin(userStore))
	router.POST("/refresh", handleRefresh())
	router.GET("/healthz", handleHealth)

	return router
}

func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func handleRegister(userStore store.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
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

		c.JSON(http.StatusCreated, TokenResponse{AccessToken: token})
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

		c.JSON(http.StatusOK, TokenResponse{AccessToken: token})
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
