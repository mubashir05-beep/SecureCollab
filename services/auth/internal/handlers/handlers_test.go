package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"securecollab/services/auth/internal/store"

	"github.com/gin-gonic/gin"
)

func newTestRouter() *gin.Engine {
	return NewRouter(store.NewInMemoryUserStore())
}

func TestRegister_WithValidRequest(t *testing.T) {
	router := newTestRouter()

	req := RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(req)

	httpReq := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, httpReq)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, res.Code)
	}

	var resp TokenResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.AccessToken == "" {
		t.Fatal("expected non-empty access_token")
	}
}

func TestRegister_WithDuplicateUsername_ReturnsConflict(t *testing.T) {
	router := newTestRouter()

	req := RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(req)

	firstReq := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	firstReq.Header.Set("Content-Type", "application/json")
	firstRes := httptest.NewRecorder()
	router.ServeHTTP(firstRes, firstReq)

	secondReq := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	secondReq.Header.Set("Content-Type", "application/json")
	secondRes := httptest.NewRecorder()
	router.ServeHTTP(secondRes, secondReq)

	if secondRes.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, secondRes.Code)
	}
}

func TestLogin_WithValidRequest(t *testing.T) {
	router := newTestRouter()

	registerReq := RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	registerBody, _ := json.Marshal(registerReq)
	registerHTTPReq := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(registerBody))
	registerHTTPReq.Header.Set("Content-Type", "application/json")
	registerRes := httptest.NewRecorder()
	router.ServeHTTP(registerRes, registerHTTPReq)
	if registerRes.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, registerRes.Code)
	}

	req := LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	body, _ := json.Marshal(req)

	httpReq := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, httpReq)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}

	var resp TokenResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.AccessToken == "" {
		t.Fatal("expected non-empty access_token")
	}
}

func TestLogin_WithInvalidPassword_ReturnsUnauthorized(t *testing.T) {
	router := newTestRouter()

	registerReq := RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	registerBody, _ := json.Marshal(registerReq)
	registerHTTPReq := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(registerBody))
	registerHTTPReq.Header.Set("Content-Type", "application/json")
	registerRes := httptest.NewRecorder()
	router.ServeHTTP(registerRes, registerHTTPReq)

	loginReq := LoginRequest{Username: "testuser", Password: "wrong-password"}
	loginBody, _ := json.Marshal(loginReq)
	loginHTTPReq := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(loginBody))
	loginHTTPReq.Header.Set("Content-Type", "application/json")
	loginRes := httptest.NewRecorder()
	router.ServeHTTP(loginRes, loginHTTPReq)

	if loginRes.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, loginRes.Code)
	}
}

func TestRefresh_WithValidToken_ReturnsFreshToken(t *testing.T) {
	router := newTestRouter()

	registerReq := RegisterRequest{
		Username: "refresh-user",
		Email:    "refresh@example.com",
		Password: "password123",
	}
	registerBody, _ := json.Marshal(registerReq)
	registerHTTPReq := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(registerBody))
	registerHTTPReq.Header.Set("Content-Type", "application/json")
	registerRes := httptest.NewRecorder()
	router.ServeHTTP(registerRes, registerHTTPReq)

	if registerRes.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, registerRes.Code)
	}

	var registerResp TokenResponse
	if err := json.Unmarshal(registerRes.Body.Bytes(), &registerResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	refreshReq := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	refreshReq.Header.Set("Authorization", "Bearer "+registerResp.AccessToken)
	refreshRes := httptest.NewRecorder()
	router.ServeHTTP(refreshRes, refreshReq)

	if refreshRes.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, refreshRes.Code)
	}

	var refreshResp TokenResponse
	if err := json.Unmarshal(refreshRes.Body.Bytes(), &refreshResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if refreshResp.AccessToken == "" {
		t.Fatal("expected non-empty access_token")
	}
}

func TestHealth_ReturnsOK(t *testing.T) {
	router := newTestRouter()

	httpReq := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, httpReq)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
}
