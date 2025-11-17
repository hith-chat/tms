-- +goose Up
-- Add tenant_id and project_id to widget_knowledge_pages
-- This enables efficient querying of knowledge pages by tenant/project without joining through widgets

-- Add tenant_id column
ALTER TABLE widget_knowledge_pages
ADD COLUMN tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

-- Add project_id column
ALTER TABLE widget_knowledge_pages
ADD COLUMN project_id UUID REFERENCES projects(id) ON DELETE CASCADE;

-- Backfill existing records with tenant_id and project_id from associated widgets
UPDATE widget_knowledge_pages wkp
SET
    tenant_id = cw.tenant_id,
    project_id = cw.project_id
FROM chat_widgets cw
WHERE wkp.widget_id = cw.id
AND (wkp.tenant_id IS NULL OR wkp.project_id IS NULL);

-- Make columns NOT NULL after backfilling
ALTER TABLE widget_knowledge_pages
ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE widget_knowledge_pages
ALTER COLUMN project_id SET NOT NULL;

-- Create indexes for efficient tenant and project-level queries
CREATE INDEX idx_widget_knowledge_pages_tenant ON widget_knowledge_pages(tenant_id);
CREATE INDEX idx_widget_knowledge_pages_project ON widget_knowledge_pages(project_id);

-- Create composite index for project + widget queries (most common pattern)
CREATE INDEX idx_widget_knowledge_pages_project_widget ON widget_knowledge_pages(project_id, widget_id);

-- +goose Down
-- Drop indexes
DROP INDEX IF EXISTS idx_widget_knowledge_pages_tenant;
DROP INDEX IF EXISTS idx_widget_knowledge_pages_project;
DROP INDEX IF EXISTS idx_widget_knowledge_pages_project_widget;

-- Drop columns
ALTER TABLE widget_knowledge_pages DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE widget_knowledge_pages DROP COLUMN IF EXISTS project_id;
