package store

import (
	"context"
	"errors"
	"testing"
)

func TestInMemoryUserStore_CreateAndReadUser(t *testing.T) {
	s := NewInMemoryUserStore()

	created, err := s.CreateUser(context.Background(), "alice", "alice@example.com", "hash")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	fetched, err := s.GetUserByUsername(context.Background(), "alice")
	if err != nil {
		t.Fatalf("get user: %v", err)
	}

	if created.ID == "" {
		t.Fatal("expected non-empty id")
	}
	if fetched.Username != "alice" {
		t.Fatalf("expected username alice, got %s", fetched.Username)
	}
}

func TestInMemoryUserStore_DuplicateUsernameOrEmail(t *testing.T) {
	s := NewInMemoryUserStore()
	_, _ = s.CreateUser(context.Background(), "alice", "alice@example.com", "hash")

	_, err := s.CreateUser(context.Background(), "alice", "alice2@example.com", "hash")
	if !errors.Is(err, ErrUserExists) {
		t.Fatalf("expected ErrUserExists for duplicate username, got %v", err)
	}

	_, err = s.CreateUser(context.Background(), "alice2", "alice@example.com", "hash")
	if !errors.Is(err, ErrUserExists) {
		t.Fatalf("expected ErrUserExists for duplicate email, got %v", err)
	}
}

func TestInMemoryUserStore_UserNotFound(t *testing.T) {
	s := NewInMemoryUserStore()

	_, err := s.GetUserByUsername(context.Background(), "missing")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
