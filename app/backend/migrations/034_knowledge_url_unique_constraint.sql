-- +goose Up
-- Add unique constraint on knowledge_scraped_pages URL to prevent duplicates
-- This migration ensures URLs are unique across the entire database (not just per tenant)
-- URLs should be normalized (query params removed) before insertion

-- Step 1: Create a function to normalize URLs (remove query parameters)
CREATE OR REPLACE FUNCTION normalize_url(url TEXT) RETURNS TEXT AS $$
DECLARE
    parsed_url TEXT;
BEGIN
    -- Remove query parameters (everything after '?')
    parsed_url := SPLIT_PART(url, '?', 1);

    -- Remove fragment (everything after '#')
    parsed_url := SPLIT_PART(parsed_url, '#', 1);

    -- Convert to lowercase
    parsed_url := LOWER(parsed_url);

    -- Trim trailing slash (except for root paths)
    IF LENGTH(parsed_url) > 1 AND RIGHT(parsed_url, 1) = '/' THEN
        parsed_url := LEFT(parsed_url, LENGTH(parsed_url) - 1);
    END IF;

    RETURN parsed_url;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Step 2: Normalize existing URLs in knowledge_scraped_pages
UPDATE knowledge_scraped_pages
SET url = normalize_url(url)
WHERE url != normalize_url(url);

-- Step 3: Remove duplicate URLs (keep the most recent one)
-- First, identify duplicates
WITH duplicates AS (
    SELECT
        url,
        id,
        ROW_NUMBER() OVER (PARTITION BY url ORDER BY scraped_at DESC, id DESC) as rn
    FROM knowledge_scraped_pages
)
DELETE FROM knowledge_scraped_pages
WHERE id IN (
    SELECT id FROM duplicates WHERE rn > 1
);

-- Step 4: Add unique constraint on URL
-- This ensures each URL can only exist once in the database
CREATE UNIQUE INDEX IF NOT EXISTS idx_knowledge_scraped_pages_url_unique
ON knowledge_scraped_pages(url);

-- Step 5: Add comment to explain the constraint
COMMENT ON INDEX idx_knowledge_scraped_pages_url_unique IS
'Ensures URL uniqueness across the entire database. URLs must be normalized (query params removed) before insertion.';

-- +goose Down
-- Remove unique constraint
DROP INDEX IF EXISTS idx_knowledge_scraped_pages_url_unique;

-- Drop the normalization function
DROP FUNCTION IF EXISTS normalize_url(TEXT);
