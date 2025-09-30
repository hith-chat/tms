package service

import (
	"testing"

	"github.com/bareuptime/tms/internal/models"
	"github.com/stretchr/testify/require"
)

func TestApplyWidgetDefaults(t *testing.T) {
	t.Parallel()

	cfg := &models.CreateChatWidgetRequest{}
	svc := &AIBuilderService{}

	svc.applyWidgetDefaults(cfg)

	require.Equal(t, "#2563eb", cfg.PrimaryColor)
	require.Equal(t, "#f3f4f6", cfg.SecondaryColor)
	require.Equal(t, "#ffffff", cfg.BackgroundColor)
	require.Equal(t, "bottom-right", cfg.Position)
	require.Equal(t, "Hello! How can we help you today?", cfg.WelcomeMessage)
	require.Equal(t, "We are currently offline. Leave us a message and we will respond soon.", cfg.OfflineMessage)
	require.Equal(t, "Support Agent", cfg.AgentName)
	require.True(t, cfg.UseAI)
}

func TestNormalizeURL(t *testing.T) {
	t.Parallel()

	require.Equal(t, "https://example.com/path", normalizeURL(" https://example.com/path/ "))
	require.Equal(t, "", normalizeURL("   "))
}

func TestStringBoolPtrHelpers(t *testing.T) {
	t.Parallel()

	require.Nil(t, stringPtr(""))
	require.Equal(t, "value", *stringPtr("value"))

	require.NotNil(t, boolPtr(true))
	require.True(t, *boolPtr(true))
}
