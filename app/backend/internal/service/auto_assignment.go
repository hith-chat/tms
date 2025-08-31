package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/models"
)

// AssignmentStatus represents the status of an agent assignment
type AssignmentStatus string

const (
	AssignmentStatusPending    AssignmentStatus = "pending"
	AssignmentStatusAccepted   AssignmentStatus = "accepted"
	AssignmentStatusDeclined   AssignmentStatus = "declined"
	AssignmentStatusTimedOut   AssignmentStatus = "timed_out"
	AssignmentStatusCancelled  AssignmentStatus = "cancelled"
	AssignmentStatusTransferred AssignmentStatus = "transferred"
)

// AssignmentPriority represents the priority level of an assignment
type AssignmentPriority string

const (
	PriorityLow      AssignmentPriority = "low"
	PriorityNormal   AssignmentPriority = "normal"
	PriorityHigh     AssignmentPriority = "high"
	PriorityCritical AssignmentPriority = "critical"
)

// AgentAssignment represents an assignment of a customer to an agent
type AgentAssignment struct {
	ID                uuid.UUID          `json:"id"`
	TenantID          uuid.UUID          `json:"tenant_id"`
	ProjectID         uuid.UUID          `json:"project_id"`
	CustomerID        uuid.UUID          `json:"customer_id"`
	AgentID           uuid.UUID          `json:"agent_id"`
	SessionID         uuid.UUID          `json:"session_id"`
	RequestType       AgentRequestType   `json:"request_type"`
	UrgencyLevel      AgentRequestUrgency `json:"urgency_level"`
	Priority          AssignmentPriority `json:"priority"`
	Status            AssignmentStatus   `json:"status"`
	AssignedAt        time.Time          `json:"assigned_at"`
	AcceptedAt        *time.Time         `json:"accepted_at,omitempty"`
	CompletedAt       *time.Time         `json:"completed_at,omitempty"`
	TimeoutAt         time.Time          `json:"timeout_at"`
	Reason            string             `json:"reason"`
	CustomerMessage   string             `json:"customer_message"`
	AgentNotes        string             `json:"agent_notes,omitempty"`
	TransferReason    string             `json:"transfer_reason,omitempty"`
}

// AssignmentRequest contains the details for creating an agent assignment
type AssignmentRequest struct {
	TenantID        uuid.UUID          `json:"tenant_id"`
	ProjectID       uuid.UUID          `json:"project_id"`
	CustomerID      uuid.UUID          `json:"customer_id"`
	SessionID       uuid.UUID          `json:"session_id"`
	CustomerMessage string             `json:"customer_message"`
	RequestType     AgentRequestType   `json:"request_type"`
	UrgencyLevel    AgentRequestUrgency `json:"urgency_level"`
	RequiredSkills  []AgentSkill       `json:"required_skills"`
	PreferredAgent  *uuid.UUID         `json:"preferred_agent_id,omitempty"`
	ExcludeAgents   []uuid.UUID        `json:"exclude_agent_ids"`
	TimeoutMinutes  int                `json:"timeout_minutes"`
}

// AssignmentResult contains the result of an assignment attempt
type AssignmentResult struct {
	Assignment     *AgentAssignment `json:"assignment"`
	Success        bool             `json:"success"`
	Reason         string           `json:"reason"`
	SelectedAgent  *OnlineAgent     `json:"selected_agent,omitempty"`
	AlternateAgent *OnlineAgent     `json:"alternate_agent,omitempty"`
	RetryAfter     *time.Duration   `json:"retry_after,omitempty"`
	ProcessingTime time.Duration    `json:"processing_time"`
}

// AutoAssignmentService handles automatic assignment of customers to agents
type AutoAssignmentService struct {
	config                  *config.AgenticConfig
	agentRequestDetection   *AgentRequestDetectionService
	onlineAgentService      *OnlineAgentService
	enhancedNotificationSvc *EnhancedNotificationService
	assignments             map[uuid.UUID]*AgentAssignment
	// In a real implementation, this would be a database repository
}

