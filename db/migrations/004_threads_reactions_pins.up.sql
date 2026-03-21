-- 004_threads_reactions_pins.up.sql
-- Rich messaging features: threads, reactions, pins, edit/delete support.

-- Thread support: parent_message_id links a reply to its parent
ALTER TABLE encrypted_messages ADD COLUMN parent_message_id UUID REFERENCES encrypted_messages(id) ON DELETE SET NULL;
ALTER TABLE encrypted_messages ADD COLUMN updated_at TIMESTAMP;
ALTER TABLE encrypted_messages ADD COLUMN deleted_at TIMESTAMP;

CREATE INDEX idx_encrypted_messages_parent ON encrypted_messages(parent_message_id, created_at) WHERE parent_message_id IS NOT NULL;

-- Reactions (emoji reactions on messages)
CREATE TABLE message_reactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL REFERENCES encrypted_messages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    emoji VARCHAR(32) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(message_id, user_id, emoji)
);

CREATE INDEX idx_message_reactions_message ON message_reactions(message_id);

-- Pins (pinned messages in a channel)
CREATE TABLE message_pins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL REFERENCES encrypted_messages(id) ON DELETE CASCADE,
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    pinned_by UUID NOT NULL,
    pinned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(message_id)
);

CREATE INDEX idx_message_pins_channel ON message_pins(channel_id, pinned_at DESC);
