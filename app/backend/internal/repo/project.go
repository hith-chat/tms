package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bareuptime/tms/internal/db"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type projectRepository struct {
	db *sqlx.DB
}

func NewProjectRepository(database *sqlx.DB) ProjectRepository {
	return &projectRepository{
		db: database,
	}
}

func (r *projectRepository) Create(ctx context.Context, project *db.Project) error {
	query := `
		INSERT INTO projects (id, tenant_id, key, name, status, is_public, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		project.ID,
		project.TenantID,
		project.Key,
		project.Name,
		project.Status,
		project.IsPublic,
		project.ExpiresAt,
		project.CreatedAt,
		project.UpdatedAt,
	)

	return err
}

func (r *projectRepository) GetByID(ctx context.Context, tenantID, projectID uuid.UUID) (*db.Project, error) {
	query := `
		SELECT id, tenant_id, key, name, status, is_public, expires_at, created_at, updated_at
		FROM projects
		WHERE id = $1 AND tenant_id = $2
			AND (expires_at IS NULL OR expires_at > NOW())
	`

	var project db.Project
	err := r.db.GetContext(ctx, &project, query, projectID, tenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &project, nil
}

func (r *projectRepository) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (*db.Project, error) {
	query := `
		SELECT id, tenant_id, key, name, status, is_public, expires_at, created_at, updated_at
		FROM projects
		WHERE key = $1 AND tenant_id = $2
			AND (expires_at IS NULL OR expires_at > NOW())
	`

	var project db.Project
	err := r.db.GetContext(ctx, &project, query, key, tenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &project, nil
}

func (r *projectRepository) Update(ctx context.Context, project *db.Project) error {
	query := `
		UPDATE projects
		SET key = $3, name = $4, status = $5, updated_at = $6
		WHERE id = $1 AND tenant_id = $2
	`

	result, err := r.db.ExecContext(ctx, query,
		project.ID,
		project.TenantID,
		project.Key,
		project.Name,
		project.Status,
		project.UpdatedAt,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no project found with id %s", project.ID)
	}

	return nil
}

func (r *projectRepository) Delete(ctx context.Context, tenantID, projectID uuid.UUID) error {
	query := `DELETE FROM projects WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, projectID, tenantID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no project found with id %s", projectID)
	}

	return nil
}

func (r *projectRepository) List(ctx context.Context, tenantID uuid.UUID) ([]*db.Project, error) {
	query := `
		SELECT id, tenant_id, key, name, status, is_public, expires_at, created_at, updated_at
		FROM projects
		WHERE tenant_id = $1
			AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY name ASC
	`

	var projects []*db.Project
	err := r.db.SelectContext(ctx, &projects, query, tenantID)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (r *projectRepository) ListForAgent(ctx context.Context, tenantID, agentID uuid.UUID) ([]*db.Project, error) {
	query := `
		SELECT DISTINCT p.id, p.tenant_id, p.key, p.name, p.status, p.is_public, p.expires_at, p.created_at, p.updated_at
		FROM projects p
		INNER JOIN agent_project_roles apr ON p.id = apr.project_id
		WHERE p.tenant_id = $1 AND apr.agent_id = $2 AND p.status = 'active'
			AND (p.expires_at IS NULL OR p.expires_at > NOW())
		ORDER BY p.name ASC
	`

	var projects []*db.Project
	err := r.db.SelectContext(ctx, &projects, query, tenantID, agentID)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (r *projectRepository) ListForAgentAdmin(ctx context.Context, tenantID uuid.UUID) ([]*db.Project, error) {
	query := `
		SELECT DISTINCT p.id, p.tenant_id, p.key, p.name, p.status, p.is_public, p.expires_at, p.created_at, p.updated_at
		FROM projects p
		WHERE p.tenant_id = $1
			AND (p.expires_at IS NULL OR p.expires_at > NOW())
	`

	var projects []*db.Project
	err := r.db.SelectContext(ctx, &projects, query, tenantID)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

// Count returns the number of projects for a tenant
func (r *projectRepository) Count(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `SELECT COUNT(1) FROM projects WHERE tenant_id = $1 AND (expires_at IS NULL OR expires_at > NOW())`
	var count int
	err := r.db.GetContext(ctx, &count, query, tenantID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetActivePublicProjectByDomain checks if an active public project exists for the given domain
func (r *projectRepository) GetActivePublicProjectByDomain(ctx context.Context, tenantID uuid.UUID, domain string) (*db.Project, error) {
	query := `
		SELECT id, tenant_id, key, name, status, is_public, expires_at, created_at, updated_at
		FROM projects
		WHERE tenant_id = $1
			AND is_public = true
			AND key LIKE $2
			AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY created_at DESC
		LIMIT 1
	`

	// Convert domain to key pattern (e.g., "example.com" -> "EXAMPLE-COM-%")
	domainKey := "%" + domain + "%"

	var project db.Project
	err := r.db.GetContext(ctx, &project, query, tenantID, domainKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &project, nil
}
