package service

import (
	"context"
	"strings"
	"testing"

	"github.com/bareuptime/tms/internal/config"
	"github.com/stretchr/testify/require"
)

func TestDetectAgentRequestDisabled(t *testing.T) {
	t.Parallel()

	svc := NewAgentRequestDetectionService(&config.AgenticConfig{
		AgentRequestDetection: false,
	})

	result, err := svc.DetectAgentRequest(context.Background(), "Please connect me to someone")
	require.NoError(t, err)

	require.False(t, result.IsAgentRequest)
	require.Contains(t, result.Reasoning, "Agent request detection is disabled")
	require.Zero(t, result.Confidence)
}

func TestDetectAgentRequestUrgentClassification(t *testing.T) {
	t.Parallel()

	svc := NewAgentRequestDetectionService(&config.AgenticConfig{
		AgentRequestDetection: true,
		AgentRequestThreshold: 0.6,
	})

	message := "I need to speak to a human agent right now, it's urgent!"

	result, err := svc.DetectAgentRequest(context.Background(), message)
	require.NoError(t, err)

	require.True(t, result.IsAgentRequest)
	require.Equal(t, AgentRequestTypeUrgent, result.RequestType)
	require.Equal(t, UrgencyHigh, result.Urgency)
	require.Contains(t, result.Keywords, "agent")
	require.GreaterOrEqual(t, result.Confidence, 0.6)
	require.Contains(t, result.Reasoning[len(result.Reasoning)-1], "Confidence")
}

func TestDetectAgentRequestBillingClassification(t *testing.T) {
	t.Parallel()

	svc := NewAgentRequestDetectionService(&config.AgenticConfig{
		AgentRequestDetection: true,
		AgentRequestThreshold: 0.6,
	})

	message := "I was double charged and need a refund, please connect me with customer service immediately"

	result, err := svc.DetectAgentRequest(context.Background(), message)
	require.NoError(t, err)

	require.True(t, result.IsAgentRequest)
	require.Equal(t, AgentRequestTypeBilling, result.RequestType)
	require.Contains(t, strings.Join(result.Keywords, " "), "billing")
	require.NotEqual(t, UrgencyLow, result.Urgency)
}