// NewAutoAssignmentService creates a new auto assignment service
func NewAutoAssignmentService(
	agenticConfig *config.AgenticConfig,
	agentRequestDetection *AgentRequestDetectionService,
	onlineAgentService *OnlineAgentService,
	enhancedNotificationSvc *EnhancedNotificationService,
) *AutoAssignmentService {
	return &AutoAssignmentService{
		config:                  agenticConfig,
		agentRequestDetection:   agentRequestDetection,
		onlineAgentService:      onlineAgentService,
		enhancedNotificationSvc: enhancedNotificationSvc,
		assignments:             make(map[uuid.UUID]*AgentAssignment),
	}
}

// ProcessMessage analyzes a message and creates an assignment if needed
func (s *AutoAssignmentService) ProcessMessage(ctx context.Context, tenantID, projectID, customerID, sessionID uuid.UUID, message string) (*AssignmentResult, error) {
	startTime := time.Now()

	// Check if agent assignment is enabled
	if !s.config.AgentAssignment {
		return &AssignmentResult{
			Success:        false,
			Reason:         "Agent assignment is disabled",
			ProcessingTime: time.Since(startTime),
		}, nil
	}

	// Detect if the message is requesting an agent
	agentRequest, err := s.agentRequestDetection.DetectAgentRequest(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to detect agent request: %w", err)
	}

	if !agentRequest.IsAgentRequest {
		return &AssignmentResult{
			Success:        false,
			Reason:         "Message does not request an agent",
			ProcessingTime: time.Since(startTime),
		}, nil
	}

	// Create assignment request
	assignmentRequest := &AssignmentRequest{
		TenantID:        tenantID,
		ProjectID:       projectID,
		CustomerID:      customerID,
		SessionID:       sessionID,
		CustomerMessage: message,
		RequestType:     agentRequest.RequestType,
		UrgencyLevel:    agentRequest.Urgency,
		RequiredSkills:  s.getRequiredSkillsForRequest(agentRequest.RequestType),
		TimeoutMinutes:  s.getTimeoutForUrgency(agentRequest.Urgency),
	}

	// Attempt to assign an agent
	result, err := s.CreateAssignment(ctx, assignmentRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create assignment: %w", err)
	}

	result.ProcessingTime = time.Since(startTime)
	return result, nil
}

