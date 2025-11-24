# TMS Integrations Documentation

Complete documentation for all third-party platform integrations.

## 📚 Available Guides

### Quick Start Guides
- **[Discord Quick Start](./DISCORD_QUICKSTART.md)** - Get Discord working in 5 minutes
- **[Slack Quick Start](./SLACK_QUICKSTART.md)** - Slack setup guide *(coming soon)*

### Detailed Guides
- **[Discord Integration](./DISCORD_INTEGRATION.md)** - Complete Discord integration guide
- **[Slack Integration](./SLACK_INTEGRATION.md)** - Full Slack setup *(coming soon)*
- **[General Integration Guide](./INTEGRATION_GUIDE.md)** - Adding new integrations

## 🚀 Supported Integrations

| Platform | Status | OAuth | Webhooks | Documentation |
|----------|--------|-------|----------|---------------|
| **Discord** | ✅ Ready | Yes | Optional | [Guide](./DISCORD_INTEGRATION.md) |
| **Slack** | ✅ Ready | Yes | Yes | [Guide](./SLACK_INTEGRATION.md) |
| **Microsoft Teams** | 🚧 Planned | Yes | Yes | - |
| **Google Chat** | 📋 Backlog | Yes | Yes | - |
| **Telegram** | 📋 Backlog | Yes | Yes | - |

## 🔧 Recent Fixes

### Discord Integration (2025-11-24)

**Issues Fixed:**
1. ✅ Missing environment variables in `.env`
2. ✅ Incorrect OAuth scopes (missing `identify` and `guilds`)
3. ✅ Guild information not being fetched after authorization

**Changes:**
- Updated `app/backend/.env` with Discord configuration
- Fixed OAuth scopes in `internal/service/integration_oauth.go`
- Added `fetchDiscordGuilds()` to retrieve server information
- Created comprehensive documentation

**Testing Status:**
- ✅ Code changes implemented
- ⏳ End-to-end testing pending (requires Discord app setup)

## 📖 Documentation Structure

```
docs/
├── INTEGRATIONS_README.md         # This file - Overview
├── DISCORD_QUICKSTART.md           # 5-minute Discord setup
├── DISCORD_INTEGRATION.md          # Detailed Discord guide
├── INTEGRATION_GUIDE.md            # How to add new integrations
└── SLACK_INTEGRATION.md            # Slack guide (coming soon)
```

## 🏗️ Architecture Overview

```
┌──────────────────┐
│  TMS Frontend    │
│ (Agent Console)  │
└────────┬─────────┘
         │ Install Integration
         ▼
┌──────────────────┐
│   TMS Backend    │
│  - Handlers      │◄───── OAuth Callback
│  - Services      │
│  - Repositories  │
└────────┬─────────┘
         │ Store Integration
         ▼
┌──────────────────┐
│   PostgreSQL     │
│ (project_        │
│  integrations)   │
└──────────────────┘

         │ OAuth Flow
         ▼
┌──────────────────┐
│ Discord / Slack  │
│ (OAuth Provider) │
└──────────────────┘
```

## 🔐 Security

All integrations follow these security practices:

- **OAuth 2.0**: Secure authorization with user consent
- **State Validation**: Redis-backed state tokens (10 min TTL)
- **Token Encryption**: Access tokens stored securely
- **HTTPS Required**: For production deployments
- **Scoped Permissions**: Minimal necessary access
- **Multi-Tenant Isolation**: Tenant-level data separation

## 🛠️ Setup Requirements

### Backend Requirements

1. **Environment Variables**:
   - `DISCORD_CLIENT_ID`
   - `DISCORD_CLIENT_SECRET`
   - `DISCORD_REDIRECT_URI`
   - Similar for other platforms

2. **Dependencies**:
   - PostgreSQL (database)
   - Redis (OAuth state management)
   - Go 1.24+ (backend runtime)

3. **Configuration**:
   - OAuth redirect URIs in provider portals
   - CORS settings for frontend domains
   - JWT secret for authentication

### Frontend Requirements

1. **Agent Console**:
   - Integration management UI
   - OAuth flow initiation
   - Status display

2. **API Client**:
   - Type-safe integration endpoints
   - Error handling
   - Loading states

## 📝 Adding a New Integration

Follow these steps to add support for a new platform:

1. **Review [Integration Guide](./INTEGRATION_GUIDE.md)**
2. **Define integration type** in models
3. **Add OAuth configuration** to config
4. **Implement OAuth flow**:
   - Generate OAuth URL
   - Exchange authorization code
   - Store access tokens
5. **Create callback handler**
6. **Register routes**
7. **Add environment variables**
8. **Write documentation**
9. **Test end-to-end**

**Estimated Time**: 2-4 hours for a standard OAuth 2.0 integration

## 🧪 Testing

### Manual Testing

