-- 002_encrypted_messages.up.sql
-- Stores encrypted message payloads and metadata only.

CREATE TABLE encrypted_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_user_id UUID NOT NULL,
    recipient_user_id UUID NOT NULL,
    channel_id UUID,
    ciphertext BYTEA NOT NULL,
    nonce BYTEA NOT NULL,
    content_type VARCHAR(50) NOT NULL DEFAULT 'ciphertext',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_encrypted_messages_recipient_created_at ON encrypted_messages(recipient_user_id, created_at DESC);
CREATE INDEX idx_encrypted_messages_sender_created_at ON encrypted_messages(sender_user_id, created_at DESC);
CREATE INDEX idx_encrypted_messages_channel_created_at ON encrypted_messages(channel_id, created_at DESC);
