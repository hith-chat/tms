package service

import (
	"context"
	"testing"
	"time"

	"github.com/bareuptime/tms/internal/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type stubProjectRepository struct {
	countResult int
	countErr    error

	activeProject    *db.Project
	activeProjectErr error
	createdProject   *db.Project
	createErr        error
}

func (s *stubProjectRepository) Create(ctx context.Context, project *db.Project) error {
	s.createdProject = project
	return s.createErr
}

func (s *stubProjectRepository) GetByID(ctx context.Context, tenantID, projectID uuid.UUID) (*db.Project, error) {
	return nil, nil
}

func (s *stubProjectRepository) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (*db.Project, error) {
	return nil, nil
}

func (s *stubProjectRepository) Update(ctx context.Context, project *db.Project) error {
	return nil
}

func (s *stubProjectRepository) Delete(ctx context.Context, tenantID, projectID uuid.UUID) error {
	return nil
}

func (s *stubProjectRepository) List(ctx context.Context, tenantID uuid.UUID) ([]*db.Project, error) {
	return nil, nil
}

func (s *stubProjectRepository) ListForAgent(ctx context.Context, tenantID, agentID uuid.UUID) ([]*db.Project, error) {
	return nil, nil
}

func (s *stubProjectRepository) Count(ctx context.Context, tenantID uuid.UUID) (int, error) {
	return s.countResult, s.countErr
}

func (s *stubProjectRepository) GetActivePublicProjectByDomain(ctx context.Context, tenantID uuid.UUID, domain string) (*db.Project, error) {
	if s.activeProject != nil {
		return s.activeProject, nil
	}
	return nil, s.activeProjectErr
}

func TestPublicAIBuilderService_BuildPublicWidget_InvalidURL(t *testing.T) {
	t.Parallel()

	repo := &stubProjectRepository{}
	svc := NewPublicAIBuilderService(repo, nil, nil, nil)

	events := make(chan AIBuilderEvent, 2)
	buildID, err := svc.BuildPublicWidget(context.Background(), ":://broken", 0, events)

	require.Error(t, err)
	require.Equal(t, uuid.Nil, buildID)

	select {
	case evt := <-events:
		require.Equal(t, "error", evt.Type)
		require.Equal(t, "initialization", evt.Stage)
	default:
		t.Fatal("expected error event to be emitted")
	}
}

func TestPublicAIBuilderService_BuildPublicWidget_HTTPNotSupported(t *testing.T) {
	t.Parallel()

	repo := &stubProjectRepository{}
	svc := NewPublicAIBuilderService(repo, nil, nil, nil)

	events := make(chan AIBuilderEvent, 2)
	buildID, err := svc.BuildPublicWidget(context.Background(), "http://example.com", 0, events)

	require.Error(t, err)
	require.Equal(t, uuid.Nil, buildID)
	require.Contains(t, err.Error(), "only HTTPS URLs are supported")

	select {
	case evt := <-events:
		require.Equal(t, "error", evt.Type)
		require.Equal(t, "initialization", evt.Stage)
	default:
		t.Fatal("expected error event to be emitted")
	}
}

func TestPublicAIBuilderService_CreatePublicProject_SetsExpectedFields(t *testing.T) {
	t.Parallel()

	repo := &stubProjectRepository{}
	svc := NewPublicAIBuilderService(repo, nil, nil, nil)

	events := make(chan AIBuilderEvent, 1)
	ctx := context.Background()

	project, err := svc.createPublicProject(ctx, "example.com", "https://example.com", events)

	require.NoError(t, err)
	require.NotNil(t, project)
	require.Equal(t, PublicTenantID, project.TenantID)
	require.True(t, project.IsPublic)
	require.Contains(t, project.Key, "EXAMPLE-COM-PUBLIC")
	require.Equal(t, "example.com Public Widget", project.Name)
	require.WithinDuration(t, time.Now().Add(6*time.Hour), *project.ExpiresAt, time.Minute)
	require.Equal(t, repo.createdProject, project)
}

func TestPublicAIBuilderService_GenerateProjectKey(t *testing.T) {
	t.Parallel()

	repo := &stubProjectRepository{}
	svc := NewPublicAIBuilderService(repo, nil, nil, nil)

	require.Equal(t, "EXAMPLE-COM-PUBLIC", svc.generateProjectKey("example.com"))
	require.Equal(t, "DOCS-ANTHROPIC-COM-PUBLIC", svc.generateProjectKey("docs.anthropic.com"))
	require.Equal(t, "LOCALHOST-PUBLIC", svc.generateProjectKey("localhost:8080"))
}
