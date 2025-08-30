-- Migration to add unique constraints and implement deduplication
-- +goose Up

-- Add a unique index on URL per project (within tenant context)
-- This prevents the same URL from being stored multiple times for the same project
CREATE UNIQUE INDEX idx_scraped_pages_unique_url_per_project 
ON knowledge_scraped_pages (url, job_id);

-- Note: This approach still allows duplicates across different jobs
-- For true deduplication, we'd need a different approach

-- Alternative: Add a content hash for true deduplication
ALTER TABLE knowledge_scraped_pages 
ADD COLUMN content_hash VARCHAR(64);

-- Create index on content hash for fast lookups
CREATE INDEX idx_scraped_pages_content_hash 
ON knowledge_scraped_pages (content_hash);

-- +goose Down
DROP INDEX IF EXISTS idx_scraped_pages_unique_url_per_project;
DROP INDEX IF EXISTS idx_scraped_pages_content_hash;
ALTER TABLE knowledge_scraped_pages DROP COLUMN IF EXISTS content_hash;
