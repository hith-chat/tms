package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/bareuptime/tms/internal/config"
)

// AgentStatus represents the current status of an agent
type AgentStatus string

const (
	AgentStatusOnline      AgentStatus = "online"
	AgentStatusAway        AgentStatus = "away"
	AgentStatusBusy        AgentStatus = "busy"
	AgentStatusOffline     AgentStatus = "offline"
	AgentStatusDoNotDisturb AgentStatus = "dnd"
)

// AgentSkill represents skills an agent possesses
type AgentSkill string

const (
	SkillGeneral    AgentSkill = "general"
	SkillTechnical  AgentSkill = "technical"
	SkillBilling    AgentSkill = "billing"
	SkillSupport    AgentSkill = "support"
	SkillSales      AgentSkill = "sales"
	SkillComplaint  AgentSkill = "complaint"
)

// OnlineAgent represents an agent that is currently available
type OnlineAgent struct {
	UserID          uuid.UUID     `json:"user_id"`
	TenantID        uuid.UUID     `json:"tenant_id"`
	Name            string        `json:"name"`
	Email           string        `json:"email"`
	Status          AgentStatus   `json:"status"`
	Skills          []AgentSkill  `json:"skills"`
	ActiveChats     int           `json:"active_chats"`
	MaxChats        int           `json:"max_chats"`
	AvgResponseTime float64       `json:"avg_response_time_seconds"`
	LastActivity    time.Time     `json:"last_activity"`
	LastAssignment  time.Time     `json:"last_assignment"`
	Workload        float64       `json:"workload"` // 0.0 to 1.0 representing capacity usage
}

// AgentSelectionCriteria defines criteria for selecting an agent
type AgentSelectionCriteria struct {
	RequiredSkills  []AgentSkill            `json:"required_skills"`
	UrgencyLevel    AgentRequestUrgency     `json:"urgency_level"`
	RequestType     AgentRequestType        `json:"request_type"`
	PreferredAgent  *uuid.UUID              `json:"preferred_agent_id,omitempty"`
	ExcludeAgents   []uuid.UUID             `json:"exclude_agent_ids"`
	MaxResponseTime float64                 `json:"max_response_time_seconds"`
	MinRating       float64                 `json:"min_rating"`
}

// AgentSelectionResult contains the result of agent selection
type AgentSelectionResult struct {
	SelectedAgent  *OnlineAgent  `json:"selected_agent"`
	AlternateAgent *OnlineAgent  `json:"alternate_agent,omitempty"`
	Reason         string        `json:"reason"`
	Score          float64       `json:"score"`
	TotalAgents    int           `json:"total_agents_available"`
	ProcessingTime time.Duration `json:"processing_time"`
}

// OnlineAgentService manages online agent discovery and selection
type OnlineAgentService struct {
	config         *config.AgenticConfig
	agents         map[uuid.UUID]*OnlineAgent
	mu             sync.RWMutex
	lastCleanup    time.Time
	cleanupInterval time.Duration
}

// NewOnlineAgentService creates a new online agent service
func NewOnlineAgentService(agenticConfig *config.AgenticConfig) *OnlineAgentService {
	service := &OnlineAgentService{
		config:          agenticConfig,
		agents:          make(map[uuid.UUID]*OnlineAgent),
		cleanupInterval: 5 * time.Minute,
		lastCleanup:     time.Now(),
	}

	// Start background cleanup routine
	go service.backgroundCleanup()

	return service
}

