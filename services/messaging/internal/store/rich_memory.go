package store

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrReactionExists = errors.New("reaction already exists")
	ErrAlreadyPinned  = errors.New("message already pinned")
	ErrNotPinned      = errors.New("message not pinned")
	ErrNotSender      = errors.New("not the message sender")
)

type inMemoryRichStore struct {
	mu        sync.Mutex
	messages  map[string]EncryptedMessage   // id -> msg
	byInbox   map[string][]string           // recipientUserID -> []id
	threads   map[string][]string           // parentID -> []childID
	reactions map[string][]Reaction         // messageID -> reactions
	pins      map[string]Pin                // messageID -> pin
	pinsByCh  map[string][]string           // channelID -> []messageID
}

func NewInMemoryRichStore() RichMessageStore {
	return &inMemoryRichStore{
		messages:  make(map[string]EncryptedMessage),
		byInbox:   make(map[string][]string),
		threads:   make(map[string][]string),
		reactions: make(map[string][]Reaction),
		pins:      make(map[string]Pin),
		pinsByCh:  make(map[string][]string),
	}
}

func (s *inMemoryRichStore) SaveEncryptedMessage(ctx context.Context, msg EncryptedMessage) (EncryptedMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveMsg(msg)
}

func (s *inMemoryRichStore) saveMsg(msg EncryptedMessage) (EncryptedMessage, error) {
	if msg.ID == "" {
		msg.ID = uuid.NewString()
	}
	if msg.ContentType == "" {
		msg.ContentType = "ciphertext"
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now().UTC()
	}
	msg.Ciphertext = append([]byte(nil), msg.Ciphertext...)
	msg.Nonce = append([]byte(nil), msg.Nonce...)
	s.messages[msg.ID] = msg
	s.byInbox[msg.RecipientUserID] = append(s.byInbox[msg.RecipientUserID], msg.ID)
	return msg, nil
}

func (s *inMemoryRichStore) ListInbox(_ context.Context, recipientUserID string, limit int) ([]EncryptedMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ids := s.byInbox[recipientUserID]
	var out []EncryptedMessage
	for _, id := range ids {
		if msg, ok := s.messages[id]; ok && msg.DeletedAt == nil {
			out = append(out, msg)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	if limit > 0 && limit < len(out) {
		out = out[:limit]
	}
	return out, nil
}

// --- Threads ---

func (s *inMemoryRichStore) SaveThreadReply(_ context.Context, msg EncryptedMessage, parentID string) (EncryptedMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.messages[parentID]; !ok {
		return EncryptedMessage{}, ErrMessageNotFound
	}
	msg.ParentMessageID = parentID
	saved, err := s.saveMsg(msg)
	if err != nil {
		return EncryptedMessage{}, err
	}
	s.threads[parentID] = append(s.threads[parentID], saved.ID)
	return saved, nil
}

func (s *inMemoryRichStore) ListThreadReplies(_ context.Context, parentID string, limit int) ([]EncryptedMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := s.threads[parentID]
	var out []EncryptedMessage
	for _, id := range ids {
		if msg, ok := s.messages[id]; ok && msg.DeletedAt == nil {
			out = append(out, msg)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	if limit > 0 && limit < len(out) {
		out = out[:limit]
	}
	return out, nil
}

// --- Reactions ---

func (s *inMemoryRichStore) AddReaction(_ context.Context, messageID, userID, emoji string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.reactions[messageID] {
		if r.UserID == userID && r.Emoji == emoji {
			return ErrReactionExists
		}
	}
	s.reactions[messageID] = append(s.reactions[messageID], Reaction{
		ID:        uuid.NewString(),
		MessageID: messageID,
		UserID:    userID,
		Emoji:     emoji,
		CreatedAt: time.Now().UTC(),
	})
	return nil
}

func (s *inMemoryRichStore) RemoveReaction(_ context.Context, messageID, userID, emoji string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	reactions := s.reactions[messageID]
	for i, r := range reactions {
		if r.UserID == userID && r.Emoji == emoji {
			s.reactions[messageID] = append(reactions[:i], reactions[i+1:]...)
			return nil
		}
	}
	return nil
}

func (s *inMemoryRichStore) ListReactions(_ context.Context, messageID string) ([]ReactionSummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	grouped := make(map[string][]string)
	for _, r := range s.reactions[messageID] {
		grouped[r.Emoji] = append(grouped[r.Emoji], r.UserID)
	}
	var out []ReactionSummary
	for emoji, users := range grouped {
		out = append(out, ReactionSummary{Emoji: emoji, Count: len(users), UserIDs: users})
	}
	return out, nil
}

// --- Pins ---

func (s *inMemoryRichStore) PinMessage(_ context.Context, messageID, channelID, pinnedBy string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.pins[messageID]; ok {
		return ErrAlreadyPinned
	}
	pin := Pin{
		ID:        uuid.NewString(),
		MessageID: messageID,
		ChannelID: channelID,
		PinnedBy:  pinnedBy,
		PinnedAt:  time.Now().UTC(),
	}
	s.pins[messageID] = pin
	s.pinsByCh[channelID] = append(s.pinsByCh[channelID], messageID)
	return nil
}

func (s *inMemoryRichStore) UnpinMessage(_ context.Context, messageID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	pin, ok := s.pins[messageID]
	if !ok {
		return ErrNotPinned
	}
	delete(s.pins, messageID)
	ids := s.pinsByCh[pin.ChannelID]
	for i, id := range ids {
		if id == messageID {
			s.pinsByCh[pin.ChannelID] = append(ids[:i], ids[i+1:]...)
			break
		}
	}
	return nil
}

func (s *inMemoryRichStore) ListPins(_ context.Context, channelID string) ([]Pin, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []Pin
	for _, mid := range s.pinsByCh[channelID] {
		if pin, ok := s.pins[mid]; ok {
			out = append(out, pin)
		}
	}
	return out, nil
}

// --- Edit / Delete ---

func (s *inMemoryRichStore) UpdateMessageCiphertext(_ context.Context, messageID, senderUserID string, ciphertext, nonce []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	msg, ok := s.messages[messageID]
	if !ok {
		return ErrMessageNotFound
	}
	if msg.SenderUserID != senderUserID {
		return ErrNotSender
	}
	now := time.Now().UTC()
	msg.Ciphertext = append([]byte(nil), ciphertext...)
	msg.Nonce = append([]byte(nil), nonce...)
	msg.UpdatedAt = &now
	s.messages[messageID] = msg
	return nil
}

func (s *inMemoryRichStore) SoftDeleteMessage(_ context.Context, messageID, senderUserID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	msg, ok := s.messages[messageID]
	if !ok {
		return ErrMessageNotFound
	}
	if msg.SenderUserID != senderUserID {
		return ErrNotSender
	}
	now := time.Now().UTC()
	msg.DeletedAt = &now
	s.messages[messageID] = msg
	return nil
}
