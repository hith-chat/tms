package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/redis"
	"github.com/bareuptime/tms/internal/repo"
)

func newTestChatSessionService(t *testing.T) (*ChatSessionService, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	chatSessionRepo := repo.NewChatSessionRepo(sqlxDB)

	mini, err := miniredis.Run()
	require.NoError(t, err)

	redisSvc := redis.NewService(redis.RedisConfig{
		URL:         fmt.Sprintf("redis://%s", mini.Addr()),
		Environment: "test",
	})

	service := &ChatSessionService{
		chatSessionRepo: chatSessionRepo,
		redisService:    redisSvc,
	}

	cleanup := func() {
		mock.ExpectClose()
		require.NoError(t, sqlxDB.Close())
		require.NoError(t, redisSvc.Close())
		mini.Close()
	}

	return service, mock, cleanup
}

func TestChatSessionService_GetChatSessionByClientSessionID_UsesCache(t *testing.T) {
	t.Parallel()

	service, mock, cleanup := newTestChatSessionService(t)
	defer cleanup()

	tenantID := uuid.New()
	projectID := uuid.New()
	widgetID := uuid.New()
	sessionID := uuid.New()
	clientSessionID := "client-session-123"
	now := time.Now().UTC().Truncate(time.Second)

	visitorInfo := models.JSONMap{"locale": "en"}
	visitorInfoJSON, err := json.Marshal(visitorInfo)
	require.NoError(t, err)

	rows := sqlmock.NewRows([]string{
		"id",
		"tenant_id",
		"project_id",
		"widget_id",
		"customer_id",
		"ticket_id",
		"status",
		"visitor_info",
		"assigned_agent_id",
		"assigned_at",
		"started_at",
		"ended_at",
		"client_session_id",
		"last_activity_at",
		"created_at",
		"updated_at",
		"assigned_agent_name",
		"customer_name",
		"customer_email",
		"widget_name",
		"use_ai",
	}).AddRow(
		sessionID,
		tenantID,
		projectID,
		widgetID,
		nil,
		nil,
		"active",
		visitorInfoJSON,
		nil,
		nil,
		now,
		nil,
		clientSessionID,
		now,
		now,
		now,
		nil,
		nil,
		nil,
		"Widget",
		true,
	)

	queryRegex := `(?s)SELECT cs.id.*FROM chat_sessions cs.*WHERE cs.client_session_id = \$1`

	mock.ExpectQuery(queryRegex).WithArgs(clientSessionID).WillReturnRows(rows)

	ctx := context.Background()

	session, err := service.GetChatSessionByClientSessionID(ctx, clientSessionID)
	require.NoError(t, err)
	require.NotNil(t, session)
	require.Equal(t, sessionID, session.ID)
	require.Equal(t, tenantID, session.TenantID)
	require.Equal(t, projectID, session.ProjectID)
	require.Equal(t, widgetID, session.WidgetID)
	require.Equal(t, clientSessionID, session.ClientSessionID)
	require.True(t, session.UseAI)
	require.Equal(t, visitorInfo, session.VisitorInfo)

	cachedSession, err := service.GetChatSessionByClientSessionID(ctx, clientSessionID)
	require.NoError(t, err)
	require.NotNil(t, cachedSession)
	require.Equal(t, session.ID, cachedSession.ID)
	require.Equal(t, session.TenantID, cachedSession.TenantID)

	require.NoError(t, mock.ExpectationsWereMet())
}
