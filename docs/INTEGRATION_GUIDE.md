# TMS Integration Guide

Complete guide for integrating third-party platforms with your Ticket Management System.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Supported Integrations](#supported-integrations)
4. [Adding New Integrations](#adding-new-integrations)
5. [Integration Types](#integration-types)
6. [Common Patterns](#common-patterns)
7. [Testing](#testing)

---

## Overview

The TMS integration system provides a flexible framework for connecting external platforms. All integrations follow a consistent OAuth 2.0 pattern with project-level scoping.

### Key Features

- **OAuth 2.0 Support**: Secure authorization flows
- **Project-Level Scoping**: Each project can have independent integrations
- **Flexible Metadata**: JSONB storage for platform-specific data
- **State Management**: Redis-backed OAuth state validation
- **Multi-Tenant**: Full tenant isolation

---

## Architecture

### Components

```
┌─────────────────┐
│   Frontend      │
│  (Agent Console)│
└────────┬────────┘
         │
         │ 1. Install Integration
         ▼
┌─────────────────┐
│   Backend API   │
│  Handler Layer  │◄──── 3. OAuth Callback
└────────┬────────┘
         │
         │ 2. OAuth URL
         ▼
┌─────────────────┐
│  OAuth Provider │
│ (Slack/Discord) │
└─────────────────┘
         │
         ▼
   [User Authorizes]
         │
         │ 3. Redirect with Code
         └──────────────►
```

### Database Schema

```sql
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
```

### Code Structure

```
internal/
├── handlers/
│   ├── integration_oauth.go       # HTTP handlers for OAuth
│   └── integration.go              # Legacy integration handlers
├── service/
│   └── integration_oauth.go        # Business logic for OAuth
├── models/
│   ├── integration.go              # Integration types & enums
│   └── project_integration.go      # Project integration models
└── repo/
    └── project_integration.go      # Database operations
```

---

## Supported Integrations

### 1. Slack

**Status**: ✅ Fully Implemented

**Features**:
- Bot user with custom scopes
- Incoming webhooks
- Channel message posting
- Team and workspace info

**Documentation**: [SLACK_INTEGRATION.md](./SLACK_INTEGRATION.md)

**Scopes**:
```go
var slackScopes = []string{
    "channels:read",
    "chat:write",
    "chat:write.public",
    "incoming-webhook",
    "users.profile:read",
    "users:read",
    "users:read.email",
    "commands",
    "channels:history",
    "groups:history",
    "im:history",
    "mpim:history",
}
```

### 2. Discord

**Status**: ✅ Fully Implemented

**Features**:
- Bot integration with guild selection
- Message posting to channels and threads
- Guild (server) information retrieval
- User authorization tracking

**Documentation**: [DISCORD_INTEGRATION.md](./DISCORD_INTEGRATION.md)

**Scopes**:
```go
var discordScopes = []string{
    "bot",
    "identify",
    "guilds",
}
```

**Permissions**: `2147485696` (Send Messages, Send Messages in Threads, Read Message History, Use Slash Commands)

### 3. Microsoft Teams

**Status**: 🚧 Planned

**Expected Features**:
- Team and channel integration
- Adaptive card messages
- Bot framework integration

### 4. Other Platforms

The system is designed to support additional integrations:

- **JIRA**: Issue tracking sync
- **Linear**: Project management
- **GitHub**: Issue and PR notifications
- **Google Calendar**: Meeting scheduling
- **Zoom**: Video call integration
- **Zendesk**: Support ticket sync

---

## Adding New Integrations

### Step 1: Define Integration Type

Add to `internal/models/project_integration.go`:

```go
const (
    ProjectIntegrationTypeSlack   ProjectIntegrationType = "slack"
    ProjectIntegrationTypeDiscord ProjectIntegrationType = "discord"
    ProjectIntegrationTypeNewApp  ProjectIntegrationType = "new_app"  // Add new type
)
```

### Step 2: Create Configuration Struct

Add to `internal/config/config.go`:

```go
type NewAppConfig struct {
    ClientID     string `mapstructure:"client_id"`
    ClientSecret string `mapstructure:"client_secret"`
    RedirectURI  string `mapstructure:"redirect_uri"`
}

type Config struct {
    // ... existing fields
    NewApp NewAppConfig `mapstructure:"new_app"`
}
```

### Step 3: Define Metadata Model

Add to `internal/models/project_integration.go`:

```go
type NewAppIntegrationMeta struct {
    AccessToken       string    `json:"access_token"`
    RefreshToken      string    `json:"refresh_token,omitempty"`
    TokenType         string    `json:"token_type"`
    Scope             string    `json:"scope"`
    
    // Platform-specific fields
    WorkspaceID       string    `json:"workspace_id,omitempty"`
    WorkspaceName     string    `json:"workspace_name,omitempty"`
    
    // Installation tracking
    InstalledByAgentID string   `json:"installed_by_agent_id,omitempty"`
    InstalledAt       time.Time `json:"installed_at"`
    LastUpdatedAt     time.Time `json:"last_updated_at"`
}

func (n *NewAppIntegrationMeta) ToMeta() (IntegrationMeta, error) {
    data, err := json.Marshal(n)
    if err != nil {
        return nil, err
    }
    var meta IntegrationMeta
    if err := json.Unmarshal(data, &meta); err != nil {
        return nil, err
    }
    return meta, nil
}
```

### Step 4: Add OAuth Scopes

Add to `internal/service/integration_oauth.go`:

```go
var newAppScopes = []string{
    "read:workspace",
    "write:messages",
    // Add required scopes
}

func (s *IntegrationOAuthService) GetNewAppOAuthURL(stateToken string) string {
    oauthURL, _ := url.Parse("https://newapp.com/oauth/authorize")
    
    params := url.Values{}
    params.Set("client_id", s.config.NewApp.ClientID)
    params.Set("scope", strings.Join(newAppScopes, " "))
    params.Set("redirect_uri", s.config.NewApp.RedirectURI)
    params.Set("response_type", "code")
    params.Set("state", stateToken)
    
    oauthURL.RawQuery = params.Encode()
    return oauthURL.String()
}
```

### Step 5: Implement Token Exchange

Add to `internal/service/integration_oauth.go`:

```go
type NewAppOAuthResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token,omitempty"`
    TokenType    string `json:"token_type"`
    ExpiresIn    int    `json:"expires_in,omitempty"`
    Scope        string `json:"scope"`
    
    Workspace    *struct {
        ID   string `json:"id"`
        Name string `json:"name"`
    } `json:"workspace,omitempty"`
}

func (s *IntegrationOAuthService) ExchangeNewAppCode(ctx context.Context, code string) (*NewAppOAuthResponse, error) {
    data := url.Values{}
    data.Set("client_id", s.config.NewApp.ClientID)
    data.Set("client_secret", s.config.NewApp.ClientSecret)
    data.Set("code", code)
    data.Set("grant_type", "authorization_code")
    data.Set("redirect_uri", s.config.NewApp.RedirectURI)

    req, err := http.NewRequestWithContext(ctx, "POST", "https://newapp.com/oauth/token", strings.NewReader(data.Encode()))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to exchange code: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("oauth error: status %d, body: %s", resp.StatusCode, string(body))
    }

    var oauthResp NewAppOAuthResponse
    if err := json.Unmarshal(body, &oauthResp); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }

    return &oauthResp, nil
}
```

### Step 6: Implement Storage

Add to `internal/service/integration_oauth.go`:

```go
func (s *IntegrationOAuthService) StoreNewAppIntegration(
    ctx context.Context,
    stateData *models.OAuthStateData,
    oauthResp *NewAppOAuthResponse,
) (*models.ProjectIntegration, error) {
    meta := &models.NewAppIntegrationMeta{
        AccessToken:        oauthResp.AccessToken,
        RefreshToken:       oauthResp.RefreshToken,
        TokenType:          oauthResp.TokenType,
        Scope:              oauthResp.Scope,
        InstalledByAgentID: stateData.AgentID.String(),
        InstalledAt:        time.Now(),
        LastUpdatedAt:      time.Now(),
    }

    if oauthResp.Workspace != nil {
        meta.WorkspaceID = oauthResp.Workspace.ID
        meta.WorkspaceName = oauthResp.Workspace.Name
    }

    metaJSON, err := meta.ToMeta()
    if err != nil {
        return nil, fmt.Errorf("failed to convert meta: %w", err)
    }

    integration := &models.ProjectIntegration{
        ID:              uuid.New(),
        TenantID:        stateData.TenantID,
        ProjectID:       stateData.ProjectID,
        IntegrationType: models.ProjectIntegrationTypeNewApp,
        Meta:            metaJSON,
        Status:          models.ProjectIntegrationStatusActive,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }

    if err := s.integrationRepo.Upsert(ctx, integration); err != nil {
        return nil, fmt.Errorf("failed to store integration: %w", err)
    }

    return integration, nil
}
```

### Step 7: Add Handler Routes

Update `internal/handlers/integration_oauth.go`:

```go
func (h *IntegrationOAuthHandler) InstallIntegration(c *gin.Context) {
    // ... existing code ...
    
    switch integrationType {
    case models.ProjectIntegrationTypeSlack:
        oauthURL = h.integrationService.GetSlackOAuthURL(stateToken)
    case models.ProjectIntegrationTypeDiscord:
        oauthURL = h.integrationService.GetDiscordOAuthURL(stateToken)
    case models.ProjectIntegrationTypeNewApp:
        oauthURL = h.integrationService.GetNewAppOAuthURL(stateToken)
    default:
        c.JSON(http.StatusBadRequest, gin.H{"error": "OAuth not implemented for this integration type"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"oauth_url": oauthURL})
}

func (h *IntegrationOAuthHandler) NewAppOAuthCallback(c *gin.Context) {
    ctx := c.Request.Context()

    // Check for errors
    if errParam := c.Query("error"); errParam != "" {
        logger.GetTxLogger(ctx).Error().Str("error", errParam).Msg("NewApp OAuth error")
        c.Redirect(http.StatusFound, h.frontendURL+"/dashboard?error=newapp_oauth_denied")
        return
    }

    code := c.Query("code")
    if code == "" {
        c.Redirect(http.StatusFound, h.frontendURL+"/dashboard?error=missing_code")
        return
    }

    stateToken := c.Query("state")
    if stateToken == "" {
        c.Redirect(http.StatusFound, h.frontendURL+"/dashboard?error=missing_state")
        return
    }

    stateData, err := h.integrationService.ValidateOAuthState(ctx, stateToken)
    if err != nil {
        logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to validate OAuth state")
        c.Redirect(http.StatusFound, h.frontendURL+"/dashboard?error=invalid_state")
        return
    }

    oauthResp, err := h.integrationService.ExchangeNewAppCode(ctx, code)
    if err != nil {
        logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to exchange NewApp code")
        c.Redirect(http.StatusFound, h.frontendURL+"/dashboard?error=token_exchange_failed")
        return
    }

    _, err = h.integrationService.StoreNewAppIntegration(ctx, stateData, oauthResp)
    if err != nil {
        logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to store NewApp integration")
        c.Redirect(http.StatusFound, h.frontendURL+"/dashboard?error=storage_failed")
        return
    }

    c.Redirect(http.StatusFound, h.frontendURL+"/dashboard?integration=newapp&status=success")
}
```

### Step 8: Register Routes

Update `cmd/api/main.go`:

```go
// Public OAuth callback routes
publicRoutes := router.Group("/api/public")
{
    publicRoutes.GET("/integrations/slack/callback", integrationOAuthHandler.SlackOAuthCallback)
    publicRoutes.GET("/integrations/discord/callback", integrationOAuthHandler.DiscordOAuthCallback)
    publicRoutes.GET("/integrations/newapp/callback", integrationOAuthHandler.NewAppOAuthCallback)
}
```

### Step 9: Add Environment Variables

Update `.env`:

```bash
# NewApp OAuth Configuration
NEWAPP_CLIENT_ID=""
NEWAPP_CLIENT_SECRET=""
NEWAPP_REDIRECT_URI="http://localhost:8080/api/public/integrations/newapp/callback"
```

### Step 10: Bind Environment Variables

Update `internal/config/config.go` in the `Load()` function:

```go
// NewApp configuration bindings
viper.BindEnv("newapp.client_id", "NEWAPP_CLIENT_ID")
viper.BindEnv("newapp.client_secret", "NEWAPP_CLIENT_SECRET")
viper.BindEnv("newapp.redirect_uri", "NEWAPP_REDIRECT_URI")
```

And in `setDefaults()`:

```go
// NewApp defaults
viper.SetDefault("newapp.client_id", "")
viper.SetDefault("newapp.client_secret", "")
viper.SetDefault("newapp.redirect_uri", "")
```

---

## Integration Types

### OAuth 2.0 Integration

Most modern platforms use OAuth 2.0:

**Characteristics**:
- User consent required
- Access tokens with refresh capability
- Scoped permissions
- State validation for security

**Examples**: Slack, Discord, Google, GitHub

### API Key Integration

Some platforms use simple API keys:

**Characteristics**:
- No user consent flow
- Static credentials
- Simpler but less secure

**Examples**: SendGrid, Stripe webhooks

### Webhook Integration

Event-driven integrations:

**Characteristics**:
- Platform sends events to your endpoint
- Signature verification required
- Asynchronous processing

**Examples**: Stripe webhooks, GitHub webhooks

---

## Common Patterns

### 1. State Token Pattern

```go
// Generate state
stateToken, err := service.GenerateOAuthState(ctx, tenantID, projectID, agentID, integrationType)

// Store in Redis with TTL
redis.Set(ctx, "oauth_state:"+stateToken, stateData, 10*time.Minute)

// Validate on callback
stateData, err := service.ValidateOAuthState(ctx, stateToken)
```

### 2. Upsert Pattern

```go
// Upsert ensures one integration per type per project
query := `
    INSERT INTO project_integrations (...)
    VALUES (...)
    ON CONFLICT (tenant_id, project_id, integration_type)
    DO UPDATE SET meta = EXCLUDED.meta, updated_at = NOW()
    RETURNING id
`
```

### 3. Metadata Extraction Pattern

```go
// Generic getter
func GetSlackMeta(integration *ProjectIntegration) (*SlackIntegrationMeta, error) {
    data, err := json.Marshal(integration.Meta)
    if err != nil {
        return nil, err
    }
    
    var meta SlackIntegrationMeta
    if err := json.Unmarshal(data, &meta); err != nil {
        return nil, err
    }
    
    return &meta, nil
}
```

### 4. Token Refresh Pattern

```go
func (s *IntegrationService) RefreshToken(ctx context.Context, integration *ProjectIntegration) error {
    // Extract current meta
    meta, err := GetPlatformMeta(integration)
    if err != nil {
        return err
    }
    
    // Exchange refresh token
    newTokens, err := s.exchangeRefreshToken(ctx, meta.RefreshToken)
    if err != nil {
        return err
    }
    
    // Update meta
    meta.AccessToken = newTokens.AccessToken
    meta.ExpiresAt = time.Now().Add(time.Duration(newTokens.ExpiresIn) * time.Second)
    
    // Save
    integration.Meta, _ = meta.ToMeta()
    return s.repo.Update(ctx, integration)
}
```

---

## Testing

### Unit Tests

```go
func TestExchangeCode(t *testing.T) {
    // Mock HTTP client
    client := &MockHTTPClient{
        Response: `{"access_token": "test_token"}`,
        StatusCode: 200,
    }
    
    service := NewIntegrationOAuthService(config, redis, repo)
    service.httpClient = client
    
    resp, err := service.ExchangeDiscordCode(ctx, "test_code")
    assert.NoError(t, err)
    assert.Equal(t, "test_token", resp.AccessToken)
}
```

### Integration Tests

```bash
# Test OAuth flow
curl -X GET \
  'http://localhost:8080/v1/tenants/{tenant_id}/projects/{project_id}/integrations/discord/install' \
  -H 'Authorization: Bearer YOUR_TOKEN'

# Should return OAuth URL
# Visit URL in browser and authorize
# Check callback succeeds
```

### Manual Testing

1. Set up test Discord/Slack app
2. Configure dev environment variables
3. Run backend: `go run cmd/api/main.go`
4. Initiate OAuth flow from frontend
5. Verify integration appears in database
6. Test sending messages/notifications

---

## Best Practices

### Security

1. **Never log secrets**: Redact tokens in logs
2. **Use HTTPS in production**: Protect token exchange
3. **Validate all inputs**: State tokens, codes, etc.
4. **Encrypt sensitive data**: At rest and in transit
5. **Implement rate limiting**: Protect OAuth endpoints

### Performance

1. **Cache OAuth URLs**: Generate once per session
2. **Async token exchange**: Don't block user flow
3. **Batch database operations**: Use transactions
4. **Index properly**: On tenant_id, project_id, type

### Reliability

1. **Implement retries**: With exponential backoff
2. **Handle token expiry**: Auto-refresh when possible
3. **Log comprehensively**: Aid debugging
4. **Monitor integration health**: Track success/failure rates

### User Experience

1. **Clear error messages**: Help users fix issues
2. **Loading states**: Show progress during OAuth
3. **Success confirmation**: Redirect with clear status
4. **Easy disconnection**: Allow users to remove integrations

---

## Resources

- [OAuth 2.0 Specification](https://oauth.net/2/)
- [Slack API Documentation](https://api.slack.com/)
- [Discord API Documentation](https://discord.com/developers/docs/)
- [Microsoft Teams Documentation](https://docs.microsoft.com/en-us/microsoftteams/)

---

## Support

For questions or issues:
- Check integration-specific guides (e.g., `DISCORD_INTEGRATION.md`)
- Review backend logs for detailed errors
- Consult API documentation for each platform
- Test with development/sandbox accounts first
