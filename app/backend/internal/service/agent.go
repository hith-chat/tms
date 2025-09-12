package service

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/bareuptime/tms/internal/db"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/rbac"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AgentService handles agent management operations
type AgentService struct {
	agentRepo   repo.AgentRepository
	projectRepo repo.ProjectRepository
	rbacService *rbac.Service
}

// NewAgentService creates a new agent service
func NewAgentService(agentRepo repo.AgentRepository, projectRepo repo.ProjectRepository, rbacService *rbac.Service) *AgentService {
	return &AgentService{
		agentRepo:   agentRepo,
		projectRepo: projectRepo,
		rbacService: rbacService,
	}
}

// CreateAgentRequest represents an agent creation request
type CreateAgentRequest struct {
	Email    string          `json:"email" validate:"required,email"`
	Name     string          `json:"name" validate:"required,min=1,max=255"`
	Password string          `json:"password" validate:"required,min=8"`
	Role     models.RoleType `json:"role" validate:"required"`
}

// CreateAgent creates a new agent
func (s *AgentService) CreateAgent(ctx context.Context, tenantID, creatorAgentID uuid.UUID, req CreateAgentRequest) (*db.Agent, error) {

	// Check if agent already exists
	existing, err := s.agentRepo.GetByEmail(ctx, tenantID, req.Email)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("agent with email %s already exists", req.Email)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create agent
	hashedPasswordStr := string(hashedPassword)
	agent := &db.Agent{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: &hashedPasswordStr,
		Status:       "active",
	}

	err = s.agentRepo.Create(ctx, agent)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// Assign default role - this would be done through RBAC service
	err = s.rbacService.AssignRole(ctx, agent.ID, tenantID, uuid.Nil, req.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to assign role: %w", err)
	}

	return agent, nil
}

// UpdateAgentRequest represents an agent update request
type UpdateAgentRequest struct {
	Name     *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	IsActive *bool   `json:"is_active,omitempty"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=8"`
}

// UpdateAgent updates an existing agent
func (s *AgentService) UpdateAgent(ctx context.Context, tenantID, agentID, updaterAgentID uuid.UUID, req UpdateAgentRequest) (*db.Agent, int, error) {

	// Get existing agent
	agent, err := s.agentRepo.GetByID(ctx, tenantID, agentID)
	if err != nil {
		return nil, http.StatusNotFound, fmt.Errorf("agent not found: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		agent.Name = *req.Name
	}
	if req.IsActive != nil {
		// Only admins can deactivate agents
		if !*req.IsActive && agentID != updaterAgentID {
			return nil, http.StatusForbidden, fmt.Errorf("insufficient permissions to deactivate agent")

		}
		if *req.IsActive {
			agent.Status = "active"
		} else {
			agent.Status = "inactive"
		}
	}
	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to hash password: %w", err)
		}
		hashedPasswordStr := string(hashedPassword)
		agent.PasswordHash = &hashedPasswordStr
	}

	err = s.agentRepo.Update(ctx, agent)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to update agent: %w", err)
	}

	return agent, http.StatusAccepted, nil
}