// RegisterAgent registers an agent as online
func (s *OnlineAgentService) RegisterAgent(ctx context.Context, agent *OnlineAgent) error {
	if agent.UserID == uuid.Nil {
		return errors.New("agent user ID cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	agent.LastActivity = time.Now()
	s.agents[agent.UserID] = agent

	log.Info().
		Str("agent_id", agent.UserID.String()).
		Str("name", agent.Name).
		Str("status", string(agent.Status)).
		Int("max_chats", agent.MaxChats).
		Msg("Agent registered as online")

	return nil
}

// UpdateAgentStatus updates an agent's status and activity
func (s *OnlineAgentService) UpdateAgentStatus(ctx context.Context, agentID uuid.UUID, status AgentStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	agent, exists := s.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	agent.Status = status
	agent.LastActivity = time.Now()

	log.Debug().
		Str("agent_id", agentID.String()).
		Str("old_status", string(agent.Status)).
		Str("new_status", string(status)).
		Msg("Agent status updated")

	return nil
}

// UpdateAgentWorkload updates an agent's current workload
func (s *OnlineAgentService) UpdateAgentWorkload(ctx context.Context, agentID uuid.UUID, activeChats int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	agent, exists := s.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	agent.ActiveChats = activeChats
	agent.LastActivity = time.Now()
	
	// Calculate workload percentage
	if agent.MaxChats > 0 {
		agent.Workload = float64(activeChats) / float64(agent.MaxChats)
	} else {
		agent.Workload = 0.0
	}

	log.Debug().
		Str("agent_id", agentID.String()).
		Int("active_chats", activeChats).
		Int("max_chats", agent.MaxChats).
		Float64("workload", agent.Workload).
		Msg("Agent workload updated")

	return nil
}

// UnregisterAgent removes an agent from online status
func (s *OnlineAgentService) UnregisterAgent(ctx context.Context, agentID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.agents[agentID]; !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	delete(s.agents, agentID)

	log.Info().
		Str("agent_id", agentID.String()).
		Msg("Agent unregistered")

	return nil
}

// GetAvailableAgents returns all currently available agents
func (s *OnlineAgentService) GetAvailableAgents(ctx context.Context, tenantID uuid.UUID) ([]*OnlineAgent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var availableAgents []*OnlineAgent

	for _, agent := range s.agents {
		if agent.TenantID == tenantID && s.isAgentAvailable(agent) {
			// Create a copy to avoid data races
			agentCopy := *agent
			availableAgents = append(availableAgents, &agentCopy)
		}
	}

	log.Debug().
		Str("tenant_id", tenantID.String()).
		Int("available_agents", len(availableAgents)).
		Msg("Retrieved available agents")

	return availableAgents, nil
}

// SelectBestAgent selects the best available agent based on criteria
func (s *OnlineAgentService) SelectBestAgent(ctx context.Context, tenantID uuid.UUID, criteria AgentSelectionCriteria) (*AgentSelectionResult, error) {
	startTime := time.Now()

	availableAgents, err := s.GetAvailableAgents(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get available agents: %w", err)
	}

	if len(availableAgents) == 0 {
		return &AgentSelectionResult{
			SelectedAgent:  nil,
			Reason:         "No agents currently available",
			TotalAgents:    0,
			ProcessingTime: time.Since(startTime),
		}, nil
	}

	// Filter agents based on criteria
	candidateAgents := s.filterAgentsByCriteria(availableAgents, criteria)

	if len(candidateAgents) == 0 {
		return &AgentSelectionResult{
			SelectedAgent:  nil,
			Reason:         "No agents match the selection criteria",
			TotalAgents:    len(availableAgents),
			ProcessingTime: time.Since(startTime),
		}, nil
	}

	// Score and rank agents
	scoredAgents := s.scoreAgents(candidateAgents, criteria)

	// Select the best agent
	bestAgent := scoredAgents[0]
	var alternateAgent *OnlineAgent
	if len(scoredAgents) > 1 {
		alternateAgent = scoredAgents[1].agent
	}

	result := &AgentSelectionResult{
		SelectedAgent:  bestAgent.agent,
		AlternateAgent: alternateAgent,
		Reason:         fmt.Sprintf("Best match based on %s", bestAgent.reason),
		Score:          bestAgent.score,
		TotalAgents:    len(availableAgents),
		ProcessingTime: time.Since(startTime),
	}

	log.Info().
		Str("tenant_id", tenantID.String()).
		Str("selected_agent_id", bestAgent.agent.UserID.String()).
		Str("agent_name", bestAgent.agent.Name).
		Float64("score", bestAgent.score).
		Str("reason", bestAgent.reason).
		Int("total_candidates", len(candidateAgents)).
		Dur("processing_time", result.ProcessingTime).
		Msg("Agent selected")

	return result, nil
}

// GetAgentStats returns statistics about online agents
func (s *OnlineAgentService) GetAgentStats(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]interface{}{
		"total_agents":     0,
		"online_agents":    0,
		"available_agents": 0,
		"busy_agents":      0,
		"avg_workload":     0.0,
		"skills_coverage":  make(map[AgentSkill]int),
	}

	totalWorkload := 0.0
	skillsCoverage := make(map[AgentSkill]int)

	for _, agent := range s.agents {
		if agent.TenantID != tenantID {
			continue
		}

		stats["total_agents"] = stats["total_agents"].(int) + 1

		if agent.Status == AgentStatusOnline || agent.Status == AgentStatusAway {
			stats["online_agents"] = stats["online_agents"].(int) + 1
		}

		if s.isAgentAvailable(agent) {
			stats["available_agents"] = stats["available_agents"].(int) + 1
		}

		if agent.Workload >= 0.8 {
			stats["busy_agents"] = stats["busy_agents"].(int) + 1
		}

		totalWorkload += agent.Workload

		for _, skill := range agent.Skills {
			skillsCoverage[skill]++
		}
	}

	if stats["total_agents"].(int) > 0 {
		stats["avg_workload"] = totalWorkload / float64(stats["total_agents"].(int))
	}

	stats["skills_coverage"] = skillsCoverage

	return stats, nil
}

// isAgentAvailable checks if an agent is available for assignment
func (s *OnlineAgentService) isAgentAvailable(agent *OnlineAgent) bool {
	// Check if agent is in available status
	if agent.Status != AgentStatusOnline && agent.Status != AgentStatusAway {
		return false
	}

	// Check if agent is not at max capacity
	if agent.ActiveChats >= agent.MaxChats {
		return false
	}

	// Check if workload is reasonable
	if agent.Workload >= 1.0 {
		return false
	}

	// Check if agent has been active recently (within 10 minutes)
	if time.Since(agent.LastActivity) > 10*time.Minute {
		return false
	}

	return true
}

// filterAgentsByCriteria filters agents based on selection criteria
func (s *OnlineAgentService) filterAgentsByCriteria(agents []*OnlineAgent, criteria AgentSelectionCriteria) []*OnlineAgent {
	var filtered []*OnlineAgent

	for _, agent := range agents {
		// Skip excluded agents
		excluded := false
		for _, excludeID := range criteria.ExcludeAgents {
			if agent.UserID == excludeID {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		// Check required skills
		if len(criteria.RequiredSkills) > 0 && !s.hasRequiredSkills(agent, criteria.RequiredSkills) {
			continue
		}

		// Check response time requirement
		if criteria.MaxResponseTime > 0 && agent.AvgResponseTime > criteria.MaxResponseTime {
			continue
		}

		filtered = append(filtered, agent)
	}

	return filtered
}

// hasRequiredSkills checks if an agent has all required skills
func (s *OnlineAgentService) hasRequiredSkills(agent *OnlineAgent, requiredSkills []AgentSkill) bool {
	agentSkillMap := make(map[AgentSkill]bool)
	for _, skill := range agent.Skills {
		agentSkillMap[skill] = true
	}

	for _, required := range requiredSkills {
		if !agentSkillMap[required] {
			return false
		}
	}

	return true
}

// scoredAgent represents an agent with a calculated score
type scoredAgent struct {
	agent  *OnlineAgent
	score  float64
	reason string
}

// scoreAgents calculates scores for agents and returns them sorted by score
func (s *OnlineAgentService) scoreAgents(agents []*OnlineAgent, criteria AgentSelectionCriteria) []scoredAgent {
	var scored []scoredAgent

	for _, agent := range agents {
		score, reason := s.calculateAgentScore(agent, criteria)
		scored = append(scored, scoredAgent{
			agent:  agent,
			score:  score,
			reason: reason,
		})
	}

	// Sort by score (highest first)
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	return scored
}

// calculateAgentScore calculates a score for an agent based on selection criteria
func (s *OnlineAgentService) calculateAgentScore(agent *OnlineAgent, criteria AgentSelectionCriteria) (float64, string) {
	score := 0.0
	var reasons []string

	// Preferred agent gets highest score
	if criteria.PreferredAgent != nil && agent.UserID == *criteria.PreferredAgent {
		return 100.0, "preferred agent"
	}

	// Workload score (lower workload = higher score)
	workloadScore := (1.0 - agent.Workload) * 30.0
	score += workloadScore
	reasons = append(reasons, fmt.Sprintf("workload %.1f", workloadScore))

	// Response time score (faster response = higher score)
	if agent.AvgResponseTime > 0 {
		responseTimeScore := (60.0 / (agent.AvgResponseTime + 1.0)) * 20.0 // Normalize to 20 points max
		score += responseTimeScore
		reasons = append(reasons, fmt.Sprintf("response time %.1f", responseTimeScore))
	}

	// Last assignment score (less recently assigned = higher score)
	timeSinceAssignment := time.Since(agent.LastAssignment)
	assignmentScore := float64(timeSinceAssignment.Minutes()) / 10.0 // 1 point per 10 minutes
	if assignmentScore > 20.0 {
		assignmentScore = 20.0 // Cap at 20 points
	}
	score += assignmentScore
	reasons = append(reasons, fmt.Sprintf("assignment recency %.1f", assignmentScore))

	// Skills match score
	skillsScore := s.calculateSkillsScore(agent, criteria) * 20.0
	score += skillsScore
	reasons = append(reasons, fmt.Sprintf("skills match %.1f", skillsScore))

	// Status bonus
	statusScore := 0.0
	switch agent.Status {
	case AgentStatusOnline:
		statusScore = 10.0
	case AgentStatusAway:
		statusScore = 5.0
	}
	score += statusScore
	if statusScore > 0 {
		reasons = append(reasons, fmt.Sprintf("status %.1f", statusScore))
	}

	return score, fmt.Sprintf("composite score from: %v", reasons)
}

// calculateSkillsScore calculates how well an agent's skills match the requirements
func (s *OnlineAgentService) calculateSkillsScore(agent *OnlineAgent, criteria AgentSelectionCriteria) float64 {
	if len(criteria.RequiredSkills) == 0 {
		return 1.0 // No specific skills required
	}

	// Check if agent has general skill (can handle any request)
	for _, skill := range agent.Skills {
		if skill == SkillGeneral {
			return 0.8 // Good match but not perfect
		}
	}

	// Map request type to preferred skills
	preferredSkills := s.getPreferredSkillsForRequest(criteria.RequestType)
	
	matchScore := 0.0
	for _, preferred := range preferredSkills {
		for _, agentSkill := range agent.Skills {
			if agentSkill == preferred {
				matchScore += 1.0
				break
			}
		}
	}

	if len(preferredSkills) > 0 {
		return matchScore / float64(len(preferredSkills))
	}

	return 0.5 // Neutral score if no specific match
}

// getPreferredSkillsForRequest returns preferred skills for a request type
func (s *OnlineAgentService) getPreferredSkillsForRequest(requestType AgentRequestType) []AgentSkill {
	switch requestType {
	case AgentRequestTypeTechnical:
		return []AgentSkill{SkillTechnical, SkillSupport}
	case AgentRequestTypeBilling:
		return []AgentSkill{SkillBilling, SkillSupport}
	case AgentRequestTypeComplaint:
		return []AgentSkill{SkillComplaint, SkillSupport}
	case AgentRequestTypeUrgent:
		return []AgentSkill{SkillSupport, SkillGeneral}
	case AgentRequestTypeSupport:
		return []AgentSkill{SkillSupport, SkillGeneral}
	default:
		return []AgentSkill{SkillGeneral, SkillSupport}
	}
}

// backgroundCleanup runs periodic cleanup of inactive agents
func (s *OnlineAgentService) backgroundCleanup() {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanupInactiveAgents()
	}
}

// cleanupInactiveAgents removes agents that have been inactive for too long
func (s *OnlineAgentService) cleanupInactiveAgents() {
	s.mu.Lock()
	defer s.mu.Unlock()

	inactiveThreshold := 15 * time.Minute
	var removed []uuid.UUID

	for agentID, agent := range s.agents {
		if time.Since(agent.LastActivity) > inactiveThreshold {
			delete(s.agents, agentID)
			removed = append(removed, agentID)
		}
	}

	if len(removed) > 0 {
		log.Info().
			Int("removed_count", len(removed)).
			Msg("Cleaned up inactive agents")
	}

	s.lastCleanup = time.Now()
}
