-- +goose Up
-- +goose StatementBegin

-- Add meta column to chat_sessions for storing integration-specific data
ALTER TABLE chat_sessions ADD COLUMN IF NOT EXISTS meta JSONB DEFAULT '{}';

-- Create GIN index for JSONB meta field queries
CREATE INDEX IF NOT EXISTS idx_chat_sessions_meta ON chat_sessions USING GIN (meta);

-- Comment
COMMENT ON COLUMN chat_sessions.meta IS 'JSONB field for storing integration metadata (Slack threads, Discord channels, etc.)';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_chat_sessions_meta;
ALTER TABLE chat_sessions DROP COLUMN IF EXISTS meta;

-- +goose StatementEnd