```bash
# 1. Start backend
cd app/backend
go run cmd/api/main.go

# 2. Get OAuth URL
curl -X GET \
  'http://localhost:8080/v1/tenants/{tenant_id}/projects/{project_id}/integrations/discord/install' \
  -H 'Authorization: Bearer YOUR_JWT'

# 3. Visit OAuth URL in browser and authorize

# 4. Verify integration stored
psql -c "SELECT * FROM project_integrations WHERE integration_type='discord';"
```

### Automated Tests

```bash
# Run integration tests
cd app/backend
go test ./internal/service/... -v

# Test specific integration
go test ./internal/service/ -run TestDiscordOAuth -v
```

## 📊 Database Schema

```sql
-- Project integrations table
CREATE TABLE project_integrations (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    project_id UUID NOT NULL,
    integration_type VARCHAR(50) NOT NULL,
    meta JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(tenant_id, project_id, integration_type)
);

-- Indexes
CREATE INDEX idx_project_integrations_project ON project_integrations(tenant_id, project_id);
CREATE INDEX idx_project_integrations_type ON project_integrations(integration_type);
CREATE INDEX idx_project_integrations_status ON project_integrations(status);
CREATE INDEX idx_project_integrations_meta ON project_integrations USING GIN (meta);
```

## 🔄 OAuth Flow Diagram

```
┌─────────┐                                    ┌──────────┐
│  Agent  │                                    │ Provider │
│ Console │                                    │(Discord) │
└────┬────┘                                    └────┬─────┘
     │                                              │
     │ 1. Click "Install Discord"                  │
     │────────────────────────────────►            │
     │                                 │            │
     │ 2. GET /integrations/discord/   │            │
     │    install                      │            │
     │────────────────────────────────►│            │
     │                                 │            │
     │ 3. Return OAuth URL             │            │
     │◄────────────────────────────────│            │
     │                                 │            │
     │ 4. Redirect to Discord          │            │
     │──────────────────────────────────────────────►
     │                                 │            │
     │ 5. User authorizes              │            │
     │                                 │            │
     │ 6. Redirect with code & state   │            │
     │◄──────────────────────────────────────────────
     │                                 │            │
     │ 7. GET /integrations/discord/   │            │
     │    callback?code=...&state=...  │            │
     │────────────────────────────────►│            │
     │                                 │            │
     │                                 │ 8. Exchange│
     │                                 │    code    │
     │                                 │───────────►│
     │                                 │            │
     │                                 │ 9. Tokens  │
     │                                 │◄───────────│
     │                                 │            │
     │                                 │10. Store   │
     │                                 │    in DB   │
     │                                 │            │
     │ 11. Redirect to dashboard       │            │
     │     ?status=success             │            │
     │◄────────────────────────────────│            │
     │                                 │            │
```

## 🐛 Troubleshooting

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Invalid Client ID | Wrong credentials | Verify `.env` matches Discord portal |
| Invalid Redirect URI | Mismatch between code and portal | Ensure exact match (http/https) |
| State token expired | Took too long or Redis down | Retry OAuth flow, check Redis |
| Missing guild info | Insufficient scopes | Verify `guilds` scope is included |
| Integration not saving | Database error | Check logs, verify migrations ran |

### Debug Commands

```bash
# Check environment variables
env | grep DISCORD

# Test Redis connection
redis-cli ping

# Check database
psql $DATABASE_URL -c "SELECT COUNT(*) FROM project_integrations;"

# View backend logs
docker-compose logs backend | grep -i discord

# Test API endpoint
curl -v http://localhost:8080/api/public/integrations/discord/callback?code=test&state=test
```

## 📞 Support

For help with integrations:

1. **Check documentation** - Review the specific integration guide
2. **Search logs** - Backend logs contain detailed error messages
3. **Verify configuration** - Double-check all environment variables
4. **Test manually** - Use curl to test API endpoints directly
5. **Review code** - Check `internal/service/integration_oauth.go`

## 🎯 Roadmap

### Q1 2025
- ✅ Discord integration
- ✅ Slack integration
- 🚧 Microsoft Teams integration
- 📋 Integration management UI

### Q2 2025
- 📋 Google Chat integration
- 📋 Telegram integration
- 📋 WhatsApp Business API
- 📋 Webhook builder

### Q3 2025
- 📋 Zapier integration
- 📋 Make.com integration
- 📋 Custom webhook templates
- 📋 Integration marketplace

## 📄 License

These integration guides are part of the TMS project. See [LICENSE](../LICENSE) for details.

## 🤝 Contributing

To contribute integration guides:

1. Follow the template in `INTEGRATION_GUIDE.md`
2. Include setup steps, code examples, and troubleshooting
3. Test all instructions on a fresh setup
4. Submit a PR with your documentation

---

**Last Updated**: November 24, 2025  
**Maintained By**: TMS Development Team
