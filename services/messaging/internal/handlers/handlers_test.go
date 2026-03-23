package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"securecollab/services/messaging/internal/store"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

func TestSendEncryptedMessageAndInbox(t *testing.T) {
	router := NewRouter(store.NewInMemoryMessageStore())
	senderToken := testToken(t, defaultDevSecret, "sender-1")
	recipientToken := testToken(t, defaultDevSecret, "recipient-1")

	payload := SendMessageRequest{
		RecipientUserID: "recipient-1",
		CiphertextB64:   base64.StdEncoding.EncodeToString([]byte("ciphertext")),
		NonceB64:        base64.StdEncoding.EncodeToString([]byte("nonce")),
	}
	body, _ := json.Marshal(payload)
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

	var resp struct {
		Messages []MessageEnvelope `json:"messages"`
	}
	if err := json.Unmarshal(inboxRes.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal inbox: %v", err)
	}
	if len(resp.Messages) != 1 {
		t.Fatalf("expected 1 message in inbox, got %d", len(resp.Messages))
	}
	if resp.Messages[0].SenderUserID != "sender-1" {
		t.Fatalf("expected sender sender-1, got %s", resp.Messages[0].SenderUserID)
	}
	if resp.Messages[0].CiphertextB64 != payload.CiphertextB64 {
		t.Fatalf("expected ciphertext %s, got %s", payload.CiphertextB64, resp.Messages[0].CiphertextB64)
	}
}

func TestWebSocketInbox_RequiresAuth(t *testing.T) {
	router := NewRouter(store.NewInMemoryMessageStore())
	req := httptest.NewRequest(http.MethodGet, "/v1/ws", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, res.Code)
	}
}

func TestWebSocketInbox_DeliversPublishedMessage(t *testing.T) {
	router := NewRouter(store.NewInMemoryMessageStore())
	server := httptest.NewServer(router)
	defer server.Close()

	recipientToken := testToken(t, defaultDevSecret, "recipient-1")
	senderToken := testToken(t, defaultDevSecret, "sender-1")

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/v1/ws?access_token=" + recipientToken
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	payload := SendMessageRequest{
		RecipientUserID: "recipient-1",
		CiphertextB64:   base64.StdEncoding.EncodeToString([]byte("ciphertext")),
		NonceB64:        base64.StdEncoding.EncodeToString([]byte("nonce")),
	}
	body, _ := json.Marshal(payload)
	sendReq, _ := http.NewRequest(http.MethodPost, server.URL+"/v1/messages", bytes.NewReader(body))
	sendReq.Header.Set("Content-Type", "application/json")
	sendReq.Header.Set("Authorization", "Bearer "+senderToken)

	sendRes, err := http.DefaultClient.Do(sendReq)
	if err != nil {
		t.Fatalf("send message request failed: %v", err)
	}
	defer func() { _ = sendRes.Body.Close() }()
	if sendRes.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, sendRes.StatusCode)
	}

	var delivered MessageEnvelope
	if err := conn.ReadJSON(&delivered); err != nil {
		t.Fatalf("read websocket message: %v", err)
	}

	if delivered.RecipientUserID != "recipient-1" {
		t.Fatalf("expected recipient recipient-1, got %s", delivered.RecipientUserID)
	}
	if delivered.SenderUserID != "sender-1" {
		t.Fatalf("expected sender sender-1, got %s", delivered.SenderUserID)
	}
	if delivered.CiphertextB64 != payload.CiphertextB64 {
		t.Fatalf("expected ciphertext %s, got %s", payload.CiphertextB64, delivered.CiphertextB64)
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
