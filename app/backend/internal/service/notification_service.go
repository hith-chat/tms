package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/models"
	ws "github.com/bareuptime/tms/internal/websocket"
)

type NotificationRepository interface {
	CreateNotification(ctx context.Context, notification *models.Notification) error
	GetNotifications(ctx context.Context, tenantID, agentID uuid.UUID, limit, offset int) ([]models.Notification, error)
	GetNotificationCount(ctx context.Context, tenantID, agentID uuid.UUID) (*models.NotificationCount, error)
	MarkNotificationAsRead(ctx context.Context, tenantID, agentID, notificationID uuid.UUID) error
	MarkAllNotificationsAsRead(ctx context.Context, tenantID, agentID uuid.UUID) error
	CleanupOldNotifications(ctx context.Context) error
}

type WebSocketBroadcaster interface {
	DeliverWebSocketMessage(sessionID uuid.UUID, message *ws.Message) error
}

type NotificationService struct {
	notificationRepo NotificationRepository
	connectionMgr    WebSocketBroadcaster
}

func NewNotificationService(notificationRepo NotificationRepository, connectionMgr WebSocketBroadcaster) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		connectionMgr:    connectionMgr,
	}
}

// CreateAndDeliverNotification creates a notification and delivers it via WebSocket
func (s *NotificationService) CreateAndDeliverNotification(ctx context.Context, notification *models.Notification) error {
	// Create notification in database
	err := s.notificationRepo.CreateNotification(ctx, notification)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Deliver via WebSocket
	s.deliverNotificationViaWebSocket(notification)

	return nil
}

// CreateTicketNotification creates a notification for ticket-related events
func (s *NotificationService) CreateTicketNotification(ctx context.Context,
	tenantID, projectID, agentID uuid.UUID,
	notificationType models.NotificationType,
	title, message string,
	ticketID *uuid.UUID,
	priority models.NotificationPriority) error {

	notification := &models.Notification{
		TenantID:  tenantID,
		ProjectID: &projectID,
		AgentID:   agentID,
		Type:      notificationType,
		Title:     title,
		Message:   message,
		Priority:  priority,
		Channels:  models.NotificationChannels{models.NotificationChannelWeb},
		Metadata: map[string]interface{}{
			"ticket_id": ticketID,
		},
	}

	if ticketID != nil {
		actionURL := fmt.Sprintf("/tickets/%s", ticketID.String())
		notification.ActionURL = &actionURL
	}

	return s.CreateAndDeliverNotification(ctx, notification)
}

// CreateSystemNotification creates a system-wide notification
func (s *NotificationService) CreateSystemNotification(ctx context.Context,
	tenantID uuid.UUID, projectID *uuid.UUID, agentID uuid.UUID,
	notificationType models.NotificationType,
	title, message string,
	priority models.NotificationPriority,
	actionURL *string) error {

	notification := &models.Notification{
		TenantID:  tenantID,
		ProjectID: projectID,
		AgentID:   agentID,
		Type:      notificationType,
		Title:     title,
		Message:   message,
		Priority:  priority,
		Channels:  models.NotificationChannels{models.NotificationChannelWeb},
		ActionURL: actionURL,
		Metadata:  map[string]interface{}{},
	}

	return s.CreateAndDeliverNotification(ctx, notification)
}

// GetNotifications retrieves notifications for an agent
func (s *NotificationService) GetNotifications(ctx context.Context, tenantID, agentID uuid.UUID, limit, offset int) ([]models.Notification, error) {
	return s.notificationRepo.GetNotifications(ctx, tenantID, agentID, limit, offset)
}

// GetNotificationCount returns the count of notifications
func (s *NotificationService) GetNotificationCount(ctx context.Context, tenantID, agentID uuid.UUID) (*models.NotificationCount, error) {
	return s.notificationRepo.GetNotificationCount(ctx, tenantID, agentID)
}

// MarkNotificationAsRead marks a notification as read
func (s *NotificationService) MarkNotificationAsRead(ctx context.Context, tenantID, agentID, notificationID uuid.UUID) error {
	err := s.notificationRepo.MarkNotificationAsRead(ctx, tenantID, agentID, notificationID)
	if err != nil {
		return err
	}

	// Send updated count via WebSocket
	s.broadcastNotificationCount(ctx, tenantID, agentID)
	return nil
}

// MarkAllNotificationsAsRead marks all notifications as read
func (s *NotificationService) MarkAllNotificationsAsRead(ctx context.Context, tenantID, agentID uuid.UUID) error {
	err := s.notificationRepo.MarkAllNotificationsAsRead(ctx, tenantID, agentID)
	if err != nil {
		return err
	}

	// Send updated count via WebSocket
	s.broadcastNotificationCount(ctx, tenantID, agentID)
	return nil
}

// deliverNotificationViaWebSocket sends notification via WebSocket using existing infrastructure
func (s *NotificationService) deliverNotificationViaWebSocket(notification *models.Notification) {
	// Marshal notification data
	notificationData, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Failed to marshal notification: %v", err)
		return
	}

	wsMessage := &ws.Message{
		Type:      string(models.WSMsgTypeNotification),
		Data:      json.RawMessage(notificationData),
		SessionID: uuid.Nil,               // Global notification
		FromType:  ws.ConnectionTypeAgent, // Use existing type
		TenantID:  &notification.TenantID,
		ProjectID: notification.ProjectID,
		AgentID:   &notification.AgentID,
	}

	// Use existing delivery mechanism to agents in the tenant
	err = s.connectionMgr.DeliverWebSocketMessage(uuid.Nil, wsMessage)
	if err != nil {
		log.Printf("Failed to deliver notification via WebSocket: %v", err)
	}
}

// broadcastNotificationCount sends updated notification count to agent
func (s *NotificationService) broadcastNotificationCount(ctx context.Context, tenantID, agentID uuid.UUID) {
	count, err := s.GetNotificationCount(ctx, tenantID, agentID)
	if err != nil {
		log.Printf("Failed to get notification count for agent %s: %v", agentID, err)
		return
	}

	// Marshal count data
	countData, err := json.Marshal(count)
	if err != nil {
		log.Printf("Failed to marshal notification count: %v", err)
		return
	}

	wsMessage := &ws.Message{
		Type:      "notification_count",
		Data:      json.RawMessage(countData),
		SessionID: uuid.Nil,
		FromType:  ws.ConnectionTypeAgent,
		TenantID:  &tenantID,
		AgentID:   &agentID,
	}

	err = s.connectionMgr.DeliverWebSocketMessage(uuid.Nil, wsMessage)
	if err != nil {
		log.Printf("Failed to deliver notification count via WebSocket: %v", err)
	}
}

// CleanupOldNotifications removes old notifications
func (s *NotificationService) CleanupOldNotifications(ctx context.Context) error {
	return s.notificationRepo.CleanupOldNotifications(ctx)
}
