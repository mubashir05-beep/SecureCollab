//go:build integration

package store

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestPostgresAnalyticsStore_GetMessageVolume(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("ANALYTICS_TEST_DATABASE_URL"))
	if databaseURL == "" {
		databaseURL = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	if databaseURL == "" {
		t.Skip("ANALYTICS_TEST_DATABASE_URL or DATABASE_URL is not set")
	}

	s, err := NewPostgresAnalyticsStore(databaseURL)
	if err != nil {
		t.Fatalf("init postgres analytics store: %v", err)
	}
	defer func() { _ = s.Close() }()

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	if err := ensureMessagesTable(ctx, db); err != nil {
		t.Fatalf("ensure encrypted_messages table: %v", err)
	}

	if _, err := db.ExecContext(ctx, "DELETE FROM encrypted_messages"); err != nil {
		t.Fatalf("cleanup messages: %v", err)
	}

	now := time.Now().UTC()
	insert := `
		INSERT INTO encrypted_messages
			(id, sender_user_id, recipient_user_id, ciphertext, nonce, content_type, created_at)
		VALUES
			($1, $2, $3, $4, $5, 'ciphertext', $6),
			($7, $8, $9, $10, $11, 'ciphertext', $12),
			($13, $14, $15, $16, $17, 'ciphertext', $18)
	`
	_, err = db.ExecContext(ctx, insert,
		"10000000-0000-0000-0000-000000000001", "20000000-0000-0000-0000-000000000001", "30000000-0000-0000-0000-000000000001", []byte("c1"), []byte("n1"), now.Add(-2*time.Hour),
		"10000000-0000-0000-0000-000000000002", "20000000-0000-0000-0000-000000000002", "30000000-0000-0000-0000-000000000002", []byte("c2"), []byte("n2"), now.Add(-25*time.Hour),
		"10000000-0000-0000-0000-000000000003", "20000000-0000-0000-0000-000000000003", "30000000-0000-0000-0000-000000000003", []byte("c3"), []byte("n3"), now.Add(-1*time.Hour),
	)
	if err != nil {
		t.Fatalf("insert fixture messages: %v", err)
	}

	volume, err := s.GetMessageVolume(ctx, 24)
	if err != nil {
		t.Fatalf("get message volume: %v", err)
	}
	if volume.TotalMessages != 3 {
		t.Fatalf("expected total 3, got %d", volume.TotalMessages)
	}
	if volume.MessagesLast24h != 2 {
		t.Fatalf("expected 2 messages in 24h, got %d", volume.MessagesLast24h)
	}
}

func ensureMessagesTable(ctx context.Context, db *sql.DB) error {
	ddl := `
		CREATE TABLE IF NOT EXISTS encrypted_messages (
			id UUID PRIMARY KEY,
			sender_user_id UUID NOT NULL,
			recipient_user_id UUID NOT NULL,
			channel_id UUID NULL,
			ciphertext BYTEA NOT NULL,
			nonce BYTEA NOT NULL,
			content_type VARCHAR(50) NOT NULL DEFAULT 'ciphertext',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.ExecContext(ctx, ddl)
	return err
}
