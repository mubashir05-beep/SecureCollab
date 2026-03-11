package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const SecretKey = "securecollab-dev-secret-key"

type JWTClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a short-lived JWT token.
func GenerateAccessToken(userID string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return tokenString, nil
}

// ValidateToken verifies and parses a JWT token.
func ValidateToken(tokenString string) (*JWTClaims, error) {
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return claims, nil
}
