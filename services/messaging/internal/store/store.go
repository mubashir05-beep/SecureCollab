package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

var ErrMessageNotFound = errors.New("message not found")

type EncryptedMessage struct {
	ID              string
	SenderUserID    string
	RecipientUserID string
	ChannelID       string
	Ciphertext      []byte
	Nonce           []byte
	ContentType     string
	CreatedAt       time.Time
}

type MessageStore interface {
	SaveEncryptedMessage(ctx context.Context, msg EncryptedMessage) (EncryptedMessage, error)
	ListInbox(ctx context.Context, recipientUserID string, limit int) ([]EncryptedMessage, error)
}

type inMemoryMessageStore struct {
	mu      sync.Mutex
	byInbox map[string][]EncryptedMessage
}

func NewInMemoryMessageStore() MessageStore {
	return &inMemoryMessageStore{byInbox: make(map[string][]EncryptedMessage)}
}

func (s *inMemoryMessageStore) SaveEncryptedMessage(_ context.Context, msg EncryptedMessage) (EncryptedMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

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

	s.byInbox[msg.RecipientUserID] = append(s.byInbox[msg.RecipientUserID], msg)
	return msg, nil
}

func (s *inMemoryMessageStore) ListInbox(_ context.Context, recipientUserID string, limit int) ([]EncryptedMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries := s.byInbox[recipientUserID]
	out := make([]EncryptedMessage, len(entries))
	copy(out, entries)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	if limit <= 0 || limit > len(out) {
		limit = len(out)
	}
	out = out[:limit]
	for i := range out {
		out[i].Ciphertext = append([]byte(nil), out[i].Ciphertext...)
		out[i].Nonce = append([]byte(nil), out[i].Nonce...)
	}
	return out, nil
}

type PostgresMessageStore struct {
	db *sql.DB
}

func NewPostgresMessageStore(databaseURL string) (*PostgresMessageStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &PostgresMessageStore{db: db}, nil
}

func (s *PostgresMessageStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *PostgresMessageStore) SaveEncryptedMessage(ctx context.Context, msg EncryptedMessage) (EncryptedMessage, error) {
	if msg.ContentType == "" {
		msg.ContentType = "ciphertext"
	}
	query := `
		INSERT INTO encrypted_messages
			(id, sender_user_id, recipient_user_id, channel_id, ciphertext, nonce, content_type)
		VALUES
			($1, $2, $3, NULLIF($4, '')::uuid, $5, $6, $7)
		RETURNING id::text, sender_user_id::text, recipient_user_id::text, COALESCE(channel_id::text, ''), ciphertext, nonce, content_type, created_at
	`
	row := s.db.QueryRowContext(ctx, query,
		uuid.NewString(),
		msg.SenderUserID,
		msg.RecipientUserID,
		msg.ChannelID,
		msg.Ciphertext,
		msg.Nonce,
		msg.ContentType,
	)
	stored := EncryptedMessage{}
	if err := row.Scan(&stored.ID, &stored.SenderUserID, &stored.RecipientUserID, &stored.ChannelID, &stored.Ciphertext, &stored.Nonce, &stored.ContentType, &stored.CreatedAt); err != nil {
		return EncryptedMessage{}, fmt.Errorf("insert encrypted message: %w", err)
	}
	return stored, nil
}

func (s *PostgresMessageStore) ListInbox(ctx context.Context, recipientUserID string, limit int) ([]EncryptedMessage, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id::text, sender_user_id::text, recipient_user_id::text, COALESCE(channel_id::text, ''), ciphertext, nonce, content_type, created_at
		FROM encrypted_messages
		WHERE recipient_user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := s.db.QueryContext(ctx, query, recipientUserID, limit)
	if err != nil {
		return nil, fmt.Errorf("query inbox: %w", err)
	}
	defer rows.Close()

	result := make([]EncryptedMessage, 0)
	for rows.Next() {
		msg := EncryptedMessage{}
		if err := rows.Scan(&msg.ID, &msg.SenderUserID, &msg.RecipientUserID, &msg.ChannelID, &msg.Ciphertext, &msg.Nonce, &msg.ContentType, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan inbox message: %w", err)
		}
		result = append(result, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate inbox rows: %w", err)
	}
	return result, nil
}

func NewMessageStoreFromEnv() (MessageStore, func() error, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return NewInMemoryMessageStore(), func() error { return nil }, nil
	}

	postgresStore, err := NewPostgresMessageStore(databaseURL)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			return nil, nil, fmt.Errorf("init postgres message store: %s (%s)", pgErr.Message, pgErr.Code)
		}
		return nil, nil, err
	}
	return postgresStore, postgresStore.Close, nil
}
