package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SettingsHandler struct {
	settingsRepo *repo.SettingsRepository
}

func NewSettingsHandler(settingsRepo *repo.SettingsRepository) *SettingsHandler {
	return &SettingsHandler{
		settingsRepo: settingsRepo,
	}
}

// BrandingSettings represents branding configuration
type BrandingSettings struct {
	CompanyName          string `json:"company_name"`
	About                string `json:"about"`
	LogoURL              string `json:"logo_url"`
	SupportURL           string `json:"support_url"`
	PrimaryColor         string `json:"primary_color"`
	AccentColor          string `json:"accent_color"`
	SecondaryColor       string `json:"secondary_color"`
	CustomCSS            string `json:"custom_css"`
	FaviconURL           string `json:"favicon_url"`
	HeaderLogoHeight     int    `json:"header_logo_height"`
	EnableCustomBranding bool   `json:"enable_custom_branding"`
}

// AutomationSettings represents automation configuration
type AutomationSettings struct {
	EnableAutoAssignment     bool   `json:"enable_auto_assignment"`
	AssignmentStrategy       string `json:"assignment_strategy"`
	MaxTicketsPerAgent       int    `json:"max_tickets_per_agent"`
	EnableEscalation         bool   `json:"enable_escalation"`
	EscalationThresholdHours int    `json:"escalation_threshold_hours"`
	EnableAutoReply          bool   `json:"enable_auto_reply"`
	AutoReplyTemplate        string `json:"auto_reply_template"`
}

// AboutMeSettings represents about me configuration
type AboutMeSettings struct {
	Content string `json:"content"`
}

// GetBrandingSettings retrieves branding settings for a tenant
// @Summary Get branding settings
// @Description Retrieve branding configuration settings for the current tenant
// @Tags settings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Success 200 {object} BrandingSettings
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/settings/branding [get]
func (h *SettingsHandler) GetBrandingSettings(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectIDStr, _ := c.Params.Get("project_id")
	projectUUID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	settings, httpStatusCode, err := h.settingsRepo.GetSetting(context.Background(), tenantID, projectUUID, "branding_settings")
	if err != nil {
		c.JSON(httpStatusCode, gin.H{"error": "Failed to retrieve branding settings"})
		return
	}

	// Convert map to BrandingSettings struct
	settingsJSON, _ := json.Marshal(settings)
	var brandingSettings BrandingSettings
	json.Unmarshal(settingsJSON, &brandingSettings)

	c.JSON(http.StatusOK, brandingSettings)
}

// UpdateBrandingSettings updates branding settings for a tenant
// @Summary Update branding settings
// @Description Update branding configuration settings for a project
// @Tags settings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id path string true "Project ID"
// @Param branding body BrandingSettings true "Branding settings to update"
// @Success 200 {object} BrandingSettings
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/settings/projects/{project_id}/branding [put]
func (h *SettingsHandler) UpdateBrandingSettings(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectIDStr, _ := c.Params.Get("project_id")
	projectUUID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var brandingSettings BrandingSettings
	if err := c.ShouldBindJSON(&brandingSettings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Convert struct to map
	settingsJSON, _ := json.Marshal(brandingSettings)
	var settingsMap map[string]interface{}
	json.Unmarshal(settingsJSON, &settingsMap)

	err = h.settingsRepo.UpdateSetting(context.Background(), tenantID, projectUUID, "branding_settings", settingsMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update branding settings"})
		return
	}

	c.JSON(http.StatusOK, brandingSettings)
}

// GetAutomationSettings retrieves automation settings for a tenant
// @Summary Get automation settings
// @Description Retrieve automation configuration settings for a project
// @Tags settings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id path string true "Project ID"
// @Success 200 {object} AutomationSettings
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/settings/projects/{project_id}/automation [get]
func (h *SettingsHandler) GetAutomationSettings(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectIDStr, _ := c.Params.Get("project_id")
	projectUUID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	settings, httpStatusCode, err := h.settingsRepo.GetSetting(context.Background(), tenantID, projectUUID, "automation_settings")
	if err != nil {
		c.JSON(httpStatusCode, gin.H{"error": "Failed to retrieve automation settings"})
		return
	}

	// Convert map to AutomationSettings struct
	settingsJSON, _ := json.Marshal(settings)
	var automationSettings AutomationSettings
	json.Unmarshal(settingsJSON, &automationSettings)

	c.JSON(http.StatusOK, automationSettings)
}

// UpdateAutomationSettings updates automation settings for a tenant
// @Summary Update automation settings
// @Description Update automation configuration settings for a project
// @Tags settings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id path string true "Project ID"
// @Param automation body AutomationSettings true "Automation settings to update"
// @Success 200 {object} AutomationSettings
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/settings/projects/{project_id}/automation [put]
func (h *SettingsHandler) UpdateAutomationSettings(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectIDStr, _ := c.Params.Get("project_id")
	projectUUID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var automationSettings AutomationSettings
	if err := c.ShouldBindJSON(&automationSettings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Convert struct to map
	settingsJSON, _ := json.Marshal(automationSettings)
	var settingsMap map[string]interface{}
	json.Unmarshal(settingsJSON, &settingsMap)

	err = h.settingsRepo.UpdateSetting(context.Background(), tenantID, projectUUID, "automation_settings", settingsMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update automation settings"})
		return
	}

	c.JSON(http.StatusOK, automationSettings)
}

// GetAboutMeSettings retrieves about me settings for a tenant
// @Summary Get about me settings
// @Description Retrieve about me configuration settings for a project
// @Tags settings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id path string true "Project ID"
// @Success 200 {object} AboutMeSettings
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/settings/projects/{project_id}/about-me [get]
func (h *SettingsHandler) GetAboutMeSettings(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectIDStr, _ := c.Params.Get("project_id")
	projectUUID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	settings, httpStatusCode, err := h.settingsRepo.GetSetting(context.Background(), tenantID, projectUUID, "about_me")
	if err != nil {
		c.JSON(httpStatusCode, gin.H{"error": "Failed to retrieve about me settings"})
		return
	}

	// Convert map to AboutMeSettings struct
	settingsJSON, _ := json.Marshal(settings)
	var aboutMeSettings AboutMeSettings
	json.Unmarshal(settingsJSON, &aboutMeSettings)

	c.JSON(http.StatusOK, aboutMeSettings)
}

// UpdateAboutMeSettings updates about me settings for a tenant
// @Summary Update about me settings
// @Description Update about me configuration settings for a project
// @Tags settings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id path string true "Project ID"
// @Param about_me body AboutMeSettings true "About me settings to update"
// @Success 200 {object} AboutMeSettings
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/settings/projects/{project_id}/about-me [put]
func (h *SettingsHandler) UpdateAboutMeSettings(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectIDStr, _ := c.Params.Get("project_id")
	projectUUID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var aboutMeSettings AboutMeSettings
	if err := c.ShouldBindJSON(&aboutMeSettings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Convert struct to map
	settingsJSON, _ := json.Marshal(aboutMeSettings)
	var settingsMap map[string]interface{}
	json.Unmarshal(settingsJSON, &settingsMap)

	err = h.settingsRepo.UpdateSetting(context.Background(), tenantID, projectUUID, "about_me", settingsMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update about me settings"})
		return
	}

	c.JSON(http.StatusOK, aboutMeSettings)
}
