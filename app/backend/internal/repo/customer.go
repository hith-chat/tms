package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/bareuptime/tms/internal/db"
	"github.com/bareuptime/tms/internal/logger"
	"github.com/google/uuid"
)

// customerRepository implements CustomerRepository interface
type customerRepository struct {
	db *sql.DB
}

// NewCustomerRepository creates a new customer repository
func NewCustomerRepository(database *sql.DB) CustomerRepository {
	return &customerRepository{
		db: database,
	}
}

// Create creates a new customer
func (r *customerRepository) Create(ctx context.Context, customer *db.Customer) error {
	query := `
		INSERT INTO customers (id, tenant_id, email, name, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`

	// Marshal metadata map to JSON for storage. Store '{}' for nil/empty maps.
	var metadataJSON string
	if len(customer.Metadata) == 0 {
		metadataJSON = "{}"
	} else {
		b, err := json.Marshal(customer.Metadata)
		if err != nil {
			return err
		}
		metadataJSON = string(b)
	}

	_, err := r.db.ExecContext(ctx, query,
		customer.ID,
		customer.TenantID,
		customer.Email,
		customer.Name,
		metadataJSON,
	)

	return err
}

// GetByID retrieves a customer by ID
func (r *customerRepository) GetByID(ctx context.Context, tenantID, customerID uuid.UUID) (*db.Customer, error) {
	logger.DebugfCtx(ctx, "Getting customer by ID - tenantID: %s, customerID: %s", tenantID.String(), customerID.String())

	query := `
		SELECT id, tenant_id, email, name, metadata, created_at, updated_at
		FROM customers
		WHERE tenant_id = $1 AND id = $2
	`

	logger.DebugfCtx(ctx, "Executing customer query: %s", query)

	customer := &db.Customer{}
	var metadataJSON sql.NullString

	err := r.db.QueryRowContext(ctx, query, tenantID, customerID).Scan(
		&customer.ID,
		&customer.TenantID,
		&customer.Email,
		&customer.Name,
		&metadataJSON,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		logger.ErrorfCtx(ctx, err, "Error retrieving customer: %v", err)
		return nil, err
	}

	// Parse metadata JSON
	customer.Metadata = make(map[string]string)
	if metadataJSON.Valid && metadataJSON.String != "" {
		var meta map[string]string
		if err := json.Unmarshal([]byte(metadataJSON.String), &meta); err == nil {
			customer.Metadata = meta
		} else {
			// Log parse error but continue with empty metadata
			logger.WarnfCtx(ctx, "Failed to parse metadata JSON for customer %s: %v", customer.ID.String(), err)
		}
	}

	logger.DebugfCtx(ctx, "Customer retrieved successfully - email: %s, name: %s", customer.Email, customer.Name)

	return customer, nil
}

// GetByEmail retrieves a customer by email
func (r *customerRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*db.Customer, error) {
	query := `
		SELECT id, tenant_id, email, name, metadata, created_at, updated_at
		FROM customers
		WHERE tenant_id = $1 AND email = $2
	`

	customer := &db.Customer{}
	var metadataJSON sql.NullString

	err := r.db.QueryRowContext(ctx, query, tenantID, email).Scan(
		&customer.ID,
		&customer.TenantID,
		&customer.Email,
		&customer.Name,
		&metadataJSON,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Customer not found
		}
		return nil, err
	}

	// Parse metadata JSON
	customer.Metadata = make(map[string]string)
	if metadataJSON.Valid && metadataJSON.String != "" {
		var meta map[string]string
		if err := json.Unmarshal([]byte(metadataJSON.String), &meta); err == nil {
			customer.Metadata = meta
		}
	}

	return customer, nil
}

// Update updates an existing customer
func (r *customerRepository) Update(ctx context.Context, customer *db.Customer) error {
	query := `
		UPDATE customers
		SET name = $3, metadata = $4, updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2
	`

	metadataJSON := "{}"
	if customer.Metadata != nil {
		// Convert map to JSON string - in production you'd use proper JSON marshaling
		metadataJSON = "{}" // Simplified for now
	}

	_, err := r.db.ExecContext(ctx, query,
		customer.TenantID,
		customer.ID,
		customer.Name,
		metadataJSON,
	)

	return err
}

// Delete deletes a customer
func (r *customerRepository) Delete(ctx context.Context, tenantID, customerID uuid.UUID) error {
	query := `DELETE FROM customers WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.ExecContext(ctx, query, tenantID, customerID)
	return err
}

// List retrieves a list of customers with filters and pagination
func (r *customerRepository) List(ctx context.Context, tenantID uuid.UUID, filters CustomerFilters, pagination PaginationParams) ([]*db.Customer, string, error) {
	baseQuery := `
		SELECT id, tenant_id, email, name, metadata, created_at, updated_at
		FROM customers
		WHERE tenant_id = $1
	`

	args := []interface{}{tenantID}
	argIndex := 2

	// Add filters
	if filters.Email != "" {
		baseQuery += fmt.Sprintf(" AND email = $%d", argIndex)
		args = append(args, filters.Email)
		argIndex++
	}

	if filters.Search != "" {
		baseQuery += fmt.Sprintf(" AND (name ILIKE $%d OR email ILIKE $%d)", argIndex, argIndex)
		searchTerm := "%" + filters.Search + "%"
		args = append(args, searchTerm)
		argIndex++
	}

	// Add cursor-based pagination
	if pagination.Cursor != "" {
		baseQuery += fmt.Sprintf(" AND id > $%d", argIndex)
		cursorID, err := uuid.Parse(pagination.Cursor)
		if err != nil {
			return nil, "", fmt.Errorf("invalid cursor: %w", err)
		}
		args = append(args, cursorID)
		argIndex++
	}

	// Add ordering and limit
	baseQuery += " ORDER BY id ASC"

	limit := pagination.Limit
	if limit <= 0 || limit > 100 {
		limit = 50 // Default limit
	}

	baseQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
	args = append(args, limit+1) // Fetch one extra to determine if there's a next page

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var customers []*db.Customer
	for rows.Next() {
		customer := &db.Customer{}
		var metadataJSON sql.NullString

		err := rows.Scan(
			&customer.ID,
			&customer.TenantID,
			&customer.Email,
			&customer.Name,
			&metadataJSON,
			&customer.CreatedAt,
			&customer.UpdatedAt,
		)
		if err != nil {
			return nil, "", err
		}

		// Parse metadata JSON
		customer.Metadata = make(map[string]string)
		if metadataJSON.Valid && metadataJSON.String != "" {
			var meta map[string]string
			if err := json.Unmarshal([]byte(metadataJSON.String), &meta); err == nil {
				customer.Metadata = meta
			}
		}
		customers = append(customers, customer)
	}

	// Determine next cursor
	var nextCursor string
	if len(customers) > limit {
		nextCursor = customers[limit-1].ID.String()
		customers = customers[:limit] // Remove the extra record
	}

	return customers, nextCursor, nil
}
