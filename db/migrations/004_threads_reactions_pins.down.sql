-- 004_threads_reactions_pins.down.sql

DROP INDEX IF EXISTS idx_message_pins_channel;
DROP TABLE IF EXISTS message_pins;

DROP INDEX IF EXISTS idx_message_reactions_message;
DROP TABLE IF EXISTS message_reactions;

DROP INDEX IF EXISTS idx_encrypted_messages_parent;
ALTER TABLE encrypted_messages DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE encrypted_messages DROP COLUMN IF EXISTS updated_at;
ALTER TABLE encrypted_messages DROP COLUMN IF EXISTS parent_message_id;
