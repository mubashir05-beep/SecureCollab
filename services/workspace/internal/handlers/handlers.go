package handlers

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"securecollab/services/workspace/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "workspace_http_requests_total", Help: "Total HTTP requests."},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "workspace_http_request_duration_seconds", Help: "Request duration.", Buckets: prometheus.DefBuckets},
		[]string{"method", "path", "status"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
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

type claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type CreateWorkspaceRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type CreateChannelRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Topic       string `json:"topic"`
	IsPrivate   bool   `json:"is_private"`
}

type JoinByInviteRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role"`
}

type UpdateTopicRequest struct {
	Topic string `json:"topic"`
}

func NewRouter(ws store.WorkspaceStore) *gin.Engine {
	if ws == nil {
		ws = store.NewInMemoryStore()
	}

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

	// Workspaces
	v1.POST("/workspaces", createWorkspace(ws))
	v1.GET("/workspaces", listWorkspaces(ws))
	v1.GET("/workspaces/:id", getWorkspace(ws))
	v1.POST("/workspaces/join", joinByInvite(ws))

	// Members
	v1.GET("/workspaces/:id/members", listMembers(ws))
	v1.POST("/workspaces/:id/members", addMember(ws))
	v1.DELETE("/workspaces/:id/members/:user_id", removeMember(ws))

	// Channels
	v1.POST("/workspaces/:id/channels", createChannel(ws))
	v1.GET("/workspaces/:id/channels", listChannels(ws))
	v1.GET("/channels/:channel_id", getChannel(ws))
	v1.PUT("/channels/:channel_id/topic", updateChannelTopic(ws))
	v1.POST("/channels/:channel_id/archive", archiveChannel(ws))

	return r
}

// --- Workspace Handlers ---

func createWorkspace(s store.WorkspaceStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		var req CreateWorkspaceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		ws, err := s.CreateWorkspace(c.Request.Context(), req.Name, req.Description, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create workspace"})
			return
		}
		c.JSON(http.StatusCreated, ws)
	}
}

func listWorkspaces(s store.WorkspaceStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		list, err := s.ListWorkspacesForUser(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list workspaces"})
			return
		}
		if list == nil {
			list = []store.Workspace{}
		}
		c.JSON(http.StatusOK, gin.H{"workspaces": list})
	}
}

func getWorkspace(s store.WorkspaceStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		ws, err := s.GetWorkspace(c.Request.Context(), c.Param("id"))
		if err != nil {
			if errors.Is(err, store.ErrWorkspaceNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
			return
		}
		c.JSON(http.StatusOK, ws)
	}
}

