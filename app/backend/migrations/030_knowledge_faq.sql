-- +goose Up
-- Create FAQ table for auto-generated knowledge question and answers

CREATE TABLE IF NOT EXISTS knowledge_faq_items (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    source_url TEXT,
    source_section TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_knowledge_faq_project ON knowledge_faq_items(project_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_faq_tenant_project ON knowledge_faq_items(tenant_id, project_id);

-- +goose Down
-- Drop FAQ table for knowledge base

DROP INDEX IF EXISTS idx_knowledge_faq_tenant_project;
DROP INDEX IF EXISTS idx_knowledge_faq_project;
DROP TABLE IF EXISTS knowledge_faq_items;
