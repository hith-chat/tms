# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

TMS (Ticket Management System) is a Go-based backend API server for a comprehensive customer support platform featuring:
- Multi-tenant ticket management
- Real-time chat with AI assistance
- Knowledge base management with vector embeddings
- Email inbox integration
- WebSocket-based agent and customer communication
- Payment processing with credits system
- Alarm/notification system

**Tech Stack**: Go 1.24+, Gin web framework, PostgreSQL with pgvector, Redis, JWT auth, WebSocket, OpenAI integration

## Development Commands

### Building & Running
- `make run` - Run the application directly
- `make dev` - Run with hot reload using Air (recommended for development)
- `make build` - Build binary to `bin/tms-backend`
- `go run cmd/api/main.go` - Run main server
- `go run cmd/migrate/main.go up` - Run database migrations

### Documentation & Testing
- `make docs` - Generate OpenAPI/Swagger documentation (requires swaggo/swag)
- `make install-swaggo` - Install Swagger generation tool
- `make test` - Run all tests
- Access Swagger UI at `http://localhost:8080/swagger/index.html` when server is running

### Code Quality
- `make format` - Format code with `go fmt`
- `make lint` - Run linter (requires golangci-lint)
- `make clean` - Clean build artifacts and docs

### Dependencies
- `make deps` - Install and tidy dependencies
- `go mod tidy` - Update go.mod/go.sum

## Architecture Overview

### Layered Architecture (Handler → Service → Repository)

The codebase follows a clean architecture with clear separation:

1. **Handlers** (`internal/handlers/`) - HTTP request handling, input validation, response formatting
   - Example: `auth.go`, `ticket.go`, `chat_widget.go`
   - WebSocket handlers: `agent_websocket.go`, `chat_websocket.go`

2. **Services** (`internal/service/`) - Business logic, orchestration, external API calls
   - Core services: `auth.go`, `ticket.go`, `chat_session.go`
   - AI services: `ai.go`, `ai_builder.go`, `document_processor.go`, `embedding.go`
   - Email: `email_provider.go`, `email_inbox.go`

3. **Repositories** (`internal/repo/`) - Database access layer
   - Implements interfaces defined in `interfaces.go`
   - Direct SQL queries using `sqlx` and `database/sql`

### Key Architectural Components

#### Multi-Tenancy
- All routes include `tenant_id` in path (`/v1/tenants/:tenant_id/...`)
- `middleware.TenantMiddleware()` validates tenant existence and adds to context
- RBAC service enforces tenant-level permissions

#### Authentication & Authorization
- JWT-based auth with access/refresh tokens (`internal/auth/`)
- API key support for external integrations
- Dual auth middleware: `AuthMiddleware` (JWT) and `ApiKeyOrJWTAuthMiddleware` (flexible)
- RBAC system (`internal/rbac/`) with roles: tenant_admin, project_admin, agent, viewer

#### Real-Time Communication
- **Agent WebSocket**: Global connection at `/v1/tenants/{tenant_id}/chat/agent/ws`
  - Agents subscribe to specific chat sessions
  - Managed by `websocket.ConnectionManager` with Redis pub/sub
- **Customer WebSocket**: Session-specific at `/api/public/chat/ws/widgets/{widget_id}/chat/{session_token}`
- Message types: `chat_message`, `typing_start`, `typing_stop`, `notification`, etc.

#### AI Integration
- OpenAI GPT-4 for chat responses (`internal/service/ai.go`)
- Vector embeddings with pgvector for RAG (Retrieval Augmented Generation)
- Knowledge base: documents + web scraping using Playwright
- Agentic behavior: greeting detection, agent request detection, auto-assignment
- Credits system tracks AI usage per tenant

#### Knowledge Management
- Document upload and processing (PDF support via `ledongthuc/pdf`)
- Web scraping with Playwright (`comprehensive_url_extractor.go`, `web_scraping.go`)
- Vector embeddings stored in PostgreSQL with pgvector extension
- Deduplication based on content hashes

#### Email Integration
- IMAP connector for inbox monitoring
- Email-to-ticket conversion
- Email provider abstraction supporting Resend and Maileroo
- Domain validation for custom email sending

### Database Schema

PostgreSQL database with key tables:
- **tenants** - Multi-tenant isolation root
- **agents** - Support staff with RBAC roles
- **customers** - End users creating tickets
- **projects** - Organizational units within tenants
- **tickets** - Support tickets with status workflow
- **ticket_messages** - Ticket conversation threads
- **chat_sessions** / **chat_messages** - Real-time chat system
- **chat_widgets** - Embeddable chat configurations
- **knowledge_base_documents** - Knowledge base entries with vector embeddings
- **scraped_pages** - Web-scraped content for knowledge base
- **alarms** - Notification/alert system
- **credits** / **credits_transactions** - AI usage billing

