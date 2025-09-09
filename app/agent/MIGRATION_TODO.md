# MIGRATION TODO: SwarmService to AgentService

## Overview
Migrating from deprecated OpenAI Swarm to OpenAI Agents Python SDK while preserving all existing functionality.

## ✅ COMPLETED
- [x] Fixed SQLAlchemy `metadata` reserved name conflict in models
- [x] Fixed SQLAlchemy database connection (text() wrapper for raw SQL)
- [x] Basic AgentService class structure
- [x] Updated Docker port configuration to use env vars
- [x] Updated imports and references

## 🚧 IN PROGRESS - CORE FEATURES TO MIGRATE

### 1. Agent System Architecture
- [ ] **Multiple Agent Types**: Support agent, Contact agent, Ticket agent
- [ ] **Agent Handoffs**: Proper handoff mechanism between agents  
- [ ] **Agent Instructions**: Migrate specific instructions for each agent
- [ ] **Agent Tools**: Recreate all function tools with proper schemas

### 2. Function Tools (HIGH PRIORITY)
- [ ] **search_knowledge_base**: Search knowledge base for relevant info
- [ ] **create_support_ticket**: Create real support tickets via TMS API
- [ ] **escalate_to_human**: Escalate cases to human agents
- [ ] **collect_contact_information**: Guide contact info collection
- [ ] **save_contact_info**: Save contact info via TMS API
- [ ] **update_contact_info**: Update existing contact information

### 3. Session Management
- [ ] **AgentSession**: In-memory session state management
- [ ] **Session Context**: Maintain context across messages
- [ ] **Temporary Data**: Store temp data during conversations
- [ ] **Session Cleanup**: Proper session cleanup and management

### 4. Service Dependencies
- [ ] **KnowledgeService**: Knowledge base search and formatting
- [ ] **ChatService**: Chat session management and message history
- [ ] **AgentAuthService**: Authentication for API calls
- [ ] **TMSApiClient**: API client for TMS backend integration

### 5. Streaming Response System
- [ ] **Message Streaming**: Stream responses using AsyncGenerator
- [ ] **SSE Format**: Proper Server-Sent Events formatting
- [ ] **Chunk Management**: Proper chunking of response content
- [ ] **Metadata Streaming**: Stream session metadata and context
- [ ] **Error Handling**: Proper error streaming and recovery

### 6. API Integration
- [ ] **Ticket Creation**: Real ticket creation via TMS API
- [ ] **Contact Info Updates**: Save/update contact info via API
- [ ] **Session Escalation**: Escalate sessions to human agents
- [ ] **Knowledge Search**: Search project knowledge base
- [ ] **Authentication**: Proper API authentication and authorization

### 7. Error Handling & Resilience
- [ ] **Graceful Degradation**: Fallback when API calls fail
- [ ] **Session Recovery**: Handle session state recovery
- [ ] **API Error Handling**: Proper API error responses
- [ ] **Logging**: Comprehensive logging for debugging

### 8. Missing Service Classes
- [ ] **KnowledgeService**: Create missing knowledge service
- [ ] **AgentAuthService**: Create missing auth service  
- [ ] **TMSApiClient**: Create missing API client
- [ ] **ApiClient**: Fix existing API client integration

## 🔧 TECHNICAL REQUIREMENTS

### Pydantic Models for Function Tools
- [ ] All function tools must use proper Pydantic models for parameters
- [ ] No `additionalProperties` in schemas (strict mode)
- [ ] Proper type hints and validation

### OpenAI Agents SDK Integration
- [ ] Use `Agent` class with proper instructions
- [ ] Use `FunctionTool` with proper input schemas
- [ ] Use `Runner.run()` for agent execution
- [ ] Handle agent handoffs properly

### Database Integration
- [ ] Proper async SQLAlchemy session handling
- [ ] Session info retrieval and management
- [ ] Message history and context storage

## 🚨 CRITICAL ISSUES TO FIX
1. **Function Tool Schema Error**: Pydantic strict schema validation failing
2. **Missing Service Dependencies**: KnowledgeService, AgentAuthService, TMSApiClient
3. **Streaming Implementation**: Need proper streaming with new SDK
4. **Agent Runner**: Need to implement proper Runner usage

## 📋 MIGRATION STRATEGY
1. Create all missing service dependencies
2. Implement proper function tools with Pydantic models
3. Set up agent system with handoffs
4. Implement streaming response system
5. Test end-to-end functionality
6. Clean up old code

## 🎯 SUCCESS CRITERIA
- [ ] All existing functionality preserved
- [ ] No Pydantic schema errors
- [ ] Proper streaming responses
- [ ] Real API integration working
- [ ] Agent handoffs working
- [ ] Session management working
- [ ] Error handling and fallbacks working
