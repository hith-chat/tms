-- Better migration: Project-level URL deduplication
-- +goose Up

-- Drop the previous approach
DROP INDEX IF EXISTS idx_scraped_pages_unique_url_per_project;

-- Create a new table for project-level page deduplication
CREATE TABLE IF NOT EXISTS knowledge_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    content_hash VARCHAR(64) NOT NULL,
    title TEXT,
    content TEXT NOT NULL,
    token_count INTEGER NOT NULL,
    embedding vector(1536),
    metadata JSONB DEFAULT '{}',
    first_scraped_at TIMESTAMP DEFAULT NOW(),
    last_scraped_at TIMESTAMP DEFAULT NOW(),
    scrape_count INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- Unique constraint: one URL per project
    UNIQUE(tenant_id, project_id, url),
    -- Unique constraint: one content hash per project (for true deduplication)
    UNIQUE(tenant_id, project_id, content_hash)
);

-- Create indexes
CREATE INDEX idx_knowledge_pages_project ON knowledge_pages(tenant_id, project_id);
CREATE INDEX idx_knowledge_pages_embedding ON knowledge_pages USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
CREATE INDEX idx_knowledge_pages_content_hash ON knowledge_pages(content_hash);

-- Link scraped pages to deduplicated pages
ALTER TABLE knowledge_scraped_pages 
ADD COLUMN page_id UUID REFERENCES knowledge_pages(id) ON DELETE SET NULL;

-- +goose Down
DROP TABLE IF EXISTS knowledge_pages CASCADE;
ALTER TABLE knowledge_scraped_pages DROP COLUMN IF EXISTS page_id;
ALTER TABLE knowledge_scraped_pages DROP COLUMN IF EXISTS content_hash;
