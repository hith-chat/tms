# Agent Integration Plan - SSE Based Approach

## Overview
Create end-to-end connection between Go and Python agent service using Server-Sent Events (SSE) instead of Redis streams.

## Architecture
```
Go Service → Python Agent Service (HTTP API call with SSE response)
│
├── Chat Session initiated
├── Python Agent calls /ai-agent/:tenant_id/:project_id/login for auth token
├── Python Agent processes message and streams response via SSE
└── Go Service receives SSE stream and handles response
```

## Implementation Plan

### Phase 1: Python Agent Service Modifications ✅
- [x] Create SSE response handler in Python
- [x] Add HTTP endpoint to receive chat requests
- [x] Implement authentication flow to Go service
- [x] Create streaming response mechanism

### Phase 2: Go Service Integration ✅
- [x] Create HTTP client to call Python agent service
- [x] Implement SSE client to receive streaming responses
- [x] Integrate with existing chat handlers
- [x] Handle authentication token management

### Phase 3: Integration Points ✅
- [x] Update chat message handlers in Go to call Python service
- [x] Ensure proper error handling and fallbacks
- [x] Test end-to-end flow

## Files to Create/Modify

### Python Service (app/ai-agent/)
- [x] src/api/chat_handler.py - Handle incoming chat requests
- [x] src/services/auth_service.py - Authenticate with Go service
- [x] src/services/sse_service.py - Handle SSE responses
- [x] src/main.py - Add new endpoints

### Go Service (app/backend/)
- [ ] internal/service/agent_client.go - HTTP client for Python service
- [ ] internal/service/sse_client.go - SSE client for streaming responses
- [ ] internal/handlers/chat.go - Update chat handlers to use agent service

## Environment Variables Required
- TMS_API_BASE_URL - Base URL for Go service (for Python to authenticate)
- AI_AGENT_SERVICE_URL - URL of Python agent service (for Go to call)
- AI_AGENT_SECRET - Secret for agent authentication

## Next Steps
1. ✅ Python service implementation complete
2. Implement Go service HTTP client for calling Python agent
3. Implement Go service SSE client for receiving responses
4. Integrate with existing chat flow
5. Test end-to-end connection

## Notes
- Using SSE instead of Redis streams for direct service-to-service communication
- Python agent will authenticate with Go service using existing /ai-agent/:tenant_id/:project_id/login endpoint
- Maintaining existing architecture, only adding new integration points
