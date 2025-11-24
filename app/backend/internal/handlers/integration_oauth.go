package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/logger"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/service"
)

// IntegrationOAuthHandler handles integration OAuth endpoints
type IntegrationOAuthHandler struct {
	integrationService *service.IntegrationOAuthService
	frontendURL        string
}

// NewIntegrationOAuthHandler creates a new integration OAuth handler
func NewIntegrationOAuthHandler(
	integrationService *service.IntegrationOAuthService,
	frontendURL string,
) *IntegrationOAuthHandler {
	return &IntegrationOAuthHandler{
		integrationService: integrationService,
		frontendURL:        frontendURL,
	}
}

// InstallIntegration initiates the OAuth flow for an integration
// @Summary Install integration
// @Description Initiates the OAuth flow for a third-party integration (e.g., Slack)
// @Tags integrations
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param project_id path string true "Project ID"
// @Param app_type path string true "Integration type (e.g., slack)"
// @Security BearerAuth
// @Success 302 {string} string "Redirect to OAuth provider"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/tenants/{tenant_id}/projects/{project_id}/integrations/{app_type}/install [get]
func (h *IntegrationOAuthHandler) InstallIntegration(c *gin.Context) {
	ctx := c.Request.Context()

	// Get tenant ID from path
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id"})
		return
	}

	// Get project ID from path
	projectIDStr := c.Param("project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	// Get integration type from path
	appType := c.Param("app_type")
	if appType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app_type is required"})
		return
	}

	// Get agent ID from context (set by auth middleware)
	var agentUUID uuid.UUID
	agentIDInterface, exists := c.Get("agent_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "agent_id not found in context"})
		return
	}

	agentIDStr, ok := agentIDInterface.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid agent_id type"})
		return
	}

	agentUUID, err = uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid agent_id format"})
		return
	}

	// Convert app_type to ProjectIntegrationType
	var integrationType models.ProjectIntegrationType
	switch appType {
	case "slack":
		integrationType = models.ProjectIntegrationTypeSlack
	case "discord":
		integrationType = models.ProjectIntegrationTypeDiscord
	case "microsoft_teams":
		integrationType = models.ProjectIntegrationTypeTeams
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported integration type"})
		return
	}

	// Generate OAuth state
	stateToken, err := h.integrationService.GenerateOAuthState(ctx, tenantID, projectID, agentUUID, integrationType)
	if err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to generate OAuth state")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initiate OAuth flow"})
		return
	}

	// Get OAuth URL based on integration type
	var oauthURL string
	switch integrationType {
	case models.ProjectIntegrationTypeSlack:
		oauthURL = h.integrationService.GetSlackOAuthURL(stateToken)
	case models.ProjectIntegrationTypeDiscord:
		oauthURL = h.integrationService.GetDiscordOAuthURL(stateToken)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "OAuth not implemented for this integration type"})
		return
	}

	// Return OAuth URL as JSON (frontend will handle redirect)
	c.JSON(http.StatusOK, gin.H{"oauth_url": oauthURL})
}

// SlackOAuthCallback handles the Slack OAuth callback
// @Summary Slack OAuth callback
// @Description Handles the callback from Slack OAuth flow
// @Tags integrations
// @Accept json
// @Produce json
// @Param code query string true "Authorization code"
// @Param state query string true "State token"
// @Success 302 {string} string "Redirect to frontend"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/integrations/slack/callback [get]
func (h *IntegrationOAuthHandler) SlackOAuthCallback(c *gin.Context) {
	ctx := c.Request.Context()

	// Check for error from Slack
	if errParam := c.Query("error"); errParam != "" {
		logger.GetTxLogger(ctx).Error().
			Str("error", errParam).
			Msg("Slack OAuth error")
		c.Redirect(http.StatusFound, h.frontendURL+"/dashboard?error=slack_oauth_denied")
		return
	}

	// Get authorization code
	code := c.Query("code")
	if code == "" {
		c.Redirect(http.StatusFound, h.frontendURL+"/dashboard?error=missing_code")
		return
	}

	// Get state token
	stateToken := c.Query("state")
	if stateToken == "" {
		c.Redirect(http.StatusFound, h.frontendURL+"/dashboard?error=missing_state")
		return
	}

	// Validate state token
	stateData, err := h.integrationService.ValidateOAuthState(ctx, stateToken)
	if err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to validate OAuth state")
		c.Redirect(http.StatusFound, h.frontendURL+"/dashboard?error=invalid_state")
		return
	}

	// Exchange code for tokens
	oauthResp, err := h.integrationService.ExchangeSlackCode(ctx, code)
	if err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to exchange Slack code")
		c.Redirect(http.StatusFound, h.frontendURL+"/dashboard?error=token_exchange_failed")
		return
	}

	// Store the integration
	_, err = h.integrationService.StoreSlackIntegration(ctx, stateData, oauthResp)
	if err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to store Slack integration")
		c.Redirect(http.StatusFound, h.frontendURL+"/dashboard?error=storage_failed")
		return
	}

	// Redirect to frontend dashboard with success
	c.Redirect(http.StatusFound, h.frontendURL+"/dashboard?integration=slack&status=success")
}

