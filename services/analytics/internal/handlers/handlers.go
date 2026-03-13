package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"securecollab/services/analytics/internal/store"
)

type Handler struct {
	store store.AnalyticsStore
}

func NewHandler(s store.AnalyticsStore) *Handler {
	return &Handler{store: s}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/healthz", h.health)
	r.GET("/v1/analytics/messages/volume", h.messageVolume)
}

func (h *Handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) messageVolume(c *gin.Context) {
	windowHours := 24
	if raw := c.Query("window_hours"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > 24*30 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "window_hours must be an integer between 1 and 720"})
			return
		}
		windowHours = parsed
	}

	volume, err := h.store.GetMessageVolume(c.Request.Context(), windowHours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load analytics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_messages":     volume.TotalMessages,
		"messages_in_window": volume.MessagesLast24h,
		"window_hours":       volume.WindowHours,
	})
}
