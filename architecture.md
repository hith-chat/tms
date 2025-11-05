# Hith TMS - System Architecture

## Overview

Hith is an **enterprise-grade, multi-tenant SaaS platform** for support ticketing and customer service management. The system combines traditional ticketing with real-time chat, AI-powered knowledge management, and email integration to provide a comprehensive customer support solution.

## Architecture Type

**Microservices with Clean Architecture**

- Multi-tenant SaaS platform
- Event-driven real-time communication
- AI-powered intelligent routing and responses
- Embeddable customer widgets

---

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                          CLIENT LAYER                                │
├─────────────────────────────────────────────────────────────────────┤
│  Agent Console    │  Public View    │  Chat Widget (Embeddable)     │
│  (React/TS)       │  (React/TS)     │  (UMD Bundle)                 │
└──────────┬────────┴────────┬────────┴────────┬────────────────────────┘
           │                 │                 │
           │                 │                 │
           ▼                 ▼                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        API GATEWAY LAYER                             │
├─────────────────────────────────────────────────────────────────────┤
│  Traefik Reverse Proxy + Load Balancer (SSL, Service Discovery)     │
└──────────┬──────────────────────────────────────────────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      APPLICATION LAYER                               │
├─────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌──────────────────┐  ┌──────────────────┐   │
│  │  Backend API    │  │  AI Agent        │  │  Email Server    │   │
│  │  (Go/Gin)       │  │  (Python/FastAPI)│  │  (Go/Guerrilla)  │   │
│  │  Port: 8080     │  │  Port: 8090      │  │  SMTP: 2525      │   │
│  └────────┬────────┘  └────────┬─────────┘  └────────┬─────────┘   │
│           │                    │                     │              │
└───────────┼────────────────────┼─────────────────────┼──────────────┘
            │                    │                     │
            ▼                    ▼                     ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        DATA LAYER                                    │
├─────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────┐  ┌───────────┐  ┌──────────────┐   │
│  │ PostgreSQL   │  │  Redis   │  │  MinIO    │  │  OpenAI API  │   │
│  │ + pgvector   │  │  Cache   │  │  Object   │  │  GPT-4       │   │
│  │              │  │  PubSub  │  │  Storage  │  │  Embeddings  │   │
│  └──────────────┘  └──────────┘  └───────────┘  └──────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Technology Stack

### Backend (Go 1.24+)

#### Core Framework
- **Web Framework**: Gin v1.11.0
- **Configuration**: Viper v1.20.1
- **Logging**: Zerolog v1.34.0
- **Validation**: go-playground/validator v10.27.0

#### Database & Storage
- **Database**: PostgreSQL 15 + pgvector v0.3.0
- **Database Driver**: lib/pq v1.10.9
- **Query Builder**: sqlx v1.3.5
- **Cache/PubSub**: Redis (go-redis v9.12.1)
- **Object Storage**: MinIO

#### Authentication & Security
- **JWT**: golang-jwt/jwt v5.2.0
- **Encryption**: Crypto utilities
- **RBAC**: Role-based access control

#### Real-Time Communication
- **WebSocket**: gorilla/websocket v1.5.3
- **PubSub**: Redis for message distribution

#### AI & Machine Learning
- **OpenAI Integration**: GPT-4
- **Vector Embeddings**: pgvector (1536 dimensions)
- **Tokenization**: tiktoken-go v0.1.8

#### Email Services
- **IMAP Client**: emersion/go-imap v1.2.1
- **Email Providers**:
  - Resend SDK v2.23.0
  - Maileroo SDK v1.0.0

#### Web Scraping
- **Browser Automation**: chromedp v0.11.2
- **Web Scraper**: Colly v2.2.0

#### Payments
- Stripe
- Cashfree

#### Testing
- testify v1.11.1
- sqlmock v1.5.2

### Frontend (React + TypeScript)

#### Framework
- **React**: 18.2.0
- **TypeScript**: 5.2.2
- **Build Tool**: Vite 5.3.4
- **Package Manager**: pnpm (monorepo)

#### UI Framework
- **UI Library**: Radix UI (20+ components)
- **Styling**: TailwindCSS 3.4.9
- **Utilities**: clsx, tailwind-merge
- **Icons**: lucide-react

#### State Management
- **API Client**: Axios 1.11.0
- **Data Fetching**: TanStack React Query 5.51.23
- **Forms**: react-hook-form 7.52.2 + Zod 3.23.8
- **State**: Zustand 4.5.7
- **Routing**: React Router DOM 6.26.0

#### Performance
- **Virtualization**: TanStack Virtual 3.13.12
- **Date Handling**: date-fns 3.6.0

#### Rich Content
- react-markdown 9.0.1
- react-syntax-highlighter 15.5.0
- recharts 2.12.7

