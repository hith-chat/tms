package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/service"
)

type AIHandler struct {
	aiService *service.AIService
}

func NewAIHandler(aiService *service.AIService) *AIHandler {
	return &AIHandler{
		aiService: aiService,
	}
}

// GetAIStatus returns the current AI service status
func (h *AIHandler) GetAIStatus(c *gin.Context) {
	// Check if user has permission to view AI settings
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	status := map[string]interface{}{
		"enabled":    h.aiService.IsEnabled(),
		"tenant_id":  tenantID,
		"project_id": projectID,
	}

	if h.aiService.IsEnabled() {
		// Don't expose sensitive config details, just availability
		status["provider"] = "configured"
		status["model"] = "available"
	}

	c.JSON(http.StatusOK, status)
}

// GetAICapabilities returns what the AI system can do
func (h *AIHandler) GetAICapabilities(c *gin.Context) {
	capabilities := map[string]interface{}{
		"features": []string{
			"automatic_responses",
			"human_handoff",
			"keyword_detection",
			"context_awareness",
		},
		"supported_providers": []string{
			"openai",
			"anthropic",
			"azure",
		},
		"handoff_triggers": []string{
			"manual_request",
			"keyword_detection",
			"timeout_based",
			"complexity_threshold",
		},
	}

	c.JSON(http.StatusOK, capabilities)
}

// GetAIMetrics returns AI usage metrics for the project
func (h *AIHandler) GetAIMetrics(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	// TODO: Implement metrics collection
	// For now, return placeholder metrics
	metrics := map[string]interface{}{
		"tenant_id":                tenantID,
		"project_id":               projectID,
		"period":                   "last_30_days",
		"ai_responses_sent":        0,
		"sessions_handled":         0,
		"handoffs_triggered":       0,
		"average_response_time_ms": 0,
		"customer_satisfaction":    nil,
	}

	c.JSON(http.StatusOK, metrics)
}

// AcceptHandoff handles an agent accepting a handoff request
func (h *AIHandler) AcceptHandoff(c *gin.Context) {
	sessionIDStr := c.Param("sessionId")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)
	agentID := middleware.GetAgentID(c)
	
	// Use the AI service to handle the handoff acceptance
	err = h.aiService.AcceptHandoff(c.Request.Context(), tenantID, projectID, sessionID, agentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}
		if strings.Contains(err.Error(), "already assigned") {
			c.JSON(http.StatusConflict, gin.H{"error": "Session already assigned to an agent"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept handoff"})
		return
	}

	response := map[string]interface{}{
		"success":     true,
		"session_id":  sessionID,
		"agent_id":    agentID,
		"tenant_id":   tenantID,
		"project_id":  projectID,
		"accepted_at": time.Now().UTC().Format(time.RFC3339),
		"message":     "Handoff accepted successfully",
	}

	c.JSON(http.StatusOK, response)
}

// DeclineHandoff handles an agent declining a handoff request
func (h *AIHandler) DeclineHandoff(c *gin.Context) {
	sessionIDStr := c.Param("sessionId")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)
	agentID := middleware.GetAgentID(c)
	
	// Use the AI service to handle the handoff decline
	err = h.aiService.DeclineHandoff(c.Request.Context(), tenantID, projectID, sessionID, agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decline handoff"})
		return
	}

	response := map[string]interface{}{
		"success":     true,
		"session_id":  sessionID,
		"agent_id":    agentID,
		"tenant_id":   tenantID,
		"project_id":  projectID,
		"declined_at": time.Now().UTC().Format(time.RFC3339),
		"message":     "Handoff declined successfully",
	}

	c.JSON(http.StatusOK, response)
}
