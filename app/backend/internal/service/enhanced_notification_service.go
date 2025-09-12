package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
	ws "github.com/bareuptime/tms/internal/websocket"
)

// EnhancedNotificationService extends the basic notification service with alarm capabilities
type EnhancedNotificationService struct {
	notificationRepo *repo.NotificationRepo
	connectionMgr    *ws.ConnectionManager
	howlingAlarmSvc  *HowlingAlarmService
	config           *config.Config
}

// NewEnhancedNotificationService creates a new enhanced notification service
func NewEnhancedNotificationService(notificationRepo *repo.NotificationRepo,
	connectionMgr *ws.ConnectionManager, howlingAlarmSvc *HowlingAlarmService,
	cfg *config.Config) *EnhancedNotificationService {

	return &EnhancedNotificationService{
		notificationRepo: notificationRepo,
		connectionMgr:    connectionMgr,
		howlingAlarmSvc:  howlingAlarmSvc,
		config:           cfg,
	}
}

// CreateAgentAssignmentNotification creates a notification for agent assignments with potential alarm
func (s *EnhancedNotificationService) CreateAgentAssignmentNotification(ctx context.Context,
	tenantID, projectID, agentID, assignmentID uuid.UUID,
	title, message string,
	priority models.NotificationPriority,
	urgency string) error {

	// Create standard notification
	notification := &models.Notification{
		TenantID:  tenantID,
		ProjectID: &projectID,
		AgentID:   agentID,
		Type:      models.NotificationTypeAgentAssignment,
		Title:     title,
		Message:   message,
		Priority:  priority,
		Channels:  s.getChannelsForPriority(priority),
		Metadata: map[string]interface{}{
			"assignment_id": assignmentID,
			"urgency":       urgency,
			"created_at":    time.Now(),
		},
	}

	actionURL := fmt.Sprintf("/assignments/%s", assignmentID.String())
	notification.ActionURL = &actionURL

	// Create and deliver standard notification
	err := s.CreateAndDeliverNotification(ctx, notification)
	if err != nil {
		return fmt.Errorf("failed to create assignment notification: %w", err)
	}

	// Trigger howling alarm for high/urgent/critical priority assignments
	if s.shouldTriggerAlarm(priority, urgency) {
		alarmMetadata := models.JSONMap{
			"notification_id": notification.ID,
			"urgency":         urgency,
			"assignment_type": "agent_assignment",
		}

		_, err = s.howlingAlarmSvc.TriggerAlarm(ctx, tenantID, projectID,
			title, message, priority, alarmMetadata)
		if err != nil {
			log.Printf("Failed to trigger alarm for assignment %s: %v", assignmentID, err)
		}
	}

	return nil
}

// CreateUrgentRequestNotification creates a notification for urgent customer requests
func (s *EnhancedNotificationService) CreateUrgentRequestNotification(ctx context.Context,
	tenantID, projectID uuid.UUID, agentIDs []uuid.UUID,
	title, message string,
	customerInfo map[string]interface{}) error {

	// Create notifications for all specified agents
	for _, agentID := range agentIDs {
		notification := &models.Notification{
			TenantID:  tenantID,
			ProjectID: &projectID,
			AgentID:   agentID,
			Type:      models.NotificationTypeUrgentRequest,
			Title:     title,
			Message:   message,
			Priority:  models.NotificationPriorityUrgent,
			Channels:  s.getChannelsForPriority(models.NotificationPriorityUrgent),
			Metadata: map[string]interface{}{
				"customer_info": customerInfo,
				"request_type":  "urgent_support",
				"created_at":    time.Now(),
			},
		}

		err := s.CreateAndDeliverNotification(ctx, notification)
		if err != nil {
			log.Printf("Failed to create urgent request notification for agent %s: %v", agentID, err)
			continue
		}

		// Trigger alarm for urgent requests
		alarmMetadata := models.JSONMap{
			"notification_id": notification.ID,
			"customer_info":   customerInfo,
			"request_type":    "urgent_support",
		}

		_, err = s.howlingAlarmSvc.TriggerAlarm(ctx, tenantID, projectID,
			title, message, models.NotificationPriorityUrgent, alarmMetadata)
		if err != nil {
			log.Printf("Failed to trigger alarm for urgent request: %v", err)
		}
	}

	return nil
}

