//go:build integration

package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"securecollab/services/keydist/internal/store"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

func TestKeyBootstrapFlow_ClientToService_WithPostgres(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("KEYDIST_TEST_DATABASE_URL"))
	if databaseURL == "" {
		databaseURL = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	if databaseURL == "" {
		t.Skip("KEYDIST_TEST_DATABASE_URL or DATABASE_URL is not set")
	}

	s, err := store.NewPostgresKeyStore(databaseURL)
	if err != nil {
		t.Fatalf("init postgres key store: %v", err)
	}
	defer func() {
		_ = s.Close()
	}()

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	if err := ensureTablesForIntegration(ctx, db); err != nil {
		t.Fatalf("ensure tables: %v", err)
	}

	userID := "11111111-1111-1111-1111-111111111111"
	if err := ensureIntegrationUser(ctx, db, userID); err != nil {
		t.Fatalf("ensure user: %v", err)
	}

	router := NewRouter(s)
	token := integrationToken(t, defaultJWTSecret, userID)

	publicKey := base64.StdEncoding.EncodeToString([]byte("bootstrap-public-key"))
	uploadBody, _ := json.Marshal(UploadKeyRequest{PublicKeyB64: publicKey})
	uploadReq := httptest.NewRequest(http.MethodPost, "/v1/keys/identity", bytes.NewReader(uploadBody))
	uploadReq.Header.Set("Content-Type", "application/json")
	uploadReq.Header.Set("Authorization", "Bearer "+token)
	uploadRes := httptest.NewRecorder()
	router.ServeHTTP(uploadRes, uploadReq)

	if uploadRes.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, uploadRes.Code)
	}

	fetchReq := httptest.NewRequest(http.MethodGet, "/v1/keys/identity/"+userID, nil)
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
	if resp.PublicKeyB64 != publicKey {
		t.Fatalf("expected public key %s, got %s", publicKey, resp.PublicKeyB64)
	}
}

func ensureTablesForIntegration(ctx context.Context, db *sql.DB) error {
	usersDDL := `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`
	if _, err := db.ExecContext(ctx, usersDDL); err != nil {
		return err
	}

	keysDDL := `
		CREATE TABLE IF NOT EXISTS public_keys (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			key_data BYTEA NOT NULL,
			key_type VARCHAR(50) NOT NULL DEFAULT 'identity',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.ExecContext(ctx, keysDDL)
	return err
}

func ensureIntegrationUser(ctx context.Context, db *sql.DB, userID string) error {
	query := `
		INSERT INTO users (id, username, email, password_hash)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	_, err := db.ExecContext(ctx, query, userID, "keydist_user_"+suffix, "keydist_"+suffix+"@example.com", "placeholder")
	return err
}

func integrationToken(t *testing.T, secret, userID string) string {
	t.Helper()
	claims := jwt.MapClaims{"user_id": userID, "iat": time.Now().Unix(), "exp": time.Now().Add(10 * time.Minute).Unix()}
	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tkn.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return s
}
