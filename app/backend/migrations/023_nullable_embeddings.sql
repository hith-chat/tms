-- +goose Up
-- Make embedding columns nullable to handle cases where embedding generation fails
-- This allows pages to be saved even when the embedding service is unavailable or fails

ALTER TABLE knowledge_scraped_pages ALTER COLUMN embedding DROP NOT NULL;
ALTER TABLE knowledge_chunks ALTER COLUMN embedding DROP NOT NULL;

-- +goose Down
-- Revert embedding columns to NOT NULL (this would require ensuring all existing records have embeddings)
-- ALTER TABLE knowledge_scraped_pages ALTER COLUMN embedding SET NOT NULL;
-- ALTER TABLE knowledge_chunks ALTER COLUMN embedding SET NOT NULL;
