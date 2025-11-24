# Discord Integration - Quick Start

Get Discord integration working in 5 minutes.

## What Was Fixed

1. ✅ Added missing Discord environment variables to `.env`
2. ✅ Fixed Discord OAuth scopes (added `identify` and `guilds`)
3. ✅ Added guild information fetching after authorization
4. ✅ Created comprehensive documentation

## Setup Steps

### 1. Create Discord Application

1. Go to https://discord.com/developers/applications
2. Click **"New Application"**
3. Name it (e.g., "TMS Bot")
4. Navigate to **OAuth2** section
5. Copy **Client ID** and **Client Secret**

### 2. Configure Redirect URI

In Discord Developer Portal → OAuth2 → Redirects:

```
http://localhost:8080/api/public/integrations/discord/callback
```

### 3. Update Backend Environment

Edit `app/backend/.env`:

```bash
# Discord OAuth Configuration
DISCORD_CLIENT_ID="your_client_id_here"
DISCORD_CLIENT_SECRET="your_client_secret_here"
DISCORD_REDIRECT_URI="http://localhost:8080/api/public/integrations/discord/callback"
```

### 4. Restart Backend

```bash
cd app/backend
# If using docker
docker-compose restart backend

# If running locally
go run cmd/api/main.go
```

### 5. Test Integration

1. Open TMS Agent Console: http://localhost:3000
2. Navigate to **Settings** → **Integrations**
3. Click **Install** on Discord card
4. Authorize on Discord
5. Select a server
6. Verify success message

## Verify Installation

Check the database:

```sql
SELECT 
  id, 
  integration_type, 
  meta->>'guild_name' as server_name,
  status,
  created_at
FROM project_integrations 
WHERE integration_type = 'discord';
```

Expected result:
```
id                                   | integration_type | server_name    | status | created_at
-------------------------------------|------------------|----------------|--------|------------
uuid-here                           | discord          | My Server      | active | 2025-11-24
```

## Troubleshooting

### "Invalid Client ID"
- Verify `DISCORD_CLIENT_ID` is correct
- Remove any quotes or extra spaces
- Restart backend

### "Invalid Redirect URI"
- Ensure redirect URI in Discord Portal exactly matches `.env`
- Use `http://` for local dev, `https://` for production

### "State Token Expired"
- Redis must be running
- Try the OAuth flow again (tokens expire after 10 minutes)

### Bot Not in Server
- Ensure `bot` scope is included
- Re-run installation and select a server

## Next Steps

- **Configure Notifications**: Set up rules to send ticket updates to Discord
- **Channel Mapping**: Map specific ticket types to Discord channels
- **Test Messages**: Send a test notification to verify connectivity

## Documentation

- [Full Discord Integration Guide](./DISCORD_INTEGRATION.md)
- [General Integration Guide](./INTEGRATION_GUIDE.md)
- [Slack Integration](./SLACK_INTEGRATION.md)

## Changes Made to Code

### 1. Environment Variables (`app/backend/.env`)

```diff
+ # Discord OAuth Configuration
+ DISCORD_CLIENT_ID=""
+ DISCORD_CLIENT_SECRET=""
+ DISCORD_REDIRECT_URI="http://localhost:8080/api/public/integrations/discord/callback"
```

### 2. Discord Scopes (`internal/service/integration_oauth.go`)

```diff
 var discordScopes = []string{
     "bot",
-    "webhook.incoming",
+    "identify",
+    "guilds",
 }
```

### 3. Guild Fetching (`internal/service/integration_oauth.go`)

Added `fetchDiscordGuilds()` function to retrieve server information after authorization.

## Production Deployment

For production, update:

```bash
DISCORD_REDIRECT_URI="https://api.yourdomain.com/api/public/integrations/discord/callback"
```

And add this redirect URI in Discord Developer Portal.

## Support

If you encounter issues:
1. Check backend logs: `docker-compose logs backend`
2. Verify Redis is running: `docker-compose ps redis`
3. Test endpoints manually with curl
4. Review [full documentation](./DISCORD_INTEGRATION.md)