// CreateAssignment creates a new agent assignment
func (s *AutoAssignmentService) CreateAssignment(ctx context.Context, request *AssignmentRequest) (*AssignmentResult, error) {
	startTime := time.Now()

	// Build selection criteria
	criteria := AgentSelectionCriteria{
		RequiredSkills:  request.RequiredSkills,
		UrgencyLevel:    request.UrgencyLevel,
		RequestType:     request.RequestType,
		PreferredAgent:  request.PreferredAgent,
		ExcludeAgents:   request.ExcludeAgents,
		MaxResponseTime: s.getMaxResponseTimeForUrgency(request.UrgencyLevel),
	}

	// Select best available agent
	selection, err := s.onlineAgentService.SelectBestAgent(ctx, request.TenantID, criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to select agent: %w", err)
	}

	if selection.SelectedAgent == nil {
		// No agents available
		retryAfter := s.getRetryAfterForNoAgents(request.UrgencyLevel)
		return &AssignmentResult{
			Success:        false,
			Reason:         "No suitable agents available",
			RetryAfter:     &retryAfter,
			ProcessingTime: time.Since(startTime),
		}, nil
	}

	// Create the assignment
	assignment := &AgentAssignment{
		ID:              uuid.New(),
		TenantID:        request.TenantID,
		ProjectID:       request.ProjectID,
		CustomerID:      request.CustomerID,
		AgentID:         selection.SelectedAgent.UserID,
		SessionID:       request.SessionID,
		RequestType:     request.RequestType,
		UrgencyLevel:    request.UrgencyLevel,
		Priority:        s.urgencyToPriority(request.UrgencyLevel),
		Status:          AssignmentStatusPending,
		AssignedAt:      time.Now(),
		TimeoutAt:       time.Now().Add(time.Duration(request.TimeoutMinutes) * time.Minute),
		Reason:          fmt.Sprintf("Auto-assigned based on %s", selection.Reason),
		CustomerMessage: request.CustomerMessage,
	}

	// Store the assignment (in real implementation, this would be in database)
	s.assignments[assignment.ID] = assignment

	// Update agent workload
	err = s.onlineAgentService.UpdateAgentWorkload(ctx, selection.SelectedAgent.UserID, selection.SelectedAgent.ActiveChats+1)
	if err != nil {
		log.Warn().Err(err).Str("agent_id", selection.SelectedAgent.UserID.String()).Msg("Failed to update agent workload after assignment")
	}

	log.Info().
		Str("assignment_id", assignment.ID.String()).
		Str("tenant_id", request.TenantID.String()).
		Str("customer_id", request.CustomerID.String()).
		Str("agent_id", assignment.AgentID.String()).
		Str("agent_name", selection.SelectedAgent.Name).
		Str("request_type", string(request.RequestType)).
		Str("urgency", string(request.UrgencyLevel)).
		Str("priority", string(assignment.Priority)).
		Dur("timeout", time.Until(assignment.TimeoutAt)).
		Msg("Agent assignment created")

	// Send enhanced notification with potential alarm
	if s.enhancedNotificationSvc != nil {
		notificationTitle := fmt.Sprintf("New %s Assignment", request.RequestType)
		notificationMessage := fmt.Sprintf("You have been assigned to assist a customer. Priority: %s", assignment.Priority)
		
		if request.CustomerMessage != "" {
			notificationMessage += fmt.Sprintf("\nCustomer message: %s", s.truncateMessage(request.CustomerMessage, 100))
		}

		priority := s.assignmentPriorityToNotificationPriority(assignment.Priority)
		err = s.enhancedNotificationSvc.CreateAgentAssignmentNotification(
			ctx, assignment.TenantID, assignment.ProjectID, assignment.AgentID, assignment.ID,
			notificationTitle, notificationMessage, priority, string(request.UrgencyLevel))
		if err != nil {
			log.Warn().Err(err).Str("assignment_id", assignment.ID.String()).Msg("Failed to send assignment notification")
		}
	}

	return &AssignmentResult{
		Assignment:     assignment,
		Success:        true,
		Reason:         fmt.Sprintf("Successfully assigned to %s", selection.SelectedAgent.Name),
		SelectedAgent:  selection.SelectedAgent,
		AlternateAgent: selection.AlternateAgent,
		ProcessingTime: time.Since(startTime),
	}, nil
}

// AcceptAssignment marks an assignment as accepted by the agent
func (s *AutoAssignmentService) AcceptAssignment(ctx context.Context, assignmentID, agentID uuid.UUID) error {
	assignment, exists := s.assignments[assignmentID]
	if !exists {
		return errors.New("assignment not found")
	}

	if assignment.AgentID != agentID {
		return errors.New("assignment not assigned to this agent")
	}

	if assignment.Status != AssignmentStatusPending {
		return fmt.Errorf("assignment status is %s, cannot accept", assignment.Status)
	}

	now := time.Now()
	assignment.Status = AssignmentStatusAccepted
	assignment.AcceptedAt = &now

	log.Info().
		Str("assignment_id", assignmentID.String()).
		Str("agent_id", agentID.String()).
		Dur("response_time", time.Since(assignment.AssignedAt)).
		Msg("Agent assignment accepted")

	return nil
}

