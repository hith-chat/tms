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

// HowlingAlarmService manages critical notifications with escalating alerts
type HowlingAlarmService struct {
	config           *config.Config
	connectionMgr    *ws.ConnectionManager
	alarmRepo        *repo.AlarmRepository
	escalationTicker *time.Ticker
	stopChannel      chan bool
}

// NewHowlingAlarmService creates a new howling alarm service
func NewHowlingAlarmService(cfg *config.Config, connectionMgr *ws.ConnectionManager, alarmRepo *repo.AlarmRepository) *HowlingAlarmService {
	service := &HowlingAlarmService{
		config:        cfg,
		connectionMgr: connectionMgr,
		alarmRepo:     alarmRepo,
		stopChannel:   make(chan bool, 1),
	}

	log.Printf("HowlingAlarmService initialized successfully")
	return service
}

// TriggerAlarm triggers a new howling alarm
func (s *HowlingAlarmService) TriggerAlarm(ctx context.Context, assignmentID, agentID, tenantID, projectID uuid.UUID,
	title, message string, priority models.NotificationPriority, metadata models.JSONMap) (*models.Alarm, error) {

	// Create alarm configuration based on priority
	config := s.getAlarmConfigForPriority(priority)

	now := time.Now()
	alarm := &models.Alarm{
		ID:              uuid.New(),
		TenantID:        tenantID,
		ProjectID:       projectID,
		AssignmentID:    &assignmentID,
		AgentID:         &agentID,
		Title:           title,
		Message:         message,
		Priority:        priority,
		CurrentLevel:    models.AlarmLevel(config.InitialLevel.String()),
		StartTime:       now,
		LastEscalation:  now,
		EscalationCount: 0,
		IsAcknowledged:  false,
		Config: models.AlarmEscalationConfig{
			InitialLevel:          models.AlarmLevel(config.InitialLevel.String()),
			EscalationInterval:    config.EscalationInterval,
			MaxLevel:              models.AlarmLevel(config.MaxLevel.String()),
			PersistUntilAcknowled: config.PersistUntilAcknowled,
			AudioEnabled:          config.AudioEnabled,
			VisualEnabled:         config.VisualEnabled,
			BroadcastToAll:        config.BroadcastToAll,
		},
		Metadata:  metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Send initial notification
	s.sendAlarmNotification(alarm)

	log.Printf("Howling alarm triggered: ID=%s, Assignment=%s, Agent=%s, Level=%s",
		alarm.ID, assignmentID, agentID, alarm.CurrentLevel)

	return alarm, nil
}

// AcknowledgeAlarm acknowledges an active alarm
func (s *HowlingAlarmService) AcknowledgeAlarm(ctx context.Context, tenantID, alarmID, agentID uuid.UUID, response string) error {
	// Use repository to acknowledge alarm
	err := s.alarmRepo.AcknowledgeAlarm(ctx, tenantID, alarmID, agentID, response)
	if err != nil {
		return fmt.Errorf("failed to acknowledge alarm: %w", err)
	}

	// Get the updated alarm for notification
	alarm, err := s.alarmRepo.GetAlarmByID(ctx, tenantID, alarmID)
	if err != nil {
		log.Printf("Failed to get acknowledged alarm for notification: %v", err)
		return nil // Don't fail the acknowledgment if we can't notify
	}

	// Send acknowledgment notification
	s.sendAcknowledgmentNotification(alarm, agentID, response)

	log.Printf("Alarm acknowledged: ID=%s, Agent=%s", alarmID, agentID)
	return nil
}

// GetActiveAlarms returns all active alarms for a project
func (s *HowlingAlarmService) GetActiveAlarms(ctx context.Context, tenantID, projectID uuid.UUID) ([]*models.Alarm, error) {
	return s.alarmRepo.GetActiveAlarms(ctx, tenantID, projectID)
}

// GetAlarmStats returns alarm statistics for a project
func (s *HowlingAlarmService) GetAlarmStats(ctx context.Context, tenantID, projectID uuid.UUID) (*models.AlarmStats, error) {
	return s.alarmRepo.GetAlarmStats(ctx, tenantID, projectID)
}

// AlarmLevel represents the intensity level of a howling alarm
type AlarmLevel int

const (
	AlarmLevelSoft AlarmLevel = iota
	AlarmLevelMedium
	AlarmLevelLoud
	AlarmLevelUrgent
	AlarmLevelCritical
)

// String returns the string representation of AlarmLevel
func (al AlarmLevel) String() string {
	switch al {
	case AlarmLevelSoft:
		return "soft"
	case AlarmLevelMedium:
		return "medium"
	case AlarmLevelLoud:
		return "loud"
	case AlarmLevelUrgent:
		return "urgent"
	case AlarmLevelCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// AlarmEscalationConfig defines how alarms escalate over time
type AlarmEscalationConfig struct {
	InitialLevel          AlarmLevel    `json:"initial_level"`
	EscalationInterval    time.Duration `json:"escalation_interval"`
	MaxLevel              AlarmLevel    `json:"max_level"`
	PersistUntilAcknowled bool          `json:"persist_until_acknowledged"`
	AudioEnabled          bool          `json:"audio_enabled"`
	VisualEnabled         bool          `json:"visual_enabled"`
	BroadcastToAll        bool          `json:"broadcast_to_all"`
}

// getAlarmConfigForPriority returns the alarm configuration for a given priority
func (s *HowlingAlarmService) getAlarmConfigForPriority(priority models.NotificationPriority) AlarmEscalationConfig {
	switch priority {
	case models.NotificationPriorityLow:
		return AlarmEscalationConfig{
			InitialLevel:          AlarmLevelSoft,
			EscalationInterval:    10 * time.Minute,
			MaxLevel:              AlarmLevelMedium,
			PersistUntilAcknowled: false,
			AudioEnabled:          false,
			VisualEnabled:         true,
			BroadcastToAll:        false,
		}
	case models.NotificationPriorityNormal:
		return AlarmEscalationConfig{
			InitialLevel:          AlarmLevelMedium,
			EscalationInterval:    5 * time.Minute,
			MaxLevel:              AlarmLevelLoud,
			PersistUntilAcknowled: true,
			AudioEnabled:          true,
			VisualEnabled:         true,
			BroadcastToAll:        false,
		}
	case models.NotificationPriorityHigh:
		return AlarmEscalationConfig{
			InitialLevel:          AlarmLevelLoud,
			EscalationInterval:    2 * time.Minute,
			MaxLevel:              AlarmLevelUrgent,
			PersistUntilAcknowled: true,
			AudioEnabled:          true,
			VisualEnabled:         true,
			BroadcastToAll:        true,
		}
	case models.NotificationPriorityUrgent:
		return AlarmEscalationConfig{
			InitialLevel:          AlarmLevelUrgent,
			EscalationInterval:    1 * time.Minute,
			MaxLevel:              AlarmLevelCritical,
			PersistUntilAcknowled: true,
			AudioEnabled:          true,
			VisualEnabled:         true,
			BroadcastToAll:        true,
		}
	case models.NotificationPriorityCritical:
		return AlarmEscalationConfig{
			InitialLevel:          AlarmLevelUrgent,
			EscalationInterval:    30 * time.Second,
			MaxLevel:              AlarmLevelCritical,
			PersistUntilAcknowled: true,
			AudioEnabled:          true,
			VisualEnabled:         true,
			BroadcastToAll:        true,
		}
	default:
		return AlarmEscalationConfig{
			InitialLevel:          AlarmLevelSoft,
			EscalationInterval:    5 * time.Minute,
			MaxLevel:              AlarmLevelMedium,
			PersistUntilAcknowled: false,
			AudioEnabled:          false,
			VisualEnabled:         true,
			BroadcastToAll:        false,
		}
	}
}

// sendAlarmNotification sends the alarm notification via websocket to agents assigned to the project
func (s *HowlingAlarmService) sendAlarmNotification(alarm *models.Alarm) {

	notification := map[string]interface{}{
		"type":             "alarm_triggered",
		"alarm_id":         alarm.ID,
		"tenant_id":        alarm.TenantID,
		"project_id":       alarm.ProjectID,
		"title":            alarm.Title,
		"message":          alarm.Message,
		"priority":         alarm.Priority,
		"current_level":    alarm.CurrentLevel,
		"start_time":       alarm.StartTime,
		"escalation_count": alarm.EscalationCount,
		"config":           alarm.Config,
		"metadata":         alarm.Metadata,
		"timestamp":        time.Now(),
	}

	data, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Failed to marshal alarm notification: %v", err)
		return
	}

	// Create websocket message for agents in the project
	wsMessage := &ws.Message{
		Type:      "alarm_triggered",
		Data:      data,
		Timestamp: time.Now(),
		TenantID:  &alarm.TenantID,
		ProjectID: &alarm.ProjectID,
		FromType:  ws.ConnectionTypeAgent, // This is an agent-targeted message
	}

	// Send to all agents connected to this project via Redis pub/sub
	// The connection manager will handle routing to the correct agent connections
	err = s.connectionMgr.SendToProjectAgents(alarm.ProjectID, wsMessage)
	if err != nil {
		log.Printf("Failed to publish alarm notification to project %s: %v", alarm.ProjectID, err)
		return
	}

	log.Printf("Alarm notification sent to project agents: ID=%s, Level=%s, Project=%s",
		alarm.ID, alarm.CurrentLevel, alarm.ProjectID)
}

// sendAcknowledgmentNotification sends acknowledgment notification
func (s *HowlingAlarmService) sendAcknowledgmentNotification(alarm *models.Alarm, agentID uuid.UUID, response string) {
	if s.connectionMgr == nil {
		return
	}

	notification := map[string]interface{}{
		"type":            "alarm_acknowledged",
		"alarm_id":        alarm.ID,
		"tenant_id":       alarm.TenantID,
		"project_id":      alarm.ProjectID,
		"acknowledged_by": agentID,
		"acknowledged_at": alarm.AcknowledgedAt,
		"response":        response,
	}

	data, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Failed to marshal acknowledgment notification: %v", err)
		return
	}

	// Create websocket message for agents in the project
	wsMessage := &ws.Message{
		Type:      "alarm_acknowledged",
		Data:      data,
		Timestamp: time.Now(),
		TenantID:  &alarm.TenantID,
		ProjectID: &alarm.ProjectID,
		FromType:  ws.ConnectionTypeAgent,
	}

	err = s.connectionMgr.SendToProjectAgents(alarm.ProjectID, wsMessage)
	if err != nil {
		log.Printf("Failed to deliver acknowledgment notification to project %s: %v", alarm.ProjectID, err)
		return
	}

	log.Printf("Alarm acknowledgment notification sent to project agents: ID=%s, Project=%s",
		alarm.ID, alarm.ProjectID)
}

// stringToAlarmLevel converts string to AlarmLevel enum
func (s *HowlingAlarmService) stringToAlarmLevel(level string) AlarmLevel {
	switch level {
	case "soft":
		return AlarmLevelSoft
	case "medium":
		return AlarmLevelMedium
	case "loud":
		return AlarmLevelLoud
	case "urgent":
		return AlarmLevelUrgent
	case "critical":
		return AlarmLevelCritical
	default:
		return AlarmLevelSoft
	}
}
