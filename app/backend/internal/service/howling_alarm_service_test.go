package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/models"
	ws "github.com/bareuptime/tms/internal/websocket"
)

type mockAlarmRepo struct {
	mock.Mock
}

func (m *mockAlarmRepo) AcknowledgeAlarm(ctx context.Context, tenantID, alarmID, agentID uuid.UUID, response string) error {
	args := m.Called(ctx, tenantID, alarmID, agentID, response)
	return args.Error(0)
}

func (m *mockAlarmRepo) GetAlarmByID(ctx context.Context, tenantID, alarmID uuid.UUID) (*models.Alarm, error) {
	args := m.Called(ctx, tenantID, alarmID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Alarm), args.Error(1)
}

func (m *mockAlarmRepo) GetActiveAlarms(ctx context.Context, tenantID, projectID uuid.UUID) ([]*models.Alarm, error) {
	args := m.Called(ctx, tenantID, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Alarm), args.Error(1)
}

func (m *mockAlarmRepo) GetAlarmStats(ctx context.Context, tenantID, projectID uuid.UUID) (*models.AlarmStats, error) {
	args := m.Called(ctx, tenantID, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AlarmStats), args.Error(1)
}

type mockProjectBroadcaster struct {
	mock.Mock
}

func (m *mockProjectBroadcaster) SendToProjectAgents(projectID uuid.UUID, message *ws.Message) error {
	args := m.Called(projectID, message)
	return args.Error(0)
}

func TestHowlingAlarmService_TriggerAlarm(t *testing.T) {
	cfg := &config.Config{}
	broadcaster := &mockProjectBroadcaster{}
	repoMock := &mockAlarmRepo{}

	service := NewHowlingAlarmService(cfg, broadcaster, repoMock)

	tenantID := uuid.New()
	projectID := uuid.New()

	broadcaster.On("SendToProjectAgents", projectID, mock.MatchedBy(func(message *ws.Message) bool {
		if message == nil {
			return false
		}
		if message.Type != "alarm_triggered" {
			return false
		}
		var payload map[string]interface{}
		if err := json.Unmarshal(message.Data, &payload); err != nil {
			return false
		}
		project, ok := payload["project_id"].(string)
		if !ok {
			return false
		}
		return project == projectID.String()
	})).Return(nil).Once()

	alarm, err := service.TriggerAlarm(context.Background(), tenantID, projectID, "Database down", "Primary database is unreachable", models.NotificationPriorityHigh, models.JSONMap{})

	assert.NoError(t, err)
	assert.Equal(t, tenantID, alarm.TenantID)
	assert.Equal(t, projectID, alarm.ProjectID)
	assert.Equal(t, models.NotificationPriorityHigh, alarm.Priority)
	assert.Equal(t, models.AlarmLevel("loud"), alarm.CurrentLevel)
	broadcaster.AssertExpectations(t)
}

func TestHowlingAlarmService_AcknowledgeAlarm(t *testing.T) {
	cfg := &config.Config{}
	broadcaster := &mockProjectBroadcaster{}
	repoMock := &mockAlarmRepo{}

	service := NewHowlingAlarmService(cfg, broadcaster, repoMock)

	tenantID := uuid.New()
	alarmID := uuid.New()
	agentID := uuid.New()

	now := time.Now()
	alarm := &models.Alarm{
		ID:        alarmID,
		TenantID:  tenantID,
		ProjectID: uuid.New(),
		Metadata:  models.JSONMap{},
	}

	repoMock.On("AcknowledgeAlarm", mock.Anything, tenantID, alarmID, agentID, "ack").Return(nil).Once()
	repoMock.On("GetAlarmByID", mock.Anything, tenantID, alarmID).Return(alarm, nil).Once()

	broadcaster.On("SendToProjectAgents", alarm.ProjectID, mock.MatchedBy(func(message *ws.Message) bool {
		if message.Type != "alarm_acknowledged" {
			return false
		}
		var payload map[string]interface{}
		if err := json.Unmarshal(message.Data, &payload); err != nil {
			return false
		}
		alarmIDStr, ok := payload["alarm_id"].(string)
		if !ok {
			return false
		}
		return alarmIDStr == alarmID.String()
	})).Return(nil).Once()

	alarm.AcknowledgedAt = &now

	err := service.AcknowledgeAlarm(context.Background(), tenantID, alarmID, agentID, "ack")
	assert.NoError(t, err)

	repoMock.AssertExpectations(t)
	broadcaster.AssertExpectations(t)
}

func TestHowlingAlarmService_AcknowledgeAlarm_Failure(t *testing.T) {
	cfg := &config.Config{}
	broadcaster := &mockProjectBroadcaster{}
	repoMock := &mockAlarmRepo{}

	service := NewHowlingAlarmService(cfg, broadcaster, repoMock)

	tenantID := uuid.New()
	alarmID := uuid.New()
	agentID := uuid.New()

	repoMock.On("AcknowledgeAlarm", mock.Anything, tenantID, alarmID, agentID, "ack").Return(assert.AnError).Once()

	err := service.AcknowledgeAlarm(context.Background(), tenantID, alarmID, agentID, "ack")
	assert.Error(t, err)

	broadcaster.AssertNotCalled(t, "SendToProjectAgents", mock.Anything, mock.Anything)
}

func TestHowlingAlarmService_GetAlarmConfigForPriority(t *testing.T) {
	cfg := &config.Config{}
	broadcaster := &mockProjectBroadcaster{}
	repoMock := &mockAlarmRepo{}

	service := NewHowlingAlarmService(cfg, broadcaster, repoMock)

	cases := map[models.NotificationPriority]AlarmEscalationConfig{
		models.NotificationPriorityLow:      {InitialLevel: AlarmLevelSoft, MaxLevel: AlarmLevelMedium},
		models.NotificationPriorityNormal:   {InitialLevel: AlarmLevelMedium, MaxLevel: AlarmLevelLoud},
		models.NotificationPriorityHigh:     {InitialLevel: AlarmLevelLoud, MaxLevel: AlarmLevelUrgent},
		models.NotificationPriorityUrgent:   {InitialLevel: AlarmLevelUrgent, MaxLevel: AlarmLevelCritical},
		models.NotificationPriorityCritical: {InitialLevel: AlarmLevelUrgent, MaxLevel: AlarmLevelCritical},
	}

	for priority, expected := range cases {
		cfg := service.getAlarmConfigForPriority(priority)
		assert.Equal(t, expected.InitialLevel, cfg.InitialLevel)
		assert.Equal(t, expected.MaxLevel, cfg.MaxLevel)
	}
}
