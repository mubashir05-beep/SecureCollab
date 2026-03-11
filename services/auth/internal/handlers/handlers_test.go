package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegister_WithValidRequest(t *testing.T) {
	router := NewRouter()

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

func TestLogin_WithValidRequest(t *testing.T) {
	router := NewRouter()

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

func TestHealth_ReturnsOK(t *testing.T) {
	router := NewRouter()

	httpReq := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, httpReq)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
}
