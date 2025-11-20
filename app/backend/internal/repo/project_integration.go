package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/bareuptime/tms/internal/models"
)

type ProjectIntegrationRepository struct {
	db *sqlx.DB
}

func NewProjectIntegrationRepository(db *sqlx.DB) *ProjectIntegrationRepository {
	return &ProjectIntegrationRepository{db: db}
}

// Create creates a new project integration
func (r *ProjectIntegrationRepository) Create(ctx context.Context, integration *models.ProjectIntegration) error {
	query := `
		INSERT INTO project_integrations (
			id, tenant_id, project_id, integration_type, meta, status, created_at, updated_at
		) VALUES (
			:id, :tenant_id, :project_id, :integration_type, :meta, :status, :created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, integration)
	return err
}

// GetByID retrieves a project integration by ID
func (r *ProjectIntegrationRepository) GetByID(ctx context.Context, tenantID, integrationID uuid.UUID) (*models.ProjectIntegration, error) {
	var integration models.ProjectIntegration
	query := `
		SELECT * FROM project_integrations
		WHERE tenant_id = $1 AND id = $2`

	err := r.db.GetContext(ctx, &integration, query, tenantID, integrationID)
	if err != nil {
		return nil, err
	}
	return &integration, nil
}

// GetByProjectAndType retrieves a project integration by project and type
func (r *ProjectIntegrationRepository) GetByProjectAndType(ctx context.Context, tenantID, projectID uuid.UUID, integrationType models.ProjectIntegrationType) (*models.ProjectIntegration, error) {
	var integration models.ProjectIntegration
	query := `
		SELECT * FROM project_integrations
		WHERE tenant_id = $1 AND project_id = $2 AND integration_type = $3`

	fmt.Println("Querying for integration:", query, tenantID, projectID, integrationType)

	fmt.Println("tenantID ---> ", tenantID, " projectID ---> ", projectID, " integrationType ---> ", integrationType)

	err := r.db.GetContext(ctx, &integration, query, tenantID, projectID, integrationType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &integration, nil
}

// ListByProject retrieves all integrations for a project
func (r *ProjectIntegrationRepository) ListByProject(ctx context.Context, tenantID, projectID uuid.UUID) ([]*models.ProjectIntegration, error) {
	var integrations []*models.ProjectIntegration
	query := `
		SELECT * FROM project_integrations
		WHERE tenant_id = $1 AND project_id = $2
		ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &integrations, query, tenantID, projectID)
	return integrations, err
}

// ListByTenant retrieves all integrations for a tenant
func (r *ProjectIntegrationRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*models.ProjectIntegration, error) {
	var integrations []*models.ProjectIntegration
	query := `
		SELECT * FROM project_integrations
		WHERE tenant_id = $1
		ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &integrations, query, tenantID)
	return integrations, err
}

// Update updates a project integration
func (r *ProjectIntegrationRepository) Update(ctx context.Context, integration *models.ProjectIntegration) error {
	query := `
		UPDATE project_integrations SET
			meta = :meta,
			status = :status,
			updated_at = NOW()
		WHERE tenant_id = :tenant_id AND id = :id`

	_, err := r.db.NamedExecContext(ctx, query, integration)
	return err
}

// Upsert creates or updates a project integration based on tenant, project, and type
func (r *ProjectIntegrationRepository) Upsert(ctx context.Context, integration *models.ProjectIntegration) error {
	query := `
		INSERT INTO project_integrations (
			id, tenant_id, project_id, integration_type, meta, status, created_at, updated_at
		) VALUES (
			:id, :tenant_id, :project_id, :integration_type, :meta, :status, :created_at, :updated_at
		)
		ON CONFLICT (tenant_id, project_id, integration_type)
		DO UPDATE SET
			meta = EXCLUDED.meta,
			status = EXCLUDED.status,
			updated_at = NOW()
		RETURNING id`

	rows, err := r.db.NamedQueryContext(ctx, query, integration)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&integration.ID); err != nil {
			return err
		}
	}

	return nil
}

// Delete deletes a project integration
func (r *ProjectIntegrationRepository) Delete(ctx context.Context, tenantID, integrationID uuid.UUID) error {
	query := `DELETE FROM project_integrations WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.ExecContext(ctx, query, tenantID, integrationID)
	return err
}

// DeleteByProjectAndType deletes a project integration by project and type
func (r *ProjectIntegrationRepository) DeleteByProjectAndType(ctx context.Context, tenantID, projectID uuid.UUID, integrationType models.ProjectIntegrationType) error {
	query := `DELETE FROM project_integrations WHERE tenant_id = $1 AND project_id = $2 AND integration_type = $3`
	_, err := r.db.ExecContext(ctx, query, tenantID, projectID, integrationType)
	return err
}

// UpdateStatus updates the status of a project integration
func (r *ProjectIntegrationRepository) UpdateStatus(ctx context.Context, tenantID, integrationID uuid.UUID, status models.ProjectIntegrationStatus) error {
	query := `
		UPDATE project_integrations
		SET status = $1, updated_at = NOW()
		WHERE tenant_id = $2 AND id = $3`

	_, err := r.db.ExecContext(ctx, query, status, tenantID, integrationID)
	return err
}
