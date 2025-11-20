package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// GetSetting retrieves a setting by tenant ID and key
func (r *SettingsRepository) GetSetting(ctx context.Context, tenantID, projectID uuid.UUID, settingKey string) (map[string]interface{}, int, error) {
	query := `
		SELECT setting_value 
		FROM tenant_project_settings 
		WHERE tenant_id = $1 AND project_id = $2 AND setting_key = $3
	`

	var settingValueJSON []byte
	err := r.db.QueryRowContext(ctx, query, tenantID, projectID, settingKey).Scan(&settingValueJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			// Setting not found is expected, not an error - return empty gracefully
			return nil, http.StatusNoContent, fmt.Errorf("setting not found: %s", settingKey)
		}
		// Only log actual database errors
		fmt.Println("Error retrieving setting:", err)
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to get setting: %w", err)
	}

	var settingValue map[string]interface{}
	err = json.Unmarshal(settingValueJSON, &settingValue)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to unmarshal setting value: %w", err)
	}

	return settingValue, http.StatusOK, nil
}

// UpdateSetting updates or creates a setting
func (r *SettingsRepository) UpdateSetting(ctx context.Context, tenantID, projectUUID uuid.UUID, settingKey string, settingValue map[string]interface{}) error {
	settingValueJSON, err := json.Marshal(settingValue)
	if err != nil {
		return fmt.Errorf("failed to marshal setting value: %w", err)
	}

	query := `
		INSERT INTO tenant_project_settings (tenant_id, project_id, setting_key, setting_value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (tenant_id, project_id, setting_key)
		DO UPDATE SET 
			setting_value = EXCLUDED.setting_value,
			updated_at = NOW()
	`

	_, err = r.db.ExecContext(ctx, query, tenantID, projectUUID, settingKey, settingValueJSON)
	if err != nil {
		fmt.Println("Error updating setting:", err)
		return fmt.Errorf("failed to update setting: %w", err)
	}

	return nil
}

// DeleteSetting removes a setting
func (r *SettingsRepository) DeleteSetting(ctx context.Context, tenantID uuid.UUID, settingKey string) error {
	query := `
		DELETE FROM tenant_project_settings 
		WHERE tenant_id = $1 AND setting_key = $2
	`

	result, err := r.db.ExecContext(ctx, query, tenantID, settingKey)
	if err != nil {
		return fmt.Errorf("failed to delete setting: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("setting not found: %s", settingKey)
	}

	return nil
}

// ListSettings retrieves all settings for a tenant
func (r *SettingsRepository) ListSettings(ctx context.Context, tenantID uuid.UUID) (map[string]map[string]interface{}, error) {
	query := `
		SELECT setting_key, setting_value 
		FROM tenant_project_settings 
		WHERE tenant_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list settings: %w", err)
	}
	defer rows.Close()

	settings := make(map[string]map[string]interface{})

	for rows.Next() {
		var settingKey string
		var settingValueJSON []byte

		err := rows.Scan(&settingKey, &settingValueJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan setting row: %w", err)
		}

		var settingValue map[string]interface{}
		err = json.Unmarshal(settingValueJSON, &settingValue)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal setting value: %w", err)
		}

		settings[settingKey] = settingValue
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate settings rows: %w", err)
	}

	return settings, nil
}
