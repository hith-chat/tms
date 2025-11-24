package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/logger"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/redis"
	"github.com/bareuptime/tms/internal/repo"
)

// IntegrationOAuthService handles OAuth flows for integrations
type IntegrationOAuthService struct {
	config          *config.Config
	redisService    *redis.Service
	integrationRepo *repo.ProjectIntegrationRepository
}

// NewIntegrationOAuthService creates a new integration OAuth service
func NewIntegrationOAuthService(
	cfg *config.Config,
	redisService *redis.Service,
	integrationRepo *repo.ProjectIntegrationRepository,
) *IntegrationOAuthService {
	return &IntegrationOAuthService{
		config:          cfg,
		redisService:    redisService,
		integrationRepo: integrationRepo,
	}
}

// Slack OAuth scopes
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

// Discord OAuth scopes - bot permissions for sending messages to threads
var discordScopes = []string{
	"bot",
	"webhook.incoming",
}

// Discord bot permissions (bitfield) - Send Messages, Send Messages in Threads, Read Message History, Use Slash Commands
const discordBotPermissions = "2147485696"

// GenerateOAuthState generates a state token and stores it in Redis
func (s *IntegrationOAuthService) GenerateOAuthState(
	ctx context.Context,
	tenantID, projectID, agentID uuid.UUID,
	integrationType models.ProjectIntegrationType,
) (string, error) {
	// Generate random state token
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate state token: %w", err)
	}
	stateToken := hex.EncodeToString(bytes)

	// Create state data
	stateData := models.OAuthStateData{
		TenantID:        tenantID,
		ProjectID:       projectID,
		IntegrationType: integrationType,
		AgentID:         agentID,
		CreatedAt:       time.Now(),
	}

	// Serialize state data
	stateJSON, err := json.Marshal(stateData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal state data: %w", err)
	}

	// Store in Redis with 10 minute TTL
	key := oauthStateKey(stateToken)
	err = s.redisService.GetClient().Set(ctx, key, stateJSON, 10*time.Minute).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store state in Redis: %w", err)
	}

	logger.GetTxLogger(ctx).Info().
		Str("state_token", stateToken[:8]+"...").
		Str("tenant_id", tenantID.String()).
		Str("project_id", projectID.String()).
		Str("integration_type", string(integrationType)).
		Msg("Generated OAuth state token")

	return stateToken, nil
}

// ValidateOAuthState validates and retrieves the state data from Redis
func (s *IntegrationOAuthService) ValidateOAuthState(ctx context.Context, stateToken string) (*models.OAuthStateData, error) {
	key := oauthStateKey(stateToken)

	// Get from Redis
	stateJSON, err := s.redisService.GetClient().Get(ctx, key).Bytes()
	if err != nil {
		if err == goredis.Nil {
			return nil, fmt.Errorf("state token not found or expired")
		}
		return nil, fmt.Errorf("failed to get state from Redis: %w", err)
	}

	// Deserialize state data
	var stateData models.OAuthStateData
	if err := json.Unmarshal(stateJSON, &stateData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state data: %w", err)
	}

	// Delete the state token after successful validation (one-time use)
	s.redisService.GetClient().Del(ctx, key)

	logger.GetTxLogger(ctx).Info().
		Str("state_token", stateToken[:8]+"...").
		Str("tenant_id", stateData.TenantID.String()).
		Str("project_id", stateData.ProjectID.String()).
		Msg("Validated OAuth state token")

	return &stateData, nil
}

// GetSlackOAuthURL generates the Slack OAuth URL for installation
func (s *IntegrationOAuthService) GetSlackOAuthURL(stateToken string) string {
	oauthURL, _ := url.Parse("https://slack.com/oauth/v2/authorize")

	params := url.Values{}
	params.Set("client_id", s.config.Slack.ClientID)
	params.Set("scope", strings.Join(slackScopes, ","))
	params.Set("redirect_uri", s.config.Slack.RedirectURI)
	params.Set("state", stateToken)

	oauthURL.RawQuery = params.Encode()
	return oauthURL.String()
}

// GetDiscordOAuthURL generates the Discord OAuth URL for bot installation
func (s *IntegrationOAuthService) GetDiscordOAuthURL(stateToken string) string {
	oauthURL, _ := url.Parse("https://discord.com/api/oauth2/authorize")

	params := url.Values{}
	params.Set("client_id", s.config.Discord.ClientID)
	params.Set("scope", strings.Join(discordScopes, " "))
	params.Set("permissions", discordBotPermissions)
	params.Set("redirect_uri", s.config.Discord.RedirectURI)
	params.Set("response_type", "code")
	params.Set("state", stateToken)

	oauthURL.RawQuery = params.Encode()
	return oauthURL.String()
}

