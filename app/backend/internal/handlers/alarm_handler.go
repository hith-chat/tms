package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/service"
)

// AlarmHandler handles alarm-related HTTP requests
type AlarmHandler struct {
	howlingAlarmService *service.HowlingAlarmService
}

// NewAlarmHandler creates a new alarm handler
func NewAlarmHandler(howlingAlarmService *service.HowlingAlarmService) *AlarmHandler {
	return &AlarmHandler{
		howlingAlarmService: howlingAlarmService,
	}
}

// GetActiveAlarms retrieves active alarms for a project
// @Summary Get active alarms
// @Description Retrieve all active alarms for a specific project
// @Tags alarms
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Success 200 {object} object{alarms=[]models.Alarm,total=int}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/alarms/active [get]
func (h *AlarmHandler) GetActiveAlarms(c *gin.Context) {
	// Extract tenant ID from JWT context
	tenantID := middleware.GetTenantID(c)

	// Extract project ID from URL parameters
	projectID := middleware.GetProjectID(c)

	// TODO: Validate that the project belongs to the tenant
	// This requires a project validation service

	// Get active alarms for the tenant and project
	activeAlarms, err := h.howlingAlarmService.GetActiveAlarms(c.Request.Context(), tenantID, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active alarms: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alarms": activeAlarms,
		"total":  len(activeAlarms),
	})
}

// GetAlarmStats retrieves alarm statistics for a project
// @Summary Get alarm statistics
// @Description Retrieve alarm statistics and metrics for a project
// @Tags alarms
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Success 200 {object} models.AlarmStats
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/alarms/stats [get]
func (h *AlarmHandler) GetAlarmStats(c *gin.Context) {
	// Extract tenant ID from JWT context
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	// TODO: Validate that the project belongs to the tenant

	// Get alarm statistics for the tenant and project
	stats, err := h.howlingAlarmService.GetAlarmStats(c.Request.Context(), tenantID, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get alarm stats: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// AcknowledgeAlarm acknowledges a specific alarm
// @Summary Acknowledge alarm
// @Description Acknowledge a specific alarm to mark it as seen/handled
// @Tags alarms
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Param alarm_id path string true "Alarm ID"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/alarms/{alarm_id}/acknowledge [post]
func (h *AlarmHandler) AcknowledgeAlarm(c *gin.Context) {
	// Extract tenant ID from JWT context
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)
	agentID := middleware.GetAgentID(c)

	// Extract alarm ID from URL parameters
	alarmIDParam := c.Param("alarmId")
	alarmID, err := uuid.Parse(alarmIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alarm ID"})
		return
	}

	// Parse optional response message
	var requestBody struct {
		Response string `json:"response,omitempty"`
	}
	c.ShouldBindJSON(&requestBody)

	// Acknowledge the alarm
	err = h.howlingAlarmService.AcknowledgeAlarm(c.Request.Context(), tenantID, alarmID, agentID, requestBody.Response)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to acknowledge alarm: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Alarm acknowledged successfully",
		"alarm_id":   alarmID,
		"agent_id":   agentID,
		"project_id": projectID,
		"tenant_id":  tenantID,
	})
}

// GetNotificationPreferences retrieves notification preferences for an agent
// @Summary Get notification preferences
// @Description Retrieve notification preferences for alarm notifications
// @Tags alarms
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param agent_id header string true "Agent ID"
// @Success 200 {object} object{agent_id=string,tenant_id=string,desktop_enabled=boolean,email_enabled=boolean,browser_enabled=boolean,sound_enabled=boolean,vibration_enabled=boolean,escalation_timeout_minutes=int}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/alarms/notification-preferences [get]
func (h *AlarmHandler) GetNotificationPreferences(c *gin.Context) {
	// Extract tenant ID from JWT context
	tenantIDInterface, exists := c.Get("tenantID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	tenantID, ok := tenantIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid tenant ID"})
		return
	}

	// Extract agent ID from URL parameters or JWT context
	agentIDParam := c.Param("agentId")
	agentID, err := uuid.Parse(agentIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	// TODO: Validate that the agent belongs to the tenant
	// TODO: Implement notification preferences service

	// For now, return default preferences
	defaultPreferences := gin.H{
		"agent_id":                   agentID,
		"tenant_id":                  tenantID,
		"desktop_enabled":            true,
		"email_enabled":              true,
		"browser_enabled":            true,
		"sound_enabled":              true,
		"vibration_enabled":          false,
		"escalation_timeout_minutes": 5,
	}

	c.JSON(http.StatusOK, defaultPreferences)
}

// UpdateNotificationPreferences updates notification preferences for an agent
// @Summary Update notification preferences
// @Description Update notification preferences for alarm notifications
// @Tags alarms
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param agent_id header string true "Agent ID"
// @Param preferences body object{desktop_enabled=boolean,email_enabled=boolean,browser_enabled=boolean,sound_enabled=boolean,vibration_enabled=boolean,escalation_timeout_minutes=int} true "Notification preferences to update"
// @Success 200 {object} object{message=string,agent_id=string,tenant_id=string,preferences=object}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/alarms/notification-preferences [put]
func (h *AlarmHandler) UpdateNotificationPreferences(c *gin.Context) {
	// Extract tenant ID from JWT context
	tenantID := middleware.GetTenantID(c)
	agentID := middleware.GetAgentID(c)

	// Parse the request body
	var preferences struct {
		DesktopEnabled           *bool `json:"desktop_enabled,omitempty"`
		EmailEnabled             *bool `json:"email_enabled,omitempty"`
		BrowserEnabled           *bool `json:"browser_enabled,omitempty"`
		SoundEnabled             *bool `json:"sound_enabled,omitempty"`
		VibrationEnabled         *bool `json:"vibration_enabled,omitempty"`
		EscalationTimeoutMinutes *int  `json:"escalation_timeout_minutes,omitempty"`
	}

	if err := c.ShouldBindJSON(&preferences); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// TODO: Validate that the agent belongs to the tenant
	// TODO: Implement notification preferences service to save to database

	// For now, just return the updated preferences
	response := gin.H{
		"message":     "Notification preferences updated successfully",
		"agent_id":    agentID,
		"tenant_id":   tenantID,
		"preferences": preferences,
	}

	c.JSON(http.StatusOK, response)
}
