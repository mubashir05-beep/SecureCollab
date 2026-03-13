package store

import (
	"context"
	"testing"
)

func TestInMemoryMessageStore_SaveAndInbox(t *testing.T) {
	s := NewInMemoryMessageStore()

	_, err := s.SaveEncryptedMessage(context.Background(), EncryptedMessage{
		SenderUserID:    "sender-1",
		RecipientUserID: "recipient-1",
		Ciphertext:      []byte("ciphertext-1"),
		Nonce:           []byte("nonce-1"),
	})
	if err != nil {
		t.Fatalf("save message: %v", err)
	}

	inbox, err := s.ListInbox(context.Background(), "recipient-1", 10)
	if err != nil {
		t.Fatalf("list inbox: %v", err)
	}
	if len(inbox) != 1 {
		t.Fatalf("expected inbox length 1, got %d", len(inbox))
	}
	if inbox[0].SenderUserID != "sender-1" {
		t.Fatalf("expected sender sender-1, got %s", inbox[0].SenderUserID)
	}
	if string(inbox[0].Ciphertext) != "ciphertext-1" {
		t.Fatalf("expected ciphertext ciphertext-1, got %s", string(inbox[0].Ciphertext))
	}
}
