package repo

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/bareuptime/tms/internal/db"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type apiKeyRepository struct {
	db *sqlx.DB
}

// NewApiKeyRepository creates a new API key repository
func NewApiKeyRepository(database *sqlx.DB) ApiKeyRepository {
	return &apiKeyRepository{
		db: database,
	}
}

// Create creates a new API key
func (r *apiKeyRepository) Create(ctx context.Context, apiKey *db.ApiKey) error {
	query := `
		INSERT INTO api_keys (
			id, tenant_id, project_id, name, key_hash, key_prefix, 
			scopes, expires_at, is_active, created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)`

	_, err := r.db.ExecContext(ctx, query,
		apiKey.ID,
		apiKey.TenantID,
		apiKey.ProjectID,
		apiKey.Name,
		apiKey.KeyHash,
		apiKey.KeyPrefix,
		pq.Array(apiKey.Scopes),
		apiKey.ExpiresAt,
		apiKey.IsActive,
		apiKey.AgentID,
		apiKey.CreatedAt,
		apiKey.UpdatedAt,
	)

	return err
}

// GetByID retrieves an API key by its ID
func (r *apiKeyRepository) GetByID(ctx context.Context, tenantID uuid.UUID, keyID uuid.UUID) (*db.ApiKey, error) {
	query := `
		SELECT id, tenant_id, project_id, name, key_hash, key_prefix,
			   scopes, last_used_at, expires_at, is_active, created_by, created_at, updated_at
		FROM api_keys 
		WHERE id = $1 AND tenant_id = $2`

	var apiKey db.ApiKey
	err := r.db.GetContext(ctx, &apiKey, query, keyID, tenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("api key not found")
		}
		return nil, err
	}

	return &apiKey, nil
}

// GetByHash retrieves an API key by its hash (for authentication)
func (r *apiKeyRepository) GetByHash(ctx context.Context, keyHash string) (*db.ApiKey, error) {
	query := `
		SELECT id, tenant_id, project_id, name, key_hash, key_prefix,
			   scopes, last_used_at, expires_at, is_active, created_by, created_at, updated_at
		FROM api_keys 
		WHERE key_hash = $1 AND is_active = true 
		AND (expires_at IS NULL OR expires_at > NOW())`

	var apiKey db.ApiKey
	err := r.db.GetContext(ctx, &apiKey, query, keyHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("api key not found or expired")
		}
		return nil, err
	}

	return &apiKey, nil
}

// List retrieves all API keys for a tenant/project
func (r *apiKeyRepository) List(ctx context.Context, tenantID uuid.UUID, projectID *uuid.UUID) ([]*db.ApiKey, error) {
	var query string
	var args []interface{}

	if projectID != nil {
		query = `
			SELECT id, tenant_id, project_id, name, key_hash, key_prefix,
				   scopes, last_used_at, expires_at, is_active, created_by, created_at, updated_at
			FROM api_keys 
			WHERE tenant_id = $1 AND project_id = $2 
			ORDER BY created_at DESC`
		args = []interface{}{tenantID, *projectID}
	} else {
		query = `
			SELECT id, tenant_id, project_id, name, key_hash, key_prefix,
				   scopes, last_used_at, expires_at, is_active, created_by, created_at, updated_at
			FROM api_keys 
			WHERE tenant_id = $1 AND project_id IS NULL 
			ORDER BY created_at DESC`
		args = []interface{}{tenantID}
	}

	var apiKeys []*db.ApiKey
	err := r.db.SelectContext(ctx, &apiKeys, query, args...)
	if err != nil {
		return nil, err
	}

	return apiKeys, nil
}

// Update updates an API key
func (r *apiKeyRepository) Update(ctx context.Context, apiKey *db.ApiKey) error {
	query := `
		UPDATE api_keys 
		SET name = $1, scopes = $2, expires_at = $3, is_active = $4, updated_at = $5
		WHERE id = $6 AND tenant_id = $7`

	apiKey.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		apiKey.Name,
		pq.Array(apiKey.Scopes),
		apiKey.ExpiresAt,
		apiKey.IsActive,
		apiKey.UpdatedAt,
		apiKey.ID,
		apiKey.TenantID,
	)

	return err
}

// Delete deletes an API key
func (r *apiKeyRepository) Delete(ctx context.Context, tenantID uuid.UUID, keyID uuid.UUID) error {
	query := `DELETE FROM api_keys WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, keyID, tenantID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("api key not found")
	}

	return nil
}

// UpdateLastUsed updates the last used timestamp for an API key
func (r *apiKeyRepository) UpdateLastUsed(ctx context.Context, keyID uuid.UUID) error {
	query := `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, keyID)
	return err
}

// HashApiKey creates a SHA-256 hash of the API key
func HashApiKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
