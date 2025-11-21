-- +goose Up
-- +goose StatementBegin

-- Create project_integrations table for simplified integration storage
CREATE TABLE project_integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    integration_type VARCHAR(50) NOT NULL,
    meta JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Unique constraint: one integration per type per project
CREATE UNIQUE INDEX idx_project_integrations_unique
    ON project_integrations(tenant_id, project_id, integration_type);

-- Index for listing integrations by project
CREATE INDEX idx_project_integrations_project
    ON project_integrations(tenant_id, project_id);

-- Index for querying by integration type
CREATE INDEX idx_project_integrations_type
    ON project_integrations(integration_type);

-- Index for status filtering
CREATE INDEX idx_project_integrations_status
    ON project_integrations(status);

-- GIN index for JSONB meta field queries
CREATE INDEX idx_project_integrations_meta
    ON project_integrations USING GIN (meta);

-- Comment on table
COMMENT ON TABLE project_integrations IS 'Simplified integration storage with flexible JSONB meta field for different integration types';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS project_integrations;

-- +goose StatementEnd
