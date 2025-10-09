-- Add root_directory column to projects table for monorepo support
ALTER TABLE projects ADD COLUMN root_directory TEXT;