// DeclineAssignment marks an assignment as declined and attempts reassignment
func (s *AutoAssignmentService) DeclineAssignment(ctx context.Context, assignmentID, agentID uuid.UUID, reason string) (*AssignmentResult, error) {
	assignment, exists := s.assignments[assignmentID]
	if !exists {
		return nil, errors.New("assignment not found")
	}

	if assignment.AgentID != agentID {
		return nil, errors.New("assignment not assigned to this agent")
	}

	if assignment.Status != AssignmentStatusPending {
		return nil, fmt.Errorf("assignment status is %s, cannot decline", assignment.Status)
	}

	assignment.Status = AssignmentStatusDeclined
	assignment.AgentNotes = reason

	log.Info().
		Str("assignment_id", assignmentID.String()).
		Str("agent_id", agentID.String()).
		Str("decline_reason", reason).
		Msg("Agent assignment declined")

	// Attempt reassignment with the declining agent excluded
	reassignmentRequest := &AssignmentRequest{
		TenantID:        assignment.TenantID,
		ProjectID:       assignment.ProjectID,
		CustomerID:      assignment.CustomerID,
		SessionID:       assignment.SessionID,
		CustomerMessage: assignment.CustomerMessage,
		RequestType:     assignment.RequestType,
		UrgencyLevel:    assignment.UrgencyLevel,
		RequiredSkills:  s.getRequiredSkillsForRequest(assignment.RequestType),
		ExcludeAgents:   []uuid.UUID{agentID},
		TimeoutMinutes:  s.getTimeoutForUrgency(assignment.UrgencyLevel),
	}

	return s.CreateAssignment(ctx, reassignmentRequest)
}

// TimeoutAssignment marks assignments as timed out
func (s *AutoAssignmentService) TimeoutAssignment(ctx context.Context, assignmentID uuid.UUID) error {
	assignment, exists := s.assignments[assignmentID]
	if !exists {
		return errors.New("assignment not found")
	}

	if assignment.Status != AssignmentStatusPending {
		return fmt.Errorf("assignment status is %s, cannot timeout", assignment.Status)
	}

	assignment.Status = AssignmentStatusTimedOut

	log.Warn().
		Str("assignment_id", assignmentID.String()).
		Str("agent_id", assignment.AgentID.String()).
		Dur("duration", time.Since(assignment.AssignedAt)).
		Msg("Agent assignment timed out")

	return nil
}

// GetAssignment retrieves an assignment by ID
func (s *AutoAssignmentService) GetAssignment(ctx context.Context, assignmentID uuid.UUID) (*AgentAssignment, error) {
	assignment, exists := s.assignments[assignmentID]
	if !exists {
		return nil, errors.New("assignment not found")
	}

	return assignment, nil
}

// GetAgentAssignments retrieves all assignments for an agent
func (s *AutoAssignmentService) GetAgentAssignments(ctx context.Context, agentID uuid.UUID) ([]*AgentAssignment, error) {
	var agentAssignments []*AgentAssignment

	for _, assignment := range s.assignments {
		if assignment.AgentID == agentID {
			agentAssignments = append(agentAssignments, assignment)
		}
	}

	return agentAssignments, nil
}

// GetPendingAssignments retrieves all pending assignments for a tenant
func (s *AutoAssignmentService) GetPendingAssignments(ctx context.Context, tenantID uuid.UUID) ([]*AgentAssignment, error) {
	var pendingAssignments []*AgentAssignment

	for _, assignment := range s.assignments {
		if assignment.TenantID == tenantID && assignment.Status == AssignmentStatusPending {
			pendingAssignments = append(pendingAssignments, assignment)
		}
	}

	return pendingAssignments, nil
}

// GetAssignmentStats returns statistics about assignments
func (s *AutoAssignmentService) GetAssignmentStats(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"total_assignments":     0,
		"pending_assignments":   0,
		"accepted_assignments":  0,
		"declined_assignments":  0,
		"timed_out_assignments": 0,
		"avg_acceptance_time":   0.0,
		"assignment_by_type":    make(map[AgentRequestType]int),
		"assignment_by_urgency": make(map[AgentRequestUrgency]int),
	}

	var acceptanceTimes []time.Duration
	assignmentByType := make(map[AgentRequestType]int)
	assignmentByUrgency := make(map[AgentRequestUrgency]int)

	for _, assignment := range s.assignments {
		if assignment.TenantID != tenantID {
			continue
		}

		stats["total_assignments"] = stats["total_assignments"].(int) + 1

		switch assignment.Status {
		case AssignmentStatusPending:
			stats["pending_assignments"] = stats["pending_assignments"].(int) + 1
		case AssignmentStatusAccepted:
			stats["accepted_assignments"] = stats["accepted_assignments"].(int) + 1
			if assignment.AcceptedAt != nil {
				acceptanceTimes = append(acceptanceTimes, assignment.AcceptedAt.Sub(assignment.AssignedAt))
			}
		case AssignmentStatusDeclined:
			stats["declined_assignments"] = stats["declined_assignments"].(int) + 1
		case AssignmentStatusTimedOut:
			stats["timed_out_assignments"] = stats["timed_out_assignments"].(int) + 1
		}

		assignmentByType[assignment.RequestType]++
		assignmentByUrgency[assignment.UrgencyLevel]++
	}

	// Calculate average acceptance time
	if len(acceptanceTimes) > 0 {
		var totalTime time.Duration
		for _, duration := range acceptanceTimes {
			totalTime += duration
		}
		avgAcceptanceTime := totalTime / time.Duration(len(acceptanceTimes))
		stats["avg_acceptance_time"] = avgAcceptanceTime.Seconds()
	}

	stats["assignment_by_type"] = assignmentByType
	stats["assignment_by_urgency"] = assignmentByUrgency

	return stats, nil
}