// CreateHowlingAlarmNotification creates a howling alarm notification
func (s *EnhancedNotificationService) CreateHowlingAlarmNotification(ctx context.Context,
	tenantID, agentID, assignmentID uuid.UUID,
	title, message string,
	alarmLevel string,
	escalationCount int) error {

	notification := &models.Notification{
		TenantID: tenantID,
		AgentID:  agentID,
		Type:     models.NotificationTypeHowlingAlarm,
		Title:    title,
		Message:  message,
		Priority: models.NotificationPriorityCritical,
		Channels: models.NotificationChannels{
			models.NotificationChannelWeb,
			models.NotificationChannelAudio,
			models.NotificationChannelDesktop,
			models.NotificationChannelOverlay,
		},
		Metadata: map[string]interface{}{
			"assignment_id":    assignmentID,
			"alarm_level":      alarmLevel,
			"escalation_count": escalationCount,
			"alarm_type":       "howling_alarm",
			"created_at":       time.Now(),
		},
	}

	actionURL := fmt.Sprintf("/assignments/%s", assignmentID.String())
	notification.ActionURL = &actionURL

	return s.CreateAndDeliverNotification(ctx, notification)
}

// CreateAndDeliverNotification creates a notification and delivers it via WebSocket
func (s *EnhancedNotificationService) CreateAndDeliverNotification(ctx context.Context, notification *models.Notification) error {
	// Create notification in database if repository is available
	if s.notificationRepo != nil {
		err := s.notificationRepo.CreateNotification(ctx, notification)
		if err != nil {
			return fmt.Errorf("failed to create notification: %w", err)
		}
	} else {
		// For testing without database - just set required fields
		notification.ID = uuid.New()
		notification.CreatedAt = time.Now()
		notification.UpdatedAt = time.Now()
	}

	// Deliver via WebSocket with enhanced channels
	s.deliverEnhancedNotification(notification)

	return nil
}

// deliverEnhancedNotification sends notification via WebSocket with enhanced channel support
func (s *EnhancedNotificationService) deliverEnhancedNotification(notification *models.Notification) {
	// Skip WebSocket delivery if connection manager is nil (for testing)
	if s.connectionMgr == nil {
		log.Printf("Skipping WebSocket delivery (no connection manager) for notification: %s", notification.Title)
		return
	}

	// Prepare enhanced notification data
	enhancedData := map[string]interface{}{
		"id":         notification.ID,
		"tenant_id":  notification.TenantID,
		"project_id": notification.ProjectID,
		"agent_id":   notification.AgentID,
		"type":       notification.Type,
		"title":      notification.Title,
		"message":    notification.Message,
		"priority":   notification.Priority,
		"channels":   notification.Channels,
		"action_url": notification.ActionURL,
		"metadata":   notification.Metadata,
		"created_at": notification.CreatedAt,
		"enhanced":   true, // Mark as enhanced notification
	}

	// Add audio configuration for audio channels
	if s.containsChannel(notification.Channels, models.NotificationChannelAudio) {
		enhancedData["audio_config"] = s.getAudioConfigForPriority(notification.Priority)
	}

	// Add visual configuration for overlay/popup channels
	if s.containsChannel(notification.Channels, models.NotificationChannelOverlay) {
		enhancedData["overlay_config"] = s.getOverlayConfigForPriority(notification.Priority)
	}

	if s.containsChannel(notification.Channels, models.NotificationChannelPopup) {
		enhancedData["popup_config"] = s.getPopupConfigForPriority(notification.Priority)
	}

	// Marshal enhanced notification data
	notificationData, err := json.Marshal(enhancedData)
	if err != nil {
		log.Printf("Failed to marshal enhanced notification: %v", err)
		return
	}

	wsMessage := &ws.Message{
		Type:      "enhanced_notification",
		Data:      json.RawMessage(notificationData),
		SessionID: uuid.Nil,
		FromType:  ws.ConnectionTypeAgent,
		TenantID:  &notification.TenantID,
		ProjectID: notification.ProjectID,
		AgentID:   &notification.AgentID,
	}

	// Use existing delivery mechanism
	err = s.connectionMgr.DeliverWebSocketMessage(uuid.Nil, wsMessage)
	if err != nil {
		log.Printf("Failed to deliver enhanced notification via WebSocket: %v", err)
	}
}

