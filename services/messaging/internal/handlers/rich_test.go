package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"securecollab/services/messaging/internal/store"

	"github.com/golang-jwt/jwt/v5"
)

const richTestSecret = "securecollab-dev-secret-key"

func richTestToken(userID string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	s, _ := token.SignedString([]byte(richTestSecret))
	return s
}

type richTestEnv struct {
	store  store.RichMessageStore
	router http.Handler
}

func setupRichRouter() *richTestEnv {
	s := store.NewInMemoryRichStore()
	r := NewRichRouter(s)
	return &richTestEnv{store: s, router: r}
}

func (te *richTestEnv) req(method, path, token string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	r := httptest.NewRequest(method, path, &buf)
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	te.router.ServeHTTP(w, r)
	return w
}

func (te *richTestEnv) sendMsg(token, recipient string) MessageEnvelope {
	ct := base64.StdEncoding.EncodeToString([]byte("test-cipher"))
	nc := base64.StdEncoding.EncodeToString([]byte("test-nonce-12345"))
	w := te.req("POST", "/v1/messages", token, SendMessageRequest{
		RecipientUserID: recipient,
		CiphertextB64:   ct,
		NonceB64:        nc,
		ContentType:     "text",
	})
	var env MessageEnvelope
	json.Unmarshal(w.Body.Bytes(), &env)
	return env
}

func TestThreadReply(t *testing.T) {
	te := setupRichRouter()
	token := richTestToken("alice")
	bobToken := richTestToken("bob")

	// Alice sends a message
	parent := te.sendMsg(token, "bob")

	// Bob replies in thread
	ct := base64.StdEncoding.EncodeToString([]byte("reply-cipher"))
	nc := base64.StdEncoding.EncodeToString([]byte("reply-nonce-1234"))
	w := te.req("POST", fmt.Sprintf("/v1/messages/%s/replies", parent.ID), bobToken, ThreadReplyRequest{
		RecipientUserID: "alice",
		CiphertextB64:   ct,
		NonceB64:        nc,
	})
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// List thread replies
	w = te.req("GET", fmt.Sprintf("/v1/messages/%s/replies", parent.ID), token, nil)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Replies []MessageEnvelope `json:"replies"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Replies) != 1 {
		t.Fatalf("expected 1 reply, got %d", len(resp.Replies))
	}
}

func TestReactions(t *testing.T) {
	te := setupRichRouter()
	token := richTestToken("alice")
	bobToken := richTestToken("bob")

	msg := te.sendMsg(token, "bob")

	// Alice reacts
	w := te.req("POST", fmt.Sprintf("/v1/messages/%s/reactions", msg.ID), token, ReactionRequest{Emoji: "thumbsup"})
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Bob reacts same emoji
	w = te.req("POST", fmt.Sprintf("/v1/messages/%s/reactions", msg.ID), bobToken, ReactionRequest{Emoji: "thumbsup"})
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	// List reactions
	w = te.req("GET", fmt.Sprintf("/v1/messages/%s/reactions", msg.ID), token, nil)
	var resp struct {
		Reactions []store.ReactionSummary `json:"reactions"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Reactions) != 1 || resp.Reactions[0].Count != 2 {
		t.Fatalf("expected 1 reaction type with count 2, got %+v", resp.Reactions)
	}

	// Duplicate reaction fails
	w = te.req("POST", fmt.Sprintf("/v1/messages/%s/reactions", msg.ID), token, ReactionRequest{Emoji: "thumbsup"})
	if w.Code != 409 {
		t.Fatalf("expected 409 for duplicate, got %d", w.Code)
	}

	// Remove reaction
	w = te.req("DELETE", fmt.Sprintf("/v1/messages/%s/reactions/thumbsup", msg.ID), token, nil)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPinUnpin(t *testing.T) {
	te := setupRichRouter()
	token := richTestToken("alice")

	msg := te.sendMsg(token, "bob")

	// Pin
	w := te.req("POST", fmt.Sprintf("/v1/messages/%s/pin", msg.ID), token, PinRequest{ChannelID: "ch-1"})
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// List pins
	w = te.req("GET", "/v1/channels/ch-1/pins", token, nil)
	var resp struct {
		Pins []store.Pin `json:"pins"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Pins) != 1 {
		t.Fatalf("expected 1 pin, got %d", len(resp.Pins))
	}

	// Duplicate pin fails
	w = te.req("POST", fmt.Sprintf("/v1/messages/%s/pin", msg.ID), token, PinRequest{ChannelID: "ch-1"})
	if w.Code != 409 {
		t.Fatalf("expected 409, got %d", w.Code)
	}

	// Unpin
	w = te.req("DELETE", fmt.Sprintf("/v1/messages/%s/pin", msg.ID), token, nil)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestEditMessage(t *testing.T) {
	te := setupRichRouter()
	token := richTestToken("alice")
	bobToken := richTestToken("bob")

	msg := te.sendMsg(token, "bob")

	newCt := base64.StdEncoding.EncodeToString([]byte("edited-cipher"))
	newNc := base64.StdEncoding.EncodeToString([]byte("edited-nonce-123"))

	// Alice edits own message
	w := te.req("PUT", fmt.Sprintf("/v1/messages/%s", msg.ID), token, EditMessageRequest{
		CiphertextB64: newCt,
		NonceB64:      newNc,
	})
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Bob cannot edit Alice's message
	w = te.req("PUT", fmt.Sprintf("/v1/messages/%s", msg.ID), bobToken, EditMessageRequest{
		CiphertextB64: newCt,
		NonceB64:      newNc,
	})
	if w.Code != 403 {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestDeleteMessage(t *testing.T) {
	te := setupRichRouter()
	token := richTestToken("alice")
	bobToken := richTestToken("bob")

	msg := te.sendMsg(token, "bob")

	// Bob cannot delete Alice's message
	w := te.req("DELETE", fmt.Sprintf("/v1/messages/%s", msg.ID), bobToken, nil)
	if w.Code != 403 {
		t.Fatalf("expected 403, got %d", w.Code)
	}

	// Alice deletes own message
	w = te.req("DELETE", fmt.Sprintf("/v1/messages/%s", msg.ID), token, nil)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
