package handlers

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"securecollab/services/messaging/internal/store"

	"github.com/gin-gonic/gin"
)

type ThreadReplyRequest struct {
	RecipientUserID string `json:"recipient_user_id" binding:"required"`
	CiphertextB64   string `json:"ciphertext_b64" binding:"required"`
	NonceB64        string `json:"nonce_b64" binding:"required"`
	ContentType     string `json:"content_type"`
}

type ReactionRequest struct {
	Emoji string `json:"emoji" binding:"required"`
}

type PinRequest struct {
	ChannelID string `json:"channel_id" binding:"required"`
}

type EditMessageRequest struct {
	CiphertextB64 string `json:"ciphertext_b64" binding:"required"`
	NonceB64      string `json:"nonce_b64" binding:"required"`
}

// RegisterRichRoutes adds thread, reaction, pin, edit/delete routes to an existing router group.
func RegisterRichRoutes(v1 *gin.RouterGroup, richStore store.RichMessageStore, hub *deliveryHub) {
	// Threads
	v1.POST("/messages/:id/replies", postThreadReply(richStore, hub))
	v1.GET("/messages/:id/replies", getThreadReplies(richStore))

	// Reactions
	v1.POST("/messages/:id/reactions", addReaction(richStore))
	v1.DELETE("/messages/:id/reactions/:emoji", removeReaction(richStore))
	v1.GET("/messages/:id/reactions", listReactions(richStore))

	// Pins
	v1.POST("/messages/:id/pin", pinMessage(richStore))
	v1.DELETE("/messages/:id/pin", unpinMessage(richStore))
	v1.GET("/channels/:channel_id/pins", listPins(richStore))

	// Edit / Delete
	v1.PUT("/messages/:id", editMessage(richStore))
	v1.DELETE("/messages/:id", deleteMessage(richStore))
}

// --- Threads ---

func postThreadReply(s store.RichMessageStore, hub *deliveryHub) gin.HandlerFunc {
	return func(c *gin.Context) {
		senderUserID := c.GetString("user_id")
		parentID := c.Param("id")

		var req ThreadReplyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		ciphertext, err := base64.StdEncoding.DecodeString(req.CiphertextB64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ciphertext_b64"})
			return
		}
		nonce, err := base64.StdEncoding.DecodeString(req.NonceB64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid nonce_b64"})
			return
		}

		msg := store.EncryptedMessage{
			SenderUserID:    senderUserID,
			RecipientUserID: strings.TrimSpace(req.RecipientUserID),
			Ciphertext:      ciphertext,
			Nonce:           nonce,
			ContentType:     req.ContentType,
		}

		saved, err := s.SaveThreadReply(c.Request.Context(), msg, parentID)
		if err != nil {
			if errors.Is(err, store.ErrMessageNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "parent message not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save reply"})
			return
		}

		envelope := toEnvelope(saved)
		hub.Publish(saved.RecipientUserID, envelope)
		c.JSON(http.StatusCreated, envelope)
	}
}

func getThreadReplies(s store.RichMessageStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		parentID := c.Param("id")
		limit := 100
		if raw := c.Query("limit"); raw != "" {
			if p, err := strconv.Atoi(raw); err == nil && p > 0 && p <= 500 {
				limit = p
			}
		}

		replies, err := s.ListThreadReplies(c.Request.Context(), parentID, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list replies"})
			return
		}

		out := make([]MessageEnvelope, 0, len(replies))
		for _, r := range replies {
			out = append(out, toEnvelope(r))
		}
		c.JSON(http.StatusOK, gin.H{"replies": out})
	}
}

// --- Reactions ---

func addReaction(s store.RichMessageStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		messageID := c.Param("id")

		var req ReactionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		if err := s.AddReaction(c.Request.Context(), messageID, userID, req.Emoji); err != nil {
			if errors.Is(err, store.ErrReactionExists) {
				c.JSON(http.StatusConflict, gin.H{"error": "already reacted"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add reaction"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "added"})
	}
}

func removeReaction(s store.RichMessageStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		messageID := c.Param("id")
		emoji := c.Param("emoji")

		if err := s.RemoveReaction(c.Request.Context(), messageID, userID, emoji); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove reaction"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "removed"})
	}
}

func listReactions(s store.RichMessageStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		messageID := c.Param("id")
		reactions, err := s.ListReactions(c.Request.Context(), messageID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list reactions"})
			return
		}
		if reactions == nil {
			reactions = []store.ReactionSummary{}
		}
		c.JSON(http.StatusOK, gin.H{"reactions": reactions})
	}
}

// --- Pins ---

func pinMessage(s store.RichMessageStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		messageID := c.Param("id")

		var req PinRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		if err := s.PinMessage(c.Request.Context(), messageID, req.ChannelID, userID); err != nil {
			if errors.Is(err, store.ErrAlreadyPinned) {
				c.JSON(http.StatusConflict, gin.H{"error": "already pinned"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to pin"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "pinned"})
	}
}

func unpinMessage(s store.RichMessageStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		messageID := c.Param("id")
		if err := s.UnpinMessage(c.Request.Context(), messageID); err != nil {
			if errors.Is(err, store.ErrNotPinned) {
				c.JSON(http.StatusNotFound, gin.H{"error": "not pinned"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unpin"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "unpinned"})
	}
}

func listPins(s store.RichMessageStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		channelID := c.Param("channel_id")
		pins, err := s.ListPins(c.Request.Context(), channelID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list pins"})
			return
		}
		if pins == nil {
			pins = []store.Pin{}
		}
		c.JSON(http.StatusOK, gin.H{"pins": pins})
	}
}

// --- Edit / Delete ---

func editMessage(s store.RichMessageStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		messageID := c.Param("id")

		var req EditMessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		ciphertext, err := base64.StdEncoding.DecodeString(req.CiphertextB64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ciphertext_b64"})
			return
		}
		nonce, err := base64.StdEncoding.DecodeString(req.NonceB64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid nonce_b64"})
			return
		}

		if err := s.UpdateMessageCiphertext(c.Request.Context(), messageID, userID, ciphertext, nonce); err != nil {
			if errors.Is(err, store.ErrMessageNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
				return
			}
			if errors.Is(err, store.ErrNotSender) {
				c.JSON(http.StatusForbidden, gin.H{"error": "can only edit own messages"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to edit"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "updated"})
	}
}

func deleteMessage(s store.RichMessageStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		messageID := c.Param("id")

		if err := s.SoftDeleteMessage(c.Request.Context(), messageID, userID); err != nil {
			if errors.Is(err, store.ErrMessageNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
				return
			}
			if errors.Is(err, store.ErrNotSender) {
				c.JSON(http.StatusForbidden, gin.H{"error": "can only delete own messages"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	}
}
