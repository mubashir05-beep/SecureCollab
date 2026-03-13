//go:build integration

package handlers

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"securecollab/services/messaging/internal/store"

	"golang.org/x/crypto/curve25519"
)

func TestE2EEncryptSendReceiveDecrypt_WithTwoIdentities(t *testing.T) {
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

	senderID := "44444444-4444-4444-4444-444444444444"
	recipientID := "55555555-5555-5555-5555-555555555555"
	if err := ensureIntegrationUser(ctx, db, senderID); err != nil {
		t.Fatalf("ensure sender: %v", err)
	}
	if err := ensureIntegrationUser(ctx, db, recipientID); err != nil {
		t.Fatalf("ensure recipient: %v", err)
	}

	senderPriv, senderPub := generateX25519KeyPair(t)
	recipientPriv, recipientPub := generateX25519KeyPair(t)

	senderShared := deriveSharedSecret(t, senderPriv, recipientPub)
	recipientShared := deriveSharedSecret(t, recipientPriv, senderPub)
	if !bytes.Equal(senderShared, recipientShared) {
		t.Fatal("derived shared secrets do not match")
	}

	plaintext := []byte("hello securecollab e2e decrypt")
	nonce, ciphertext := encryptAESGCM(t, senderShared, plaintext)

	router := NewRouter(s)
	senderToken := integrationToken(t, defaultJWTSecret, senderID)
	recipientToken := integrationToken(t, defaultJWTSecret, recipientID)

	body, _ := json.Marshal(SendMessageRequest{
		RecipientUserID: recipientID,
		CiphertextB64:   base64.StdEncoding.EncodeToString(ciphertext),
		NonceB64:        base64.StdEncoding.EncodeToString(nonce),
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

	inboxReq := httptest.NewRequest(http.MethodGet, "/v1/messages/inbox?limit=1", nil)
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
	if len(inbox.Messages) != 1 {
		t.Fatalf("expected exactly one inbox message, got %d", len(inbox.Messages))
	}

	receivedCiphertext, err := base64.StdEncoding.DecodeString(inbox.Messages[0].CiphertextB64)
	if err != nil {
		t.Fatalf("decode ciphertext: %v", err)
	}
	receivedNonce, err := base64.StdEncoding.DecodeString(inbox.Messages[0].NonceB64)
	if err != nil {
		t.Fatalf("decode nonce: %v", err)
	}

	decrypted := decryptAESGCM(t, recipientShared, receivedNonce, receivedCiphertext)
	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("decrypted plaintext mismatch: expected %q, got %q", plaintext, decrypted)
	}

	var containsPlaintext int
	err = db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(POSITION($1::bytea IN ciphertext)), 0) FROM encrypted_messages WHERE recipient_user_id = $2",
		plaintext,
		recipientID,
	).Scan(&containsPlaintext)
	if err != nil {
		t.Fatalf("query plaintext presence: %v", err)
	}
	if containsPlaintext > 0 {
		t.Fatal("plaintext bytes found in ciphertext column")
	}
}

func generateX25519KeyPair(t *testing.T) ([]byte, []byte) {
	t.Helper()
	priv := make([]byte, 32)
	if _, err := rand.Read(priv); err != nil {
		t.Fatalf("rand private key: %v", err)
	}
	pub, err := curve25519.X25519(priv, curve25519.Basepoint)
	if err != nil {
		t.Fatalf("derive public key: %v", err)
	}
	return priv, pub
}

func deriveSharedSecret(t *testing.T, priv, peerPub []byte) []byte {
	t.Helper()
	shared, err := curve25519.X25519(priv, peerPub)
	if err != nil {
		t.Fatalf("derive shared secret: %v", err)
	}
	key := sha256.Sum256(shared)
	return key[:]
}

func encryptAESGCM(t *testing.T, key, plaintext []byte) ([]byte, []byte) {
	t.Helper()
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("new aes cipher: %v", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("new aes gcm: %v", err)
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		t.Fatalf("rand nonce: %v", err)
	}
	ciphertext := aead.Seal(nil, nonce, plaintext, nil)
	return nonce, ciphertext
}

func decryptAESGCM(t *testing.T, key, nonce, ciphertext []byte) []byte {
	t.Helper()
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("new aes cipher: %v", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("new aes gcm: %v", err)
	}
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		t.Fatalf("decrypt ciphertext: %v", err)
	}
	return plaintext
}
