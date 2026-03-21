-- 003_workspace_channels_enhancements.down.sql

DROP INDEX IF EXISTS idx_channel_members_user_id;
DROP INDEX IF EXISTS idx_channel_members_channel_id;
DROP INDEX IF EXISTS idx_workspaces_invite_code;

DROP TABLE IF EXISTS channel_members;

ALTER TABLE channels DROP COLUMN IF EXISTS created_by;
ALTER TABLE channels DROP COLUMN IF EXISTS archived_at;
ALTER TABLE channels DROP COLUMN IF EXISTS topic;
ALTER TABLE channels DROP COLUMN IF EXISTS is_private;

ALTER TABLE workspaces DROP COLUMN IF EXISTS invite_code;
ALTER TABLE workspaces DROP COLUMN IF EXISTS description;
