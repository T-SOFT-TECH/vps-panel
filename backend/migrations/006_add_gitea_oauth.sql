-- Add Gitea OAuth fields to users table
ALTER TABLE users ADD COLUMN gitea_connected BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN gitea_token TEXT;
ALTER TABLE users ADD COLUMN gitea_username TEXT;
ALTER TABLE users ADD COLUMN gitea_url TEXT;
