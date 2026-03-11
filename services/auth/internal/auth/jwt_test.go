package auth

import (
	"testing"
)

func TestGenerateAccessToken_ReturnsNonEmptyString(t *testing.T) {
	token, err := GenerateAccessToken("user-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestValidateToken_ValidToken(t *testing.T) {
	userID := "user-456"
	token, err := GenerateAccessToken(userID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if claims.UserID != userID {
		t.Fatalf("expected user_id %q, got %q", userID, claims.UserID)
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	_, err := ValidateToken("invalid.token.here")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
