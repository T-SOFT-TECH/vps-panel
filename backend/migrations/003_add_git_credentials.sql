-- Add git credentials columns to projects table
ALTER TABLE projects ADD COLUMN IF NOT EXISTS git_username VARCHAR(255);
ALTER TABLE projects ADD COLUMN IF NOT EXISTS git_token TEXT;
