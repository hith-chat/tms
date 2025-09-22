package service

import (
	"context"
	"fmt"
	"math"

	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/google/uuid"
)

// TokenUsageMetrics captures raw token usage reported by AI providers.
type TokenUsageMetrics struct {
	PromptTokens     int64
	CompletionTokens int64
	TotalTokens      int64
}

// UsageDeductionInput contains the data required to deduct credits for AI usage.
type UsageDeductionInput struct {
	TenantID  uuid.UUID
	ProjectID uuid.UUID
	Model     string
	SessionID *uuid.UUID
	RequestID string
	Metrics   TokenUsageMetrics
}

// UsageDeductionResult contains the outcome of a usage deduction attempt.
type UsageDeductionResult struct {
	TransactionID  int64
	ChargedCredits int64
	MarkupPercent  float64
	Metrics        TokenUsageMetrics
	BalanceAfter   int64
}

// AIUsageService handles converting token usage into credit deductions.
type AIUsageService struct {
	creditsRepo   repo.CreditsRepository
	markupPercent float64
}

const defaultMarkupPercent = 0.21 // 21% markup on consumed tokens

// NewAIUsageService creates a new instance of AIUsageService.
func NewAIUsageService(creditsRepo repo.CreditsRepository) *AIUsageService {
	return &AIUsageService{
		creditsRepo:   creditsRepo,
		markupPercent: defaultMarkupPercent,
	}
}

// DeductUsage converts the supplied token metrics into a credit deduction and records it.
func (s *AIUsageService) DeductUsage(ctx context.Context, input UsageDeductionInput) (*UsageDeductionResult, error) {
	if s == nil || s.creditsRepo == nil {
		return nil, fmt.Errorf("usage service not configured")
	}

	metrics := input.Metrics
	if metrics.TotalTokens <= 0 {
		metrics.TotalTokens = metrics.PromptTokens + metrics.CompletionTokens
	}

	if metrics.TotalTokens <= 0 {
		return nil, fmt.Errorf("no token usage reported")
	}

	chargedCredits := s.calculateCharge(metrics.TotalTokens)
	if chargedCredits <= 0 {
		return nil, fmt.Errorf("calculated charge is zero")
	}

	description := fmt.Sprintf(
		"AI usage for model %s (prompt=%d, completion=%d, total=%d, markup=%.0f%%)",
		input.Model,
		metrics.PromptTokens,
		metrics.CompletionTokens,
		metrics.TotalTokens,
		s.markupPercent*100,
	)

	if input.SessionID != nil {
		description = fmt.Sprintf("%s | session=%s", description, input.SessionID.String())
	}

	if input.RequestID != "" {
		description = fmt.Sprintf("%s | request=%s", description, input.RequestID)
	}

	tx, err := s.creditsRepo.DeductCredits(ctx, input.TenantID, chargedCredits, models.TransactionTypeAIUsage, description)
	if err != nil {
		return nil, err
	}

	result := &UsageDeductionResult{
		TransactionID:  tx.ID,
		ChargedCredits: chargedCredits,
		MarkupPercent:  s.markupPercent * 100,
		Metrics:        metrics,
		BalanceAfter:   tx.BalanceAfter,
	}

	return result, nil
}

func (s *AIUsageService) calculateCharge(totalTokens int64) int64 {
	multiplier := 1 + s.markupPercent
	charged := math.Ceil(float64(totalTokens) * multiplier)
	if charged < 1 {
		charged = 1
	}
	return int64(charged)
}
