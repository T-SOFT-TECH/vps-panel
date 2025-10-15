-- Add PocketBase version tracking field to projects table
-- This allows us to show current version and check for updates

ALTER TABLE projects ADD COLUMN IF NOT EXISTS pocketbase_version TEXT;

-- Add index for faster queries filtering by BaaS type with version
CREATE INDEX IF NOT EXISTS idx_projects_baas_version ON projects(baas_type, pocketbase_version) WHERE baas_type = 'pocketbase';