// getChannelsForPriority returns appropriate channels based on priority
func (s *EnhancedNotificationService) getChannelsForPriority(priority models.NotificationPriority) models.NotificationChannels {
	switch priority {
	case models.NotificationPriorityLow:
		return models.NotificationChannels{models.NotificationChannelWeb}
	case models.NotificationPriorityNormal:
		return models.NotificationChannels{
			models.NotificationChannelWeb,
			models.NotificationChannelDesktop,
		}
	case models.NotificationPriorityHigh:
		return models.NotificationChannels{
			models.NotificationChannelWeb,
			models.NotificationChannelDesktop,
			models.NotificationChannelAudio,
		}
	case models.NotificationPriorityUrgent:
		return models.NotificationChannels{
			models.NotificationChannelWeb,
			models.NotificationChannelDesktop,
			models.NotificationChannelAudio,
			models.NotificationChannelPopup,
		}
	case models.NotificationPriorityCritical:
		return models.NotificationChannels{
			models.NotificationChannelWeb,
			models.NotificationChannelDesktop,
			models.NotificationChannelAudio,
			models.NotificationChannelPopup,
			models.NotificationChannelOverlay,
		}
	default:
		return models.NotificationChannels{models.NotificationChannelWeb}
	}
}

// shouldTriggerAlarm determines if an alarm should be triggered
func (s *EnhancedNotificationService) shouldTriggerAlarm(priority models.NotificationPriority, urgency string) bool {
	// Trigger alarms for high priority or above
	if priority == models.NotificationPriorityHigh ||
		priority == models.NotificationPriorityUrgent ||
		priority == models.NotificationPriorityCritical {
		return true
	}

	// Trigger alarms for critical urgency regardless of priority
	if urgency == "critical" {
		return true
	}

	return false
}

// containsChannel checks if a channel is in the channels list
func (s *EnhancedNotificationService) containsChannel(channels models.NotificationChannels, channel models.NotificationChannel) bool {
	for _, c := range channels {
		if c == channel {
			return true
		}
	}
	return false
}

// getAudioConfigForPriority returns audio configuration based on priority
func (s *EnhancedNotificationService) getAudioConfigForPriority(priority models.NotificationPriority) map[string]interface{} {
	switch priority {
	case models.NotificationPriorityHigh:
		return map[string]interface{}{
			"sound_type": "notification",
			"volume":     0.7,
			"repeat":     1,
			"duration":   2000, // 2 seconds
		}
	case models.NotificationPriorityUrgent:
		return map[string]interface{}{
			"sound_type": "alert",
			"volume":     0.8,
			"repeat":     2,
			"duration":   3000, // 3 seconds
		}
	case models.NotificationPriorityCritical:
		return map[string]interface{}{
			"sound_type": "alarm",
			"volume":     0.9,
			"repeat":     3,
			"duration":   5000, // 5 seconds
		}
	default:
		return map[string]interface{}{
			"sound_type": "soft",
			"volume":     0.5,
			"repeat":     1,
			"duration":   1000, // 1 second
		}
	}
}

