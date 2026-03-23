package httpserver

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const defaultDevSecret = "securecollab-dev-secret-key"

type gatewayClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func authMiddleware(secret string) gin.HandlerFunc {
	if secret == "" {
		secret = defaultDevSecret
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		tokenString, ok := strings.CutPrefix(authHeader, "Bearer ")
		if !ok || tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		claims := &gatewayClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		if claims.UserID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token missing user_id"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

func jwtSecretFromEnv() string {
	if value := os.Getenv("JWT_SECRET"); value != "" {
		return value
	}
	log.Println("WARNING: JWT_SECRET not set, using insecure default. Set JWT_SECRET env var for production.")
	return defaultDevSecret
}
