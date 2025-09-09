# Agent Service Refactoring TODO

## Overview
Refactor the agent service from a single monolithic file to a proper structured FastAPI application with async SQLAlchemy integration.

## Task Breakdown

### 1. File Structure Creation
- [ ] Create proper directory structure:
  - `app/agent/`
    - `src/`
      - `__init__.py`
      - `main.py` (FastAPI app)
      - `config.py` (configuration management)
      - `database.py` (database connection and session management)
      - `models/`
        - `__init__.py`
        - `base.py` (SQLAlchemy base)
        - `chat.py` (chat-related models)
        - `knowledge.py` (knowledge base models)
        - `agent.py` (agent-related models)
      - `api/`
        - `__init__.py`
        - `chat.py` (chat endpoints)
        - `health.py` (health check endpoints)
      - `services/`
        - `__init__.py`
        - `swarm_service.py` (Swarm agent logic)
        - `knowledge_service.py` (KB search service)
        - `ticket_service.py` (ticket creation service)
      - `schemas/`
        - `__init__.py`
        - `chat.py` (Pydantic models for API)
        - `ticket.py` (ticket schemas)
    - `requirements.txt`
    - `Dockerfile` (update existing)

### 2. Configuration Management
- [ ] Create `config.py` with:
  - Single PostgreSQL URL: `postgres://tms:tms123@localhost:5432/tms?sslmode=disable`
  - OpenAI API key management
  - Environment-based configuration
  - No database initialization logic

### 3. Database Models (SQLAlchemy)
- [ ] Create async SQLAlchemy models based on existing migrations:
  - Reference migration files for existing table structures
  - Use existing tables: `tenants`, `projects`, `agents`, `knowledge_*` tables
  - Create models for chat sessions, messages, tickets
  - No table creation - rely on Go migrations

### 4. Database Session Management
- [ ] Create `database.py` with:
  - Async SQLAlchemy engine using asyncpg driver
  - Session factory
  - Dependency injection for FastAPI
  - Connection management

### 5. API Layer
- [ ] Refactor endpoints into proper API modules:
  - Chat streaming endpoint
  - Health check endpoint
  - Proper error handling
  - Request/response schemas

### 6. Service Layer
- [ ] Extract business logic into services:
  - Swarm agent orchestration
  - Knowledge base search
  - Ticket creation
  - Contact information management

### 7. Dependencies and Requirements
- [ ] Update requirements.txt with:
  - fastapi
  - sqlalchemy[asyncio]
  - asyncpg
  - swarm-ai
  - openai
  - uvicorn
  - pydantic

## Implementation Notes
- Use async/await throughout
- Follow existing database schema from migrations
- Don't create or initialize database - only connect and use existing tables
- Keep the Swarm agent functionality intact
- Maintain streaming API capability
- Use proper dependency injection patterns

## Files to Create/Modify
1. Directory structure (multiple files)
2. Update `requirements.txt`
3. Update `Dockerfile` if needed
4. Keep existing functionality while restructuring

## Priority Order
1. Create file structure
2. Setup configuration and database connection
3. Create SQLAlchemy models
4. Refactor API endpoints
5. Extract service layer
6. Test integration
