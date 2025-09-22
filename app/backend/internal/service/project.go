package service

import (
	"context"
	"time"

	"fmt"

	"github.com/bareuptime/tms/internal/db"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/google/uuid"
)

type ProjectService struct {
	projectRepo repo.ProjectRepository
}

func NewProjectService(projectRepo repo.ProjectRepository) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
	}
}

// ErrProjectLimitReached is returned when tenant has reached maximum allowed projects
var ErrProjectLimitReached = fmt.Errorf("tenant project limit reached")

func (s *ProjectService) GetProject(ctx context.Context, tenantID, projectID uuid.UUID) (*db.Project, error) {
	return s.projectRepo.GetByID(ctx, tenantID, projectID)
}

func (s *ProjectService) GetProjectByKey(ctx context.Context, tenantID uuid.UUID, key string) (*db.Project, error) {
	return s.projectRepo.GetByKey(ctx, tenantID, key)
}

func (s *ProjectService) ListProjects(ctx context.Context, tenantID uuid.UUID) ([]*db.Project, error) {
	return s.projectRepo.List(ctx, tenantID)
}

func (s *ProjectService) ListProjectsForAgent(ctx context.Context, tenantID, agentID uuid.UUID) ([]*db.Project, error) {
	return s.projectRepo.ListForAgent(ctx, tenantID, agentID)
}

func (s *ProjectService) CreateProject(ctx context.Context, tenantID uuid.UUID, key, name string) (*db.Project, error) {
	// Enforce a maximum of 5 projects per tenant
	count, err := s.projectRepo.Count(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if count >= 5 {
		return nil, ErrProjectLimitReached
	}
	now := time.Now()
	project := &db.Project{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Key:       key,
		Name:      name,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = s.projectRepo.Create(ctx, project)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, tenantID, projectID uuid.UUID, key, name, status string) (*db.Project, error) {
	project, err := s.projectRepo.GetByID(ctx, tenantID, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, nil
	}

	project.Key = key
	project.Name = name
	project.Status = status
	project.UpdatedAt = time.Now()

	err = s.projectRepo.Update(ctx, project)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, tenantID, projectID uuid.UUID) error {
	return s.projectRepo.Delete(ctx, tenantID, projectID)
}
