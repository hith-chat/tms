# Hith TMS - Quick Reference Guide

## Table of Contents

1. [Project Structure](#project-structure)
2. [Getting Started](#getting-started)
3. [Development Commands](#development-commands)
4. [API Endpoints](#api-endpoints)
5. [Database](#database)
6. [Configuration](#configuration)
7. [Common Tasks](#common-tasks)
8. [Testing](#testing)
9. [Deployment](#deployment)
10. [Troubleshooting](#troubleshooting)

---

## Project Structure

```
tms/
├── app/
│   ├── backend/              # Go API server
│   │   ├── cmd/api/         # Main entry point
│   │   ├── internal/        # Application code
│   │   │   ├── handlers/    # HTTP handlers
│   │   │   ├── service/     # Business logic
│   │   │   ├── repo/        # Data access
│   │   │   └── middleware/  # Middleware
│   │   ├── migrations/      # Database migrations
│   │   └── tests/           # Tests
│   │
│   ├── frontend/            # React monorepo (pnpm)
│   │   ├── agent-console/  # Agent dashboard (Port 3000)
│   │   ├── public-view/    # Customer view (Port 3002)
│   │   ├── chat-widget/    # Embeddable widget
│   │   └── shared/         # @tms/shared library
│   │
│   ├── ai-agent/           # Python AI service (Port 8090)
│   └── email-server/       # Go email-to-ticket (SMTP 2525)
│
├── deploy/                 # Docker Compose
├── manifests/              # Nomad deployment
├── tests/                  # Integration tests
├── scripts/                # Utility scripts
└── docs/                   # Documentation
```

---

## Getting Started

### Prerequisites

- **Go**: 1.24+
- **Node.js**: 18+
- **pnpm**: Latest
- **PostgreSQL**: 15+
- **Redis**: 7+
- **Docker** (optional)

### Quick Setup

```bash
# 1. Clone repository
git clone <repo-url>
cd tms

# 2. Start infrastructure (Docker)
cd deploy
docker-compose up -d postgres redis

# 3. Setup backend
cd ../app/backend
cp config.yaml.example config.yaml  # Edit with your settings
go mod download
go run cmd/migrate/main.go up      # Run migrations
go run cmd/api/main.go             # Start API (port 8080)

# 4. Setup frontend
cd ../frontend
pnpm install
pnpm --filter agent-console dev    # Port 5173
pnpm --filter public-view dev      # Port 5174

# 5. (Optional) AI Agent
cd ../ai-agent
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
uvicorn main:app --reload --port 8090
```

### Environment Variables

**Backend** (`app/backend/config.yaml`):
```yaml
database:
  url: postgres://tms:tms123@localhost:5432/tms

redis:
  url: redis://localhost:6379/0

jwt:
  secret: your-secret-key

ai:
  openai:
    api_key: sk-...
```

**Frontend** (`app/frontend/agent-console/.env`):
```env
VITE_API_URL=http://localhost:8080
```

---

## Development Commands

### Backend (Go)

```bash
# Navigate to backend
cd app/backend

# Run API server
go run cmd/api/main.go

# Run tests
go test ./...
go test -v ./internal/service/...

# Run specific test
go test -run TestTicketService_Create

# Build binary
go build -o tms-api cmd/api/main.go

# Run migrations
go run cmd/migrate/main.go up
go run cmd/migrate/main.go down
go run cmd/migrate/main.go status

# Generate Swagger docs
swag init -g cmd/api/main.go -o docs

# Format code
go fmt ./...

# Lint
golangci-lint run
```

### Frontend (pnpm)

```bash
# Navigate to frontend
cd app/frontend

# Install dependencies
pnpm install

# Run agent console (dev)
pnpm --filter agent-console dev         # Port 5173

# Run public view (dev)
pnpm --filter public-view dev           # Port 5174

# Build chat widget
pnpm --filter chat-widget build         # Creates UMD bundle

# Run tests
pnpm test
pnpm --filter agent-console test

# Type check
pnpm --filter agent-console typecheck

# Lint
pnpm --filter agent-console lint

# Format
pnpm --filter agent-console format

# Build for production
pnpm --filter agent-console build
```

### Docker

```bash
# Start all services
cd deploy
docker-compose up -d

# View logs
docker-compose logs -f backend

# Restart service
docker-compose restart backend

# Stop all
docker-compose down

# Rebuild image
docker-compose build backend
docker-compose up -d backend
```

---

## API Endpoints

### Base URLs

- **Local**: `http://localhost:8080`
- **Production**: `https://api.hith.chat`

### Authentication

**Login**
```bash
POST /v1/auth/login
Content-Type: application/json

{
  "email": "agent@example.com",
  "password": "password123"
}

# Response
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ...",
  "agent": {...}
}
```

**Use Token**
```bash
Authorization: Bearer <access_token>
```

### Common Endpoints

**Health Check**
```bash
GET /api/public/health
```

**Get Tickets**
```bash
GET /v1/tenants/{tenant_id}/projects/{project_id}/tickets
Authorization: Bearer <token>
```

**Create Ticket**
```bash
POST /v1/tenants/{tenant_id}/projects/{project_id}/tickets
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Issue with login",
  "description": "Cannot login to account",
  "customer_id": "cust_123",
  "priority": "high",
  "status": "open"
}
```

**Send Message to Ticket**
```bash
POST /v1/tenants/{tenant_id}/projects/{project_id}/tickets/{ticket_id}/messages
Authorization: Bearer <token>
Content-Type: application/json

{
  "content": "We're looking into this issue.",
  "is_internal": false
}
```

**Agent WebSocket**
```bash
ws://localhost:8080/v1/tenants/{tenant_id}/chat/agent/ws?token=<jwt_token>

# Once connected, send:
{
  "action": "subscribe",
  "session_id": "sess_123"
}
```

**Customer WebSocket** (Chat Widget)
```bash
ws://localhost:8080/api/public/chat/ws/widgets/{widget_id}/chat/{session_token}

# Send message:
{
  "type": "message",
  "content": "Hello, I need help"
}
```

### API Documentation

- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **Status**: See `docs/API_DOCUMENTATION_STATUS.md`

---

## Database

### Connection

```bash
# Local connection
psql postgres://tms:tms123@localhost:5432/tms

# Docker connection
docker exec -it tms-postgres psql -U tms -d tms
```

### Migrations

**Location**: `app/backend/migrations/`

**Run Migrations**
```bash
cd app/backend
go run cmd/migrate/main.go up      # Apply all
go run cmd/migrate/main.go down    # Rollback last
go run cmd/migrate/main.go status  # View status
```

**Create Migration**
```bash
# Create new migration file
touch migrations/032_add_new_feature.sql
```

**Migration Format** (Goose-style):
```sql
-- +goose Up
CREATE TABLE example (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- +goose Down
DROP TABLE example;
```

### Common Queries

**Find Tenant**
```sql
SELECT * FROM tenants WHERE id = 1;
```

**List Projects**
```sql
SELECT * FROM projects WHERE tenant_id = 1;
```

**Get Tickets**
```sql
SELECT t.*, c.name as customer_name
FROM tickets t
JOIN customers c ON t.customer_id = c.id
WHERE t.project_id = 1
ORDER BY t.created_at DESC
LIMIT 20;
```

**Vector Search** (Knowledge Base)
```sql
SELECT
    content,
    embedding <-> '[0.1, 0.2, ...]'::vector AS distance
FROM knowledge_chunks
WHERE project_id = 1
ORDER BY distance
LIMIT 5;
```

**Active Chat Sessions**
```sql
SELECT
    cs.*,
    cw.name as widget_name,
    c.name as customer_name
FROM chat_sessions cs
JOIN chat_widgets cw ON cs.widget_id = cw.id
LEFT JOIN customers c ON cs.customer_id = c.id
WHERE cs.status = 'active';
```

---

## Configuration

### Backend Config (`app/backend/config.yaml`)

```yaml
server:
  port: 8080
  environment: development  # development | staging | production

database:
  url: postgres://user:pass@host:5432/dbname
  max_open_conns: 25
  max_idle_conns: 5

redis:
  url: redis://localhost:6379/0

jwt:
  secret: your-jwt-secret-key
  access_token_expiry: 150m     # 2.5 hours
  refresh_token_expiry: 240h    # 10 days

features:
  enable_registration: true
  require_corporate_email: true

agentic:
  enabled: true
  greeting_detection: true
  knowledge_responses: true
  agent_assignment: true
  greeting_confidence: 0.4
  knowledge_confidence: 0.6
  agent_request_confidence: 0.7

ai:
  openai:
    api_key: sk-proj-...
    model: gpt-4
    embedding_model: text-embedding-ada-002
    max_tokens: 500

email:
  provider: resend  # resend | maileroo
  from_address: support@yourdomain.com
  resend:
    api_key: re_...
  maileroo:
    api_key: ...

cors:
  enabled: true
  origins:
    - http://localhost:5173
    - http://localhost:5174
```

### Redis Keys

```
# Rate limiting
rate_limit:{ip}:{endpoint}

# Session cache
session:{token}

# WebSocket pub/sub channels
chat:session:{session_id}
notifications:{agent_id}

# Cache
cache:project:{id}
cache:widget:{id}
```

---

## Common Tasks

### Create a New Tenant

```bash
# Via psql
INSERT INTO tenants (name, created_at)
VALUES ('New Company', NOW())
RETURNING id;

# Create first agent
INSERT INTO agents (tenant_id, email, name, password_hash, role, status)
VALUES (1, 'admin@newcompany.com', 'Admin', '$2a$...', 'tenant_admin', 'active');
```

### Add a New Project

```bash
curl -X POST http://localhost:8080/v1/tenants/1/projects \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Customer Support",
    "description": "Main support project"
  }'
```

### Generate Magic Link for Customer

```bash
curl -X POST http://localhost:8080/v1/tenants/1/public/tickets/request-access \
  -H "Content-Type: application/json" \
  -d '{
    "ticket_id": "ticket_123",
    "email": "customer@example.com"
  }'
```

### Test Chat Widget Locally

```html
<!-- Create test.html -->
<!DOCTYPE html>
<html>
<head>
  <title>Widget Test</title>
</head>
<body>
  <h1>Test Page</h1>

  <script>
    window.TMSChatConfig = {
      widgetId: "your_widget_id",
      domain: "localhost",
      apiUrl: "http://localhost:8080"
    };
  </script>
  <script src="http://localhost:5175/chat-widget.umd.js"></script>
</body>
</html>
```

### Scrape Website for Knowledge Base

```bash
curl -X POST http://localhost:8080/v1/tenants/1/projects/1/knowledge/scrape \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://docs.example.com",
    "depth": 2,
    "max_pages": 50
  }'
```

### Upload Document to Knowledge Base

```bash
curl -X POST http://localhost:8080/v1/tenants/1/projects/1/knowledge/documents \
  -H "Authorization: Bearer <token>" \
  -F "file=@manual.pdf" \
  -F "title=Product Manual"
```

---

## Testing

### Backend Tests

```bash
cd app/backend

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package
go test ./internal/service/...

# Run specific test
go test -run TestAuthService_Login ./internal/service/

# Verbose output
go test -v ./...

# Race detection
go test -race ./...
```

### Frontend Tests

```bash
cd app/frontend

# Run all tests
pnpm test

# Run specific app
pnpm --filter agent-console test

# Watch mode
pnpm --filter agent-console test -- --watch

# Coverage
pnpm --filter agent-console test -- --coverage

# UI mode
pnpm --filter agent-console test -- --ui
```

### Integration Tests

```bash
cd tests

# Run E2E tests
./run_e2e_tests.sh

# Run performance tests
./run_performance_tests.sh
```

### Manual Testing

```bash
# Generate test magic link
./scripts/test-magic-link.sh ticket_123 customer@example.com

# Create test data
./scripts/create-test-data.sh
```

---

## Deployment

### Build Docker Images

```bash
# Backend
docker build -t tms-backend:latest -f deploy/Dockerfile.backend .

# Frontend (agent console)
docker build -t tms-agent-console:latest -f deploy/Dockerfile.frontend .

# AI Agent
docker build -t tms-ai-agent:latest -f app/ai-agent/Dockerfile .
```

### Deploy to Nomad

```bash
# Via CI/CD (automatic on push to main)
git push origin main

# Manual deployment
nomad job run manifests/tms.nomad

# Check status
nomad job status tms

# View logs
nomad alloc logs <alloc-id>

# Restart
nomad job restart tms
```

### Environment-Specific Configs

**Development**: `config.yaml`
**Staging**: Environment variables in Nomad/Docker
**Production**: Vault/Secrets Manager

---

## Troubleshooting

### Backend Issues

**Database Connection Failed**
```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Test connection
psql postgres://tms:tms123@localhost:5432/tms

# Check logs
docker logs tms-postgres
```

**Redis Connection Failed**
```bash
# Check Redis is running
docker ps | grep redis

# Test connection
redis-cli -h localhost -p 6379 ping

# Check logs
docker logs tms-redis
```

**JWT Token Invalid**
```bash
# Verify JWT secret in config.yaml matches
# Check token expiry
# Ensure Bearer token format: "Bearer <token>"
```

**WebSocket Connection Drops**
```bash
# Check Redis pub/sub is working
redis-cli
> PSUBSCRIBE chat:*

# Verify agent is subscribed to session
# Check connection manager logs
```

### Frontend Issues

**API Connection Refused**
```bash
# Check VITE_API_URL in .env
# Verify backend is running on port 8080
# Check CORS configuration in backend
```

**Chat Widget Not Loading**
```bash
# Check widget ID is correct
# Verify domain is whitelisted in widget config
# Check browser console for errors
# Ensure widget build is latest: pnpm --filter chat-widget build
```

**Build Failures**
```bash
# Clear node_modules
rm -rf node_modules
pnpm install

# Clear pnpm cache
pnpm store prune

# Check TypeScript errors
pnpm typecheck
```

### Database Issues

**Migration Failed**
```bash
# Check current migration status
go run cmd/migrate/main.go status

# Rollback last migration
go run cmd/migrate/main.go down

# Fix migration file and re-run
go run cmd/migrate/main.go up
```

**Slow Queries**
```sql
-- Enable query logging
ALTER DATABASE tms SET log_statement = 'all';

-- Check slow queries
SELECT * FROM pg_stat_statements
ORDER BY total_exec_time DESC
LIMIT 10;

-- Add missing indexes
CREATE INDEX idx_tickets_project_id ON tickets(project_id);
```

### Performance Issues

**High Memory Usage**
```bash
# Check Go memory profile
go tool pprof http://localhost:8080/debug/pprof/heap

# Reduce database connection pool
# Check for memory leaks in WebSocket connections
```

**Slow API Responses**
```bash
# Enable request logging
# Check database query performance
# Verify Redis is caching properly
# Add indexes to frequently queried columns
```

---

## Useful Scripts

### Backend

```bash
# Generate JWT token for testing
go run scripts/generate_jwt.go

# Reset database
go run scripts/reset_db.go

# Seed test data
go run scripts/seed_data.go
```

### Database

```bash
# Backup database
pg_dump tms > backup.sql

# Restore database
psql tms < backup.sql

# Export specific table
pg_dump -t tickets tms > tickets.sql
```

### Logs

```bash
# View backend logs (Docker)
docker logs -f tms-backend

# View specific service logs (Nomad)
nomad alloc logs -f <alloc-id> tms-backend

# Search logs
docker logs tms-backend 2>&1 | grep ERROR
```

---

## Keyboard Shortcuts (Agent Console)

- `Ctrl/Cmd + K` - Quick command palette
- `Ctrl/Cmd + /` - Search tickets
- `Ctrl/Cmd + Enter` - Send message
- `Esc` - Close modal/dialog

---

## Support & Resources

- **Documentation**: `/docs`
- **API Docs**: `http://localhost:8080/swagger/index.html`
- **GitHub**: [Repository URL]
- **Issues**: [Issues URL]

---

## Quick Command Reference

```bash
# Backend
cd app/backend && go run cmd/api/main.go

# Frontend (Agent Console)
cd app/frontend && pnpm --filter agent-console dev

# Frontend (Public View)
cd app/frontend && pnpm --filter public-view dev

# Docker (All services)
cd deploy && docker-compose up -d

# Tests
go test ./...                              # Backend
pnpm test                                  # Frontend

# Migrations
go run cmd/migrate/main.go up              # Apply
go run cmd/migrate/main.go down            # Rollback

# Database
psql postgres://tms:tms123@localhost:5432/tms

# Redis
redis-cli -h localhost -p 6379
```

---

## Port Reference

| Service         | Port | URL                          |
|-----------------|------|------------------------------|
| Backend API     | 8080 | http://localhost:8080        |
| Agent Console   | 5173 | http://localhost:5173        |
| Public View     | 5174 | http://localhost:5174        |
| Chat Widget Dev | 5175 | http://localhost:5175        |
| AI Agent        | 8090 | http://localhost:8090        |
| PostgreSQL      | 5432 | postgres://localhost:5432    |
| Redis           | 6379 | redis://localhost:6379       |
| Email Server    | 2525 | SMTP on localhost:2525       |
| MailHog Web     | 8025 | http://localhost:8025        |

---

## File Locations

| Item                | Path                                    |
|---------------------|-----------------------------------------|
| Backend Config      | `app/backend/config.yaml`               |
| Database Migrations | `app/backend/migrations/`               |
| Email Templates     | `app/backend/templates/`                |
| Frontend Env        | `app/frontend/agent-console/.env`       |
| Docker Compose      | `deploy/docker-compose.yml`             |
| Nomad Manifest      | `manifests/tms.nomad`                   |
| Swagger Docs        | `app/backend/docs/`                     |
| Test Scripts        | `scripts/`                              |

---

## RBAC Roles Quick Reference

| Role           | Permissions                                    |
|----------------|------------------------------------------------|
| tenant_admin   | Full tenant access, manage all projects       |
| project_admin  | Manage project, agents, settings               |
| supervisor     | Assign tickets, view reports, manage agents    |
| agent          | Handle tickets, respond to chat, view KB       |
| read_only      | View-only access to tickets and conversations  |

---

## Status Values

**Ticket Status**: `open`, `in_progress`, `waiting_customer`, `resolved`, `closed`

**Priority**: `low`, `medium`, `high`, `urgent`

**Chat Session Status**: `active`, `assigned`, `waiting`, `ended`

**Agent Status**: `active`, `inactive`, `busy`, `away`

---

This quick reference should help you navigate and work with the Hith TMS codebase efficiently!
