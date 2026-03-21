-- 003_workspace_channels_enhancements.up.sql
-- Add fields needed for workspace/channel CRUD features.

ALTER TABLE workspaces ADD COLUMN description TEXT DEFAULT '';
ALTER TABLE workspaces ADD COLUMN invite_code VARCHAR(32) UNIQUE;

ALTER TABLE channels ADD COLUMN is_private BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE channels ADD COLUMN topic VARCHAR(500) DEFAULT '';
ALTER TABLE channels ADD COLUMN archived_at TIMESTAMP;
ALTER TABLE channels ADD COLUMN created_by UUID REFERENCES users(id);

CREATE TABLE channel_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(channel_id, user_id)
);

CREATE INDEX idx_channel_members_channel_id ON channel_members(channel_id);
CREATE INDEX idx_channel_members_user_id ON channel_members(user_id);
CREATE INDEX idx_workspaces_invite_code ON workspaces(invite_code);
