package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/models"
	ws "github.com/bareuptime/tms/internal/websocket"
)

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
	InitialLevel          AlarmLevel        `json:"initial_level"`
	EscalationInterval    time.Duration     `json:"escalation_interval"`
	MaxLevel              AlarmLevel        `json:"max_level"`
	PersistUntilAcknowled bool              `json:"persist_until_acknowledged"`
	AudioEnabled          bool              `json:"audio_enabled"`
	VisualEnabled         bool              `json:"visual_enabled"`
	BroadcastToAll        bool              `json:"broadcast_to_all"`
}

// ActiveAlarm represents an ongoing alarm
type ActiveAlarm struct {
	ID               uuid.UUID             `json:"id"`
	TenantID         uuid.UUID             `json:"tenant_id"`
	AssignmentID     uuid.UUID             `json:"assignment_id"`
	AgentID          uuid.UUID             `json:"agent_id"`
	Title            string                `json:"title"`
	Message          string                `json:"message"`
	Priority         models.NotificationPriority `json:"priority"`
	CurrentLevel     AlarmLevel            `json:"current_level"`
	StartTime        time.Time             `json:"start_time"`
	LastEscalation   time.Time             `json:"last_escalation"`
	EscalationCount  int                   `json:"escalation_count"`
	IsAcknowledged   bool                  `json:"is_acknowledged"`
	AcknowledgedAt   *time.Time            `json:"acknowledged_at,omitempty"`
	AcknowledgedBy   *uuid.UUID            `json:"acknowledged_by,omitempty"`
	Config           AlarmEscalationConfig `json:"config"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// AlarmAcknowledgment represents an alarm acknowledgment
type AlarmAcknowledgment struct {
	AlarmID       uuid.UUID  `json:"alarm_id"`
	AgentID       uuid.UUID  `json:"agent_id"`
	AcknowledgedAt time.Time `json:"acknowledged_at"`
	Response      string     `json:"response,omitempty"`
}

// HowlingAlarmService manages critical notifications with escalating alerts
type HowlingAlarmService struct {
	config        *config.Config
	connectionMgr *ws.ConnectionManager
	activeAlarms  map[uuid.UUID]*ActiveAlarm
	alarmMutex    sync.RWMutex
	escalationTicker *time.Ticker
	stopChannel   chan bool
}

// NewHowlingAlarmService creates a new howling alarm service
func NewHowlingAlarmService(cfg *config.Config, connectionMgr *ws.ConnectionManager) *HowlingAlarmService {
	service := &HowlingAlarmService{
		config:        cfg,
		connectionMgr: connectionMgr,
		activeAlarms:  make(map[uuid.UUID]*ActiveAlarm),
		stopChannel:   make(chan bool, 1),
	}

	// Start the escalation ticker
	service.startEscalationTicker()

	log.Printf("HowlingAlarmService initialized successfully")
	return service
}

// TriggerAlarm triggers a new howling alarm
func (s *HowlingAlarmService) TriggerAlarm(ctx context.Context, assignmentID, agentID, tenantID uuid.UUID, 
	title, message string, priority models.NotificationPriority, metadata map[string]interface{}) (*ActiveAlarm, error) {

	s.alarmMutex.Lock()
	defer s.alarmMutex.Unlock()

	// Create alarm configuration based on priority
	config := s.getAlarmConfigForPriority(priority)

	alarm := &ActiveAlarm{
		ID:               uuid.New(),
		TenantID:         tenantID,
		AssignmentID:     assignmentID,
		AgentID:          agentID,
		Title:            title,
		Message:          message,
		Priority:         priority,
		CurrentLevel:     config.InitialLevel,
		StartTime:        time.Now(),
		LastEscalation:   time.Now(),
		EscalationCount:  0,
		IsAcknowledged:   false,
		Config:           config,
		Metadata:         metadata,
	}

	// Store active alarm
	s.activeAlarms[alarm.ID] = alarm

	// Send initial notification
	s.sendAlarmNotification(alarm)

	log.Printf("Howling alarm triggered: ID=%s, Assignment=%s, Agent=%s, Level=%s", 
		alarm.ID, assignmentID, agentID, alarm.CurrentLevel)

	return alarm, nil
}

// AcknowledgeAlarm acknowledges an alarm and stops escalation
func (s *HowlingAlarmService) AcknowledgeAlarm(ctx context.Context, alarmID, agentID uuid.UUID, response string) error {
	s.alarmMutex.Lock()
	defer s.alarmMutex.Unlock()

	alarm, exists := s.activeAlarms[alarmID]
	if !exists {
		return fmt.Errorf("alarm not found: %s", alarmID)
	}

	if alarm.IsAcknowledged {
		return fmt.Errorf("alarm already acknowledged: %s", alarmID)
	}

	// Mark as acknowledged
	now := time.Now()
	alarm.IsAcknowledged = true
	alarm.AcknowledgedAt = &now
	alarm.AcknowledgedBy = &agentID

	// Send acknowledgment notification
	s.sendAlarmAcknowledgment(alarm, response)

	// Remove from active alarms
	delete(s.activeAlarms, alarmID)

	log.Printf("Alarm acknowledged: ID=%s, Agent=%s, Response=%s", alarmID, agentID, response)

	return nil
}

// GetActiveAlarms returns all active alarms for a tenant
func (s *HowlingAlarmService) GetActiveAlarms(tenantID uuid.UUID) []*ActiveAlarm {
	s.alarmMutex.RLock()
	defer s.alarmMutex.RUnlock()

	var alarms []*ActiveAlarm
	for _, alarm := range s.activeAlarms {
		if alarm.TenantID == tenantID && !alarm.IsAcknowledged {
			alarms = append(alarms, alarm)
		}
	}

	return alarms
}

// GetAlarmStats returns statistics about alarms
func (s *HowlingAlarmService) GetAlarmStats(tenantID uuid.UUID) map[string]interface{} {
	s.alarmMutex.RLock()
	defer s.alarmMutex.RUnlock()

	stats := map[string]interface{}{
		"total_active":      0,
		"by_level":          make(map[string]int),
		"by_priority":       make(map[string]int),
		"average_duration":  0,
		"escalation_counts": make(map[string]int),
	}

	var totalDuration time.Duration
	activeCount := 0

	for _, alarm := range s.activeAlarms {
		if alarm.TenantID == tenantID && !alarm.IsAcknowledged {
			activeCount++
			
			// Count by level
			levelKey := alarm.CurrentLevel.String()
			if count, ok := stats["by_level"].(map[string]int)[levelKey]; ok {
				stats["by_level"].(map[string]int)[levelKey] = count + 1
			} else {
				stats["by_level"].(map[string]int)[levelKey] = 1
			}

			// Count by priority
			priorityKey := string(alarm.Priority)
			if count, ok := stats["by_priority"].(map[string]int)[priorityKey]; ok {
				stats["by_priority"].(map[string]int)[priorityKey] = count + 1
			} else {
				stats["by_priority"].(map[string]int)[priorityKey] = 1
			}

			// Calculate duration
			duration := time.Since(alarm.StartTime)
			totalDuration += duration

			// Count escalations
			escalationKey := fmt.Sprintf("%d_escalations", alarm.EscalationCount)
			if count, ok := stats["escalation_counts"].(map[string]int)[escalationKey]; ok {
				stats["escalation_counts"].(map[string]int)[escalationKey] = count + 1
			} else {
				stats["escalation_counts"].(map[string]int)[escalationKey] = 1
			}
		}
	}

	stats["total_active"] = activeCount
	if activeCount > 0 {
		stats["average_duration"] = totalDuration.Seconds() / float64(activeCount)
	}

	return stats
}

// startEscalationTicker starts the background escalation process
func (s *HowlingAlarmService) startEscalationTicker() {
	s.escalationTicker = time.NewTicker(30 * time.Second) // Check every 30 seconds
	
	go func() {
		for {
			select {
			case <-s.escalationTicker.C:
				s.processEscalations()
			case <-s.stopChannel:
				s.escalationTicker.Stop()
				return
			}
		}
	}()
}

// processEscalations processes escalation for all active alarms
func (s *HowlingAlarmService) processEscalations() {
	s.alarmMutex.Lock()
	defer s.alarmMutex.Unlock()

	now := time.Now()
	
	for _, alarm := range s.activeAlarms {
		if alarm.IsAcknowledged {
			continue
		}

		// Check if escalation is needed
		timeSinceLastEscalation := now.Sub(alarm.LastEscalation)
		if timeSinceLastEscalation >= alarm.Config.EscalationInterval {
			s.escalateAlarm(alarm)
		}
	}
}

// escalateAlarm escalates an alarm to the next level
func (s *HowlingAlarmService) escalateAlarm(alarm *ActiveAlarm) {
	if alarm.CurrentLevel < alarm.Config.MaxLevel {
		alarm.CurrentLevel++
		alarm.EscalationCount++
		alarm.LastEscalation = time.Now()

		// Send escalated notification
		s.sendAlarmNotification(alarm)

		log.Printf("Alarm escalated: ID=%s, New Level=%s, Escalation Count=%d", 
			alarm.ID, alarm.CurrentLevel, alarm.EscalationCount)
	}
}

// getAlarmConfigForPriority returns alarm configuration based on priority
func (s *HowlingAlarmService) getAlarmConfigForPriority(priority models.NotificationPriority) AlarmEscalationConfig {
	switch priority {
	case models.NotificationPriorityLow:
		return AlarmEscalationConfig{
			InitialLevel:          AlarmLevelSoft,
			EscalationInterval:    5 * time.Minute,
			MaxLevel:              AlarmLevelMedium,
			PersistUntilAcknowled: true,
			AudioEnabled:          false,
			VisualEnabled:         true,
			BroadcastToAll:        false,
		}
	case models.NotificationPriorityNormal:
		return AlarmEscalationConfig{
			InitialLevel:          AlarmLevelMedium,
			EscalationInterval:    3 * time.Minute,
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
	default:
		return AlarmEscalationConfig{
			InitialLevel:          AlarmLevelSoft,
			EscalationInterval:    5 * time.Minute,
			MaxLevel:              AlarmLevelMedium,
			PersistUntilAcknowled: true,
			AudioEnabled:          false,
			VisualEnabled:         true,
			BroadcastToAll:        false,
		}
	}
}

// sendAlarmNotification sends an alarm notification via WebSocket
func (s *HowlingAlarmService) sendAlarmNotification(alarm *ActiveAlarm) {
	// Skip WebSocket delivery if connection manager is nil (for testing)
	if s.connectionMgr == nil {
		log.Printf("Skipping alarm notification delivery (no connection manager) for alarm: %s", alarm.Title)
		return
	}

	// Create alarm notification message
	alarmData := map[string]interface{}{
		"type":              "howling_alarm",
		"alarm_id":          alarm.ID,
		"assignment_id":     alarm.AssignmentID,
		"title":             alarm.Title,
		"message":           alarm.Message,
		"priority":          alarm.Priority,
		"current_level":     alarm.CurrentLevel.String(),
		"escalation_count":  alarm.EscalationCount,
		"start_time":        alarm.StartTime,
		"config":            alarm.Config,
		"metadata":          alarm.Metadata,
	}

	// Marshal alarm data
	alarmDataBytes, err := json.Marshal(alarmData)
	if err != nil {
		log.Printf("Failed to marshal alarm data: %v", err)
		return
	}

	// Determine target agents
	var targetAgentID *uuid.UUID
	if !alarm.Config.BroadcastToAll {
		targetAgentID = &alarm.AgentID
	}

	wsMessage := &ws.Message{
		Type:      "howling_alarm",
		Data:      json.RawMessage(alarmDataBytes),
		SessionID: uuid.Nil,
		FromType:  ws.ConnectionTypeAgent,
		TenantID:  &alarm.TenantID,
		AgentID:   targetAgentID,
	}

	// Deliver via WebSocket
	err = s.connectionMgr.DeliverWebSocketMessage(uuid.Nil, wsMessage)
	if err != nil {
		log.Printf("Failed to deliver alarm notification via WebSocket: %v", err)
	}
}

// sendAlarmAcknowledgment sends alarm acknowledgment notification
func (s *HowlingAlarmService) sendAlarmAcknowledgment(alarm *ActiveAlarm, response string) {
	// Skip WebSocket delivery if connection manager is nil (for testing)
	if s.connectionMgr == nil {
		log.Printf("Skipping alarm acknowledgment delivery (no connection manager) for alarm: %s", alarm.ID)
		return
	}

	ackData := map[string]interface{}{
		"type":             "alarm_acknowledged",
		"alarm_id":         alarm.ID,
		"assignment_id":    alarm.AssignmentID,
		"acknowledged_by":  alarm.AcknowledgedBy,
		"acknowledged_at":  alarm.AcknowledgedAt,
		"response":         response,
		"final_level":      alarm.CurrentLevel.String(),
		"total_duration":   time.Since(alarm.StartTime).Seconds(),
		"escalation_count": alarm.EscalationCount,
	}

	// Marshal acknowledgment data
	ackDataBytes, err := json.Marshal(ackData)
	if err != nil {
		log.Printf("Failed to marshal acknowledgment data: %v", err)
		return
	}

	wsMessage := &ws.Message{
		Type:      "alarm_acknowledged",
		Data:      json.RawMessage(ackDataBytes),
		SessionID: uuid.Nil,
		FromType:  ws.ConnectionTypeAgent,
		TenantID:  &alarm.TenantID,
	}

	// Broadcast acknowledgment to all agents in tenant
	err = s.connectionMgr.DeliverWebSocketMessage(uuid.Nil, wsMessage)
	if err != nil {
		log.Printf("Failed to deliver acknowledgment notification via WebSocket: %v", err)
	}
}

// Stop stops the howling alarm service
func (s *HowlingAlarmService) Stop() {
	if s.escalationTicker != nil {
		s.stopChannel <- true
	}
	log.Printf("HowlingAlarmService stopped")
}
