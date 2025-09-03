-- +goose Up
-- Add missing agent fields for enhanced assignment system

-- Add agent configuration columns
ALTER TABLE agents ADD COLUMN IF NOT EXISTS max_chats INTEGER DEFAULT 5;
ALTER TABLE agents ADD COLUMN IF NOT EXISTS last_activity_at TIMESTAMP;
ALTER TABLE agents ADD COLUMN IF NOT EXISTS last_assignment_at TIMESTAMP;

-- Create agent skills table
CREATE TABLE IF NOT EXISTS agent_skills (
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    skill VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (agent_id, skill)
);

-- Create index for efficient skill filtering
CREATE INDEX IF NOT EXISTS idx_agent_skills_skill ON agent_skills(skill);
CREATE INDEX IF NOT EXISTS idx_agent_skills_agent_id ON agent_skills(agent_id);

-- Add default skills for existing agents (optional)
-- INSERT INTO agent_skills (agent_id, tenant_id, skill) 
-- SELECT id, tenant_id, 'general' FROM agents WHERE id NOT IN (SELECT agent_id FROM agent_skills);

-- +goose Down
-- Remove agent enhancements

DROP INDEX IF EXISTS idx_agent_skills_skill;
DROP INDEX IF EXISTS idx_agent_skills_agent_id;
DROP TABLE IF EXISTS agent_skills;

ALTER TABLE agents DROP COLUMN IF EXISTS max_chats;
ALTER TABLE agents DROP COLUMN IF EXISTS last_activity_at;
ALTER TABLE agents DROP COLUMN IF EXISTS last_assignment_at;
