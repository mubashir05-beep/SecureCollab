package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"securecollab/services/keydist/internal/store"

	"github.com/golang-jwt/jwt/v5"
)

func TestUploadAndFetchIdentityKey_WithValidToken(t *testing.T) {
	router := NewRouter(store.NewInMemoryKeyStore())
	token := testToken(t, defaultDevSecret, "user-123")

	payload := UploadKeyRequest{PublicKeyB64: base64.StdEncoding.EncodeToString([]byte("public-key-1"))}
	body, _ := json.Marshal(payload)
	uploadReq := httptest.NewRequest(http.MethodPost, "/v1/keys/identity", bytes.NewReader(body))
	uploadReq.Header.Set("Content-Type", "application/json")
	uploadReq.Header.Set("Authorization", "Bearer "+token)
	uploadRes := httptest.NewRecorder()
	router.ServeHTTP(uploadRes, uploadReq)

	if uploadRes.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, uploadRes.Code)
	}

	fetchReq := httptest.NewRequest(http.MethodGet, "/v1/keys/identity/user-123", nil)
	fetchReq.Header.Set("Authorization", "Bearer "+token)
	fetchRes := httptest.NewRecorder()
	router.ServeHTTP(fetchRes, fetchReq)

	if fetchRes.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, fetchRes.Code)
	}

	var resp KeyResponse
	if err := json.Unmarshal(fetchRes.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.UserID != "user-123" {
		t.Fatalf("expected user_id user-123, got %s", resp.UserID)
	}
	if resp.PublicKeyB64 != payload.PublicKeyB64 {
		t.Fatalf("expected public key %s, got %s", payload.PublicKeyB64, resp.PublicKeyB64)
	}
}

func TestUploadIdentityKey_WithoutToken_ReturnsUnauthorized(t *testing.T) {
	router := NewRouter(store.NewInMemoryKeyStore())
	payload := UploadKeyRequest{PublicKeyB64: base64.StdEncoding.EncodeToString([]byte("public-key-1"))}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/keys/identity", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, res.Code)
	}
}

func TestFetchIdentityKey_NotFound_Returns404(t *testing.T) {
	router := NewRouter(store.NewInMemoryKeyStore())
	token := testToken(t, defaultDevSecret, "user-123")

	req := httptest.NewRequest(http.MethodGet, "/v1/keys/identity/missing-user", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, res.Code)
	}
}

func testToken(t *testing.T, secret, userID string) string {
	t.Helper()
	claims := jwt.MapClaims{"user_id": userID, "iat": time.Now().Unix(), "exp": time.Now().Add(10 * time.Minute).Unix()}
	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tkn.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return s
}