// Helper methods

func (s *AutoAssignmentService) getRequiredSkillsForRequest(requestType AgentRequestType) []AgentSkill {
	switch requestType {
	case AgentRequestTypeTechnical:
		return []AgentSkill{SkillTechnical}
	case AgentRequestTypeBilling:
		return []AgentSkill{SkillBilling}
	case AgentRequestTypeComplaint:
		return []AgentSkill{SkillComplaint}
	case AgentRequestTypeSupport:
		return []AgentSkill{SkillSupport}
	default:
		return []AgentSkill{SkillGeneral}
	}
}

func (s *AutoAssignmentService) getTimeoutForUrgency(urgency AgentRequestUrgency) int {
	switch urgency {
	case UrgencyCritical:
		return 1 // 1 minute
	case UrgencyHigh:
		return 3 // 3 minutes
	case UrgencyNormal:
		return 5 // 5 minutes
	case UrgencyLow:
		return 10 // 10 minutes
	default:
		return 5
	}
}

func (s *AutoAssignmentService) getMaxResponseTimeForUrgency(urgency AgentRequestUrgency) float64 {
	switch urgency {
	case UrgencyCritical:
		return 10.0 // 10 seconds
	case UrgencyHigh:
		return 30.0 // 30 seconds
	case UrgencyNormal:
		return 60.0 // 60 seconds
	case UrgencyLow:
		return 120.0 // 2 minutes
	default:
		return 60.0
	}
}

func (s *AutoAssignmentService) getRetryAfterForNoAgents(urgency AgentRequestUrgency) time.Duration {
	switch urgency {
	case UrgencyCritical:
		return 30 * time.Second
	case UrgencyHigh:
		return 1 * time.Minute
	case UrgencyNormal:
		return 2 * time.Minute
	case UrgencyLow:
		return 5 * time.Minute
	default:
		return 2 * time.Minute
	}
}

func (s *AutoAssignmentService) urgencyToPriority(urgency AgentRequestUrgency) AssignmentPriority {
	switch urgency {
	case UrgencyCritical:
		return PriorityCritical
	case UrgencyHigh:
		return PriorityHigh
	case UrgencyNormal:
		return PriorityNormal
	case UrgencyLow:
		return PriorityLow
	default:
		return PriorityNormal
	}
}

// truncateMessage truncates a message to the specified length
func (s *AutoAssignmentService) truncateMessage(message string, maxLength int) string {
	if len(message) <= maxLength {
		return message
	}
	return message[:maxLength-3] + "..."
}

// assignmentPriorityToNotificationPriority converts assignment priority to notification priority
func (s *AutoAssignmentService) assignmentPriorityToNotificationPriority(priority AssignmentPriority) models.NotificationPriority {
	switch priority {
	case PriorityLow:
		return models.NotificationPriorityLow
	case PriorityNormal:
		return models.NotificationPriorityNormal
	case PriorityHigh:
		return models.NotificationPriorityHigh
	case PriorityCritical:
		return models.NotificationPriorityCritical
	default:
		return models.NotificationPriorityNormal
	}
}
