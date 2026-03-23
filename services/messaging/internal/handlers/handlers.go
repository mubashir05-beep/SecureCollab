package handlers

import (
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"securecollab/services/messaging/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "messaging_http_requests_total", Help: "Total HTTP requests."},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "messaging_http_request_duration_seconds", Help: "Request duration.", Buckets: prometheus.DefBuckets},
		[]string{"method", "path", "status"},
	)
	wsConnectionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{Name: "messaging_websocket_connections_active", Help: "Active WebSocket connections."},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration, wsConnectionsActive)
}

func metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		status := http.StatusText(c.Writer.Status())
		httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath(), status).Observe(time.Since(start).Seconds())
	}
}

const defaultDevSecret = "securecollab-dev-secret-key"

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
	r.Use(corsMiddleware())
	r.Use(metricsMiddleware())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := r.Group("/v1")
	v1.Use(authMiddleware(jwtSecretFromEnv()))
	v1.POST("/messages", sendEncryptedMessage(messageStore, hub))
	v1.GET("/messages/inbox", getInbox(messageStore))
	v1.GET("/ws", websocketInbox(hub))

	// Register rich messaging routes if store supports it
	if richStore, ok := messageStore.(store.RichMessageStore); ok {
		RegisterRichRoutes(v1, richStore, hub)
	}

	return r
}

func NewRichRouter(richStore store.RichMessageStore) *gin.Engine {
	return NewRouter(richStore)
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
		wsConnectionsActive.Inc()
		defer func() {
			wsConnectionsActive.Dec()
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

func corsMiddleware() gin.HandlerFunc {
	allowedOrigin := corsOriginFromEnv()
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if allowedOrigin == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if origin != "" && originAllowed(origin, allowedOrigin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func corsOriginFromEnv() string {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "prod") {
		if origins := strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS")); origins != "" {
			return origins
		}
	}
	return "*"
}

func originAllowed(origin, allowed string) bool {
	for _, o := range strings.Split(allowed, ",") {
		if strings.TrimSpace(o) == origin {
			return true
		}
	}
	return false
}

func authMiddleware(secret string) gin.HandlerFunc {
	if secret == "" {
		secret = defaultDevSecret
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
	log.Println("WARNING: JWT_SECRET not set, using insecure default. Set JWT_SECRET env var for production.")
	return defaultDevSecret
}