Migrations are in `migrations/*.sql` using Goose format (with `-- +goose Up/Down` markers).

### Configuration

Configuration via `config.yaml` (see `config.yaml.example`):
- Server, database, Redis, JWT settings
- Email providers (SMTP, Resend, Maileroo)
- AI configuration (OpenAI API key, models)
- Feature flags (agentic behavior, greeting detection, etc.)
- CORS, rate limiting

Loaded via Viper in `internal/config/config.go`.

## Important Development Patterns

### Adding New Endpoints

1. **Define route** in `setupRouter()` in `cmd/api/main.go`
2. **Create handler** in `internal/handlers/` with method signature like:
   ```go
   func (h *Handler) MethodName(c *gin.Context) { ... }
   ```
3. **Add Swagger annotations** above handler method (see existing examples in `auth.go`, `ticket.go`)
4. **Implement service logic** in `internal/service/`
5. **Add repository methods** in `internal/repo/` if database access needed
6. **Run** `make docs` to regenerate OpenAPI spec

### WebSocket Message Handling

Both agent and customer WebSocket handlers follow a pattern:
1. Upgrade HTTP connection to WebSocket
2. Authenticate (JWT for agents, session token for customers)
3. Register connection with `ConnectionManager`
4. Listen for Redis pub/sub messages and client messages concurrently
5. Route messages based on `type` field in JSON payload

### Middleware Stack

Global middleware (applied to all routes):
1. `ErrorHandlerMiddleware()` - Panic recovery and error handling
2. `TransactionLoggingMiddleware()` - Structured logging with zerolog
3. `RequestIDMiddleware()` - Unique request ID tracking
4. `CORSMiddleware()` - CORS headers
5. `TenantMiddleware()` - Tenant validation

Route-specific middleware:
- `AuthMiddleware()` - JWT validation (most authenticated routes)
- `ApiKeyOrJWTAuthMiddleware()` - Flexible auth for public-facing APIs
- `TenantAdminMiddleware()` - Restricts to tenant admins
- `ProjectAdminMiddleware()` - Restricts to project admins
- `TicketAccessMiddleware()` - Ticket-level RBAC
- Rate limiting: `AuthRateLimit()`, `PublicAPIRateLimit()`, `StrictRateLimit()`

### RBAC Permission Checks

Services use `rbacService` to check permissions:
```go
err := h.rbacService.CheckTenantPermission(ctx, agentID, tenantID, rbac.PermissionReadTickets)
if err != nil {
    return err
}
```

Common permissions: `PermissionReadTickets`, `PermissionWriteTickets`, `PermissionManageAgents`, etc.

## API Documentation Status

OpenAPI/Swagger documentation is partially complete. See `docs/API_DOCUMENTATION_STATUS.md` for current status.

**Fully documented**: auth, projects, tickets, knowledge (partial), agent websocket, customers (partial), agents (partial)

**Missing documentation**: tenant management, public API, api_key management, settings, chat widgets, chat sessions, email handlers, notifications, integrations, payments, alarms, AI builder

When adding documentation, use swaggo annotations format. Run `make docs` after changes.

## Testing Approach

- Unit tests should be in `*_test.go` files alongside source
- Integration tests in `tests/` directory
- Test files in project root (`test_*.go`) appear to be integration/e2e tests
- Use `testify` for assertions (`github.com/stretchr/testify`)

## Payment & Credits System

- Stripe and Cashfree webhook handlers for payment processing
- Credits system (`credits`, `credits_transactions` tables) tracks AI usage
- `AIUsageService` deducts credits when AI features are used
- Payment session creation with automatic gateway selection based on IP geolocation

## External Dependencies

- **Playwright** (headless browser) - installed in Docker containers for web scraping
- **Redis** - pub/sub for WebSocket messaging, caching, rate limiting
- **PostgreSQL with pgvector** - main database with vector similarity search
- **OpenAI API** - GPT-4 for chat and embeddings
- **Email providers** - Resend, Maileroo for transactional emails
- **Payment gateways** - Stripe, Cashfree

## Docker Development

- `Dockerfile` - Production build (multi-stage with Playwright)
- `Dockerfile.dev` - Development with Air hot reload and Playwright
- Uses Playwright Docker image as base for browser automation

## Notes for Claude

- The project uses structured logging with `zerolog` - prefer `logger.Info()`, `logger.Error()` over `fmt.Println()`
- Always check RBAC permissions in service layer, not just middleware
- WebSocket connection management is centralized in `websocket.ConnectionManager`
- AI responses should check credit balance before processing
- Email operations require validated domains for sending
- The project is actively being documented - add Swagger annotations when creating new endpoints