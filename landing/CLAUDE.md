# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

This project uses **pnpm** as the package manager. Common commands:

```bash
# Install dependencies
pnpm install

# Development server (Vite dev server)
pnpm vite dev

# Build for production
pnpm vite build

# Start production server
pnpm tsx server.ts
```

The production server runs on port 3344 by default.

## Architecture Overview

This is a **Ticket Management and AI Chat Platform** built with:

- **Frontend**: React 19 + Vite + TypeScript + Radix UI components
- **Backend**: Hono.js server with TypeScript
- **Database**: PostgreSQL with Kysely ORM
- **Deployment**: Built with Floot platform conventions

### Key Architecture Patterns

1. **File-based Routing**: Routes are generated from files in `pages/` directory with automatic layout composition
2. **API Endpoints**: Located in `endpoints/` directory with convention-based routing (`_api/` prefix)
3. **Component System**: Modular components in `components/` with CSS modules
4. **Database Schema**: Auto-generated types from Kysely codegen in `helpers/schema.tsx`

### Directory Structure

- `components/` - Reusable UI components with CSS modules
- `endpoints/` - API endpoint handlers (GET/POST methods by file)
- `pages/` - React Router pages with layout composition
- `helpers/` - Utilities including database schema, hooks, and configurations
- `server.ts` - Hono.js production server with API routing
- `App.tsx` - Main React app with routing configuration

### Database

The app uses PostgreSQL with three main tables:
- `tickets` - Ticket management (id, title, description, status, priority, timestamps)
- `chatSessions` - AI chat sessions (id, sessionId, createdAt)
- `chatMessages` - Chat messages (id, sessionId, message, role, createdAt)

Database connection is configured via `FLOOT_DATABASE_URL` in `env.json`.

### API Conventions

- All API routes use `_api/` prefix
- Endpoints follow REST conventions with file naming (e.g., `tickets_GET.ts`, `tickets_POST.ts`)
- Each endpoint exports a `handle` function that returns a Response object
- Dynamic imports are used for endpoint loading in the server

### Environment Configuration

Configure `env.json` with:
- `FLOOT_DATABASE_URL` - PostgreSQL connection string

Generate JWT secrets with:
```bash
node -e "console.log(require('crypto').randomBytes(32).toString('hex'))"
```