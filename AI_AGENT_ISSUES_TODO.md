# AI Agent Service Issues - TODO List

## Critical Issues Found

### 1. **Knowledge Service Method Issue**
- **File**: `app/ai-agent/src/services/agent_service.py`
- **Issue**: Calling `self.knowledge_service.search()` which doesn't exist
- **Correct Method**: `self.knowledge_service.search_knowledge_base()`
- **Line**: ~107
- **Fix**: Update method call and parameters


### 3. **Database Session Context Issue**
- **File**: `app/ai-agent/src/services/agent_service.py`
- **Issue**: `_current_db_session` is never set, always `None`
- **Lines**: Multiple locations checking for `db_session`
- **Fix**: Pass database session through function calls or initialize properly

### 4. **Auth Token Flow Issue**
- **File**: `app/ai-agent/src/services/agent_service.py`
- **Issue**: Auth token is passed but not properly used in TMS API calls
- **Problem**: `_current_auth_token` is set but never accessed by API client
- **Fix**: Pass auth token to TMS API client methods

### 5. **Session Authentication Context Issue**
- **File**: `app/ai-agent/src/services/agent_service.py`
- **Issue**: Calling `auth_service.get_session_auth_context()` which may not exist
- **Lines**: ~143, 184, 246
- **Fix**: Verify method exists or create proper auth context retrieval

### 6. **Missing Database Session Injection**
- **File**: `app/ai-agent/src/api/chat_handler.py`
- **Issue**: Database session is never created or passed to agent service
- **Fix**: Use dependency injection to get database session

### 7. **Agent Service Constructor Issue**
- **File**: `app/ai-agent/src/services/agent_service.py`
- **Issue**: `TMSApiClient(self.auth_service)` may not match constructor signature
- **Fix**: Verify TMSApiClient constructor parameters

### 8. **OpenAI API Key Configuration Issue**  
- **File**: `app/ai-agent/src/services/agent_service.py`
- **Issue**: OpenAI client not initialized with API key for Runner
- **Error**: "api_key client option must be set"
- **Fix**: Initialize AsyncOpenAI client properly for Runner

### 9. **Response Processing Issue**
- **File**: `app/ai-agent/src/api/chat_handler.py`
- **Issue**: Trying to call `.dict()` on response object, but should be `.model_dump()`
- **Line**: 79 (in `_format_sse_message`)
- **Fix**: Use proper Pydantic v2 method

### 10. **Import Issues**
- **File**: `app/ai-agent/src/services/agent_service.py`
- **Issue**: Incorrect imports for OpenAI Agents SDK
- **Fix**: Use correct import: `from agents import function_tool`

## Required Fixes - Priority Order

### High Priority (Blocking functionality)

1. **Fix OpenAI Agents SDK Tool Definition**
   - Replace `FunctionTool` with `@function_tool` decorator
   - Update all tool implementations
   - Fix import statements

2. **Fix Knowledge Service Method Call**
   - Change `search()` to `search_knowledge_base()`
   - Update parameters to match method signature

3. **Fix Database Session Context**
   - Add database session dependency injection in chat handler
   - Pass database session to agent service methods

4. **Fix Auth Token Flow**
   - Pass auth token from chat handler to agent service
   - Use auth token in TMS API client calls

5. **Fix OpenAI Client Configuration**
   - Initialize AsyncOpenAI client with API key for Runner
   - Set proper environment variables

### Medium Priority

6. **Fix Response Processing**
   - Use `.model_dump()` instead of `.dict()`
   - Handle response object types properly

7. **Fix API Client Constructor**
   - Verify TMSApiClient parameters match constructor
   - Update if needed

8. **Fix Session Auth Context**
   - Verify `get_session_auth_context()` method exists
   - Create proper auth context retrieval mechanism

### Low Priority (Performance/cleanup)

9. **Clean up Error Handling**
   - Add proper exception handling for all service calls
   - Log errors consistently

10. **Add Type Hints**
    - Complete type hints for all methods
    - Add proper return type annotations

## Reference Files to Study

### Go Backend API Structure
- `app/backend/internal/handlers/ticket.go` - For ticket creation API
- `app/backend/internal/handlers/alarm_handler.go` - For alarm/escalation API  
- `app/backend/internal/repo/knowledge.go` - For knowledge base operations
- `app/backend/internal/service/` - For service layer patterns

### Python Service Structure
- `app/ai-agent/src/services/knowledge_service.py` - For proper method calls
- `app/ai-agent/src/services/tms_api_client.py` - For API client usage
- `app/ai-agent/src/services/auth_service.py` - For authentication flow
- `app/ai-agent/src/database.py` - For database session management

### OpenAI Agents SDK Examples
- Use `@function_tool` decorator pattern
- Import from `agents` package
- Use `Runner.run()` for async execution

## Next Steps

1. Start with High Priority fixes first
2. Test each fix incrementally
3. Verify authentication flow works end-to-end
4. Test knowledge base search functionality
5. Test ticket creation and escalation workflows
6. Add comprehensive error handling
7. Add logging for debugging

## Notes

- The OpenAI Agents SDK syntax used is mostly correct, but tool definition needs decorator pattern
- Auth token flow exists but needs proper context passing
- Database connectivity needs proper dependency injection
- TMS API endpoints exist in Go backend and need proper calling from Python