func joinByInvite(s store.WorkspaceStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		var req JoinByInviteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		ws, err := s.JoinWorkspaceByInvite(c.Request.Context(), req.InviteCode, userID)
		if err != nil {
			if errors.Is(err, store.ErrInvalidInvite) {
				c.JSON(http.StatusNotFound, gin.H{"error": "invalid invite code"})
				return
			}
			if errors.Is(err, store.ErrAlreadyMember) {
				c.JSON(http.StatusConflict, gin.H{"error": "already a member"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to join workspace"})
			return
		}
		c.JSON(http.StatusOK, ws)
	}
}

// --- Member Handlers ---

func listMembers(s store.WorkspaceStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		members, err := s.ListMembers(c.Request.Context(), c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list members"})
			return
		}
		if members == nil {
			members = []store.WorkspaceMember{}
		}
		c.JSON(http.StatusOK, gin.H{"members": members})
	}
}

func addMember(s store.WorkspaceStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		callerID := c.GetString("user_id")
		wsID := c.Param("id")

		role, err := s.GetMemberRole(c.Request.Context(), wsID, callerID)
		if err != nil || (role != "owner" && role != "admin") {
			c.JSON(http.StatusForbidden, gin.H{"error": "only owner or admin can add members"})
			return
		}

		var req AddMemberRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// Resolve username or UUID to a valid user UUID.
		resolvedID, err := s.ResolveUserID(c.Request.Context(), req.UserID)
		if err != nil {
			if errors.Is(err, store.ErrUserNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve user"})
			return
		}

		memberRole := req.Role
		if memberRole == "" {
			memberRole = "member"
		}
		if err := s.AddMember(c.Request.Context(), wsID, resolvedID, memberRole); err != nil {
			if errors.Is(err, store.ErrAlreadyMember) {
				c.JSON(http.StatusConflict, gin.H{"error": "already a member"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add member"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "added"})
	}
}

func removeMember(s store.WorkspaceStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		callerID := c.GetString("user_id")
		wsID := c.Param("id")
		targetID := c.Param("user_id")

		role, err := s.GetMemberRole(c.Request.Context(), wsID, callerID)
		if err != nil || (role != "owner" && role != "admin") {
			c.JSON(http.StatusForbidden, gin.H{"error": "only owner or admin can remove members"})
			return
		}

		if err := s.RemoveMember(c.Request.Context(), wsID, targetID); err != nil {
			if errors.Is(err, store.ErrNotMember) {
				c.JSON(http.StatusNotFound, gin.H{"error": "not a member"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove member"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "removed"})
	}
}

// --- Channel Handlers ---

func createChannel(s store.WorkspaceStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		wsID := c.Param("id")

		_, err := s.GetMemberRole(c.Request.Context(), wsID, userID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "must be a workspace member"})
			return
		}

		var req CreateChannelRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		ch, err := s.CreateChannel(c.Request.Context(), wsID, req.Name, req.Description, req.Topic, userID, req.IsPrivate)
		if err != nil {
			if errors.Is(err, store.ErrChannelExists) {
				c.JSON(http.StatusConflict, gin.H{"error": "channel already exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create channel"})
			return
		}
		c.JSON(http.StatusCreated, ch)
	}
}

func listChannels(s store.WorkspaceStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		channels, err := s.ListChannels(c.Request.Context(), c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list channels"})
			return
		}
		if channels == nil {
			channels = []store.Channel{}
		}
		c.JSON(http.StatusOK, gin.H{"channels": channels})
	}
}

func getChannel(s store.WorkspaceStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		ch, err := s.GetChannel(c.Request.Context(), c.Param("channel_id"))
		if err != nil {
			if errors.Is(err, store.ErrChannelNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get channel"})
			return
		}
		c.JSON(http.StatusOK, ch)
	}
}

func updateChannelTopic(s store.WorkspaceStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateTopicRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		if err := s.UpdateChannelTopic(c.Request.Context(), c.Param("channel_id"), req.Topic); err != nil {
			if errors.Is(err, store.ErrChannelNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update topic"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "updated"})
	}
}

func archiveChannel(s store.WorkspaceStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		callerID := c.GetString("user_id")
		ch, err := s.GetChannel(c.Request.Context(), c.Param("channel_id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
			return
		}
		role, err := s.GetMemberRole(c.Request.Context(), ch.WorkspaceID, callerID)
		if err != nil || (role != "owner" && role != "admin") {
			c.JSON(http.StatusForbidden, gin.H{"error": "only owner or admin can archive channels"})
			return
		}
		if err := s.ArchiveChannel(c.Request.Context(), c.Param("channel_id")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to archive channel"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "archived"})
	}
}

// --- Auth Middleware ---

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
		authHeader := c.GetHeader("Authorization")
		tokenString, ok := strings.CutPrefix(authHeader, "Bearer ")
		if !ok || tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}
		parsed := &claims{}
		token, err := jwt.ParseWithClaims(tokenString, parsed, func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid || parsed.UserID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("user_id", parsed.UserID)
		c.Next()
	}
}

func jwtSecretFromEnv() string {
	if v := strings.TrimSpace(os.Getenv("JWT_SECRET")); v != "" {
		return v
	}
	log.Println("WARNING: JWT_SECRET not set, using insecure default. Set JWT_SECRET env var for production.")
	return defaultDevSecret
}