#### Testing
- Vitest 3.2.4
- React Testing Library 16.3.0
- jsdom 26.1.0

### AI Agent Service (Python 3.11+)
- **Framework**: FastAPI
- **Database**: Async SQLAlchemy
- **AI**: OpenAI Agents SDK
- **Pattern**: READ-ONLY (returns JSON, no DB writes)

---

## Backend Architecture (Clean Architecture)

### Directory Structure

```
app/backend/
├── cmd/
│   ├── api/          # Main API server entry point
│   └── migrate/      # Database migration tool
├── internal/
│   ├── auth/         # JWT authentication
│   ├── config/       # Configuration management (Viper)
│   ├── crypto/       # Encryption utilities
│   ├── db/           # Database connection pool
│   ├── handlers/     # HTTP request handlers (27 files)
│   ├── http/         # HTTP utilities
│   ├── logger/       # Structured logging (zerolog)
│   ├── mail/         # Email service abstraction
│   ├── middleware/   # Gin middleware (Auth, RBAC, CORS, Rate Limit)
│   ├── models/       # Data models (6 files)
│   ├── observability/# Monitoring & tracing
│   ├── queue/        # Background job queue
│   ├── rate/         # Rate limiting
│   ├── rbac/         # Role-based access control
│   ├── redis/        # Redis client wrapper
│   ├── repo/         # Data repositories (21 files)
│   ├── search/       # Search functionality
│   ├── service/      # Business logic (71 services)
│   ├── store/        # Storage layer
│   ├── util/         # Utility functions
│   └── websocket/    # WebSocket connection manager
├── migrations/       # SQL migrations (31 migrations)
├── templates/        # Email HTML templates
├── docs/            # Swagger API documentation
└── tests/           # Unit & integration tests
```

### Layered Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP LAYER                               │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐            │
│  │ Middleware │→ │  Handlers  │→ │   Router   │            │
│  └────────────┘  └────────────┘  └────────────┘            │
│   • Auth (JWT)     • Input        • Gin Router             │
│   • RBAC           • Validation   • Route Groups           │
│   • CORS           • Error        • API Versioning         │
│   • Rate Limit       Handling                              │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                   SERVICE LAYER                             │
│  Business Logic (71 services)                               │
│  ┌────────────────┐ ┌────────────────┐ ┌─────────────────┐ │
│  │  AuthService   │ │ TicketService  │ │  AIService      │ │
│  │  ChatService   │ │ EmailService   │ │  PaymentService │ │
│  │  RBACService   │ │ KnowledgeServ. │ │  NotifService   │ │
│  └────────────────┘ └────────────────┘ └─────────────────┘ │
│   • Transaction management                                  │
│   • Business rules validation                               │
│   • Cross-entity operations                                 │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                  REPOSITORY LAYER                           │
│  Data Access (21 repositories)                              │
│  ┌────────────────┐ ┌────────────────┐ ┌─────────────────┐ │
│  │  TenantRepo    │ │  TicketRepo    │ │  AgentRepo      │ │
│  │  ProjectRepo   │ │  CustomerRepo  │ │  ChatRepo       │ │
│  │  MessageRepo   │ │  KnowledgeRepo │ │  EmailRepo      │ │
│  └────────────────┘ └────────────────┘ └─────────────────┘ │
│   • SQL query construction                                  │
│   • Database operations (CRUD)                              │
│   • Result mapping to models                                │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                    DATA LAYER                               │
│  ┌──────────────┐  ┌──────────┐  ┌──────────┐              │
│  │ PostgreSQL   │  │  Redis   │  │  MinIO   │              │
│  └──────────────┘  └──────────┘  └──────────┘              │
└─────────────────────────────────────────────────────────────┘
```

### Request Flow

```
HTTP Request → Middleware → Handler → Service → Repository → Database
                  ↓
            1. Auth (JWT)
            2. RBAC Check
            3. Rate Limit
            4. CORS
                  ↓
            Validate Input → Business Logic → Data Access → Response
```

---

## Multi-Tenancy Architecture

### Tenant Isolation

All resources are scoped by `tenant_id`:

```
Tenant (Organization)
  ├── Projects (Multiple)
  │   ├── Tickets
  │   ├── Agents (with roles)
  │   ├── Customers
  │   ├── Chat Widgets
  │   └── Knowledge Base
  └── Settings
```

### Route Structure

```
/v1/tenants/:tenant_id/
  /projects/:project_id/
    /tickets/:ticket_id/
      /messages
