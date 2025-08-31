# AI Knowledge Management Integration Plan

## Overview
Implement a comprehensive AI knowledge management system that allows users to upload documents, scrape websites, and use this knowledge to enhance AI chat responses with contextual information.

## Architecture Components

### 1. Document Processing Service
- PDF document upload and processing
- Text extraction from PDFs
- Document chunking for embeddings
- Metadata storage (file name, upload date, project association)

### 2. Web Scraping Service
- URL-based content extraction (up to 5 levels deep)
- Sitemap parsing
- Content cleaning and normalization
- Rate limiting and respectful crawling

### 3. Embedding Service
- Text-to-vector conversion using OpenAI/cheaper alternatives
- Batch processing for large documents
- Embedding storage and indexing
- Similarity search capabilities

### 4. Vector Database
- PostgreSQL with pgvector extension (Docker-supported)
- Efficient similarity search
- Document metadata associations
- Scalable storage

### 5. Knowledge-Enhanced Chat
- Context retrieval based on chat query
- RAG (Retrieval-Augmented Generation) implementation
- Dynamic context injection
- Source attribution in responses

## Implementation Phases

## Phase 1: Database Foundation
### Todo Items:
- [x] **DONE** ✅ Add pgvector extension to PostgreSQL setup
- [x] **DONE** ✅ Create knowledge base database schema
- [x] **DONE** ✅ Create document storage tables
- [x] **DONE** ✅ Create embedding storage tables
- [x] **DONE** ✅ Create web scraping job tables
- [x] **DONE** ✅ Update Docker configuration for pgvector

## Phase 2: Core Knowledge Services
### Todo Items:
- [x] **DONE** ✅ Create document processing service
- [x] **DONE** ✅ Create embedding generation service
- [x] **DONE** ✅ Create web scraping service
- [x] **DONE** ✅ Create knowledge retrieval service
- [x] **DONE** ✅ Create vector similarity search
- [x] **DONE** ✅ Add document upload handlers

## Phase 3: PDF Processing
### Todo Items:
- [x] **DONE** ✅ Implement PDF text extraction
- [x] **DONE** ✅ Add file upload validation
- [x] **DONE** ✅ Implement document chunking strategy
- [x] **DONE** ✅ Add document metadata processing
- [x] **DONE** ✅ Create document storage repository

## Phase 4: Web Scraping
### Todo Items:
- [x] **DONE** ✅ Implement website content extraction
- [x] **DONE** ✅ Add URL validation and sanitization
- [x] **DONE** ✅ Implement recursive crawling (5 levels)
- [x] **DONE** ✅ Add robots.txt respect
- [x] **DONE** ✅ Implement rate limiting
- [x] **DONE** ✅ Add scraping job management

## Phase 5: Embedding Integration
### Todo Items:
- [x] **DONE** ✅ Integrate OpenAI embedding API
- [x] **DONE** ✅ Add fallback to cheaper embedding services
- [x] **DONE** ✅ Implement batch embedding processing
- [x] **DONE** ✅ Add embedding storage optimization
- [x] **DONE** ✅ Create similarity search algorithms

## Phase 6: RAG Implementation
### Todo Items:
- [x] **DONE** ✅ Modify AI service for context injection
- [x] **DONE** ✅ Implement context retrieval logic
- [x] **DONE** ✅ Add relevance scoring
- [x] **DONE** ✅ Implement source attribution
- [x] **DONE** ✅ Add context window management

## Phase 7: Frontend Integration
### Todo Items:
- [x] **DONE** ✅ Create knowledge management settings page
- [x] **DONE** ✅ Add document upload interface
- [x] **DONE** ✅ Add website URL input interface
- [x] **DONE** ✅ Create document management UI
- [x] **DONE** ✅ Add scraping job status display
- [x] **DONE** ✅ Add knowledge source indicators in chat

## Phase 8: API Endpoints
### Todo Items:
- [x] **DONE** ✅ Create document upload API
- [x] **DONE** ✅ Create website scraping API
- [x] **DONE** ✅ Create knowledge source listing API
- [x] **DONE** ✅ Create document deletion API
- [x] **DONE** ✅ Create scraping job status API
- [x] **DONE** ✅ Add knowledge search API

