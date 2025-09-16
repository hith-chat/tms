package repo

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/bareuptime/tms/internal/db"
)

type paymentWebhookRepository struct {
	db *sql.DB
}

// NewPaymentWebhookRepository creates a new payment webhook repository
func NewPaymentWebhookRepository(database *sql.DB) PaymentWebhookRepository {
	return &paymentWebhookRepository{db: database}
}

// Create creates a new payment webhook event
func (r *paymentWebhookRepository) Create(event *db.PaymentWebhookEvent) error {
	event.PreCreate()

	query := `
		INSERT INTO payment_webhook_events (event_id, event_type, object_id, tenant_email, tenant_id, amount, currency, status, payload_data, processed_at, created_at, updated_at, error, payment_gateway)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id
	`

	err := r.db.QueryRow(
		query,
		event.EventID,
		event.EventType,
		event.ObjectID,
		event.TenantEmail,
		event.TenantID,
		event.Amount,
		event.Currency,
		event.Status,
		event.PayloadData,
		event.ProcessedAt,
		event.CreatedAt,
		event.UpdatedAt,
		event.Error,
		event.PaymentGateway,
	).Scan(&event.ID)

	if err != nil {
		return fmt.Errorf("failed to create payment webhook event: %w", err)
	}

	return nil
}

// FindByPaymentEventID finds a payment webhook event by event ID
func (r *paymentWebhookRepository) FindByPaymentEventID(eventID string) (*db.PaymentWebhookEvent, error) {
	query := `
		SELECT id, event_id, event_type, object_id, tenant_email, tenant_id, amount, currency, status, payload_data, processed_at, created_at, updated_at, error, payment_gateway
		FROM payment_webhook_events
		WHERE event_id = $1
	`

	event := &db.PaymentWebhookEvent{}
	var payloadData []byte

	err := r.db.QueryRow(query, eventID).Scan(
		&event.ID,
		&event.EventID,
		&event.EventType,
		&event.ObjectID,
		&event.TenantEmail,
		&event.TenantID,
		&event.Amount,
		&event.Currency,
		&event.Status,
		&payloadData,
		&event.ProcessedAt,
		&event.CreatedAt,
		&event.UpdatedAt,
		&event.Error,
		&event.PaymentGateway,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Event not found
		}
		return nil, fmt.Errorf("failed to find payment webhook event: %w", err)
	}

	// Unmarshal payload data
	event.PayloadData = json.RawMessage(payloadData)

	return event, nil
}

// Update updates an existing payment webhook event
func (r *paymentWebhookRepository) Update(event *db.PaymentWebhookEvent) error {
	query := `
		UPDATE payment_webhook_events
		SET event_type = $2, object_id = $3, tenant_email = $4, tenant_id = $5, amount = $6, currency = $7, status = $8, payload_data = $9, processed_at = $10, updated_at = $11, error = $12, payment_gateway = $13
		WHERE id = $1
	`

	_, err := r.db.Exec(
		query,
		event.ID,
		event.EventType,
		event.ObjectID,
		event.TenantEmail,
		event.TenantID,
		event.Amount,
		event.Currency,
		event.Status,
		event.PayloadData,
		event.ProcessedAt,
		event.UpdatedAt,
		event.Error,
		event.PaymentGateway,
	)

	if err != nil {
		return fmt.Errorf("failed to update payment webhook event: %w", err)
	}

	return nil
}

// FindTenantByEmail finds a tenant by email (this is a simple implementation)
// In a real system, you might have a more sophisticated tenant-email relationship
func (r *paymentWebhookRepository) FindTenantByEmail(email string) (*db.Tenant, error) {
	query := `
		SELECT id, name, status, region, kms_key_id, created_at, updated_at
		FROM tenants
		WHERE name = $1 OR id::text = $1
		LIMIT 1
	`

	var tenant db.Tenant
	err := r.db.QueryRow(query, email).Scan(
		&tenant.ID, &tenant.Name, &tenant.Status, &tenant.Region, &tenant.KMSKeyID,
		&tenant.CreatedAt, &tenant.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &tenant, nil
}
