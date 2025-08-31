-- Test data for knowledge management system tests
-- This file contains mock data for testing database operations

-- Clean up existing test data
DELETE FROM knowledge_scraped_pages WHERE job_id IN (
    SELECT id FROM knowledge_scraping_jobs WHERE tenant_id = '123e4567-e89b-12d3-a456-426614174000'
);
DELETE FROM knowledge_chunks WHERE document_id IN (
    SELECT id FROM knowledge_documents WHERE tenant_id = '123e4567-e89b-12d3-a456-426614174000'
);
DELETE FROM knowledge_scraping_jobs WHERE tenant_id = '123e4567-e89b-12d3-a456-426614174000';
DELETE FROM knowledge_documents WHERE tenant_id = '123e4567-e89b-12d3-a456-426614174000';

-- Insert test tenant and project (if not exists)
INSERT INTO tenants (id, name, created_at, updated_at) 
VALUES ('123e4567-e89b-12d3-a456-426614174000', 'Test Tenant', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO projects (id, tenant_id, name, created_at, updated_at)
VALUES ('123e4567-e89b-12d3-a456-426614174001', '123e4567-e89b-12d3-a456-426614174000', 'Test Project', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Insert test documents
INSERT INTO knowledge_documents (
    id, tenant_id, project_id, filename, content_type, file_size, 
    file_path, status, created_at, updated_at
) VALUES 
(
    '123e4567-e89b-12d3-a456-426614174003',
    '123e4567-e89b-12d3-a456-426614174000', 
    '123e4567-e89b-12d3-a456-426614174001',
    'test-document.pdf',
    'application/pdf',
    1024000,
    '/uploads/test-document.pdf',
    'completed',
    NOW(),
    NOW()
),
(
    '123e4567-e89b-12d3-a456-426614174005',
    '123e4567-e89b-12d3-a456-426614174000', 
    '123e4567-e89b-12d3-a456-426614174001',
    'second-document.pdf',
    'application/pdf',
    2048000,
    '/uploads/second-document.pdf',
    'processing',
    NOW(),
    NOW()
);

-- Insert test chunks with embeddings
INSERT INTO knowledge_chunks (
    id, document_id, chunk_index, content, token_count, embedding, created_at
) VALUES 
(
    '123e4567-e89b-12d3-a456-426614174006',
    '123e4567-e89b-12d3-a456-426614174003',
    0,
    'This is the first chunk of test content from the PDF document.',
    15,
    '[0.1,0.2,0.3]'::vector,
    NOW()
),
(
    '123e4567-e89b-12d3-a456-426614174007',
    '123e4567-e89b-12d3-a456-426614174003',
    1,
    'This is the second chunk containing more detailed information.',
    12,
    '[0.2,0.3,0.4]'::vector,
    NOW()
);

-- Insert test scraping job
INSERT INTO knowledge_scraping_jobs (
    id, tenant_id, project_id, url, max_depth, status, 
    pages_scraped, total_pages, started_at, completed_at, created_at, updated_at
) VALUES (
    '123e4567-e89b-12d3-a456-426614174004',
    '123e4567-e89b-12d3-a456-426614174000',
    '123e4567-e89b-12d3-a456-426614174001',
    'https://example.com',
    3,
    'completed',
    5,
    5,
    NOW() - INTERVAL '1 hour',
    NOW() - INTERVAL '30 minutes',
    NOW() - INTERVAL '2 hours',
    NOW() - INTERVAL '30 minutes'
);

-- Insert test scraped pages
INSERT INTO knowledge_scraped_pages (
    id, job_id, url, title, content, scraped_at, embedding
) VALUES 
(
    '123e4567-e89b-12d3-a456-426614174008',
    '123e4567-e89b-12d3-a456-426614174004',
    'https://example.com',
    'Example Home Page',
    'This is the content from the example.com homepage.',
    NOW() - INTERVAL '1 hour',
    '[0.3,0.4,0.5]'::vector
),
(
    '123e4567-e89b-12d3-a456-426614174009',
    '123e4567-e89b-12d3-a456-426614174004',
    'https://example.com/about',
    'About Us',
    'This is the content from the about page with company information.',
    NOW() - INTERVAL '50 minutes',
    '[0.4,0.5,0.6]'::vector
);

-- Create test indexes for performance testing
CREATE INDEX IF NOT EXISTS idx_test_chunks_embedding 
ON knowledge_chunks USING ivfflat (embedding vector_cosine_ops);

CREATE INDEX IF NOT EXISTS idx_test_pages_embedding 
ON knowledge_scraped_pages USING ivfflat (embedding vector_cosine_ops);
