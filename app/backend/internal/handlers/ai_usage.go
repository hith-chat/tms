package handlers

import (
	"fmt"
	"net/http"

	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AIUsageHandler exposes endpoints for AI usage billing.
type AIUsageHandler struct {
	usageService *service.AIUsageService
	s2sKey       string
}

// NewAIUsageHandler constructs a new AIUsageHandler.
func NewAIUsageHandler(usageService *service.AIUsageService, s2sKey string) *AIUsageHandler {
	return &AIUsageHandler{
		usageService: usageService,
		s2sKey:       s2sKey,
	}
}

type deductUsageRequest struct {
	Model            string `json:"model" binding:"required"`
	PromptTokens     *int64 `json:"prompt_tokens"`
	CompletionTokens *int64 `json:"completion_tokens"`
	TotalTokens      *int64 `json:"total_tokens"`
	SessionID        string `json:"session_id"`
	RequestID        string `json:"request_id"`
}

// DeductUsage handles credit deductions for AI token usage.
// @Summary Deduct AI usage credits
// @Description Deduct credits based on AI token usage for a specific tenant
// @Tags ai-usage
// @Accept json
// @Produce json
// @Security S2SAuth
// @Param tenant_id path string true "Tenant ID"
// @Param usage body deductUsageRequest true "AI usage deduction request"
// @Success 200 {object} object{message=string,remaining_credits=int64}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/ai-usage/{tenant_id}/deduct [post]
func (h *AIUsageHandler) DeductUsage(c *gin.Context) {
	if h.usageService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "usage service unavailable"})
		return
	}

	if err := h.ensureS2SAuthorized(c); err != nil {
		return
	}

	tenantIDStr := c.Param("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant id"})
		return
	}

	projectID := middleware.GetProjectID(c)
	if projectID == uuid.Nil {
		projectParam := c.Param("project_id")
		projectID, err = uuid.Parse(projectParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
			return
		}
	}

	var req deductUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	promptTokens := int64(0)
	if req.PromptTokens != nil {
		promptTokens = *req.PromptTokens
	}

	completionTokens := int64(0)
	if req.CompletionTokens != nil {
		completionTokens = *req.CompletionTokens
	}

	totalTokens := int64(0)
	if req.TotalTokens != nil {
		totalTokens = *req.TotalTokens
	}

	sessionID := (*uuid.UUID)(nil)
	if req.SessionID != "" {
		parsed, err := uuid.Parse(req.SessionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
			return
		}
		sessionID = &parsed
	}

	result, err := h.usageService.DeductUsage(c.Request.Context(), service.UsageDeductionInput{
		TenantID:  tenantID,
		ProjectID: projectID,
		Model:     req.Model,
		SessionID: sessionID,
		RequestID: req.RequestID,
		Metrics: service.TokenUsageMetrics{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      totalTokens,
		},
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transaction_id":    result.TransactionID,
		"charged_credits":   result.ChargedCredits,
		"markup_percent":    result.MarkupPercent,
		"prompt_tokens":     result.Metrics.PromptTokens,
		"completion_tokens": result.Metrics.CompletionTokens,
		"total_tokens":      result.Metrics.TotalTokens,
		"balance_after":     result.BalanceAfter,
		"request_id":        req.RequestID,
		"session_id":        req.SessionID,
	})
}

func (h *AIUsageHandler) ensureS2SAuthorized(c *gin.Context) error {
	if h.s2sKey == "" {
		return nil
	}

	key := c.GetHeader("X-S2S-KEY")
	if key == "" || key != h.s2sKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing s2s key"})
		return fmt.Errorf("unauthorized")
	}

	return nil
}
