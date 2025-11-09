-- +goose Up
-- Widget Knowledge Pages Mapping Table
-- This migration creates a junction table to associate widgets with knowledge pages,
-- enabling embedding reuse across multiple widgets and automatic content updates

-- Create widget_knowledge_pages mapping table
CREATE TABLE widget_knowledge_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    widget_id UUID NOT NULL REFERENCES chat_widgets(id) ON DELETE CASCADE,
    page_id UUID NOT NULL REFERENCES knowledge_scraped_pages(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    -- Ensure a widget doesn't reference the same page twice
    UNIQUE(widget_id, page_id)
);

-- Indexes for performance
CREATE INDEX idx_widget_knowledge_pages_widget ON widget_knowledge_pages(widget_id);
CREATE INDEX idx_widget_knowledge_pages_page ON widget_knowledge_pages(page_id);
CREATE INDEX idx_widget_knowledge_pages_created ON widget_knowledge_pages(created_at);

-- Add tenant_id to knowledge_scraped_pages for tenant-level queries
-- (job_id still links to project, but tenant_id enables cross-project deduplication)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'knowledge_scraped_pages'
                   AND column_name = 'tenant_id') THEN
        ALTER TABLE knowledge_scraped_pages
        ADD COLUMN tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Make job_id nullable since pages are now shared resources
-- job_id tracks the FIRST job that discovered the page, but pages can be reused by multiple jobs/widgets
ALTER TABLE knowledge_scraped_pages
ALTER COLUMN job_id DROP NOT NULL;

-- Populate tenant_id from job_id relationship
UPDATE knowledge_scraped_pages ksp
SET tenant_id = ksj.tenant_id
FROM knowledge_scraping_jobs ksj
WHERE ksp.job_id = ksj.id
AND ksp.tenant_id IS NULL;

-- Create index for tenant-level queries
CREATE INDEX IF NOT EXISTS idx_scraped_pages_tenant ON knowledge_scraped_pages(tenant_id);

-- Create composite index for tenant + URL + content_hash lookups (key for deduplication)
CREATE INDEX IF NOT EXISTS idx_scraped_pages_tenant_url_hash
ON knowledge_scraped_pages(tenant_id, url, content_hash);

-- Add updated_at trigger for widget_knowledge_pages
CREATE TRIGGER update_widget_knowledge_pages_updated_at
BEFORE UPDATE ON widget_knowledge_pages
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- +goose Down
-- Drop indexes
DROP INDEX IF EXISTS idx_widget_knowledge_pages_widget;
DROP INDEX IF EXISTS idx_widget_knowledge_pages_page;
DROP INDEX IF EXISTS idx_widget_knowledge_pages_created;
DROP INDEX IF EXISTS idx_scraped_pages_tenant;
DROP INDEX IF EXISTS idx_scraped_pages_tenant_url_hash;

-- Drop trigger
DROP TRIGGER IF EXISTS update_widget_knowledge_pages_updated_at ON widget_knowledge_pages;

-- Drop table
DROP TABLE IF EXISTS widget_knowledge_pages CASCADE;

-- Remove tenant_id column from knowledge_scraped_pages
ALTER TABLE knowledge_scraped_pages DROP COLUMN IF EXISTS tenant_id;
