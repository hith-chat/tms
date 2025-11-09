-- Better migration: Project-level URL deduplication
-- +goose Up

-- Drop the previous approach
DROP INDEX IF EXISTS idx_scraped_pages_unique_url_per_project;

-- Link scraped pages to deduplicated pages
ALTER TABLE knowledge_scraped_pages 

-- +goose Down
ALTER TABLE knowledge_scraped_pages DROP COLUMN IF EXISTS page_id;
ALTER TABLE knowledge_scraped_pages DROP COLUMN IF EXISTS content_hash;
