# Discord Integration Guide

This guide explains how to integrate Discord with your TMS (Ticket Management System) to enable automated notifications and bot interactions.

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Discord Application Setup](#discord-application-setup)
4. [Backend Configuration](#backend-configuration)
5. [OAuth Flow](#oauth-flow)
6. [Integration Features](#integration-features)
7. [API Reference](#api-reference)
8. [Troubleshooting](#troubleshooting)

---

## Overview

The Discord integration allows you to:
- Connect your TMS projects to Discord servers
- Send automated notifications to Discord channels
- Use Discord bots to interact with your ticket system
- Manage customer support tickets via Discord

## Prerequisites

- A Discord account
- Admin access to a Discord server (or ability to create one)
- TMS backend running (default: `http://localhost:8080`)
- TMS frontend running (default: `http://localhost:3000`)

---

## Discord Application Setup

### Step 1: Create a Discord Application

1. Go to the [Discord Developer Portal](https://discord.com/developers/applications)
2. Click **"New Application"**
3. Enter your application name (e.g., "TMS Bot")
4. Click **"Create"**

### Step 2: Get Your Client Credentials

1. In your application, navigate to the **"OAuth2"** section in the left sidebar
2. Copy the **Client ID** - you'll need this for `DISCORD_CLIENT_ID`
3. Click **"Reset Secret"** (or reveal if it's your first time)
4. Copy the **Client Secret** - you'll need this for `DISCORD_CLIENT_SECRET`

⚠️ **Important**: Never share your Client Secret publicly or commit it to version control.

### Step 3: Configure OAuth2 Redirects

1. Still in the **"OAuth2"** section, scroll to **"Redirects"**
2. Click **"Add Redirect"**
3. Add your callback URL:
   - For development: `http://localhost:8080/api/public/integrations/discord/callback`
   - For production: `https://your-domain.com/api/public/integrations/discord/callback`
4. Click **"Save Changes"**

### Step 4: Configure Bot Settings

1. Navigate to the **"Bot"** section in the left sidebar
2. Click **"Add Bot"** if you haven't already
3. Customize your bot:
   - **Username**: Set a friendly name (e.g., "TMS Support Bot")
   - **Icon**: Upload a bot avatar (optional)
4. Under **"Privileged Gateway Intents"**, enable:
   - ✅ Server Members Intent (if you need to read member info)
   - ✅ Message Content Intent (if you need to read message contents)

### Step 5: Set Bot Permissions

The TMS integration requires the following permissions:
- `Send Messages` - To post notifications
- `Send Messages in Threads` - To post in threads
- `Read Message History` - To read context
- `Use Slash Commands` - For bot commands (future)

These are automatically configured in the OAuth URL with permissions code: `2147485696`

---

## Backend Configuration

### Environment Variables

Update your `.env` file in the backend directory (`app/backend/.env`):

```bash
# Discord OAuth Configuration
DISCORD_CLIENT_ID="your_client_id_here"
DISCORD_CLIENT_SECRET="your_client_secret_here"
DISCORD_REDIRECT_URI="http://localhost:8080/api/public/integrations/discord/callback"
```

**For Production:**
```bash
DISCORD_CLIENT_ID="your_client_id_here"
DISCORD_CLIENT_SECRET="your_client_secret_here"
DISCORD_REDIRECT_URI="https://api.yourdomain.com/api/public/integrations/discord/callback"
```

### Configuration File (Alternative)

You can also configure Discord in `app/backend/config.yaml`:

```yaml
discord:
  client_id: "your_client_id_here"
  client_secret: "your_client_secret_here"
  redirect_uri: "http://localhost:8080/api/public/integrations/discord/callback"
```

### Restart Backend

After updating configuration:

```bash
cd app/backend
# If using docker-compose
docker-compose restart backend

# If running locally
go run cmd/api/main.go
```

---

## OAuth Flow

### Step 1: Initiate Installation

From your TMS frontend (Agent Console):

1. Navigate to **Settings** → **Integrations**
2. Find the **Discord** integration card
3. Click **"Install"** or **"Connect"**

**API Endpoint:**
```http
GET /v1/tenants/{tenant_id}/projects/{project_id}/integrations/discord/install
Authorization: Bearer {your_jwt_token}
```

**Response:**
```json
{
  "oauth_url": "https://discord.com/api/oauth2/authorize?client_id=...&scope=bot+identify+guilds&permissions=2147485696&redirect_uri=...&response_type=code&state=..."
}
```

### Step 2: Authorize on Discord

1. You'll be redirected to Discord's authorization page
2. **Select a server** where you want to add the bot
3. Review the requested permissions
4. Click **"Authorize"**

### Step 3: Bot is Added

Discord will redirect back to your TMS backend callback URL with:
- Authorization code
- State token (for security)
- Guild ID (server ID) - if applicable

### Step 4: Token Exchange

The backend automatically:
1. Validates the state token
2. Exchanges the authorization code for access tokens
3. Fetches guild information
4. Stores the integration in the database

### Step 5: Confirmation

You'll be redirected to:
```
http://localhost:3000/dashboard?integration=discord&status=success
```

---

## Integration Features

### Stored Information

The integration stores the following metadata:

```json
{
  "access_token": "encrypted_access_token",
  "refresh_token": "encrypted_refresh_token",
  "token_type": "Bearer",
  "expires_in": 604800,
  "scope": "bot identify guilds",
  "guild_id": "123456789012345678",
  "guild_name": "My Awesome Server",
  "guild_icon": "a_1234567890abcdef",
  "bot_user_id": "987654321098765432",
  "installed_by_agent_id": "agent-uuid",
  "installed_at": "2025-11-24T10:00:00Z",
  "last_updated_at": "2025-11-24T10:00:00Z"
}
```

### Webhook Support (Optional)

If configured with webhook scopes:
```json
{
  "webhook_id": "webhook-id",
  "webhook_token": "webhook-token",
  "webhook_url": "https://discord.com/api/webhooks/...",
  "channel_id": "channel-id",
  "channel_name": "#general"
}
```

---

## API Reference

### List Project Integrations

Get all integrations for a project:

```http
GET /v1/tenants/{tenant_id}/projects/{project_id}/integrations/project
Authorization: Bearer {jwt_token}
```

**Response:**
```json
[
  {
    "id": "integration-uuid",
    "tenant_id": "tenant-uuid",
    "project_id": "project-uuid",
    "integration_type": "discord",
    "meta": {
      "guild_name": "My Server",
      "guild_id": "123456789012345678"
    },
    "status": "active",
    "created_at": "2025-11-24T10:00:00Z",
    "updated_at": "2025-11-24T10:00:00Z"
  }
]
```

### Install Integration

Initiate OAuth flow:

```http
GET /v1/tenants/{tenant_id}/projects/{project_id}/integrations/discord/install
Authorization: Bearer {jwt_token}
```

### Delete Integration

Remove Discord integration:

```http
DELETE /v1/tenants/{tenant_id}/projects/{project_id}/integrations/project/discord
Authorization: Bearer {jwt_token}
```

**Response:**
```json
{
  "message": "integration deleted successfully"
}
```

### OAuth Callback (Public)

This endpoint is called by Discord (not by your frontend):

```http
GET /api/public/integrations/discord/callback?code={code}&state={state}&guild_id={guild_id}
```

---

## Troubleshooting

### Common Issues

#### 1. "Invalid Client ID" Error

**Cause**: Client ID is incorrect or not set.

**Solution**:
- Verify `DISCORD_CLIENT_ID` in your `.env` file
- Ensure no extra spaces or quotes
- Restart backend after changes

#### 2. "Invalid Redirect URI" Error

**Cause**: The redirect URI in your code doesn't match Discord's configuration.

**Solution**:
- Go to Discord Developer Portal → Your App → OAuth2 → Redirects
- Ensure the callback URL exactly matches (including http/https)
- For local development: `http://localhost:8080/api/public/integrations/discord/callback`

#### 3. "Invalid State Token" Error

**Cause**: State token expired or Redis connection issue.

**Solution**:
- State tokens expire after 10 minutes
- Check Redis is running and accessible
- Try the OAuth flow again

#### 4. "Missing Permissions" Error

**Cause**: Bot doesn't have required permissions in the server.

**Solution**:
- Go to your Discord server
- Right-click the bot → **"Edit Server Profile"** → **Roles**
- Ensure bot role has required permissions
- Or re-invite the bot with correct permission code

#### 5. Bot Not Showing in Server

**Cause**: Bot wasn't added during OAuth flow.

**Solution**:
- The `bot` scope must be included in the OAuth URL
- Re-run the installation process
- Make sure you select a server during authorization

### Debug Tips

#### Enable Debug Logging

In your backend logs, look for:
```
Successfully exchanged Discord OAuth code
Successfully stored Discord integration
```

#### Check Database

Verify integration was stored:
```sql
SELECT * FROM project_integrations 
WHERE integration_type = 'discord' 
AND project_id = 'your-project-id';
```

#### Test API Endpoints

```bash
# Get OAuth URL
curl -X GET \
  'http://localhost:8080/v1/tenants/{tenant_id}/projects/{project_id}/integrations/discord/install' \
  -H 'Authorization: Bearer YOUR_JWT_TOKEN'

# List integrations
curl -X GET \
  'http://localhost:8080/v1/tenants/{tenant_id}/projects/{project_id}/integrations/project' \
  -H 'Authorization: Bearer YOUR_JWT_TOKEN'
```

### Getting Help

If you continue to experience issues:

1. Check backend logs for detailed error messages
2. Verify all environment variables are set correctly
3. Ensure Redis is running (required for OAuth state)
4. Confirm Discord application settings in Developer Portal
5. Check the [Discord API Documentation](https://discord.com/developers/docs/intro)

---

## Security Best Practices

1. **Never expose secrets**: Keep `DISCORD_CLIENT_SECRET` secure
2. **Use HTTPS in production**: Always use `https://` for redirect URIs
3. **Rotate tokens**: Periodically regenerate your client secret
4. **Limit bot permissions**: Only grant permissions your bot needs
5. **Validate state tokens**: The system does this automatically
6. **Encrypt stored tokens**: Access tokens are stored in encrypted form

---

## Next Steps

After setting up Discord integration:

1. Configure notification rules in TMS
2. Set up ticket-to-channel mappings
3. Test sending notifications to Discord
4. Customize bot responses and commands
5. Monitor integration health in the dashboard

For more integrations, see:
- [Slack Integration Guide](./SLACK_INTEGRATION.md)
- [General Integration Guide](./INTEGRATION_GUIDE.md)