## Phase 9: Testing & Optimization
### Todo Items:
- [x] **DONE** ✅ Add comprehensive tests for all services
- [x] **DONE** ✅ Performance optimization for large documents
- [x] **DONE** ✅ Memory usage optimization
- [x] **DONE** ✅ Database query optimization
- [x] **DONE** ✅ Error handling and retry mechanisms

## Phase 10: Production Readiness
### Todo Items:
- [x] **DONE** ✅ Add monitoring and logging
- [x] **DONE** ✅ Implement backup strategies
- [x] **DONE** ✅ Add configuration management
- [x] **DONE** ✅ Security auditing
- [x] **DONE** ✅ Documentation completion

## Technical Specifications

### Database Schema
```sql
-- Documents table
CREATE TABLE knowledge_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    project_id UUID NOT NULL REFERENCES projects(id),
    filename VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    file_path TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'processing',
    error_message TEXT,
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
    created_at TIMESTAMP DEFAULT NOW()
);

-- Web scraping jobs table
CREATE TABLE knowledge_scraping_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    project_id UUID NOT NULL REFERENCES projects(id),
    url TEXT NOT NULL,
    max_depth INTEGER NOT NULL DEFAULT 5,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
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
    scraped_at TIMESTAMP DEFAULT NOW(),
    embedding vector(1536)
);

-- Indexes for performance
CREATE INDEX idx_knowledge_chunks_embedding ON knowledge_chunks USING ivfflat (embedding vector_cosine_ops);
CREATE INDEX idx_scraped_pages_embedding ON knowledge_scraped_pages USING ivfflat (embedding vector_cosine_ops);
CREATE INDEX idx_knowledge_documents_project ON knowledge_documents(project_id);
CREATE INDEX idx_scraping_jobs_project ON knowledge_scraping_jobs(project_id);
```

### API Endpoints
```
POST /v1/tenants/:tenant_id/projects/:project_id/knowledge/documents
GET /v1/tenants/:tenant_id/projects/:project_id/knowledge/documents
DELETE /v1/tenants/:tenant_id/projects/:project_id/knowledge/documents/:document_id

POST /v1/tenants/:tenant_id/projects/:project_id/knowledge/scrape
GET /v1/tenants/:tenant_id/projects/:project_id/knowledge/scraping-jobs
GET /v1/tenants/:tenant_id/projects/:project_id/knowledge/scraping-jobs/:job_id

GET /v1/tenants/:tenant_id/projects/:project_id/knowledge/search?q=query
```

### Configuration Extensions
```yaml
knowledge:
  enabled: true
  max_file_size: 10485760  # 10MB
  max_files_per_project: 100
  embedding_service: "openai"  # openai, sentence-transformers
  openai_embedding_model: "text-embedding-ada-002"
  chunk_size: 1000
  chunk_overlap: 200
  scraping:
    max_depth: 5
    rate_limit: 1000ms  # delay between requests
    user_agent: "TMS Knowledge Bot 1.0"
    timeout: 30s
```

## Current Status: ✅ **COMPLETE - READY FOR PRODUCTION**

**All core phases have been completed:**
1. ✅ Database Foundation with pgvector
2. ✅ Core Knowledge Services  
3. ✅ PDF Processing with chunking
4. ✅ Web Scraping with depth control
5. ✅ Embedding Integration with OpenAI
6. ✅ RAG Implementation with context injection
7. ✅ Frontend Integration with full UI
8. ✅ API Endpoints fully implemented
9. ✅ Testing & Optimization
10. ✅ Production Readiness

**The AI Knowledge Management System is now fully functional with:**
- Document upload and processing (PDF text extraction and chunking)
- Web scraping with configurable depth and rate limiting  
- Vector embeddings using OpenAI text-embedding-ada-002
- PostgreSQL with pgvector for similarity search
- RAG-enhanced AI chat responses with context injection
- Full REST API with security controls
- Comprehensive error handling and logging
- Production-ready configuration management

**Available Endpoints:**
- `POST /knowledge/documents` - Upload and process documents
- `GET /knowledge/documents` - List documents
- `DELETE /knowledge/documents/{id}` - Delete documents
- `POST /knowledge/scrape` - Create web scraping jobs
- `GET /knowledge/scraping-jobs` - List scraping jobs
- `GET /knowledge/search` - Search knowledge base
- Enhanced chat with automatic context injection

**Next Steps:** System is complete and ready for production deployment!

---

**Last Updated:** August 30, 2025
**Status:** ✅ Complete - Full-Stack AI Knowledge Management System Ready
