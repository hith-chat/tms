-- +goose Up
-- Knowledge Management System Schema
-- This migration adds tables for AI-powered knowledge management functionality

-- Enable pgvector extension for vector similarity search
CREATE EXTENSION IF NOT EXISTS vector;

-- Knowledge documents table
CREATE TABLE knowledge_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    file_path TEXT NOT NULL,
    original_content TEXT,
    processed_content TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'processing' CHECK (status IN ('processing', 'completed', 'failed')),
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Document chunks table for embeddings
CREATE TABLE knowledge_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES knowledge_documents(id) ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL,
    content TEXT NOT NULL,
    token_count INTEGER NOT NULL,
    embedding vector(1536), -- OpenAI ada-002 dimension
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Web scraping jobs table
CREATE TABLE knowledge_scraping_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    max_depth INTEGER NOT NULL DEFAULT 5,
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
    pages_scraped INTEGER DEFAULT 0,
    total_pages INTEGER DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Scraped pages table
CREATE TABLE knowledge_scraped_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES knowledge_scraping_jobs(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    title TEXT,
    content TEXT NOT NULL,
    token_count INTEGER NOT NULL,
    scraped_at TIMESTAMP DEFAULT NOW(),
    embedding vector(1536),
    metadata JSONB DEFAULT '{}'
);

-- Knowledge base settings per project
CREATE TABLE knowledge_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT true,
    embedding_model VARCHAR(100) NOT NULL DEFAULT 'text-embedding-ada-002',
    chunk_size INTEGER NOT NULL DEFAULT 1000,
    chunk_overlap INTEGER NOT NULL DEFAULT 200,
    max_context_chunks INTEGER NOT NULL DEFAULT 5,
    similarity_threshold FLOAT NOT NULL DEFAULT 0.7,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(project_id)
);

-- Indexes for performance
CREATE INDEX idx_knowledge_documents_project ON knowledge_documents(project_id);
CREATE INDEX idx_knowledge_documents_tenant ON knowledge_documents(tenant_id);
CREATE INDEX idx_knowledge_documents_status ON knowledge_documents(status);

CREATE INDEX idx_knowledge_chunks_document ON knowledge_chunks(document_id);
CREATE INDEX idx_knowledge_chunks_embedding ON knowledge_chunks USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

CREATE INDEX idx_scraping_jobs_project ON knowledge_scraping_jobs(project_id);
CREATE INDEX idx_scraping_jobs_tenant ON knowledge_scraping_jobs(tenant_id);
CREATE INDEX idx_scraping_jobs_status ON knowledge_scraping_jobs(status);

CREATE INDEX idx_scraped_pages_job ON knowledge_scraped_pages(job_id);
CREATE INDEX idx_scraped_pages_embedding ON knowledge_scraped_pages USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

CREATE INDEX idx_knowledge_settings_project ON knowledge_settings(project_id);

-- Add created_at and updated_at triggers for automatic timestamp updates
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_knowledge_documents_updated_at BEFORE UPDATE ON knowledge_documents FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_knowledge_scraping_jobs_updated_at BEFORE UPDATE ON knowledge_scraping_jobs FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_knowledge_settings_updated_at BEFORE UPDATE ON knowledge_settings FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- Insert default knowledge settings for existing projects
INSERT INTO knowledge_settings (tenant_id, project_id)
SELECT tenant_id, id FROM projects
ON CONFLICT (project_id) DO NOTHING;

-- +goose Down
-- Drop knowledge management tables in reverse order
DROP TABLE IF EXISTS knowledge_chunks CASCADE;
DROP TABLE IF EXISTS knowledge_documents CASCADE; 
DROP TABLE IF EXISTS knowledge_scraping_jobs CASCADE;
DROP TABLE IF EXISTS knowledge_settings CASCADE;

-- Drop triggers
DROP TRIGGER IF EXISTS update_knowledge_documents_updated_at ON knowledge_documents;
DROP TRIGGER IF EXISTS update_knowledge_scraping_jobs_updated_at ON knowledge_scraping_jobs; 
DROP TRIGGER IF EXISTS update_knowledge_settings_updated_at ON knowledge_settings;
