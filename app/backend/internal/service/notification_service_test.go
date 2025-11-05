package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/bareuptime/tms/internal/models"
	ws "github.com/bareuptime/tms/internal/websocket"
)

type mockNotificationRepository struct {
	mock.Mock
}

func (m *mockNotificationRepository) CreateNotification(ctx context.Context, notification *models.Notification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *mockNotificationRepository) GetNotifications(ctx context.Context, tenantID, agentID uuid.UUID, limit, offset int) ([]models.Notification, error) {
	args := m.Called(ctx, tenantID, agentID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Notification), args.Error(1)
}

func (m *mockNotificationRepository) GetNotificationCount(ctx context.Context, tenantID, agentID uuid.UUID) (*models.NotificationCount, error) {
	args := m.Called(ctx, tenantID, agentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NotificationCount), args.Error(1)
}

func (m *mockNotificationRepository) MarkNotificationAsRead(ctx context.Context, tenantID, agentID, notificationID uuid.UUID) error {
	args := m.Called(ctx, tenantID, agentID, notificationID)
	return args.Error(0)
}

func (m *mockNotificationRepository) MarkAllNotificationsAsRead(ctx context.Context, tenantID, agentID uuid.UUID) error {
	args := m.Called(ctx, tenantID, agentID)
	return args.Error(0)
}

func (m *mockNotificationRepository) CleanupOldNotifications(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type mockWebSocketBroadcaster struct {
	mock.Mock
}

func (m *mockWebSocketBroadcaster) DeliverWebSocketMessage(sessionID uuid.UUID, message *ws.Message) error {
	args := m.Called(sessionID, message)
	return args.Error(0)
}

func TestNotificationService_CreateAndDeliverNotification_Success(t *testing.T) {
	ctx := context.Background()
	repoMock := &mockNotificationRepository{}
	wsMock := &mockWebSocketBroadcaster{}

	service := NewNotificationService(repoMock, wsMock)

	notification := &models.Notification{
		TenantID: uuid.New(),
		AgentID:  uuid.New(),
		Title:    "Test title",
		Message:  "Test message",
		Channels: models.NotificationChannels{models.NotificationChannelWeb},
	}

	repoMock.On("CreateNotification", mock.Anything, notification).Return(nil).Once()
	wsMock.On("DeliverWebSocketMessage", uuid.Nil, mock.MatchedBy(func(message *ws.Message) bool {
		if message == nil {
			return false
		}
		if message.Type != string(models.WSMsgTypeNotification) {
			return false
		}
		var payload models.Notification
		if err := json.Unmarshal(message.Data, &payload); err != nil {
			return false
		}
		return payload.Title == notification.Title && payload.Message == notification.Message
	})).Return(nil).Once()

	err := service.CreateAndDeliverNotification(ctx, notification)

	assert.NoError(t, err)
	repoMock.AssertExpectations(t)
	wsMock.AssertExpectations(t)
}

func TestNotificationService_CreateAndDeliverNotification_RepoError(t *testing.T) {
	ctx := context.Background()
	repoMock := &mockNotificationRepository{}
	wsMock := &mockWebSocketBroadcaster{}

	service := NewNotificationService(repoMock, wsMock)

	notification := &models.Notification{TenantID: uuid.New(), AgentID: uuid.New()}

	repoMock.On("CreateNotification", mock.Anything, notification).Return(assert.AnError).Once()

	err := service.CreateAndDeliverNotification(ctx, notification)

	assert.Error(t, err)
	wsMock.AssertNotCalled(t, "DeliverWebSocketMessage", mock.Anything, mock.Anything)
}

func TestNotificationService_MarkNotificationAsRead_BroadcastsCount(t *testing.T) {
	ctx := context.Background()
	repoMock := &mockNotificationRepository{}
	wsMock := &mockWebSocketBroadcaster{}

	service := NewNotificationService(repoMock, wsMock)

	tenantID := uuid.New()
	agentID := uuid.New()
	notificationID := uuid.New()

	repoMock.On("MarkNotificationAsRead", mock.Anything, tenantID, agentID, notificationID).Return(nil).Once()
	repoMock.On("GetNotificationCount", mock.Anything, tenantID, agentID).
		Return(&models.NotificationCount{Total: 3, Unread: 1}, nil).Once()

	wsMock.On("DeliverWebSocketMessage", uuid.Nil, mock.MatchedBy(func(message *ws.Message) bool {
		if message.Type != "notification_count" {
			return false
		}
		var payload models.NotificationCount
		if err := json.Unmarshal(message.Data, &payload); err != nil {
			return false
		}
		return payload.Total == 3 && payload.Unread == 1
	})).Return(nil).Once()

	err := service.MarkNotificationAsRead(ctx, tenantID, agentID, notificationID)
	assert.NoError(t, err)

	repoMock.AssertExpectations(t)
	wsMock.AssertExpectations(t)
}

func TestNotificationService_MarkNotificationAsRead_RepoError(t *testing.T) {
	ctx := context.Background()
	repoMock := &mockNotificationRepository{}
	wsMock := &mockWebSocketBroadcaster{}

	service := NewNotificationService(repoMock, wsMock)

	tenantID := uuid.New()
	agentID := uuid.New()
	notificationID := uuid.New()

	repoMock.On("MarkNotificationAsRead", mock.Anything, tenantID, agentID, notificationID).Return(assert.AnError).Once()

	err := service.MarkNotificationAsRead(ctx, tenantID, agentID, notificationID)
	assert.Error(t, err)
	repoMock.AssertExpectations(t)
	wsMock.AssertNotCalled(t, "DeliverWebSocketMessage", mock.Anything, mock.Anything)
}

func TestNotificationService_CreateTicketNotification_SetsActionURL(t *testing.T) {
	ctx := context.Background()
	repoMock := &mockNotificationRepository{}
	wsMock := &mockWebSocketBroadcaster{}

	service := NewNotificationService(repoMock, wsMock)

	tenantID := uuid.New()
	projectID := uuid.New()
	agentID := uuid.New()
	ticketID := uuid.New()

	var captured *models.Notification
	repoMock.On("CreateNotification", mock.Anything, mock.MatchedBy(func(n *models.Notification) bool {
		captured = n
		return true
	})).Return(nil).Once()

	wsMock.On("DeliverWebSocketMessage", uuid.Nil, mock.Anything).Return(nil).Once()

	err := service.CreateTicketNotification(ctx, tenantID, projectID, agentID, models.NotificationTypeTicketAssigned, "Title", "Body", &ticketID, models.NotificationPriorityHigh)
	assert.NoError(t, err)

	if assert.NotNil(t, captured) {
		assert.NotNil(t, captured.ActionURL)
		assert.Contains(t, *captured.ActionURL, ticketID.String())
		metadata, ok := captured.Metadata.(map[string]interface{})
		if assert.True(t, ok, "metadata should be a map") {
			value, exists := metadata["ticket_id"]
			if assert.True(t, exists) {
				pointer, ok := value.(*uuid.UUID)
				if assert.True(t, ok, "ticket_id should be *uuid.UUID") {
					assert.Equal(t, ticketID, *pointer)
				}
			}
		}
	}
}

func TestNotificationService_CleanupOldNotifications(t *testing.T) {
	ctx := context.Background()
	repoMock := &mockNotificationRepository{}
	wsMock := &mockWebSocketBroadcaster{}

	service := NewNotificationService(repoMock, wsMock)

	repoMock.On("CleanupOldNotifications", mock.Anything).Return(nil).Once()

	err := service.CleanupOldNotifications(ctx)
	assert.NoError(t, err)
	repoMock.AssertExpectations(t)
}
