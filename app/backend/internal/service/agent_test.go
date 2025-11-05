package service

import (
	"context"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/bareuptime/tms/internal/db"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/rbac"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type mockAgentRepository struct {
	repo.AgentRepository
	createErr        error
	getByEmailErr    error
	getByEmailResult *db.Agent
	createdAgents    []*db.Agent
	getByEmailCalls  int
}

func (m *mockAgentRepository) Create(ctx context.Context, agent *db.Agent) error {
	m.createdAgents = append(m.createdAgents, agent)
	return m.createErr
}

func (m *mockAgentRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*db.Agent, error) {
	m.getByEmailCalls++
	return m.getByEmailResult, m.getByEmailErr
}

func (m *mockAgentRepository) GetByID(ctx context.Context, tenantID, agentID uuid.UUID) (*db.Agent, error) {
	panic("unexpected call to GetByID")
}

func (m *mockAgentRepository) GetByEmailWithoutTenantID(ctx context.Context, email string) (*db.Agent, error) {
	panic("unexpected call to GetByEmailWithoutTenantID")
}

func (m *mockAgentRepository) Update(ctx context.Context, agent *db.Agent) error {
	panic("unexpected call to Update")
}

func (m *mockAgentRepository) Delete(ctx context.Context, tenantID, agentID uuid.UUID) error {
	panic("unexpected call to Delete")
}

func (m *mockAgentRepository) List(ctx context.Context, tenantID uuid.UUID, filters repo.AgentFilters, pagination repo.PaginationParams) ([]*db.Agent, string, error) {
	panic("unexpected call to List")
}

func (m *mockAgentRepository) GetTenantAdmins(ctx context.Context, tenantID uuid.UUID) ([]*db.Agent, error) {
	panic("unexpected call to GetTenantAdmins")
}

type noopProjectRepository struct{}

func (n *noopProjectRepository) Create(ctx context.Context, project *db.Project) error {
	panic("unexpected call to ProjectRepository.Create")
}
func (n *noopProjectRepository) GetByID(ctx context.Context, tenantID, projectID uuid.UUID) (*db.Project, error) {
	panic("unexpected call to ProjectRepository.GetByID")
}
func (n *noopProjectRepository) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (*db.Project, error) {
	panic("unexpected call to ProjectRepository.GetByKey")
}
func (n *noopProjectRepository) Update(ctx context.Context, project *db.Project) error {
	panic("unexpected call to ProjectRepository.Update")
}
func (n *noopProjectRepository) Delete(ctx context.Context, tenantID, projectID uuid.UUID) error {
	panic("unexpected call to ProjectRepository.Delete")
}
func (n *noopProjectRepository) List(ctx context.Context, tenantID uuid.UUID) ([]*db.Project, error) {
	panic("unexpected call to ProjectRepository.List")
}
func (n *noopProjectRepository) ListForAgent(ctx context.Context, tenantID, agentID uuid.UUID) ([]*db.Project, error) {
	panic("unexpected call to ProjectRepository.ListForAgent")
}
func (n *noopProjectRepository) Count(ctx context.Context, tenantID uuid.UUID) (int, error) {
	panic("unexpected call to ProjectRepository.Count")
}
func (n *noopProjectRepository) GetActivePublicProjectByDomain(ctx context.Context, tenantID uuid.UUID, domain string) (*db.Project, error) {
	panic("unexpected call to ProjectRepository.GetActivePublicProjectByDomain")
}

func TestAgentServiceCreateAgentSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := uuid.New()

	mockRepo := &mockAgentRepository{
		getByEmailErr: errors.New("not found"),
	}

	dbConn, sqlMock, err := sqlmock.New()
	require.NoError(t, err)
	defer dbConn.Close()

	sqlMock.ExpectExec("INSERT INTO agent_project_roles").
		WithArgs(sqlmock.AnyArg(), tenantID, uuid.Nil, models.RoleAgent).
		WillReturnResult(sqlmock.NewResult(1, 1))

	rbacService := rbac.NewService(dbConn)

	svc := NewAgentService(mockRepo, nil, rbacService)

	agent, err := svc.CreateAgent(ctx, tenantID, uuid.New(), CreateAgentRequest{
		Email:    "agent@example.com",
		Name:     "Agent Smith",
		Password: "SuperSecret123",
		Role:     models.RoleAgent,
	})
	require.NoError(t, err)

	require.Len(t, mockRepo.createdAgents, 1)
	created := mockRepo.createdAgents[0]
	require.Equal(t, agent.ID, created.ID)
	require.Equal(t, "agent@example.com", created.Email)
	require.NotNil(t, created.PasswordHash)
	require.NotEqual(t, "SuperSecret123", *created.PasswordHash)
	require.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestAgentServiceCreateAgentDuplicate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := uuid.New()

	mockRepo := &mockAgentRepository{
		getByEmailResult: &db.Agent{},
	}

	dbConn, _, err := sqlmock.New()
	require.NoError(t, err)
	defer dbConn.Close()

	svc := NewAgentService(mockRepo, nil, rbac.NewService(dbConn))

	_, err = svc.CreateAgent(ctx, tenantID, uuid.New(), CreateAgentRequest{
		Email:    "agent@example.com",
		Name:     "Agent Smith",
		Password: "password123",
		Role:     models.RoleAgent,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "already exists")
	require.Zero(t, len(mockRepo.createdAgents))
}
