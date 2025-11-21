-- +goose Up
-- +goose StatementBegin

-- Add dedicated Slack columns to chat_sessions for efficient querying
ALTER TABLE chat_sessions ADD COLUMN IF NOT EXISTS slack_thread_ts VARCHAR(255);
ALTER TABLE chat_sessions ADD COLUMN IF NOT EXISTS slack_channel_id VARCHAR(255);

-- Create indexes for fast Slack thread lookups
CREATE INDEX IF NOT EXISTS idx_chat_sessions_slack_thread ON chat_sessions(slack_thread_ts, slack_channel_id) WHERE slack_thread_ts IS NOT NULL;

-- Migrate existing Slack metadata from JSONB meta field to dedicated columns
UPDATE chat_sessions
SET
    slack_thread_ts = meta->'slack'->>'thread_ts',
    slack_channel_id = meta->'slack'->>'channel_id'
WHERE meta->'slack'->>'thread_ts' IS NOT NULL;

-- Comments
COMMENT ON COLUMN chat_sessions.slack_thread_ts IS 'Slack thread timestamp for bidirectional message sync';
COMMENT ON COLUMN chat_sessions.slack_channel_id IS 'Slack channel ID where thread is located';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_chat_sessions_slack_thread;
ALTER TABLE chat_sessions DROP COLUMN IF EXISTS slack_channel_id;
ALTER TABLE chat_sessions DROP COLUMN IF EXISTS slack_thread_ts;

-- +goose StatementEnd
