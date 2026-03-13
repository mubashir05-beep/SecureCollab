//go:build integration

package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"securecollab/services/auth/internal/store"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func TestAuthHTTP_RegisterLoginRefresh_WithPostgresStore(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("AUTH_TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("AUTH_TEST_DATABASE_URL is not set")
	}

	router, cleanup := newIntegrationRouter(t, databaseURL)
	defer cleanup()

	suffix := time.Now().UnixNano()
	username := fmt.Sprintf("http_user_%d", suffix)
	email := fmt.Sprintf("http_%d@example.com", suffix)
	password := "password123"

	registerReq := RegisterRequest{Username: username, Email: email, Password: password}
	registerRes := performJSONRequest(t, router, http.MethodPost, "/register", registerReq)
	if registerRes.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, registerRes.Code)
	}

	loginReq := LoginRequest{Username: username, Password: password}
	loginRes := performJSONRequest(t, router, http.MethodPost, "/login", loginReq)
	if loginRes.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, loginRes.Code)
	}
	loginToken := tokenFromResponse(t, loginRes)
	if loginToken == "" {
		t.Fatal("expected login access token")
	}

	refreshReq := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	refreshReq.Header.Set("Authorization", "Bearer "+loginToken)
	refreshRes := httptest.NewRecorder()
	router.ServeHTTP(refreshRes, refreshReq)
	if refreshRes.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, refreshRes.Code)
	}
	if tokenFromResponse(t, refreshRes) == "" {
		t.Fatal("expected refresh access token")
	}
}

func TestAuthHTTP_RegisterDuplicate_WithPostgresStore(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("AUTH_TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("AUTH_TEST_DATABASE_URL is not set")
	}

	router, cleanup := newIntegrationRouter(t, databaseURL)
	defer cleanup()

	suffix := time.Now().UnixNano()
	req := RegisterRequest{
		Username: fmt.Sprintf("dup_user_%d", suffix),
		Email:    fmt.Sprintf("dup_%d@example.com", suffix),
		Password: "password123",
	}

	firstRes := performJSONRequest(t, router, http.MethodPost, "/register", req)
	if firstRes.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, firstRes.Code)
	}

	secondRes := performJSONRequest(t, router, http.MethodPost, "/register", req)
	if secondRes.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, secondRes.Code)
	}
}

func TestAuthHTTP_LoginInvalidPassword_WithPostgresStore(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("AUTH_TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("AUTH_TEST_DATABASE_URL is not set")
	}

	router, cleanup := newIntegrationRouter(t, databaseURL)
	defer cleanup()

	suffix := time.Now().UnixNano()
	username := fmt.Sprintf("badpw_user_%d", suffix)
	registerReq := RegisterRequest{
		Username: username,
		Email:    fmt.Sprintf("badpw_%d@example.com", suffix),
		Password: "password123",
	}

	registerRes := performJSONRequest(t, router, http.MethodPost, "/register", registerReq)
	if registerRes.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, registerRes.Code)
	}

	loginRes := performJSONRequest(t, router, http.MethodPost, "/login", LoginRequest{Username: username, Password: "wrong-password"})
	if loginRes.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, loginRes.Code)
	}
}

func newIntegrationRouter(t *testing.T, databaseURL string) (*gin.Engine, func()) {
	t.Helper()

	postgresStore, err := store.NewPostgresUserStore(databaseURL)
	if err != nil {
		t.Fatalf("init postgres store: %v", err)
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatalf("open cleanup db: %v", err)
	}

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping cleanup db: %v", err)
	}

	if err := ensureUsersTableForHandlerIntegration(ctx, db); err != nil {
		t.Fatalf("ensure users table: %v", err)
	}

	router := NewRouter(postgresStore)

	cleanup := func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE username LIKE 'http_%' OR username LIKE 'dup_%' OR username LIKE 'badpw_%'")
		_ = db.Close()
		_ = postgresStore.Close()
	}

	return router, cleanup
}

func ensureUsersTableForHandlerIntegration(ctx context.Context, db *sql.DB) error {
	const ddl = `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.ExecContext(ctx, ddl)
	return err
}

func performJSONRequest(t *testing.T, router *gin.Engine, method, path string, payload interface{}) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}

func tokenFromResponse(t *testing.T, res *httptest.ResponseRecorder) string {
	t.Helper()
	var resp TokenResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal token response: %v", err)
	}
	return resp.AccessToken
}
