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

	"securecollab/services/messaging/internal/store"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

func TestMessagingFlow_StoresCiphertextOnly_WithPostgres(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("MESSAGING_TEST_DATABASE_URL"))
	if databaseURL == "" {
		databaseURL = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	if databaseURL == "" {
		t.Skip("MESSAGING_TEST_DATABASE_URL or DATABASE_URL is not set")
	}

	s, err := store.NewPostgresMessageStore(databaseURL)
	if err != nil {
		t.Fatalf("init postgres message store: %v", err)
	}
	defer func() { _ = s.Close() }()

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	if err := ensureTablesForIntegration(ctx, db); err != nil {
		t.Fatalf("ensure tables: %v", err)
	}

	senderID := "22222222-2222-2222-2222-222222222222"
	recipientID := "33333333-3333-3333-3333-333333333333"
	if err := ensureIntegrationUser(ctx, db, senderID); err != nil {
		t.Fatalf("ensure sender: %v", err)
	}
	if err := ensureIntegrationUser(ctx, db, recipientID); err != nil {
		t.Fatalf("ensure recipient: %v", err)
	}

	router := NewRouter(s)
	senderToken := integrationToken(t, defaultDevSecret, senderID)
	recipientToken := integrationToken(t, defaultDevSecret, recipientID)

	plaintext := "hello-plaintext-never-store"
	ciphertextB64 := base64.StdEncoding.EncodeToString([]byte("ciphertext-blob-123"))
	nonceB64 := base64.StdEncoding.EncodeToString([]byte("nonce-xyz"))
	body, _ := json.Marshal(SendMessageRequest{
		RecipientUserID: recipientID,
		CiphertextB64:   ciphertextB64,
		NonceB64:        nonceB64,
		ContentType:     "ciphertext",
	})

	sendReq := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body))
	sendReq.Header.Set("Content-Type", "application/json")
	sendReq.Header.Set("Authorization", "Bearer "+senderToken)
	sendRes := httptest.NewRecorder()
	router.ServeHTTP(sendRes, sendReq)
	if sendRes.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, sendRes.Code)
	}

	inboxReq := httptest.NewRequest(http.MethodGet, "/v1/messages/inbox", nil)
	inboxReq.Header.Set("Authorization", "Bearer "+recipientToken)
	inboxRes := httptest.NewRecorder()
	router.ServeHTTP(inboxRes, inboxReq)
	if inboxRes.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, inboxRes.Code)
	}

	var inbox struct {
		Messages []MessageEnvelope `json:"messages"`
	}
	if err := json.Unmarshal(inboxRes.Body.Bytes(), &inbox); err != nil {
		t.Fatalf("unmarshal inbox: %v", err)
	}
	if len(inbox.Messages) == 0 {
		t.Fatal("expected at least one inbox message")
	}
	if inbox.Messages[0].CiphertextB64 != ciphertextB64 {
		t.Fatalf("expected ciphertext %s, got %s", ciphertextB64, inbox.Messages[0].CiphertextB64)
	}

	var storedCiphertext string
	err = db.QueryRowContext(ctx,
		"SELECT convert_from(ciphertext, 'UTF8') FROM encrypted_messages WHERE recipient_user_id = $1 ORDER BY created_at DESC LIMIT 1",
		recipientID,
	).Scan(&storedCiphertext)
	if err != nil {
		t.Fatalf("query stored ciphertext: %v", err)
	}
	if strings.Contains(storedCiphertext, plaintext) {
		t.Fatal("plaintext detected in encrypted_messages ciphertext column")
	}
	if storedCiphertext != "ciphertext-blob-123" {
		t.Fatalf("expected stored ciphertext blob, got %s", storedCiphertext)
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

	messagesDDL := `
		CREATE TABLE IF NOT EXISTS encrypted_messages (
			id UUID PRIMARY KEY,
			sender_user_id UUID NOT NULL,
			recipient_user_id UUID NOT NULL,
			channel_id UUID NULL,
			ciphertext BYTEA NOT NULL,
			nonce BYTEA NOT NULL,
			content_type VARCHAR(50) NOT NULL DEFAULT 'ciphertext',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.ExecContext(ctx, messagesDDL)
	return err
}

func ensureIntegrationUser(ctx context.Context, db *sql.DB, userID string) error {
	query := `
		INSERT INTO users (id, username, email, password_hash)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	_, err := db.ExecContext(ctx, query, userID, "messaging_user_"+suffix, "messaging_"+suffix+"@example.com", "placeholder")
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