// getOverlayConfigForPriority returns overlay configuration based on priority
func (s *EnhancedNotificationService) getOverlayConfigForPriority(priority models.NotificationPriority) map[string]interface{} {
	return map[string]interface{}{
		"backdrop":     true,
		"persistent":   priority == models.NotificationPriorityCritical,
		"auto_dismiss": priority != models.NotificationPriorityCritical,
		"dismiss_timeout": func() int {
			switch priority {
			case models.NotificationPriorityUrgent:
				return 10000 // 10 seconds
			case models.NotificationPriorityCritical:
				return 0 // No auto dismiss
			default:
				return 5000 // 5 seconds
			}
		}(),
		"color_scheme": func() string {
			switch priority {
			case models.NotificationPriorityUrgent:
				return "orange"
			case models.NotificationPriorityCritical:
				return "red"
			default:
				return "blue"
			}
		}(),
	}
}

// getPopupConfigForPriority returns popup configuration based on priority
func (s *EnhancedNotificationService) getPopupConfigForPriority(priority models.NotificationPriority) map[string]interface{} {
	return map[string]interface{}{
		"position":       "top-right",
		"persistent":     priority == models.NotificationPriorityCritical,
		"require_action": priority == models.NotificationPriorityCritical,
		"auto_dismiss": func() bool {
			return priority != models.NotificationPriorityCritical
		}(),
		"dismiss_timeout": func() int {
			switch priority {
			case models.NotificationPriorityUrgent:
				return 15000 // 15 seconds
			case models.NotificationPriorityCritical:
				return 0 // No auto dismiss
			default:
				return 8000 // 8 seconds
			}
		}(),
		"animation": "bounce",
		"size": func() string {
			switch priority {
			case models.NotificationPriorityCritical:
				return "large"
			case models.NotificationPriorityUrgent:
				return "medium"
			default:
				return "small"
			}
		}(),
	}
}

// GetNotifications retrieves notifications for an agent
func (s *EnhancedNotificationService) GetNotifications(ctx context.Context, tenantID, agentID uuid.UUID, limit, offset int) ([]models.Notification, error) {
	if s.notificationRepo != nil {
		return s.notificationRepo.GetNotifications(ctx, tenantID, agentID, limit, offset)
	}
	return []models.Notification{}, nil
}

// GetNotificationCount returns the count of notifications
func (s *EnhancedNotificationService) GetNotificationCount(ctx context.Context, tenantID, agentID uuid.UUID) (*models.NotificationCount, error) {
	if s.notificationRepo != nil {
		return s.notificationRepo.GetNotificationCount(ctx, tenantID, agentID)
	}
	return &models.NotificationCount{Total: 0, Unread: 0}, nil
}

// MarkNotificationAsRead marks a notification as read
func (s *EnhancedNotificationService) MarkNotificationAsRead(ctx context.Context, tenantID, agentID, notificationID uuid.UUID) error {
	if s.notificationRepo == nil {
		return nil // No-op for testing
	}

	err := s.notificationRepo.MarkNotificationAsRead(ctx, tenantID, agentID, notificationID)
	if err != nil {
		return err
	}

	// Send updated count via WebSocket
	s.broadcastNotificationCount(ctx, tenantID, agentID)
	return nil
}

// MarkAllNotificationsAsRead marks all notifications as read
func (s *EnhancedNotificationService) MarkAllNotificationsAsRead(ctx context.Context, tenantID, agentID uuid.UUID) error {
	if s.notificationRepo == nil {
		return nil // No-op for testing
	}

	err := s.notificationRepo.MarkAllNotificationsAsRead(ctx, tenantID, agentID)
	if err != nil {
		return err
	}

	// Send updated count via WebSocket
	s.broadcastNotificationCount(ctx, tenantID, agentID)
	return nil
}

// broadcastNotificationCount sends updated notification count to agent
func (s *EnhancedNotificationService) broadcastNotificationCount(ctx context.Context, tenantID, agentID uuid.UUID) {
	if s.notificationRepo == nil {
		return // No-op for testing
	}

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
