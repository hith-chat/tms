package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bareuptime/tms/internal/db"
	"github.com/google/uuid"
)

type agentRepository struct {
	db *sql.DB
}

// NewAgentRepository creates a new agent repository
func NewAgentRepository(database *sql.DB) AgentRepository {
	return &agentRepository{db: database}
}

// Create creates a new agent
func (r *agentRepository) Create(ctx context.Context, agent *db.Agent) error {
	query := `
		INSERT INTO agents (id, tenant_id, email, name, status, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`

	_, err := r.db.ExecContext(ctx, query,
		agent.ID, agent.TenantID, agent.Email, agent.Name,
		agent.Status, agent.PasswordHash)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	return nil
}

// GetByID retrieves an agent by ID
func (r *agentRepository) GetByID(ctx context.Context, tenantID, agentID uuid.UUID) (*db.Agent, error) {
	query := `
		SELECT id, tenant_id, email, name, status, password_hash, created_at, updated_at
		FROM agents
		WHERE tenant_id = $1 AND id = $2
	`

	var agent db.Agent
	err := r.db.QueryRowContext(ctx, query, tenantID, agentID).Scan(
		&agent.ID, &agent.TenantID, &agent.Email, &agent.Name,
		&agent.Status, &agent.PasswordHash, &agent.CreatedAt, &agent.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("agent not found")
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	return &agent, nil
}

// GetByEmail retrieves an agent by email
func (r *agentRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*db.Agent, error) {
	query := `
		SELECT id, tenant_id, email, name, status, password_hash, created_at, updated_at
		FROM agents
		WHERE tenant_id = $1 AND email = $2
	`

	var agent db.Agent
	err := r.db.QueryRowContext(ctx, query, tenantID, email).Scan(
		&agent.ID, &agent.TenantID, &agent.Email, &agent.Name,
		&agent.Status, &agent.PasswordHash, &agent.CreatedAt, &agent.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("agent not found")
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	return &agent, nil
}

// GetByEmail retrieves an agent by email
func (r *agentRepository) GetByEmailWithoutTenantID(ctx context.Context, email string) (*db.Agent, error) {
	query := `
		SELECT id, tenant_id, email, name, status, password_hash, created_at, updated_at
		FROM agents
		WHERE email = $1
	`

	var agent db.Agent
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&agent.ID, &agent.TenantID, &agent.Email, &agent.Name,
		&agent.Status, &agent.PasswordHash, &agent.CreatedAt, &agent.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("agent not found")
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	return &agent, nil
}

// Update updates an existing agent
func (r *agentRepository) Update(ctx context.Context, agent *db.Agent) error {
	query := `
		UPDATE agents
		SET email = $3, name = $4, status = $5, password_hash = $6, updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2
	`

	result, err := r.db.ExecContext(ctx, query,
		agent.TenantID, agent.ID, agent.Email, agent.Name,
		agent.Status, agent.PasswordHash)
	if err != nil {
		return fmt.Errorf("failed to update agent: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("agent not found")
	}

	return nil
}

// Delete deletes an agent
func (r *agentRepository) Delete(ctx context.Context, tenantID, agentID uuid.UUID) error {
	query := `DELETE FROM agents WHERE tenant_id = $1 AND id = $2`

	result, err := r.db.ExecContext(ctx, query, tenantID, agentID)
	if err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("agent not found")
	}

	return nil
}

// List retrieves a list of agents with filtering and pagination
func (r *agentRepository) List(ctx context.Context, tenantID uuid.UUID, filters AgentFilters, pagination PaginationParams) ([]*db.Agent, string, error) {
	query := `
		SELECT id, tenant_id, email, name, status, password_hash, created_at, updated_at
		FROM agents
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	argCount := 1

	// Apply filters
	if filters.Email != "" {
		argCount++
		query += fmt.Sprintf(" AND email = $%d", argCount)
		args = append(args, filters.Email)
	}

	if filters.IsActive != nil {
		argCount++
		status := "active"
		if !*filters.IsActive {
			status = "inactive"
		}
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}

	if filters.AgentID != uuid.Nil {
		argCount++
		query += fmt.Sprintf(" AND id = $%d", argCount)
		args = append(args, filters.AgentID)
	}

	if filters.Search != "" {
		argCount++
		query += fmt.Sprintf(" AND (name ILIKE $%d OR email ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+filters.Search+"%")
	}

	// Apply pagination
	if pagination.Cursor != "" {
		argCount++
		query += fmt.Sprintf(" AND id > $%d", argCount)
		cursorID, err := uuid.Parse(pagination.Cursor)
		if err != nil {
			return nil, "", fmt.Errorf("invalid cursor")
		}
		args = append(args, cursorID)
	}

	// Set default limit
	limit := pagination.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	argCount++
	query += fmt.Sprintf(" ORDER BY id LIMIT $%d", argCount)
	args = append(args, limit+1) // Get one extra to determine if there's a next page

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list agents: %w", err)
	}
	defer rows.Close()

	var agents []*db.Agent
	for rows.Next() {
		var agent db.Agent
		err := rows.Scan(
			&agent.ID, &agent.TenantID, &agent.Email, &agent.Name,
			&agent.Status, &agent.PasswordHash, &agent.CreatedAt, &agent.UpdatedAt)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan agent: %w", err)
		}
		agents = append(agents, &agent)
	}

	// Determine next cursor
	var nextCursor string
	if len(agents) > limit {
		// Remove the extra record and set the cursor
		agents = agents[:limit]
		nextCursor = agents[len(agents)-1].ID.String()
	}

	return agents, nextCursor, nil
}

// GetTenantAdmins retrieves all agents with tenant_admin role for a given tenant
func (r *agentRepository) GetTenantAdmins(ctx context.Context, tenantID uuid.UUID) ([]*db.Agent, error) {
	query := `
		SELECT DISTINCT a.id, a.tenant_id, a.email, a.name, a.status, a.password_hash, a.created_at, a.updated_at
		FROM agents a
		INNER JOIN agent_project_roles apr ON a.id = apr.agent_id
		WHERE a.tenant_id = $1 AND apr.role = 'tenant_admin' AND a.status = 'active'
		ORDER BY a.name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tenant admins: %w", err)
	}
	defer rows.Close()

	var agents []*db.Agent
	for rows.Next() {
		var agent db.Agent
		err := rows.Scan(
			&agent.ID, &agent.TenantID, &agent.Email, &agent.Name,
			&agent.Status, &agent.PasswordHash, &agent.CreatedAt, &agent.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}
		agents = append(agents, &agent)
	}

	return agents, nil
}
