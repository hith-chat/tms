package service

import (
	"context"
	"testing"
	"time"

	"github.com/bareuptime/tms/internal/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type mockProjectRepo struct {
	countResult int
	countErr    error

	createdProject *db.Project
	createErr      error

	getByIDResult *db.Project
	getByIDErr    error

	updatedProject *db.Project
	updateErr      error
}

func (m *mockProjectRepo) Create(ctx context.Context, project *db.Project) error {
	m.createdProject = project
	return m.createErr
}

func (m *mockProjectRepo) GetByID(ctx context.Context, tenantID, projectID uuid.UUID) (*db.Project, error) {
	return m.getByIDResult, m.getByIDErr
}

func (m *mockProjectRepo) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (*db.Project, error) {
	return nil, nil
}

func (m *mockProjectRepo) Update(ctx context.Context, project *db.Project) error {
	m.updatedProject = project
	return m.updateErr
}

func (m *mockProjectRepo) Delete(ctx context.Context, tenantID, projectID uuid.UUID) error {
	return nil
}

func (m *mockProjectRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*db.Project, error) {
	return nil, nil
}

func (m *mockProjectRepo) ListForAgent(ctx context.Context, tenantID, agentID uuid.UUID) ([]*db.Project, error) {
	return nil, nil
}

func (m *mockProjectRepo) Count(ctx context.Context, tenantID uuid.UUID) (int, error) {
	return m.countResult, m.countErr
}

func (m *mockProjectRepo) GetActivePublicProjectByDomain(ctx context.Context, tenantID uuid.UUID, domain string) (*db.Project, error) {
	return nil, nil
}

func TestProjectService_CreateProject_LimitReached(t *testing.T) {
	t.Parallel()

	repo := &mockProjectRepo{countResult: 5}
	svc := NewProjectService(repo)

	project, err := svc.CreateProject(context.Background(), uuid.New(), "KEY", "Name")

	require.ErrorIs(t, err, ErrProjectLimitReached)
	require.Nil(t, project)
	require.Nil(t, repo.createdProject)
}

func TestProjectService_CreateProject_Success(t *testing.T) {
	t.Parallel()

	repo := &mockProjectRepo{countResult: 2}
	svc := NewProjectService(repo)

	tenantID := uuid.New()
	project, err := svc.CreateProject(context.Background(), tenantID, "SUP", "Support")

	require.NoError(t, err)
	require.NotNil(t, project)
	require.NotEqual(t, uuid.Nil, project.ID)
	require.Equal(t, tenantID, project.TenantID)
	require.Equal(t, "SUP", project.Key)
	require.Equal(t, "Support", project.Name)
	require.Equal(t, "active", project.Status)
	require.WithinDuration(t, time.Now(), project.CreatedAt, time.Second)
	require.WithinDuration(t, time.Now(), project.UpdatedAt, time.Second)

	require.Equal(t, repo.createdProject, project)
}

func TestProjectService_UpdateProject_NotFound(t *testing.T) {
	t.Parallel()

	repo := &mockProjectRepo{}
	svc := NewProjectService(repo)

	updated, err := svc.UpdateProject(context.Background(), uuid.New(), uuid.New(), "K", "Name", "archived")

	require.NoError(t, err)
	require.Nil(t, updated)
}

func TestProjectService_UpdateProject_Success(t *testing.T) {
	t.Parallel()

	original := &db.Project{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Key:      "OLD",
		Name:     "Old name",
		Status:   "active",
	}

	repo := &mockProjectRepo{getByIDResult: original}
	svc := NewProjectService(repo)

	newKey := "NEW"
	newName := "New name"
	newStatus := "inactive"

	updated, err := svc.UpdateProject(context.Background(), original.TenantID, original.ID, newKey, newName, newStatus)

	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Equal(t, newKey, updated.Key)
	require.Equal(t, newName, updated.Name)
	require.Equal(t, newStatus, updated.Status)
	require.WithinDuration(t, time.Now(), updated.UpdatedAt, time.Second)
	require.Equal(t, repo.updatedProject, updated)
}
