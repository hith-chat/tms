# TMS API OpenAPI Definition

This document provides a comprehensive OpenAPI definition for all handlers in the TMS (Ticket Management System) backend.

## Overview

The TMS API is organized into the following main categories:

### Authentication & Authorization
- JWT-based authentication
- API key authentication for external integrations
- Role-based access control (RBAC)
- Multi-tenant architecture

### Main Resource Categories

1. **Authentication** (`/v1/auth/*`)
2. **Tenants** (`/v1/tenants/*` and `/v1/enterprise/*`)
3. **Projects** (`/v1/tenants/{tenant_id}/projects/*`)
4. **Tickets** (`/v1/tenants/{tenant_id}/projects/{project_id}/tickets/*`)
5. **Agents** (`/v1/tenants/{tenant_id}/agents/*`)
6. **Customers** (`/v1/tenants/{tenant_id}/customers/*`)
7. **Chat System** (`/v1/tenants/{tenant_id}/projects/{project_id}/chat/*`)
8. **Knowledge Management** (`/v1/tenants/{tenant_id}/projects/{project_id}/knowledge/*`)
9. **Email Management** (`/v1/tenants/{tenant_id}/projects/{project_id}/email/*`)
10. **Notifications** (`/v1/tenants/{tenant_id}/projects/{project_id}/notifications/*`)
11. **Integrations** (`/v1/tenants/{tenant_id}/projects/{project_id}/integrations/*`)
12. **Settings** (`/v1/tenants/{tenant_id}/projects/{project_id}/settings/*`)
13. **Payments** (`/v1/payments/*`)
14. **Webhooks** (`/webhooks/*`)
15. **Public API** (`/api/public/*`)

## Current Swagger Documentation Status

### Fully Documented Handlers (with Swagger annotations):
✅ **auth.go** - Authentication endpoints
✅ **project.go** - Project management
✅ **knowledge.go** - Knowledge management (partially)
✅ **agent_websocket.go** - Agent WebSocket connection
✅ **ticket.go** - Ticket management (newly added)
✅ **customer.go** - Customer management (partially added)
✅ **agent.go** - Agent management (partially added)

### Handlers Missing Swagger Documentation:

#### Core Functionality
❌ **agent.go** - Remaining agent endpoints (CreateAgent, GetAgent, UpdateAgent, DeleteAgent, etc.)
❌ **customer.go** - Remaining customer endpoints (ListCustomers, DeleteCustomer)
❌ **tenant.go** - Tenant management
❌ **public.go** - Public ticket access endpoints
❌ **api_key.go** - API key management
❌ **settings.go** - Project settings management

#### Chat System
❌ **chat_widget.go** - Chat widget configuration
❌ **chat_session.go** - Chat session management
❌ **chat_websocket.go** - Customer WebSocket connections

#### Communication
❌ **email.go** - Email connector management
❌ **email_inbox.go** - Email inbox management
❌ **email_templates.go** - Email template management
❌ **domain_validation.go** - Domain validation

#### Advanced Features
❌ **notification_handler.go** - Notification management
❌ **integration.go** - Third-party integrations
❌ **alarm_handler.go** - Alarm system
❌ **ai_builder.go** - AI builder functionality
❌ **ai_usage.go** - AI usage tracking

#### Payment System
❌ **payment.go** - Payment processing
❌ **stripe_webhook.go** - Stripe webhook handling
❌ **cashfree_webhook.go** - Cashfree webhook handling

## Handler Documentation Priority

### High Priority (Core API functionality):
1. **ticket.go** - ✅ DONE - Core ticket management
2. **agent.go** - 🔄 IN PROGRESS - Agent management
3. **customer.go** - 🔄 IN PROGRESS - Customer management
4. **public.go** - Public API endpoints
5. **api_key.go** - API key management
6. **tenant.go** - Tenant management

### Medium Priority (Chat & Communication):
7. **chat_widget.go** - Chat widget APIs
8. **chat_session.go** - Chat session management
9. **email.go** - Email management
10. **settings.go** - Project settings

### Lower Priority (Advanced Features):
11. **notification_handler.go** - Notifications
12. **integration.go** - Integrations
13. **payment.go** - Payments
14. **alarm_handler.go** - Alarms
15. **ai_builder.go** - AI features

## Security Schemes

The API supports multiple authentication methods:

```yaml
securityDefinitions:
  BearerAuth:
    type: apiKey
    name: Authorization
    in: header
    description: JWT Bearer token (format: "Bearer {token}")
  
  ApiKeyAuth:
    type: apiKey
    name: X-API-Key
    in: header
    description: API key for external integrations
```

## WebSocket Endpoints

### Agent WebSocket
- **Endpoint**: `GET /v1/tenants/{tenant_id}/chat/agent/ws`
- **Purpose**: Real-time communication for agents
- **Authentication**: JWT Bearer token
- **Status**: ✅ Documented

### Customer WebSocket
- **Endpoint**: `GET /api/public/chat/ws/widgets/{widget_id}/chat/{session_token}`
- **Purpose**: Real-time communication for customers
- **Authentication**: Session token
- **Status**: ❌ Not documented

## Common Response Patterns

### Success Responses
- `200 OK` - Resource retrieved successfully
- `201 Created` - Resource created successfully
- `204 No Content` - Resource updated/deleted successfully

### Error Responses
- `400 Bad Request` - Invalid input data
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource already exists
- `422 Unprocessable Entity` - Validation failed
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

## Next Steps

To complete the OpenAPI documentation:

1. **Add remaining swagger annotations** to all handler methods
2. **Define request/response models** for all endpoints
3. **Document WebSocket message schemas** for chat system
4. **Add comprehensive examples** for all endpoints
5. **Generate complete OpenAPI 3.0 specification**
6. **Set up API documentation hosting** (e.g., Swagger UI)

## Generated Files

The swagger generation creates these files:
- `docs/docs.go` - Go source with embedded spec
- `docs/swagger.json` - JSON format specification
- `docs/swagger.yaml` - YAML format specification

Access the documentation at: `http://localhost:8080/swagger/index.html`