```

### Database-Level Isolation

- Every table has a `tenant_id` column
- All queries include `WHERE tenant_id = $1`
- Row-level security enforced at application layer
- No cross-tenant data leakage

---

## Real-Time Communication Architecture

### WebSocket Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                    AGENT WEBSOCKET                           │
│  /v1/tenants/{tenant_id}/chat/agent/ws                       │
│                                                               │
│  ┌────────┐    ┌────────┐    ┌────────┐    ┌────────┐       │
│  │ Agent1 │    │ Agent2 │    │ Agent3 │    │ Agent4 │       │
│  └───┬────┘    └───┬────┘    └───┬────┘    └───┬────┘       │
│      │             │             │             │             │
│      └─────────────┼─────────────┼─────────────┘             │
│                    │                                         │
│                    ▼                                         │
│         ┌──────────────────────┐                             │
│         │ ConnectionManager    │                             │
│         │  (in-memory pool)    │                             │
│         └──────────┬───────────┘                             │
│                    │                                         │
└────────────────────┼─────────────────────────────────────────┘
                     │
                     ▼
         ┌───────────────────────┐
         │   Redis Pub/Sub       │
         │  (message broker)     │
         └───────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────────┐
│                  CUSTOMER WEBSOCKET                          │
│  /api/public/chat/ws/widgets/{widget_id}/chat/{token}       │
│                                                               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                   │
│  │Customer1 │  │Customer2 │  │Customer3 │                   │
│  └──────────┘  └──────────┘  └──────────┘                   │
└──────────────────────────────────────────────────────────────┘
```

### Message Flow

1. **Customer sends message** → WebSocket → Backend
2. **Backend saves message** → PostgreSQL
3. **Backend publishes** → Redis (channel: `chat:session:{id}`)
4. **All instances subscribe** → Redis channel
5. **ConnectionManager routes** → Active agent connections
6. **Agent receives** → Real-time update

---

## AI & Knowledge Architecture

### AI-Powered Features

1. **Greeting Detection** (40% confidence threshold)
2. **Knowledge-Based Responses** (60% confidence threshold) - RAG
3. **Agent Request Detection** (70% threshold)
4. **Auto-Assignment** - Intelligent agent routing

### Vector Search

```sql
-- pgvector similarity search
SELECT *, embedding <-> $1 AS distance
FROM knowledge_chunks
WHERE project_id = $2
ORDER BY distance
LIMIT 5;
```

---

## Database Schema

### Core Entities

```sql
tenants → projects → tickets → ticket_messages
                  → agents → agent_project_roles
                  → customers → organizations
                  → chat_widgets → chat_sessions → chat_messages
                  → knowledge_documents → knowledge_chunks (vector embeddings)
```

---

## API Architecture

### API Versioning

```
/v1/                # Authenticated APIs
/api/public/        # Public (no auth) APIs
/webhooks/          # External webhooks
```

### Authentication Flow

- **Agents**: JWT (access: 150m, refresh: 240h)
- **Customers**: Magic links (1h expiry)
- **Widgets**: Session tokens (7 days)
- **Integrations**: API keys

### Rate Limiting

- Public: 5/minute
- Auth: 10/minute
- Authenticated: 100/minute
- Webhooks: 100/hour

---

## Deployment Architecture

### Infrastructure (Nomad + Docker)

```
Production (Falkenstein)
├── Traefik (Load Balancer + SSL)
├── Backend API (4 replicas, 1 CPU, 2GB each)
├── AI Agent (Python)
├── Email Server (Guerrilla)
├── PostgreSQL + pgvector
└── Redis
```

### CI/CD Pipeline

```
Push → Test → Build Docker → Push to GHCR → Deploy to Nomad
```

---

## Security

- **Authentication**: JWT, Magic Links, API Keys
- **Authorization**: RBAC (5 roles, granular permissions)
- **Encryption**: AES-256 for sensitive data
- **Rate Limiting**: Redis-based per-IP
- **HTTPS**: Enforced (Let's Encrypt)

---

## Performance

- Connection pooling (25 connections)
- Redis caching
- pgvector indexed search
- Frontend virtualization (5k+ items)
- Code splitting & lazy loading

---

## Key Features

1. **Multi-Tenant Support Ticketing**
2. **Real-Time Chat** (WebSockets + Redis PubSub)
3. **AI-Powered Knowledge Base** (RAG with pgvector)
4. **Email Integration** (IMAP polling, email-to-ticket)
5. **Embeddable Chat Widget** (UMD, 7-day sessions)
6. **Payment Integration** (Stripe, Cashfree)
7. **Credits System** (AI usage tracking)
8. **Advanced RBAC** (5 roles, fine-grained permissions)
9. **Notifications & Alarms** (Real-time alerts)
10. **AI Widget Theme Builder** (URL analysis)

---

## Conclusion

Hith TMS is a **production-ready, enterprise-grade SaaS platform** with clean architecture, real-time capabilities, AI integration, and scalable infrastructure designed for high availability and performance.
