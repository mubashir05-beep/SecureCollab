//go:build integration

package store

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestPostgresUserStore_CreateAndReadUser(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("AUTH_TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("AUTH_TEST_DATABASE_URL is not set")
	}

	s, err := NewPostgresUserStore(databaseURL)
	if err != nil {
		t.Fatalf("init postgres store: %v", err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Fatalf("close postgres store: %v", err)
		}
	}()

	ctx := context.Background()
	if err := ensureUsersTableForIntegration(ctx, s); err != nil {
		t.Fatalf("ensure users table: %v", err)
	}

	suffix := time.Now().UnixNano()
	username := fmt.Sprintf("itest_user_%d", suffix)
	email := fmt.Sprintf("itest_%d@example.com", suffix)
	defer cleanupIntegrationUser(ctx, t, s, username)

	created, err := s.CreateUser(ctx, username, email, "hash-value")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	fetched, err := s.GetUserByUsername(ctx, username)
	if err != nil {
		t.Fatalf("fetch user: %v", err)
	}

	if created.ID == "" {
		t.Fatal("expected created user id")
	}
	if fetched.Username != username {
		t.Fatalf("expected username %s, got %s", username, fetched.Username)
	}
	if fetched.Email != email {
		t.Fatalf("expected email %s, got %s", email, fetched.Email)
	}
}

func TestPostgresUserStore_DuplicateConstraint(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("AUTH_TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("AUTH_TEST_DATABASE_URL is not set")
	}

	s, err := NewPostgresUserStore(databaseURL)
	if err != nil {
		t.Fatalf("init postgres store: %v", err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Fatalf("close postgres store: %v", err)
		}
	}()

	ctx := context.Background()
	if err := ensureUsersTableForIntegration(ctx, s); err != nil {
		t.Fatalf("ensure users table: %v", err)
	}

	suffix := time.Now().UnixNano()
	username := fmt.Sprintf("itest_dupe_%d", suffix)
	email := fmt.Sprintf("itest_dupe_%d@example.com", suffix)
	defer cleanupIntegrationUser(ctx, t, s, username)

	if _, err := s.CreateUser(ctx, username, email, "hash-value"); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	_, err = s.CreateUser(ctx, username, fmt.Sprintf("other_%d@example.com", suffix), "hash-value")
	if err != ErrUserExists {
		t.Fatalf("expected ErrUserExists for duplicate username, got %v", err)
	}

	_, err = s.CreateUser(ctx, fmt.Sprintf("other_%d", suffix), email, "hash-value")
	if err != ErrUserExists {
		t.Fatalf("expected ErrUserExists for duplicate email, got %v", err)
	}
}

func ensureUsersTableForIntegration(ctx context.Context, s *PostgresUserStore) error {
	const ddl = `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := s.db.ExecContext(ctx, ddl)
	return err
}

func cleanupIntegrationUser(ctx context.Context, t *testing.T, s *PostgresUserStore, username string) {
	t.Helper()
	if _, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE username = $1", username); err != nil {
		t.Fatalf("cleanup test user: %v", err)
	}
}
