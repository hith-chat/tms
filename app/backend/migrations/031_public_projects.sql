-- +goose Up
-- +goose StatementBegin
-- Add is_public and expires_at columns to projects table for public AI widget builder feature
ALTER TABLE projects
ADD COLUMN is_public BOOLEAN DEFAULT FALSE NOT NULL,
ADD COLUMN expires_at TIMESTAMP NULL;

-- Add index on expires_at for efficient expiration queries
CREATE INDEX idx_projects_expires_at ON projects(expires_at) WHERE expires_at IS NOT NULL;

-- Add composite index for public project lookups
CREATE INDEX idx_projects_public_active ON projects(is_public, expires_at) WHERE is_public = TRUE;

-- Add comment for documentation
COMMENT ON COLUMN projects.is_public IS 'Indicates if this is a public project created via /api/public/ai-widget-builder';
COMMENT ON COLUMN projects.expires_at IS 'Timestamp when this public project expires (6 hours after creation). NULL for non-public projects.';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove indexes
DROP INDEX IF EXISTS idx_projects_public_active;
DROP INDEX IF EXISTS idx_projects_expires_at;

-- Remove columns
ALTER TABLE projects
DROP COLUMN IF EXISTS expires_at,
DROP COLUMN IF EXISTS is_public;
-- +goose StatementEnd