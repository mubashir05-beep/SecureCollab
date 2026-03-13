package handlers

import (
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"securecollab/services/keydist/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const defaultJWTSecret = "securecollab-dev-secret-key"

type UploadKeyRequest struct {
	PublicKeyB64 string `json:"public_key_b64" binding:"required"`
	KeyType      string `json:"key_type"`
}

type KeyResponse struct {
	UserID       string    `json:"user_id"`
	KeyType      string    `json:"key_type"`
	PublicKeyB64 string    `json:"public_key_b64"`
	CreatedAt    time.Time `json:"created_at"`
}

type claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func NewRouter(keyStore store.KeyStore) *gin.Engine {
	if keyStore == nil {
		keyStore = store.NewInMemoryKeyStore()
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := r.Group("/v1")
	v1.Use(authMiddleware(jwtSecretFromEnv()))
	v1.POST("/keys/identity", uploadIdentityKey(keyStore))
	v1.GET("/keys/identity/:user_id", getIdentityKey(keyStore))

	return r
}

func uploadIdentityKey(keyStore store.KeyStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
			return
		}

		var req UploadKeyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		keyType := strings.TrimSpace(req.KeyType)
		if keyType == "" {
			keyType = "identity"
		}

		keyData, err := base64.StdEncoding.DecodeString(req.PublicKeyB64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "public_key_b64 must be valid base64"})
			return
		}

		saved, err := keyStore.SavePublicKey(c.Request.Context(), userID, keyType, keyData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save public key"})
			return
		}

		c.JSON(http.StatusCreated, toKeyResponse(saved))
	}
}

func getIdentityKey(keyStore store.KeyStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := strings.TrimSpace(c.Param("user_id"))
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
			return
		}

		keyType := strings.TrimSpace(c.Query("key_type"))
		if keyType == "" {
			keyType = "identity"
		}

		key, err := keyStore.GetLatestPublicKey(c.Request.Context(), userID, keyType)
		if err != nil {
			if errors.Is(err, store.ErrKeyNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "public key not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch public key"})
			return
		}

		c.JSON(http.StatusOK, toKeyResponse(key))
	}
}

func toKeyResponse(key store.PublicKey) KeyResponse {
	return KeyResponse{
		UserID:       key.UserID,
		KeyType:      key.KeyType,
		PublicKeyB64: base64.StdEncoding.EncodeToString(key.KeyData),
		CreatedAt:    key.CreatedAt,
	}
}

func authMiddleware(secret string) gin.HandlerFunc {
	if secret == "" {
		secret = defaultJWTSecret
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenString, ok := strings.CutPrefix(authHeader, "Bearer ")
		if !ok || tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		parsedClaims := &claims{}
		token, err := jwt.ParseWithClaims(tokenString, parsedClaims, func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		if parsedClaims.UserID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token missing user_id"})
			return
		}

		c.Set("user_id", parsedClaims.UserID)
		c.Next()
	}
}

func jwtSecretFromEnv() string {
	if value := strings.TrimSpace(os.Getenv("JWT_SECRET")); value != "" {
		return value
	}
	return defaultJWTSecret
}
