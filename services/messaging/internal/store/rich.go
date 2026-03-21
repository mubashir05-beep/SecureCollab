package store

import (
	"context"
	"time"
)

// Reaction represents an emoji reaction on a message.
type Reaction struct {
	ID        string    `json:"id"`
	MessageID string    `json:"message_id"`
	UserID    string    `json:"user_id"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

// ReactionSummary is a grouped view: emoji + count + list of user IDs.
type ReactionSummary struct {
	Emoji   string   `json:"emoji"`
	Count   int      `json:"count"`
	UserIDs []string `json:"user_ids"`
}

// Pin represents a pinned message in a channel.
type Pin struct {
	ID        string    `json:"id"`
	MessageID string    `json:"message_id"`
	ChannelID string    `json:"channel_id"`
	PinnedBy  string    `json:"pinned_by"`
	PinnedAt  time.Time `json:"pinned_at"`
}

// RichMessageStore extends MessageStore with threads, reactions, pins, edit/delete.
type RichMessageStore interface {
	MessageStore

	// Threads
	SaveThreadReply(ctx context.Context, msg EncryptedMessage, parentID string) (EncryptedMessage, error)
	ListThreadReplies(ctx context.Context, parentID string, limit int) ([]EncryptedMessage, error)

	// Reactions
	AddReaction(ctx context.Context, messageID, userID, emoji string) error
	RemoveReaction(ctx context.Context, messageID, userID, emoji string) error
	ListReactions(ctx context.Context, messageID string) ([]ReactionSummary, error)

	// Pins
	PinMessage(ctx context.Context, messageID, channelID, pinnedBy string) error
	UnpinMessage(ctx context.Context, messageID string) error
	ListPins(ctx context.Context, channelID string) ([]Pin, error)

	// Edit / Delete
	UpdateMessageCiphertext(ctx context.Context, messageID, senderUserID string, ciphertext, nonce []byte) error
	SoftDeleteMessage(ctx context.Context, messageID, senderUserID string) error
}
