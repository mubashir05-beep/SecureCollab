package store

import (
	"context"
	"errors"
	"testing"
)

func TestInMemoryKeyStore_SaveAndFetchLatest(t *testing.T) {
	s := NewInMemoryKeyStore()

	_, err := s.SavePublicKey(context.Background(), "user-1", "identity", []byte{1, 2, 3})
	if err != nil {
		t.Fatalf("save first key: %v", err)
	}
	second, err := s.SavePublicKey(context.Background(), "user-1", "identity", []byte{4, 5, 6})
	if err != nil {
		t.Fatalf("save second key: %v", err)
	}

	latest, err := s.GetLatestPublicKey(context.Background(), "user-1", "identity")
	if err != nil {
		t.Fatalf("fetch latest key: %v", err)
	}

	if latest.ID != second.ID {
		t.Fatalf("expected latest key id %s, got %s", second.ID, latest.ID)
	}
	if string(latest.KeyData) != string([]byte{4, 5, 6}) {
		t.Fatalf("expected latest key data %v, got %v", []byte{4, 5, 6}, latest.KeyData)
	}
}

func TestInMemoryKeyStore_KeyNotFound(t *testing.T) {
	s := NewInMemoryKeyStore()

	_, err := s.GetLatestPublicKey(context.Background(), "missing", "identity")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}