// DiscordOAuthCallback handles the Discord OAuth callback
// @Summary Discord OAuth callback
// @Description Handles the callback from Discord OAuth flow
// @Tags integrations
// @Accept json
// @Produce json
// @Param code query string true "Authorization code"
// @Param state query string true "State token"
// @Param guild_id query string false "Guild ID"
// @Success 302 {string} string "Redirect to frontend"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/public/integrations/discord/callback [get]
func (h *IntegrationOAuthHandler) DiscordOAuthCallback(c *gin.Context) {
	ctx := c.Request.Context()

	// Check for error from Discord
	if errParam := c.Query("error"); errParam != "" {
		errDesc := c.Query("error_description")
		logger.GetTxLogger(ctx).Error().
			Str("error", errParam).
			Str("error_description", errDesc).
			Msg("Discord OAuth error")
		c.Redirect(http.StatusFound, h.frontendURL+"/dashboard?error=discord_oauth_denied")
		return
	}

	// Get authorization code
	code := c.Query("code")
	if code == "" {
		c.Redirect(http.StatusFound, h.frontendURL+"/dashboard?error=missing_code")
		return
	}

	// Get state token
	stateToken := c.Query("state")
	if stateToken == "" {
		c.Redirect(http.StatusFound, h.frontendURL+"/dashboard?error=missing_state")
		return
	}

	// Validate state token
	stateData, err := h.integrationService.ValidateOAuthState(ctx, stateToken)
	if err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to validate OAuth state")
		c.Redirect(http.StatusFound, h.frontendURL+"/dashboard?error=invalid_state")
		return
	}

	// Exchange code for tokens
	oauthResp, err := h.integrationService.ExchangeDiscordCode(ctx, code)
	if err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to exchange Discord code")
		c.Redirect(http.StatusFound, h.frontendURL+"/dashboard?error=token_exchange_failed")
		return
	}

	// Store the integration
	_, err = h.integrationService.StoreDiscordIntegration(ctx, stateData, oauthResp)
	if err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to store Discord integration")
		c.Redirect(http.StatusFound, h.frontendURL+"/dashboard?error=storage_failed")
		return
	}

	// Redirect to frontend dashboard with success
	c.Redirect(http.StatusFound, h.frontendURL+"/dashboard?integration=discord&status=success")
}

// ListProjectIntegrations lists all integrations for a project
// @Summary List project integrations
// @Description Lists all integrations configured for a project
// @Tags integrations
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param project_id path string true "Project ID"
// @Security BearerAuth
// @Success 200 {array} models.ProjectIntegration
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/tenants/{tenant_id}/projects/{project_id}/integrations [get]
func (h *IntegrationOAuthHandler) ListProjectIntegrations(c *gin.Context) {
	ctx := c.Request.Context()

	// Get tenant ID from path
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id"})
		return
	}

	// Get project ID from path
	projectIDStr := c.Param("project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	// Get integrations
	integrations, err := h.integrationService.ListProjectIntegrations(ctx, tenantID, projectID)
	if err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to list integrations")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list integrations"})
		return
	}

	c.JSON(http.StatusOK, integrations)
}

// DeleteProjectIntegration deletes an integration
// @Summary Delete project integration
// @Description Deletes a specific integration from a project
// @Tags integrations
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param project_id path string true "Project ID"
// @Param app_type path string true "Integration type (e.g., slack)"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/tenants/{tenant_id}/projects/{project_id}/integrations/{app_type} [delete]
func (h *IntegrationOAuthHandler) DeleteProjectIntegration(c *gin.Context) {
	ctx := c.Request.Context()

	// Get tenant ID from path
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id"})
		return
	}

	// Get project ID from path
	projectIDStr := c.Param("project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	// Get integration type from path
	appType := c.Param("app_type")
	if appType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app_type is required"})
		return
	}

	// Convert app_type to ProjectIntegrationType
	var integrationType models.ProjectIntegrationType
	switch appType {
	case "slack":
		integrationType = models.ProjectIntegrationTypeSlack
	case "discord":
		integrationType = models.ProjectIntegrationTypeDiscord
	case "microsoft_teams":
		integrationType = models.ProjectIntegrationTypeTeams
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported integration type"})
		return
	}

	// Delete the integration
	err = h.integrationService.DeleteProjectIntegration(ctx, tenantID, projectID, integrationType)
	if err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to delete integration")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete integration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "integration deleted successfully"})
}
