# TMS Agent Service

A READ-ONLY AI-powered customer support agent service built with FastAPI, OpenAI Agents SDK, and async SQLAlchemy.

## 🏗️ Architecture

The agent service is designed to be **READ-ONLY** - it queries the database for information but does not perform any CREATE, UPDATE, or DELETE operations. Instead, it returns structured JSON instructions that external systems can process.

### Directory Structure

```
app/agent/
├── src/
│   ├── __init__.py
│   ├── main.py                 # FastAPI application
│   ├── config.py              # Configuration management
│   ├── database.py            # Async SQLAlchemy setup
│   ├── models/                # SQLAlchemy models
│   │   ├── __init__.py
│   │   ├── base.py
│   │   ├── chat.py
│   │   ├── knowledge.py
│   │   └── agent.py
│   ├── schemas/               # Pydantic schemas
│   │   ├── __init__.py
│   │   ├── chat.py
│   │   └── ticket.py
│   ├── services/              # Business logic
│   │   ├── __init__.py
│   │   ├── chat_service.py
│   │   ├── knowledge_service.py
│   │   └── swarm_service.py
│   └── api/                   # API endpoints
│       ├── __init__.py
│       ├── chat.py
│       └── health.py
├── requirements.txt
├── Dockerfile
├── run.py                     # Development server
├── demo_agent.py             # Usage examples
└── TODO.md                   # Implementation checklist
```

## 🚀 Quick Start

### Prerequisites

1. PostgreSQL database with existing schema (managed by Go migrations)
2. OpenAI API key
3. Python 3.11+

### Environment Variables

```bash
# Required
AI_API_KEY=your_AI_API_KEY_here

# Optional (with defaults)
DATABASE_URL=postgresql+asyncpg://tms:tms123@localhost:5432/tms?sslmode=disable
HOST=0.0.0.0
PORT=5000
DEBUG=false
```

### Installation

```bash
# Install dependencies
pip install -r requirements.txt

# Run development server
python run.py

# Or with uvicorn directly
uvicorn src.main:app --host 0.0.0.0 --port 5000 --reload
```

### Docker

```bash
# Build image
docker build -t tms-agent .

# Run container
docker run -p 5000:5000 \
  -e DATABASE_URL="postgresql+asyncpg://tms:tms123@host.docker.internal:5432/tms?sslmode=disable" \
  -e AI_API_KEY="your_key_here" \
  tms-agent
```

## 📡 API Endpoints

### Chat

- `POST /chat/stream` - Streaming chat endpoint with Server-Sent Events
- `GET /chat/sessions/{session_id}` - Get session information
- `GET /chat/sessions/{session_id}/context` - Get agent context and pending actions
- `POST /chat/sessions/{session_id}/clear-context` - Clear session context

### Health

- `GET /health` - Health check with database connectivity
- `GET /ready` - Readiness check for Kubernetes

## 🤖 Agent Behavior

The agent is **READ-ONLY** and follows these principles:

1. **Knowledge Base Search**: Searches existing knowledge for answers
2. **Instruction Generation**: Creates JSON instructions for actions
3. **No Direct Operations**: Never performs database writes
4. **Structured Responses**: Returns actionable data for external systems

### Example Responses

#### Ticket Creation
```json
{
  "content": "I'll help you create a support ticket...",
  "pending_actions": {
    "ticket": {
      "action": "create_ticket",
      "data": {
        "title": "Account access issue",
        "description": "User unable to access account",
        "priority": "medium",
        "category": "general",
        "session_id": "session-123",
        "user_email": "user@example.com"
      }
    }
  }
}
```

#### Human Escalation
```json
{
  "content": "I'll escalate your case for priority handling...",
  "pending_actions": {
    "escalation": {
      "action": "escalate_to_human",
      "data": {
        "reason": "Complex technical issue",
        "session_id": "session-123",
        "priority": "high",
        "requires_immediate_attention": true
      }
    }
  }
}
```

## 🔧 Configuration

### Database Configuration

The service uses a single PostgreSQL connection URL:

```python
DATABASE_URL = "postgresql+asyncpg://user:pass@host:port/db?sslmode=disable"
```

### Database Schema

The service relies on existing database tables created by Go migrations:

- `chat_sessions` - Chat session data
- `messages` - Chat messages  
- `tickets` - Support tickets
- `knowledge_*` - Knowledge base tables
- `agents` - Agent information
- `alarms` - Alarm system

## 🧪 Testing

Run the demo to see example interactions:

```bash
python demo_agent.py
```

This shows how the agent responds with structured JSON for different scenarios.

## 📊 Monitoring

The service provides health endpoints for monitoring:

- Database connectivity checks
- OpenAI API configuration validation
- Service readiness status

## 🔒 Security Considerations

- No database write operations
- Environment-based configuration
- CORS configuration for API access
- Input validation with Pydantic

## 🛠️ Development

### Adding New Agent Functions

1. Create function in `SwarmService` class
2. Add to agent's `functions` list in `_setup_agents()`
3. Return structured JSON responses, not database operations

### Database Models

Models reflect existing database schema. To add new models:

1. Create model in appropriate file under `models/`
2. Import in `models/__init__.py`
3. Use for READ operations only

### API Endpoints

Add new endpoints in `api/` directory following existing patterns.

## 🚀 Deployment

The service is designed to be stateless and horizontally scalable:

- Use environment variables for configuration
- Database connection pooling built-in
- Health checks for load balancers
- Docker support for containerization

## 📝 License

[Add your license information here]