// GetAgent retrieves an agent by ID
func (s *AgentService) GetAgent(ctx context.Context, tenantID, requestorAgentID uuid.UUID) (*db.Agent, error) {

	agent, err := s.agentRepo.GetByID(ctx, tenantID, requestorAgentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	return agent, nil
}

// ListAgentsRequest represents an agent list request
type ListAgentsRequest struct {
	Email    string `json:"email,omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
	Search   string `json:"search,omitempty"`
	Cursor   string `json:"cursor,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

// ListAgents retrieves a list of agents
func (s *AgentService) ListAgents(ctx context.Context, tenantID, requestorAgentID uuid.UUID, req ListAgentsRequest) ([]*db.Agent, string, error) {

	filters := repo.AgentFilters{
		Email:    req.Email,
		IsActive: req.IsActive,
		Search:   req.Search,
		AgentID:  requestorAgentID,
	}

	pagination := repo.PaginationParams{
		Cursor: req.Cursor,
		Limit:  req.Limit,
	}

	agents, nextCursor, err := s.agentRepo.List(ctx, tenantID, filters, pagination)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list agents: %w", err)
	}

	return agents, nextCursor, nil
}

// AssignRoleRequest represents a role assignment request
type AssignRoleRequest struct {
	Role      models.RoleType `json:"role" validate:"required"`
	ProjectID *string         `json:"project_id,omitempty"`
}

// AssignRole assigns a role to an agent
func (s *AgentService) AssignRole(ctx context.Context, tenantID, agentID, assignerAgentID uuid.UUID, req AssignRoleRequest) error {

	_, err := s.agentRepo.GetByID(ctx, tenantID, agentID)
	if err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	projectID := uuid.Nil
	if req.ProjectID != nil {
		projectID, _ = uuid.Parse(*req.ProjectID)
	}

	err = s.rbacService.AssignRole(ctx, agentID, tenantID, projectID, req.Role)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// RemoveRoleRequest represents a role removal request
type RemoveRoleRequest struct {
	Role      models.RoleType `json:"role" validate:"required"`
	ProjectID *string         `json:"project_id,omitempty"`
}

// RemoveRole removes a role from an agent
func (s *AgentService) RemoveRole(ctx context.Context, tenantID, agentID uuid.UUID, req RemoveRoleRequest) error {
	projectID := uuid.Nil
	if req.ProjectID != nil {
		projectID, _ = uuid.Parse(*req.ProjectID)
	}

	err := s.rbacService.RemoveRole(ctx, agentID, tenantID, projectID, req.Role)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	return nil
}

// GetAgentRoles retrieves roles for an agent
func (s *AgentService) GetAgentRoles(ctx context.Context, tenantID, agentID uuid.UUID) ([]*db.RoleBinding, error) {
	// Agents can view their own roles, admins can view any agent's roles
	roleBindings, err := s.rbacService.GetAgentRoleBindings(ctx, agentID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role bindings: %w", err)
	}

	return roleBindings, nil
}

// DeleteAgent deletes an agent
func (s *AgentService) DeleteAgent(ctx context.Context, tenantID, agentID, deleterAgentID string) error {
	// Check permissions - only admins can delete agents
	// Prevent self-deletion
	if agentID == deleterAgentID {
		return fmt.Errorf("cannot delete yourself")
	}

	// Parse UUIDs
	tenantUUID, _ := uuid.Parse(tenantID)
	agentUUID, _ := uuid.Parse(agentID)
	// Delete the agent
	err := s.agentRepo.Delete(ctx, tenantUUID, agentUUID)
	if err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}

	return nil
}

// AssignToProjectRequest represents a project assignment request
type AssignToProjectRequest struct {
	ProjectID uuid.UUID       `json:"project_id" validate:"required"`
	Role      models.RoleType `json:"role" validate:"required"`
}

// AssignToProject assigns an agent to a project with a specific role
func (s *AgentService) AssignToProject(ctx context.Context, tenantID, agentID uuid.UUID, req AssignToProjectRequest) error {

	// Assign the role for the specific project
	err := s.rbacService.AssignRole(ctx, agentID, tenantID, req.ProjectID, req.Role)
	if err != nil {
		return fmt.Errorf("failed to assign agent to project: %w", err)
	}

	return nil
}

// RemoveFromProject removes an agent from a project
func (s *AgentService) RemoveFromProject(ctx context.Context, tenantID, agentID, projectID uuid.UUID) error {

	// Get the agent's current role in this project first
	roleBindings, err := s.rbacService.GetAgentRoleBindings(ctx, agentID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get role bindings: %w", err)
	}

	// Find the role for this specific project
	var roleToRemove models.RoleType
	for _, binding := range roleBindings {
		if binding.ProjectID != nil && *binding.ProjectID == projectID {
			roleToRemove = binding.Role
			break
		}
	}

	if roleToRemove == "" {
		return fmt.Errorf("agent is not assigned to this project")
	}

	// Remove the role for the specific project
	err = s.rbacService.RemoveRole(ctx, agentID, tenantID, projectID, roleToRemove)
	if err != nil {
		return fmt.Errorf("failed to remove agent from project: %w", err)
	}

	return nil
}

// AgentProject represents a project that an agent is assigned to
type AgentProject struct {
	ID   string          `json:"id"`
	Name string          `json:"name"`
	Role models.RoleType `json:"role"`
}

// GetAgentProjects retrieves all projects an agent is assigned to
func (s *AgentService) GetAgentProjects(ctx context.Context, tenantID, agentID uuid.UUID) ([]*AgentProject, error) {
	// Get the agent's role bindings
	roleBindings, err := s.rbacService.GetAgentRoleBindings(ctx, agentID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role bindings: %w", err)
	}

	// Get project details for each role binding
	var projects []*AgentProject
	for _, binding := range roleBindings {
		if binding.ProjectID != nil {
			project, err := s.projectRepo.GetByID(ctx, tenantID, *binding.ProjectID)
			if err != nil {
				// Log error but continue - project might have been deleted
				log.Printf("Failed to get project %s: %v", binding.ProjectID.String(), err)
				continue
			}

			projects = append(projects, &AgentProject{
				ID:   project.ID.String(),
				Name: project.Name,
				Role: binding.Role,
			})
		}
	}

	return projects, nil
}

func (s *AgentService) GetAgentProjectsList(ctx context.Context, tenantID, agentID uuid.UUID) ([]uuid.UUID, error) {
	// Get the agent's role bindings
	roleBindings, err := s.rbacService.GetAgentRoleBindings(ctx, agentID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role bindings: %w", err)
	}

	// Get project details for each role binding
	var projectIDs []uuid.UUID
	for _, binding := range roleBindings {
		if binding.ProjectID != nil {
			projectIDs = append(projectIDs, *binding.ProjectID)
		}
	}

	return projectIDs, nil
}

// GetTenantAdmins retrieves all tenant admin agents for a given tenant
func (s *AgentService) GetTenantAdmins(ctx context.Context, tenantID uuid.UUID) ([]*db.Agent, error) {
	return s.agentRepo.GetTenantAdmins(ctx, tenantID)
}
