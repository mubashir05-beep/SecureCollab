package handlers

import (
	"net/http"

	"securecollab/services/auth/internal/auth"

	"github.com/gin-gonic/gin"
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

func NewRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.POST("/register", handleRegister)
	router.POST("/login", handleLogin)
	router.POST("/refresh", handleRefresh)
	router.GET("/healthz", handleHealth)

	return router
}

func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func handleRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Placeholder: in production, validate and store credentials.
	// For now, return success.
	token, err := auth.GenerateAccessToken(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, TokenResponse{AccessToken: token})
}

func handleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Placeholder: in production, validate credentials against database.
	// For now, return success.
	token, err := auth.GenerateAccessToken(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{AccessToken: token})
}

func handleRefresh(c *gin.Context) {
	// Placeholder: in production, validate refresh token and issue new access token.
	c.JSON(http.StatusOK, gin.H{"message": "refresh pending implementation"})
}
