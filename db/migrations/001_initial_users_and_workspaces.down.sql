-- 001_initial_users_and_workspaces.down.sql
-- Rollback initial schema migration.

DROP INDEX IF EXISTS idx_public_keys_user_id;
DROP INDEX IF EXISTS idx_channels_workspace_id;
DROP INDEX IF EXISTS idx_workspace_members_user_id;
DROP INDEX IF EXISTS idx_workspace_members_workspace_id;
DROP INDEX IF EXISTS idx_workspaces_owner_id;
DROP INDEX IF EXISTS idx_users_username;

DROP TABLE IF EXISTS public_keys;
DROP TABLE IF EXISTS channels;
DROP TABLE IF EXISTS workspace_members;
DROP TABLE IF EXISTS workspaces;
DROP TABLE IF EXISTS users;
