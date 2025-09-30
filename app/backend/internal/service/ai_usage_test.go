package service

import (
	"context"
	"errors"
	"testing"

	"github.com/bareuptime/tms/internal/db"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type mockCreditsRepository struct {
	repo.CreditsRepository
	expectedAmount int64
	result         *db.CreditTransaction
	err            error
	calls          int
}

func (m *mockCreditsRepository) DeductCredits(ctx context.Context, tenantID uuid.UUID, amount int64, transactionType, description string) (*db.CreditTransaction, error) {
	m.calls++
	m.expectedAmount = amount
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func TestAIUsageServiceDeductUsageSuccess(t *testing.T) {
	t.Parallel()

	tenantID := uuid.New()
	creditsRepo := &mockCreditsRepository{
		result: &db.CreditTransaction{
			ID:           42,
			BalanceAfter: 880,
		},
	}

	svc := NewAIUsageService(creditsRepo)

	res, err := svc.DeductUsage(context.Background(), UsageDeductionInput{
		TenantID: tenantID,
		Model:    "gpt-4",
		Metrics: TokenUsageMetrics{
			PromptTokens:     60,
			CompletionTokens: 40,
		},
	})
	require.NoError(t, err)

	require.EqualValues(t, 42, res.TransactionID)
	require.EqualValues(t, 121, creditsRepo.expectedAmount)
	require.Equal(t, float64(21), res.MarkupPercent)
	require.Equal(t, int64(880), res.BalanceAfter)
	require.Equal(t, 1, creditsRepo.calls)
}

func TestAIUsageServiceRejectsZeroUsage(t *testing.T) {
	t.Parallel()

	svc := NewAIUsageService(&mockCreditsRepository{})

	_, err := svc.DeductUsage(context.Background(), UsageDeductionInput{
		TenantID: uuid.New(),
		Model:    "gpt-4",
		Metrics:  TokenUsageMetrics{},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "no token usage")
}

func TestAIUsageServicePropagatesRepositoryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("deduct failed")
	creditsRepo := &mockCreditsRepository{
		err: expectedErr,
	}

	svc := NewAIUsageService(creditsRepo)

	_, err := svc.DeductUsage(context.Background(), UsageDeductionInput{
		TenantID: uuid.New(),
		Model:    "gpt-4",
		Metrics: TokenUsageMetrics{
			PromptTokens:     10,
			CompletionTokens: 5,
		},
	})
	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}
