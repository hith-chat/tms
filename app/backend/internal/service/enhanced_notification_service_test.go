package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/models"
	ws "github.com/bareuptime/tms/internal/websocket"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type stubProjectBroadcaster struct {
	messages []*ws.Message
}

func (s *stubProjectBroadcaster) SendToProjectAgents(projectID uuid.UUID, message *ws.Message) error {
	s.messages = append(s.messages, message)
	return nil
}

type stubAlarmRepo struct{}

func (stubAlarmRepo) AcknowledgeAlarm(ctx context.Context, tenantID, alarmID, agentID uuid.UUID, response string) error {
	return nil
}

func (stubAlarmRepo) GetAlarmByID(ctx context.Context, tenantID, alarmID uuid.UUID) (*models.Alarm, error) {
	return &models.Alarm{}, nil
}

func (stubAlarmRepo) GetActiveAlarms(ctx context.Context, tenantID, projectID uuid.UUID) ([]*models.Alarm, error) {
	return nil, nil
}

func (stubAlarmRepo) GetAlarmStats(ctx context.Context, tenantID, projectID uuid.UUID) (*models.AlarmStats, error) {
	return &models.AlarmStats{}, nil
}

func TestEnhancedNotificationService_CreateAgentAssignmentNotification_TriggersAlarmForHighPriority(t *testing.T) {
	t.Parallel()

	broadcaster := &stubProjectBroadcaster{}
	howlingSvc := NewHowlingAlarmService(&config.Config{}, broadcaster, stubAlarmRepo{})
	svc := NewEnhancedNotificationService(nil, nil, howlingSvc, &config.Config{})

	tenantID := uuid.New()
	projectID := uuid.New()
	agentID := uuid.New()
	assignmentID := uuid.New()

	err := svc.CreateAgentAssignmentNotification(
		context.Background(),
		tenantID,
		projectID,
		agentID,
		assignmentID,
		"Agent assigned",
		"You have a new assignment",
		models.NotificationPriorityHigh,
		"urgent",
	)
	require.NoError(t, err)
	require.Len(t, broadcaster.messages, 1)

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(broadcaster.messages[0].Data, &payload))
	require.Equal(t, "alarm_triggered", payload["type"])
	require.Equal(t, "high", payload["priority"])

	metadata, ok := payload["metadata"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "agent_assignment", metadata["assignment_type"])
	require.Equal(t, "urgent", metadata["urgency"])
}

func TestEnhancedNotificationService_CreateAgentAssignmentNotification_SkipsAlarmForLowPriority(t *testing.T) {
	t.Parallel()

	broadcaster := &stubProjectBroadcaster{}
	howlingSvc := NewHowlingAlarmService(&config.Config{}, broadcaster, stubAlarmRepo{})
	svc := NewEnhancedNotificationService(nil, nil, howlingSvc, &config.Config{})

	err := svc.CreateAgentAssignmentNotification(
		context.Background(),
		uuid.New(),
		uuid.New(),
		uuid.New(),
		uuid.New(),
		"Routine assignment",
		"No alarm expected",
		models.NotificationPriorityLow,
		"normal",
	)
	require.NoError(t, err)
	require.Len(t, broadcaster.messages, 0)
}

func TestEnhancedNotificationService_ShouldTriggerAlarm(t *testing.T) {
	t.Parallel()

	svc := &EnhancedNotificationService{}

	require.True(t, svc.shouldTriggerAlarm(models.NotificationPriorityUrgent, "standard"))
	require.True(t, svc.shouldTriggerAlarm(models.NotificationPriorityLow, "critical"))
	require.False(t, svc.shouldTriggerAlarm(models.NotificationPriorityLow, "normal"))
}

func TestEnhancedNotificationService_GetChannelsForPriority(t *testing.T) {
	t.Parallel()

	svc := &EnhancedNotificationService{}

	criticalChannels := svc.getChannelsForPriority(models.NotificationPriorityCritical)
	require.Contains(t, criticalChannels, models.NotificationChannelOverlay)
	require.Contains(t, criticalChannels, models.NotificationChannelAudio)

	lowChannels := svc.getChannelsForPriority(models.NotificationPriorityLow)
	require.Equal(t, models.NotificationChannels{models.NotificationChannelWeb}, lowChannels)
}
