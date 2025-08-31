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
