package handlers

import (
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"securecollab/services/messaging/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

const defaultJWTSecret = "securecollab-dev-secret-key"

type SendMessageRequest struct {
	RecipientUserID string `json:"recipient_user_id" binding:"required"`
	ChannelID       string `json:"channel_id"`
	CiphertextB64   string `json:"ciphertext_b64" binding:"required"`
	NonceB64        string `json:"nonce_b64" binding:"required"`
	ContentType     string `json:"content_type"`
}

type MessageEnvelope struct {
	ID              string    `json:"id"`
	SenderUserID    string    `json:"sender_user_id"`
	RecipientUserID string    `json:"recipient_user_id"`
	ChannelID       string    `json:"channel_id,omitempty"`
	CiphertextB64   string    `json:"ciphertext_b64"`
	NonceB64        string    `json:"nonce_b64"`
	ContentType     string    `json:"content_type"`
	CreatedAt       time.Time `json:"created_at"`
}

type inboxResponse struct {
	Messages []MessageEnvelope `json:"messages"`
}

type claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

var wsUpgrader = websocket.Upgrader{CheckOrigin: func(_ *http.Request) bool { return true }}

func NewRouter(messageStore store.MessageStore) *gin.Engine {
	if messageStore == nil {
		messageStore = store.NewInMemoryMessageStore()
	}
	hub := newDeliveryHub()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := r.Group("/v1")
	v1.Use(authMiddleware(jwtSecretFromEnv()))
	v1.POST("/messages", sendEncryptedMessage(messageStore, hub))
	v1.GET("/messages/inbox", getInbox(messageStore))
	v1.GET("/ws", websocketInbox(hub))

	return r
}

func sendEncryptedMessage(messageStore store.MessageStore, hub *deliveryHub) gin.HandlerFunc {
	return func(c *gin.Context) {
		senderUserID := c.GetString("user_id")
		if senderUserID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
			return
		}

		var req SendMessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		ciphertext, err := base64.StdEncoding.DecodeString(req.CiphertextB64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ciphertext_b64 must be valid base64"})
			return
		}
		nonce, err := base64.StdEncoding.DecodeString(req.NonceB64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "nonce_b64 must be valid base64"})
			return
		}

		stored, err := messageStore.SaveEncryptedMessage(c.Request.Context(), store.EncryptedMessage{
			SenderUserID:    senderUserID,
			RecipientUserID: strings.TrimSpace(req.RecipientUserID),
			ChannelID:       strings.TrimSpace(req.ChannelID),
			Ciphertext:      ciphertext,
			Nonce:           nonce,
			ContentType:     strings.TrimSpace(req.ContentType),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist encrypted message"})
			return
		}

		envelope := toEnvelope(stored)
		hub.Publish(stored.RecipientUserID, envelope)
		c.JSON(http.StatusCreated, envelope)
	}
}

func websocketInbox(hub *deliveryHub) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
			return
		}

		conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		ch := hub.Subscribe(userID)
		defer func() {
			hub.Unsubscribe(userID, ch)
			_ = conn.Close()
		}()

		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()

		for {
			select {
			case msg := <-ch:
				if err := conn.WriteJSON(msg); err != nil {
					return
				}
			case <-done:
				return
			}
		}
	}
}

func getInbox(messageStore store.MessageStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
			return
		}

		limit := 50
		if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 500 {
				limit = parsed
			}
		}

		messages, err := messageStore.ListInbox(c.Request.Context(), userID, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch inbox"})
			return
		}

		out := make([]MessageEnvelope, 0, len(messages))
		for _, msg := range messages {
			out = append(out, toEnvelope(msg))
		}
		c.JSON(http.StatusOK, inboxResponse{Messages: out})
	}
}

func toEnvelope(msg store.EncryptedMessage) MessageEnvelope {
	return MessageEnvelope{
		ID:              msg.ID,
		SenderUserID:    msg.SenderUserID,
		RecipientUserID: msg.RecipientUserID,
		ChannelID:       msg.ChannelID,
		CiphertextB64:   base64.StdEncoding.EncodeToString(msg.Ciphertext),
		NonceB64:        base64.StdEncoding.EncodeToString(msg.Nonce),
		ContentType:     msg.ContentType,
		CreatedAt:       msg.CreatedAt,
	}
}

func authMiddleware(secret string) gin.HandlerFunc {
	if secret == "" {
		secret = defaultJWTSecret
	}

	return func(c *gin.Context) {
		tokenString := strings.TrimSpace(c.Query("access_token"))
		if tokenString == "" {
			authHeader := c.GetHeader("Authorization")
			var ok bool
			tokenString, ok = strings.CutPrefix(authHeader, "Bearer ")
			if !ok || tokenString == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
				return
			}
		}

		parsedClaims := &claims{}
		token, err := jwt.ParseWithClaims(tokenString, parsedClaims, func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		if parsedClaims.UserID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token missing user_id"})
			return
		}

		c.Set("user_id", parsedClaims.UserID)
		c.Next()
	}
}

func jwtSecretFromEnv() string {
	if value := strings.TrimSpace(os.Getenv("JWT_SECRET")); value != "" {
		return value
	}
	return defaultJWTSecret
}
