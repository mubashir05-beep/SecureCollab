//go:build integration

package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"securecollab/services/analytics/internal/store"
)

func TestMessageVolume_HTTP_WithPostgresStore(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("ANALYTICS_TEST_DATABASE_URL"))
	if databaseURL == "" {
		databaseURL = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	if databaseURL == "" {
		t.Skip("ANALYTICS_TEST_DATABASE_URL or DATABASE_URL is not set")
	}

	s, err := store.NewPostgresAnalyticsStore(databaseURL)
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
			($7, $8, $9, $10, $11, 'ciphertext', $12)
	`
	_, err = db.ExecContext(ctx, insert,
		"11000000-0000-0000-0000-000000000001", "21000000-0000-0000-0000-000000000001", "31000000-0000-0000-0000-000000000001", []byte("c1"), []byte("n1"), now.Add(-2*time.Hour),
		"11000000-0000-0000-0000-000000000002", "21000000-0000-0000-0000-000000000002", "31000000-0000-0000-0000-000000000002", []byte("c2"), []byte("n2"), now.Add(-1*time.Hour),
	)
	if err != nil {
		t.Fatalf("insert fixture messages: %v", err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(s)
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/v1/analytics/messages/volume?window_hours=24", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
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
