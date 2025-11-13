-- +goose Up
-- Add domain_url column to chat_widgets table
ALTER TABLE chat_widgets
ADD COLUMN domain_url TEXT NOT NULL DEFAULT 'hith.chat';

-- Add index for faster domain lookups
CREATE INDEX idx_chat_widgets_domain_url ON chat_widgets(domain_url);

-- +goose Down
-- Remove the index and column
DROP INDEX IF EXISTS idx_chat_widgets_domain_url;
ALTER TABLE chat_widgets DROP COLUMN IF EXISTS domain_url;
