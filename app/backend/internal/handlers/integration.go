package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/service"
)

type IntegrationHandler struct {
	integrationService *service.IntegrationService
}

func NewIntegrationHandler(integrationService *service.IntegrationService) *IntegrationHandler {
	return &IntegrationHandler{
		integrationService: integrationService,
	}
}

// CreateIntegration creates a new integration
// @Summary Create integration
// @Description Create a new third-party integration for a project
// @Tags integrations
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Param integration body models.CreateIntegrationRequest true "Integration creation request"
// @Success 201 {object} models.Integration
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/integrations [post]
func (h *IntegrationHandler) CreateIntegration(c *gin.Context) {
	tenantID := getTenantID(c)
	projectID := getProjectID(c)

	var req models.CreateIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	integration, err := h.integrationService.CreateIntegration(c.Request.Context(), tenantID, projectID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create integration", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, integration)
}

// GetIntegration retrieves a specific integration
// @Summary Get integration
// @Description Retrieve a specific integration by its ID
// @Tags integrations
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param id path string true "Integration ID"
// @Success 200 {object} models.Integration
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/integrations/{id} [get]
func (h *IntegrationHandler) GetIntegration(c *gin.Context) {
	tenantID := getTenantID(c)
	integrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	integration, err := h.integrationService.GetIntegration(c.Request.Context(), tenantID, integrationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	c.JSON(http.StatusOK, integration)
}

// ListIntegrations lists all integrations for a project
// @Summary List integrations
// @Description Retrieve all integrations for a project
// @Tags integrations
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Success 200 {object} object{integrations=[]models.Integration}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/integrations [get]
func (h *IntegrationHandler) ListIntegrations(c *gin.Context) {
	tenantID := getTenantID(c)
	projectID := getProjectID(c)

	// Parse query parameters
	var integrationType *models.IntegrationType
	if typeParam := c.Query("type"); typeParam != "" {
		iType := models.IntegrationType(typeParam)
		integrationType = &iType
	}

	var status *models.IntegrationStatus
	if statusParam := c.Query("status"); statusParam != "" {
		iStatus := models.IntegrationStatus(statusParam)
		status = &iStatus
	}

	integrations, err := h.integrationService.ListIntegrations(c.Request.Context(), tenantID, projectID, integrationType, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list integrations", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": integrations})
}

func (h *IntegrationHandler) UpdateIntegration(c *gin.Context) {
	tenantID := getTenantID(c)
	integrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	var req models.UpdateIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	integration, err := h.integrationService.UpdateIntegration(c.Request.Context(), tenantID, integrationID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update integration", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, integration)
}

func (h *IntegrationHandler) DeleteIntegration(c *gin.Context) {
	tenantID := getTenantID(c)
	integrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	if err := h.integrationService.DeleteIntegration(c.Request.Context(), tenantID, integrationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete integration", "details": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *IntegrationHandler) TestIntegration(c *gin.Context) {
	tenantID := getTenantID(c)
	integrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	if err := h.integrationService.TestIntegration(c.Request.Context(), tenantID, integrationID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Integration test failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Integration test successful"})
}

// Helper functions (these would be implemented elsewhere in your codebase)
func getTenantID(c *gin.Context) uuid.UUID {
	// Extract tenant ID from context (set by middleware)
	if tenantID, exists := c.Get("tenant_id"); exists {
		if id, ok := tenantID.(uuid.UUID); ok {
			return id
		}
	}
	// Return a default or handle error appropriately
	return uuid.Nil
}

func getProjectID(c *gin.Context) uuid.UUID {
	// Extract project ID from context or query parameter
	if projectID, exists := c.Get("project_id"); exists {
		if id, ok := projectID.(uuid.UUID); ok {
			return id
		}
	}

	// Try to get from query parameter
	if projectIDStr := c.Query("project_id"); projectIDStr != "" {
		if id, err := uuid.Parse(projectIDStr); err == nil {
			return id
		}
	}

	// Return a default or handle error appropriately
	return uuid.Nil
}

// New handler methods for enhanced integration system

// List integration categories with templates
func (h *IntegrationHandler) ListIntegrationCategories(c *gin.Context) {
	// Return static/config-based categories for supported integrations
	categories := []map[string]interface{}{
		{
			"id":           "communication",
			"name":         "communication",
			"display_name": "Communication",
			"description":  "Connect with your team's communication tools",
			"icon":         "message-square",
			"sort_order":   1,
			"is_active":    true,
			"templates": []map[string]interface{}{
				{
					"id":                uuid.New().String(),
					"type":              "slack",
					"name":              "slack",
					"display_name":      "Slack",
					"description":       "Send notifications and updates to Slack channels",
					"logo_url":          "https://cdn.brandfetch.io/idSUrLOWbH/theme/dark/symbol.svg",
					"website_url":       "https://slack.com",
					"documentation_url": "https://api.slack.com/docs",
					"auth_method":       "oauth",
					"is_featured":       true,
					"is_active":         true,
					"sort_order":        1,
				},
				{
					"id":                uuid.New().String(),
					"type":              "discord",
					"name":              "discord",
					"display_name":      "Discord",
					"description":       "Send notifications to Discord servers",
					"logo_url":          "https://cdn.brandfetch.io/idS7wJN6pI/theme/dark/symbol.svg",
					"website_url":       "https://discord.com",
					"documentation_url": "https://discord.com/developers/docs",
					"auth_method":       "oauth",
					"is_featured":       false,
					"is_active":         false, // Coming soon
					"sort_order":        2,
				},
				{
					"id":                uuid.New().String(),
					"type":              "microsoft_teams",
					"name":              "microsoft_teams",
					"display_name":      "Microsoft Teams",
					"description":       "Send notifications to Microsoft Teams channels",
					"logo_url":          "https://cdn.brandfetch.io/idchmboHEZ/theme/dark/symbol.svg",
					"website_url":       "https://teams.microsoft.com",
					"documentation_url": "https://docs.microsoft.com/en-us/microsoftteams/",
					"auth_method":       "oauth",
					"is_featured":       false,
					"is_active":         false, // Coming soon
					"sort_order":        3,
				},
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

// List integration templates
func (h *IntegrationHandler) ListIntegrationTemplates(c *gin.Context) {
	var categoryID *uuid.UUID
	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		if id, err := uuid.Parse(categoryIDStr); err == nil {
			categoryID = &id
		}
	}

	var featured *bool
	if featuredStr := c.Query("featured"); featuredStr != "" {
		f := featuredStr == "true"
		featured = &f
	}

	templates, err := h.integrationService.ListIntegrationTemplates(c.Request.Context(), categoryID, featured)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list integration templates", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

// Get integration template by type
func (h *IntegrationHandler) GetIntegrationTemplate(c *gin.Context) {
	integrationType := models.IntegrationType(c.Param("type"))

	template, err := h.integrationService.GetIntegrationTemplateByType(c.Request.Context(), integrationType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration template not found"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// Start OAuth flow
func (h *IntegrationHandler) StartOAuth(c *gin.Context) {
	tenantID := getTenantID(c)
	projectID := getProjectID(c)

	var req models.OAuthStartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	oauthURL, state, err := h.integrationService.StartOAuthFlow(c.Request.Context(), tenantID, projectID, req.IntegrationType, req.RedirectURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start OAuth flow", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"oauth_url": oauthURL,
		"state":     state,
	})
}

// Handle OAuth callback
func (h *IntegrationHandler) HandleOAuthCallback(c *gin.Context) {
	integrationType := models.IntegrationType(c.Param("type"))

	var req models.OAuthCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	integration, err := h.integrationService.HandleOAuthCallback(c.Request.Context(), integrationType, req.Code, req.State)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete OAuth flow", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, integration)
}

// Test integration connection
func (h *IntegrationHandler) TestIntegrationConnection(c *gin.Context) {
	tenantID := getTenantID(c)
	integrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	err = h.integrationService.TestIntegration(c.Request.Context(), tenantID, integrationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Integration test failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "success", "message": "Integration connection test passed"})
}

// Get integration metrics
func (h *IntegrationHandler) GetIntegrationMetrics(c *gin.Context) {
	tenantID := getTenantID(c)
	integrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	metrics, err := h.integrationService.GetIntegrationMetrics(c.Request.Context(), tenantID, integrationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get integration metrics", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// Enhanced list integrations with template information
func (h *IntegrationHandler) ListIntegrationsWithTemplates(c *gin.Context) {
	tenantID := getTenantID(c)
	projectID := getProjectID(c)

	// Parse query parameters
	var integrationType *models.IntegrationType
	if typeParam := c.Query("type"); typeParam != "" {
		iType := models.IntegrationType(typeParam)
		integrationType = &iType
	}

	var status *models.IntegrationStatus
	if statusParam := c.Query("status"); statusParam != "" {
		iStatus := models.IntegrationStatus(statusParam)
		status = &iStatus
	}

	integrations, err := h.integrationService.ListIntegrationsWithTemplates(c.Request.Context(), tenantID, projectID, integrationType, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list integrations", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"integrations": integrations})
}
