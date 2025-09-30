# TMS API OpenAPI Definition Summary

This document summarizes the current OpenAPI definition for the agent_websocket.go handler and the overall API documentation status.

## Agent WebSocket Handler Documentation

### Endpoint: `/v1/tenants/{tenant_id}/chat/agent/ws`

**Method**: GET (WebSocket upgrade)  
**Summary**: Establish WebSocket connection for agent  
**Description**: Establishes a WebSocket connection for an agent to handle real-time chat across all sessions they have access to. The agent can send messages, typing indicators, and subscribe to specific chat sessions.

**Authentication**: Bearer JWT token required

**Parameters**:
- `tenant_id` (path, required): Tenant ID in UUID format

**Responses**:
- `101 Switching Protocols`: WebSocket connection established
- `400 Bad Request`: Invalid request
- `401 Unauthorized`: Invalid or missing JWT token
- `500 Internal Server Error`: Server error

**WebSocket Message Types**:

### Outbound Messages (Agent → Server)
- `chat_message`: Send a chat message to a customer
- `typing_start`: Indicate agent is typing
- `typing_stop`: Indicate agent stopped typing
- `ping`: Keep connection alive
- `session_subscribe`: Subscribe to updates for a specific session
- `session_unsubscribe`: Unsubscribe from session updates

### Inbound Messages (Server → Agent)
- `agent_connected`: Connection confirmation
- `pong`: Response to ping
- `chat_message`: New message in subscribed sessions
- `typing_start`/`typing_stop`: Typing indicators from customers
- `notification`: System notifications
- `error`: Error messages

### Message Schema

All WebSocket messages follow this structure:

```json
{
  "type": "string",
  "session_id": "uuid",
  "timestamp": "2023-09-29T10:00:00Z",
  "data": {},
  "from_type": "agent|customer|system"
}
```

### Example Messages

#### Agent Sends Chat Message
```json
{
  "type": "chat_message",
  "agent_session_id": "123e4567-e89b-12d3-a456-426614174000",
  "project_id": "987fcdeb-51a2-43d1-9f12-345678901234",
  "timestamp": "2023-09-29T10:00:00Z",
  "data": {
    "content": "Hello, how can I help you today?",
    "message_type": "text"
  }
}
```

#### Server Sends Connection Confirmation
```json
{
  "type": "agent_connected",
  "session_id": "123e4567-e89b-12d3-a456-426614174000",
  "timestamp": "2023-09-29T10:00:00Z",
  "from_type": "system",
  "data": {
    "type": "connected",
    "message": "Connected to console",
    "agent_id": "123e4567-e89b-12d3-a456-426614174000"
  }
}
```

## Currently Documented API Endpoints

The following 15 endpoints are now documented in the OpenAPI specification:

### Authentication (5 endpoints)
1. `POST /v1/auth/login` - User login
2. `POST /v1/auth/refresh` - Refresh JWT token
3. `POST /v1/auth/signup` - User registration
4. `POST /v1/auth/verify-signup-otp` - Verify signup OTP
5. `POST /v1/auth/resend-signup-otp` - Resend signup OTP
6. `POST /v1/auth/ai-agent/tenant/{tenant_id}/project/{project_id}/login` - AI agent login

### Projects (2 endpoints)
7. `GET /api/v1/tenants/{tenant_id}/projects` - List projects
8. `GET /api/v1/tenants/{tenant_id}/projects/{project_id}` - Get project details
9. `POST /api/v1/tenants/{tenant_id}/projects` - Create project
10. `PUT /api/v1/tenants/{tenant_id}/projects/{project_id}` - Update project
11. `DELETE /api/v1/tenants/{tenant_id}/projects/{project_id}` - Delete project

### Tickets (2 endpoints) 
12. `GET /v1/tenants/{tenant_id}/projects/{project_id}/tickets` - List tickets
13. `POST /v1/tenants/{tenant_id}/projects/{project_id}/tickets` - Create ticket
14. `GET /v1/tenants/{tenant_id}/projects/{project_id}/tickets/{ticket_id}` - Get ticket
15. `PATCH /v1/tenants/{tenant_id}/projects/{project_id}/tickets/{ticket_id}` - Update ticket

### Customers (2 endpoints)
16. `GET /v1/tenants/{tenant_id}/customers` - List customers
17. `POST /v1/tenants/{tenant_id}/customers` - Create customer
18. `PUT /v1/tenants/{tenant_id}/customers/{customer_id}` - Update customer

### Agents (1 endpoint)
19. `GET /v1/tenants/{tenant_id}/agents` - List agents

### WebSocket (1 endpoint)
20. `GET /v1/tenants/{tenant_id}/chat/agent/ws` - Agent WebSocket connection

### Public API (1 endpoint)
21. `POST /api/public/analyze-url` - Analyze public URL

## API Documentation Access

The Swagger UI is available at: `http://localhost:8080/swagger/index.html`

## Security Schemes

The API supports two authentication methods:

1. **Bearer Authentication**: JWT tokens for regular API access
2. **API Key Authentication**: For external integrations (not yet fully documented)

## Remaining Handlers to Document

The following handlers still need OpenAPI annotations:

### High Priority
- ❌ **public.go** - Public ticket access
- ❌ **api_key.go** - API key management  
- ❌ **tenant.go** - Tenant management
- ❌ **settings.go** - Project settings

### Medium Priority
- ❌ **chat_widget.go** - Chat widget management
- ❌ **chat_session.go** - Chat session endpoints
- ❌ **email.go** - Email connectors
- ❌ **notification_handler.go** - Notifications

### Lower Priority
- ❌ **integration.go** - Third-party integrations
- ❌ **payment.go** - Payment processing
- ❌ **alarm_handler.go** - Alarm system
- ❌ **ai_builder.go** - AI builder features

## Files Generated

The swagger generation process creates:
- `docs/docs.go` - Embedded Go specification
- `docs/swagger.json` - JSON format OpenAPI spec
- `docs/swagger.yaml` - YAML format OpenAPI spec
- `docs/agent_websocket_openapi.yaml` - Detailed WebSocket documentation

## WebSocket Documentation

The agent WebSocket handler is fully documented with:
- ✅ Endpoint specification
- ✅ Authentication requirements
- ✅ Message type definitions
- ✅ Request/response schemas
- ✅ Example messages
- ✅ Error handling

This provides comprehensive documentation for real-time agent communication functionality.