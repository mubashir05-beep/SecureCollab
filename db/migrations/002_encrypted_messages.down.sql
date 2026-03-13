-- 002_encrypted_messages.down.sql

DROP INDEX IF EXISTS idx_encrypted_messages_channel_created_at;
DROP INDEX IF EXISTS idx_encrypted_messages_sender_created_at;
DROP INDEX IF EXISTS idx_encrypted_messages_recipient_created_at;
DROP TABLE IF EXISTS encrypted_messages;
