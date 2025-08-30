-- +goose Up
-- Add content hash column for better deduplication that detects content changes
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'knowledge_scraped_pages' 
                   AND column_name = 'content_hash') THEN
        ALTER TABLE knowledge_scraped_pages 
        ADD COLUMN content_hash VARCHAR(64);
    END IF;
END $$;

-- Create index on content hash for fast lookups (if not exists)
CREATE INDEX IF NOT EXISTS idx_scraped_pages_content_hash 
ON knowledge_scraped_pages (content_hash);

-- Create a composite index for URL + content hash lookups (if not exists)
CREATE INDEX IF NOT EXISTS idx_scraped_pages_url_content_hash 
ON knowledge_scraped_pages (url, content_hash);

-- +goose Down
DROP INDEX IF EXISTS idx_scraped_pages_content_hash;
DROP INDEX IF EXISTS idx_scraped_pages_url_content_hash;
ALTER TABLE knowledge_scraped_pages DROP COLUMN IF EXISTS content_hash;
