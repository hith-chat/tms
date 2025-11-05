package service

import (
	"context"
	"testing"
	"time"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAIServiceFeatureToggles(t *testing.T) {
	t.Parallel()

	aiCfg := &config.AIConfig{Enabled: true, APIKey: "key"}
	agenticCfg := &config.AgenticConfig{Enabled: true, GreetingDetection: true, KnowledgeResponses: true}

	svc := &AIService{
		config:        aiCfg,
		agenticConfig: agenticCfg,
	}

	require.True(t, svc.IsEnabled())
	require.True(t, svc.IsAgenticBehaviorEnabled())
	require.True(t, svc.IsGreetingDetectionEnabled())
	require.True(t, svc.IsKnowledgeResponsesEnabled())
}

func TestAIServiceFeatureTogglesDisabled(t *testing.T) {
	t.Parallel()

	svc := &AIService{config: &config.AIConfig{Enabled: false}, agenticConfig: &config.AgenticConfig{Enabled: false}}

	require.False(t, svc.IsEnabled())
	require.False(t, svc.IsAgenticBehaviorEnabled())
	require.False(t, svc.IsGreetingDetectionEnabled())
	require.False(t, svc.IsKnowledgeResponsesEnabled())
}

func TestAIServiceShouldHandleSession(t *testing.T) {
	t.Parallel()

	tenantID := uuid.New()
	session := &models.ChatSession{
		ID:        uuid.New(),
		TenantID:  tenantID,
		ProjectID: uuid.New(),
		CreatedAt: time.Now().Add(-time.Minute),
	}

	svc := &AIService{
		config: &config.AIConfig{Enabled: true, APIKey: "key"},
	}

	require.True(t, svc.ShouldHandleSession(context.Background(), session))

	session.AssignedAgentID = new(uuid.UUID)
	require.False(t, svc.ShouldHandleSession(context.Background(), session))

	svc.config.Enabled = false
	require.False(t, svc.ShouldHandleSession(context.Background(), session))
}

func TestAIServiceGenerateAgentIDDeterministic(t *testing.T) {
	t.Parallel()

	svc := &AIService{}
	sessionID := uuid.MustParse("d9428888-122b-11e1-b85c-61cd3cbb3210")

	id1 := svc.generateAIAgentID(sessionID)
	id2 := svc.generateAIAgentID(sessionID)

	require.Equal(t, id1, id2)
	require.NotEqual(t, uuid.Nil, id1)
}
