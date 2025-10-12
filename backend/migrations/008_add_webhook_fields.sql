-- Migration: Add webhook fields for auto-deploy functionality
-- This enables projects to receive webhooks from GitHub, GitLab, and Gitea
-- and automatically redeploy when code is pushed

-- Add webhook_secret column for secure webhook verification
ALTER TABLE projects ADD COLUMN IF NOT EXISTS webhook_secret TEXT;

-- Add auto_deploy_branch column to specify which branch triggers auto-deploy
-- Defaults to the project's git_branch if not specified
ALTER TABLE projects ADD COLUMN IF NOT EXISTS auto_deploy_branch TEXT;

-- Update existing auto_deploy column comment (no schema change needed)
-- auto_deploy now specifically controls webhook-based auto-deployment

-- Create index for faster webhook lookups
CREATE INDEX IF NOT EXISTS idx_projects_auto_deploy ON projects(auto_deploy) WHERE auto_deploy = true;
