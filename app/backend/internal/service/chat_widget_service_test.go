package service

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
)

func newTestChatWidgetService(t *testing.T) (*ChatWidgetService, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	chatWidgetRepo := repo.NewChatWidgetRepo(sqlxDB)

	service := NewChatWidgetService(chatWidgetRepo, nil)

	cleanup := func() {
		mock.ExpectClose()
		require.NoError(t, sqlxDB.Close())
	}

	return service, mock, cleanup
}

func TestChatWidgetService_CreateChatWidget_AppliesDefaultsAndSetsEmbedCode(t *testing.T) {
	t.Parallel()

	service, mock, cleanup := newTestChatWidgetService(t)
	defer cleanup()

	mock.ExpectExec("INSERT INTO chat_widgets").WillReturnResult(sqlmock.NewResult(1, 1))

	tenantID := uuid.New()
	projectID := uuid.New()
	req := &models.CreateChatWidgetRequest{}

	widget, err := service.CreateChatWidget(context.Background(), tenantID, projectID, req)
	require.NoError(t, err)
	require.NotNil(t, widget)

	require.Equal(t, "#2563eb", widget.PrimaryColor)
	require.Equal(t, "#f3f4f6", widget.SecondaryColor)
	require.Equal(t, "#ffffff", widget.BackgroundColor)
	require.Equal(t, "bottom-right", widget.Position)
	require.Equal(t, "Support Agent", widget.AgentName)
	require.Equal(t, "Hello! How can we help you?", widget.WelcomeMessage)
	require.Equal(t, "We are currently offline. Please leave a message.", widget.OfflineMessage)
	require.NotNil(t, widget.EmbedCode)
	require.True(t, strings.Contains(*widget.EmbedCode, widget.ID.String()))
	require.True(t, strings.Contains(*widget.EmbedCode, "chat-widget.js"))

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestChatWidgetService_CreateChatWidget_PreservesProvidedFields(t *testing.T) {
	t.Parallel()

	service, mock, cleanup := newTestChatWidgetService(t)
	defer cleanup()

	mock.ExpectExec("INSERT INTO chat_widgets").WillReturnResult(sqlmock.NewResult(1, 1))

	tenantID := uuid.New()
	projectID := uuid.New()
	avatar := "https://example.com/avatar.png"
	businessHours := models.JSONMap{"enabled": true}
	req := &models.CreateChatWidgetRequest{
		PrimaryColor:     "#000000",
		SecondaryColor:   "#111111",
		BackgroundColor:  "#222222",
		Position:         "bottom-left",
		WelcomeMessage:   "Welcome!",
		OfflineMessage:   "Ping us later",
		AgentName:        "Jess",
		AgentAvatarURL:   &avatar,
		BusinessHours:    businessHours,
		ChatBubbleStyle:  "classic",
		WidgetShape:      "square",
		UseAI:            true,
		AutoOpenDelay:    5,
		ShowAgentAvatars: true,
		AllowFileUploads: true,
		RequireEmail:     true,
	}

	widget, err := service.CreateChatWidget(context.Background(), tenantID, projectID, req)
	require.NoError(t, err)
	require.NotNil(t, widget)

	require.Equal(t, "#000000", widget.PrimaryColor)
	require.Equal(t, "#111111", widget.SecondaryColor)
	require.Equal(t, "#222222", widget.BackgroundColor)
	require.Equal(t, "bottom-left", widget.Position)
	require.Equal(t, "Welcome!", widget.WelcomeMessage)
	require.Equal(t, "Ping us later", widget.OfflineMessage)
	require.Equal(t, "Jess", widget.AgentName)
	require.Equal(t, &avatar, widget.AgentAvatarURL)
	require.Equal(t, businessHours, widget.BusinessHours)
	require.True(t, widget.UseAI)
	require.Equal(t, 5, widget.AutoOpenDelay)
	require.True(t, widget.ShowAgentAvatars)
	require.True(t, widget.AllowFileUploads)
	require.True(t, widget.RequireEmail)

	require.NoError(t, mock.ExpectationsWereMet())
}
