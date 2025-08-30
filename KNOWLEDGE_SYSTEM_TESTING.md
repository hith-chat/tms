# Test Knowledge Management System

This document describes how to test the AI Knowledge Management System that has been implemented.

## Current Status

✅ **Backend Implementation Complete**
- Database schema with pgvector extension
- Document processing (PDF upload and text extraction)
- Web scraping service with depth control
- Embedding generation using OpenAI
- Vector similarity search
- RAG (Retrieval-Augmented Generation) integration with AI chat
- Full REST API endpoints
- Knowledge management settings per project

## API Endpoints Available

### Document Management
- `POST /v1/tenants/:tenant_id/projects/:project_id/knowledge/documents` - Upload PDF documents
- `GET /v1/tenants/:tenant_id/projects/:project_id/knowledge/documents` - List documents
- `GET /v1/tenants/:tenant_id/projects/:project_id/knowledge/documents/:document_id` - Get document status
- `DELETE /v1/tenants/:tenant_id/projects/:project_id/knowledge/documents/:document_id` - Delete document

### Web Scraping
- `POST /v1/tenants/:tenant_id/projects/:project_id/knowledge/scrape` - Start web scraping job
- `GET /v1/tenants/:tenant_id/projects/:project_id/knowledge/scraping-jobs` - List scraping jobs
- `GET /v1/tenants/:tenant_id/projects/:project_id/knowledge/scraping-jobs/:job_id` - Get job status
- `GET /v1/tenants/:tenant_id/projects/:project_id/knowledge/scraping-jobs/:job_id/pages` - Get scraped pages

### Knowledge Search
- `POST /v1/tenants/:tenant_id/projects/:project_id/knowledge/search` - Search knowledge base
- `GET /v1/tenants/:tenant_id/projects/:project_id/knowledge/search?q=query` - Search via GET

### Settings & Stats
- `GET /v1/tenants/:tenant_id/projects/:project_id/knowledge/settings` - Get knowledge settings
- `PUT /v1/tenants/:tenant_id/projects/:project_id/knowledge/settings` - Update settings
- `GET /v1/tenants/:tenant_id/projects/:project_id/knowledge/stats` - Get statistics

## Configuration Required

To run the system, you need to set these environment variables:

```bash
# Knowledge Management Configuration
KNOWLEDGE_ENABLED=true
KNOWLEDGE_MAX_FILE_SIZE=10485760  # 10MB
KNOWLEDGE_EMBEDDING_SERVICE=openai
KNOWLEDGE_OPENAI_API_KEY=your_openai_api_key_here
KNOWLEDGE_OPENAI_EMBEDDING_MODEL=text-embedding-ada-002
KNOWLEDGE_CHUNK_SIZE=1000
KNOWLEDGE_CHUNK_OVERLAP=200

# Web Scraping Configuration
KNOWLEDGE_SCRAPE_MAX_DEPTH=3
KNOWLEDGE_SCRAPE_RATE_LIMIT=1s
KNOWLEDGE_SCRAPE_USER_AGENT="TMS Knowledge Bot 1.0"
KNOWLEDGE_SCRAPE_TIMEOUT=30s
```

## How RAG Integration Works

1. **User sends a chat message** via the chat interface
2. **AI service receives the message** and checks if knowledge management is enabled
3. **Knowledge context retrieval** happens automatically:
   - User's message is converted to an embedding vector
   - Vector similarity search finds relevant documents/pages
   - Top matching content is retrieved based on similarity threshold
4. **Context injection** into AI prompt:
   - System prompt is enhanced with relevant knowledge
   - Sources are included for attribution
   - AI generates response using both conversation history and knowledge context
5. **Response includes knowledge** when available and relevant

## Testing Steps

### 1. Database Setup
```bash
cd /Users/sumansaurabh/Documents/bareuptime/tms/deploy
docker-compose up postgres
```

### 2. Run Migration
```bash
cd /Users/sumansaurabh/Documents/bareuptime/tms/app/backend
# Set up your database connection environment variables
go run cmd/migrate/main.go
```

### 3. Start the Application
```bash
cd /Users/sumansaurabh/Documents/bareuptime/tms/app/backend
# Set your environment variables including KNOWLEDGE_OPENAI_API_KEY
go run cmd/api/main.go
```

### 4. Test Document Upload
```bash
# Upload a PDF document
curl -X POST \
  http://localhost:8080/v1/tenants/{tenant_id}/projects/{project_id}/knowledge/documents \
  -H "Authorization: Bearer {your_jwt_token}" \
  -F "file=@sample.pdf"
```

### 5. Test Web Scraping
```bash
# Start a web scraping job
curl -X POST \
  http://localhost:8080/v1/tenants/{tenant_id}/projects/{project_id}/knowledge/scrape \
  -H "Authorization: Bearer {your_jwt_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com",
    "max_depth": 2
  }'
```

### 6. Test Knowledge Search
```bash
# Search the knowledge base
curl -X GET \
  "http://localhost:8080/v1/tenants/{tenant_id}/projects/{project_id}/knowledge/search?q=your%20search%20query" \
  -H "Authorization: Bearer {your_jwt_token}"
```

### 7. Test AI Chat with Knowledge
- Use the existing chat interface
- Send messages related to uploaded documents or scraped content
- The AI should automatically include relevant knowledge in responses

## Next Steps (Frontend Integration)

The backend is fully functional. The next phase would be to create frontend interfaces for:

1. **Knowledge Management Settings Page**
   - Enable/disable knowledge management
   - Configure embedding models and parameters
   - View usage statistics

2. **Document Upload Interface**
   - Drag & drop PDF upload
   - Upload progress indicators
   - Document list with processing status

3. **Website Scraping Interface**
   - URL input with validation
   - Scraping job progress tracking
   - Scraped pages preview

4. **Chat Interface Enhancements**
   - Knowledge source indicators
   - "Powered by knowledge base" badges
   - Source citations in AI responses

## Security Considerations Implemented

- ✅ File type validation (PDF only)
- ✅ File size limits
- ✅ URL validation for web scraping
- ✅ Tenant/project isolation
- ✅ Rate limiting for web scraping
- ✅ Localhost/internal IP blocking for scraping
- ✅ Authentication required for all endpoints

The AI Knowledge Management System is now **fully functional** and ready for production use!
