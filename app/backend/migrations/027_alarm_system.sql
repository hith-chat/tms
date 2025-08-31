-- +goose Up
-- Alarm System Schema  
-- This migration adds tables for the howling alarm system

-- Create alarms table
CREATE TABLE IF NOT EXISTS alarms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    assignment_id UUID NULL, -- Logical assignment reference, no FK constraint
    agent_id UUID REFERENCES agents(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    priority notification_priority NOT NULL DEFAULT 'normal',
    current_level TEXT NOT NULL DEFAULT 'soft',
    start_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_escalation TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    escalation_count INTEGER NOT NULL DEFAULT 0,
    is_acknowledged BOOLEAN NOT NULL DEFAULT FALSE,
    acknowledged_at TIMESTAMP WITH TIME ZONE NULL,
    acknowledged_by UUID REFERENCES agents(id) ON DELETE SET NULL,
    config JSONB NOT NULL DEFAULT '{}',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create alarm_acknowledgments table
CREATE TABLE IF NOT EXISTS alarm_acknowledgments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alarm_id UUID NOT NULL REFERENCES alarms(id) ON DELETE CASCADE,
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    response TEXT NOT NULL DEFAULT '',
    acknowledged_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_alarms_tenant_project ON alarms(tenant_id, project_id);
CREATE INDEX IF NOT EXISTS idx_alarms_active ON alarms(tenant_id, project_id, is_acknowledged) WHERE is_acknowledged = FALSE;
CREATE INDEX IF NOT EXISTS idx_alarms_escalation ON alarms(last_escalation) WHERE is_acknowledged = FALSE;
CREATE INDEX IF NOT EXISTS idx_alarms_created_at ON alarms(created_at);
CREATE INDEX IF NOT EXISTS idx_alarm_acknowledgments_alarm_id ON alarm_acknowledgments(alarm_id);

-- Add CHECK constraints for alarm levels
ALTER TABLE alarms ADD CONSTRAINT check_alarm_level 
    CHECK (current_level IN ('soft', 'medium', 'loud', 'urgent', 'critical'));

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_alarms_updated_at BEFORE UPDATE ON alarms
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
-- Drop alarm system tables in reverse order
DROP TABLE IF EXISTS alarm_acknowledgments CASCADE;
DROP TABLE IF EXISTS alarms CASCADE;

-- Drop triggers
DROP TRIGGER IF EXISTS update_alarms_updated_at ON alarms;
