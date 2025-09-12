package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/bareuptime/tms/internal/models"
)

// AlarmRepository handles database operations for alarms
type AlarmRepository struct {
	db *sqlx.DB
}

// NewAlarmRepository creates a new alarm repository
func NewAlarmRepository(db *sqlx.DB) *AlarmRepository {
	return &AlarmRepository{db: db}
}

// CreateAlarm creates a new alarm in the database
func (r *AlarmRepository) CreateAlarm(ctx context.Context, alarm *models.Alarm) error {
	query := `
		INSERT INTO alarms (
			id, tenant_id, project_id, agent_id, title, message, 
			priority, current_level, start_time, last_escalation, escalation_count,
			is_acknowledged, config, metadata, created_at, updated_at
		) VALUES (
			:id, :tenant_id, :project_id, :agent_id, :title, :message,
			:priority, :current_level, :start_time, :last_escalation, :escalation_count,
			:is_acknowledged, :config, :metadata, :created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, alarm)
	return err
}

// GetAlarmByID retrieves an alarm by ID
func (r *AlarmRepository) GetAlarmByID(ctx context.Context, tenantID, alarmID uuid.UUID) (*models.Alarm, error) {
	var alarm models.Alarm
	query := `
		SELECT id, tenant_id, project_id, agent_id, title, message,
			   priority, current_level, start_time, last_escalation, escalation_count,
			   is_acknowledged, acknowledged_at, acknowledged_by, config, metadata,
			   created_at, updated_at
		FROM alarms 
		WHERE id = $1 AND tenant_id = $2`

	err := r.db.GetContext(ctx, &alarm, query, alarmID, tenantID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("alarm not found")
	}
	return &alarm, err
}

// GetActiveAlarms retrieves all active (unacknowledged) alarms for a project
func (r *AlarmRepository) GetActiveAlarms(ctx context.Context, tenantID, projectID uuid.UUID) ([]*models.Alarm, error) {
	var alarms []*models.Alarm
	query := `
		SELECT id, tenant_id, project_id, agent_id, title, message,
			   priority, current_level, start_time, last_escalation, escalation_count,
			   is_acknowledged, acknowledged_at, acknowledged_by, config, metadata,
			   created_at, updated_at
		FROM alarms 
		WHERE tenant_id = $1 AND project_id = $2 AND is_acknowledged = false
		ORDER BY start_time DESC`

	err := r.db.SelectContext(ctx, &alarms, query, tenantID, projectID)
	return alarms, err
}

// GetAllActiveAlarms retrieves all active alarms across all projects for a tenant
func (r *AlarmRepository) GetAllActiveAlarms(ctx context.Context, tenantID uuid.UUID) ([]*models.Alarm, error) {
	var alarms []*models.Alarm
	query := `
		SELECT id, tenant_id, project_id, agent_id, title, message,
			   priority, current_level, start_time, last_escalation, escalation_count,
			   is_acknowledged, acknowledged_at, acknowledged_by, config, metadata,
			   created_at, updated_at
		FROM alarms 
		WHERE tenant_id = $1 AND is_acknowledged = false
		ORDER BY start_time DESC`

	err := r.db.SelectContext(ctx, &alarms, query, tenantID)
	return alarms, err
}

// UpdateAlarm updates an existing alarm
func (r *AlarmRepository) UpdateAlarm(ctx context.Context, alarm *models.Alarm) error {
	query := `
		UPDATE alarms SET
			current_level = :current_level,
			last_escalation = :last_escalation,
			escalation_count = :escalation_count,
			is_acknowledged = :is_acknowledged,
			acknowledged_at = :acknowledged_at,
			acknowledged_by = :acknowledged_by,
			config = :config,
			metadata = :metadata,
			updated_at = :updated_at
		WHERE id = :id AND tenant_id = :tenant_id`

	result, err := r.db.NamedExecContext(ctx, query, alarm)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alarm not found or not updated")
	}

	return nil
}

// AcknowledgeAlarm marks an alarm as acknowledged
func (r *AlarmRepository) AcknowledgeAlarm(ctx context.Context, tenantID, alarmID, agentID uuid.UUID, response string) error {
	now := time.Now()

	// Start transaction for both alarm update and acknowledgment record
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update alarm as acknowledged
	updateQuery := `
		UPDATE alarms SET
			is_acknowledged = true,
			acknowledged_at = $1,
			acknowledged_by = $2,
			updated_at = $1
		WHERE id = $3 AND tenant_id = $4 AND is_acknowledged = false`

	result, err := tx.ExecContext(ctx, updateQuery, now, agentID, alarmID, tenantID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alarm not found or already acknowledged")
	}

	// Create acknowledgment record
	acknowledgmentID := uuid.New()
	ackQuery := `
		INSERT INTO alarm_acknowledgments (
			id, alarm_id, agent_id, response, acknowledged_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $5)`

	_, err = tx.ExecContext(ctx, ackQuery, acknowledgmentID, alarmID, agentID, response, now)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetAlarmStats gets alarm statistics for a project
func (r *AlarmRepository) GetAlarmStats(ctx context.Context, tenantID, projectID uuid.UUID) (*models.AlarmStats, error) {
	stats := &models.AlarmStats{}

	// Get active alarms count
	activeQuery := `
		SELECT COUNT(*) 
		FROM alarms 
		WHERE tenant_id = $1 AND project_id = $2 AND is_acknowledged = false`

	err := r.db.GetContext(ctx, &stats.ActiveCount, activeQuery, tenantID, projectID)
	if err != nil {
		return nil, err
	}

	// Get critical alarms count
	criticalQuery := `
		SELECT COUNT(*) 
		FROM alarms 
		WHERE tenant_id = $1 AND project_id = $2 AND is_acknowledged = false 
		AND current_level = 'critical'`

	err = r.db.GetContext(ctx, &stats.CriticalCount, criticalQuery, tenantID, projectID)
	if err != nil {
		return nil, err
	}

	// Get unacknowledged count (same as active for now)
	stats.UnacknowledgedCount = stats.ActiveCount

	// Get total alarms today
	todayQuery := `
		SELECT COUNT(*) 
		FROM alarms 
		WHERE tenant_id = $1 AND project_id = $2 
		AND DATE(created_at) = CURRENT_DATE`

	err = r.db.GetContext(ctx, &stats.TotalToday, todayQuery, tenantID, projectID)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// DeleteAlarm deletes an alarm (soft delete could be implemented instead)
func (r *AlarmRepository) DeleteAlarm(ctx context.Context, tenantID, alarmID uuid.UUID) error {
	query := `DELETE FROM alarms WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, alarmID, tenantID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alarm not found")
	}

	return nil
}

// GetAlarmsForEscalation gets alarms that need escalation
func (r *AlarmRepository) GetAlarmsForEscalation(ctx context.Context) ([]*models.Alarm, error) {
	var alarms []*models.Alarm
	query := `
		SELECT id, tenant_id, project_id, agent_id, title, message,
			   priority, current_level, start_time, last_escalation, escalation_count,
			   is_acknowledged, acknowledged_at, acknowledged_by, config, metadata,
			   created_at, updated_at
		FROM alarms 
		WHERE is_acknowledged = false
		ORDER BY last_escalation ASC`

	err := r.db.SelectContext(ctx, &alarms, query)
	return alarms, err
}