// SlackOAuthResponse represents the response from Slack OAuth token exchange
type SlackOAuthResponse struct {
	OK          bool   `json:"ok"`
	Error       string `json:"error,omitempty"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	BotUserID   string `json:"bot_user_id"`
	AppID       string `json:"app_id"`
	Team        struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"team"`
	AuthedUser struct {
		ID string `json:"id"`
	} `json:"authed_user"`
	IncomingWebhook *struct {
		URL              string `json:"url"`
		Channel          string `json:"channel"`
		ChannelID        string `json:"channel_id"`
		ConfigurationURL string `json:"configuration_url"`
	} `json:"incoming_webhook,omitempty"`
}

// ExchangeSlackCode exchanges the authorization code for tokens
func (s *IntegrationOAuthService) ExchangeSlackCode(ctx context.Context, code string) (*SlackOAuthResponse, error) {
	// Build the request
	data := url.Values{}
	data.Set("client_id", s.config.Slack.ClientID)
	data.Set("client_secret", s.config.Slack.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", s.config.Slack.RedirectURI)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/oauth.v2.access", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Make the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var oauthResp SlackOAuthResponse
	if err := json.Unmarshal(body, &oauthResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !oauthResp.OK {
		return nil, fmt.Errorf("slack oauth error: %s", oauthResp.Error)
	}

	logger.GetTxLogger(ctx).Info().
		Str("team_id", oauthResp.Team.ID).
		Str("team_name", oauthResp.Team.Name).
		Str("bot_user_id", oauthResp.BotUserID).
		Msg("Successfully exchanged Slack OAuth code")

	return &oauthResp, nil
}

// StoreSlackIntegration stores the Slack integration in the database
func (s *IntegrationOAuthService) StoreSlackIntegration(
	ctx context.Context,
	stateData *models.OAuthStateData,
	oauthResp *SlackOAuthResponse,
) (*models.ProjectIntegration, error) {
	// Build webhooks array
	var webhooks []models.SlackWebhookConfig
	if oauthResp.IncomingWebhook != nil {
		webhooks = append(webhooks, models.SlackWebhookConfig{
			URL:              oauthResp.IncomingWebhook.URL,
			Channel:          oauthResp.IncomingWebhook.Channel,
			ChannelID:        oauthResp.IncomingWebhook.ChannelID,
			ConfigurationURL: oauthResp.IncomingWebhook.ConfigurationURL,
		})
	}

	// Create Slack meta
	slackMeta := &models.SlackIntegrationMeta{
		AccessToken:        oauthResp.AccessToken,
		TokenType:          oauthResp.TokenType,
		BotUserID:          oauthResp.BotUserID,
		AppID:              oauthResp.AppID,
		Scope:              oauthResp.Scope,
		TeamID:             oauthResp.Team.ID,
		TeamName:           oauthResp.Team.Name,
		AuthedUserID:       oauthResp.AuthedUser.ID,
		Webhooks:           webhooks,
		InstalledByAgentID: stateData.AgentID.String(),
		InstalledAt:        time.Now(),
		LastUpdatedAt:      time.Now(),
	}

	// Convert to generic meta
	meta, err := slackMeta.ToMeta()
	if err != nil {
		return nil, fmt.Errorf("failed to convert meta: %w", err)
	}

	// Create integration record
	integration := &models.ProjectIntegration{
		ID:              uuid.New(),
		TenantID:        stateData.TenantID,
		ProjectID:       stateData.ProjectID,
		IntegrationType: models.ProjectIntegrationTypeSlack,
		Meta:            meta,
		Status:          models.ProjectIntegrationStatusActive,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Upsert (create or update if exists)
	if err := s.integrationRepo.Upsert(ctx, integration); err != nil {
		return nil, fmt.Errorf("failed to store integration: %w", err)
	}

	logger.GetTxLogger(ctx).Info().
		Str("integration_id", integration.ID.String()).
		Str("tenant_id", stateData.TenantID.String()).
		Str("project_id", stateData.ProjectID.String()).
		Str("team_name", oauthResp.Team.Name).
		Msg("Successfully stored Slack integration")

	return integration, nil
}

// DiscordOAuthResponse represents the response from Discord OAuth token exchange
type DiscordOAuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	Guild        *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Icon string `json:"icon,omitempty"`
	} `json:"guild,omitempty"`
	Webhook *struct {
		ID        string `json:"id"`
		Token     string `json:"token"`
		URL       string `json:"url"`
		ChannelID string `json:"channel_id"`
	} `json:"webhook,omitempty"`
}

// ExchangeDiscordCode exchanges the authorization code for tokens
func (s *IntegrationOAuthService) ExchangeDiscordCode(ctx context.Context, code string) (*DiscordOAuthResponse, error) {
	// Build the request
	data := url.Values{}
	data.Set("client_id", s.config.Discord.ClientID)
	data.Set("client_secret", s.config.Discord.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", s.config.Discord.RedirectURI)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://discord.com/api/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Make the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discord oauth error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var oauthResp DiscordOAuthResponse
	if err := json.Unmarshal(body, &oauthResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.GetTxLogger(ctx).Info().
		Str("scope", oauthResp.Scope).
		Msg("Successfully exchanged Discord OAuth code")

	return &oauthResp, nil
}

// StoreDiscordIntegration stores the Discord integration in the database
func (s *IntegrationOAuthService) StoreDiscordIntegration(
	ctx context.Context,
	stateData *models.OAuthStateData,
	oauthResp *DiscordOAuthResponse,
) (*models.ProjectIntegration, error) {
	// Create Discord meta
	discordMeta := &models.DiscordIntegrationMeta{
		AccessToken:        oauthResp.AccessToken,
		RefreshToken:       oauthResp.RefreshToken,
		TokenType:          oauthResp.TokenType,
		ExpiresIn:          oauthResp.ExpiresIn,
		Scope:              oauthResp.Scope,
		InstalledByAgentID: stateData.AgentID.String(),
		InstalledAt:        time.Now(),
		LastUpdatedAt:      time.Now(),
	}

	// Add guild info if available
	if oauthResp.Guild != nil {
		discordMeta.GuildID = oauthResp.Guild.ID
		discordMeta.GuildName = oauthResp.Guild.Name
		discordMeta.GuildIcon = oauthResp.Guild.Icon
	}

	// Add webhook info if available
	if oauthResp.Webhook != nil {
		discordMeta.WebhookID = oauthResp.Webhook.ID
		discordMeta.WebhookToken = oauthResp.Webhook.Token
		discordMeta.WebhookURL = oauthResp.Webhook.URL
		discordMeta.ChannelID = oauthResp.Webhook.ChannelID
	}

	// Convert to generic meta
	meta, err := discordMeta.ToMeta()
	if err != nil {
		return nil, fmt.Errorf("failed to convert meta: %w", err)
	}

	// Create integration record
	integration := &models.ProjectIntegration{
		ID:              uuid.New(),
		TenantID:        stateData.TenantID,
		ProjectID:       stateData.ProjectID,
		IntegrationType: models.ProjectIntegrationTypeDiscord,
		Meta:            meta,
		Status:          models.ProjectIntegrationStatusActive,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Upsert (create or update if exists)
	if err := s.integrationRepo.Upsert(ctx, integration); err != nil {
		return nil, fmt.Errorf("failed to store integration: %w", err)
	}

	guildName := ""
	if oauthResp.Guild != nil {
		guildName = oauthResp.Guild.Name
	}

	logger.GetTxLogger(ctx).Info().
		Str("integration_id", integration.ID.String()).
		Str("tenant_id", stateData.TenantID.String()).
		Str("project_id", stateData.ProjectID.String()).
		Str("guild_name", guildName).
		Msg("Successfully stored Discord integration")

	return integration, nil
}

// GetProjectIntegration retrieves an integration by project and type
func (s *IntegrationOAuthService) GetProjectIntegration(
	ctx context.Context,
	tenantID, projectID uuid.UUID,
	integrationType models.ProjectIntegrationType,
) (*models.ProjectIntegration, error) {
	return s.integrationRepo.GetByProjectAndType(ctx, tenantID, projectID, integrationType)
}

// ListProjectIntegrations retrieves all integrations for a project
func (s *IntegrationOAuthService) ListProjectIntegrations(
	ctx context.Context,
	tenantID, projectID uuid.UUID,
) ([]*models.ProjectIntegration, error) {
	return s.integrationRepo.ListByProject(ctx, tenantID, projectID)
}

// DeleteProjectIntegration deletes an integration
func (s *IntegrationOAuthService) DeleteProjectIntegration(
	ctx context.Context,
	tenantID, projectID uuid.UUID,
	integrationType models.ProjectIntegrationType,
) error {
	return s.integrationRepo.DeleteByProjectAndType(ctx, tenantID, projectID, integrationType)
}

// oauthStateKey generates the Redis key for OAuth state
func oauthStateKey(stateToken string) string {
	return fmt.Sprintf("oauth_state:%s", stateToken)
}
