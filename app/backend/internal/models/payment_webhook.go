package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PaymentWebhookEvent represents a webhook event received from payment gateways
type PaymentWebhookEvent struct {
	ID             int64           `db:"id" json:"id"`                               // Auto-increment ID
	EventID        string          `db:"event_id" json:"event_id"`                   // Gateway's event ID
	EventType      string          `db:"event_type" json:"event_type"`               // Event type
	ObjectID       string          `db:"object_id" json:"object_id"`                 // The ID of the object this event is about
	TenantEmail    string          `db:"tenant_email" json:"tenant_email"`           // Associated tenant email (for identification)
	TenantID       *uuid.UUID      `db:"tenant_id" json:"tenant_id,omitempty"`       // Associated tenant ID if found
	Amount         *int64          `db:"amount" json:"amount,omitempty"`             // Amount in cents
	Currency       string          `db:"currency" json:"currency"`                   // Currency code
	Status         string          `db:"status" json:"status"`                       // Status of the event processing
	PayloadData    json.RawMessage `db:"payload_data" json:"payload_data"`           // Full webhook payload as JSON
	ProcessedAt    *time.Time      `db:"processed_at" json:"processed_at,omitempty"` // When the event was processed
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`               // When the webhook was received
	UpdatedAt      time.Time       `db:"updated_at" json:"updated_at"`               // Last update time
	Error          *string         `db:"error" json:"error,omitempty"`               // Error message if processing failed
	PaymentGateway string          `db:"payment_gateway" json:"payment_gateway"`     // Payment gateway used
}

// WebhookEventStatus constants
const (
	WebhookStatusPending   = "pending"
	WebhookStatusProcessed = "processed"
	WebhookStatusFailed    = "failed"
	WebhookStatusIgnored   = "ignored"
)

// PreCreate sets default values before creation
func (w *PaymentWebhookEvent) PreCreate() {
	now := time.Now()
	w.CreatedAt = now
	w.UpdatedAt = now

	if w.Status == "" {
		w.Status = WebhookStatusPending
	}
}

// MarkProcessed marks the webhook event as successfully processed
func (w *PaymentWebhookEvent) MarkProcessed() {
	now := time.Now()
	w.Status = WebhookStatusProcessed
	w.ProcessedAt = &now
	w.UpdatedAt = now
}

// MarkFailed marks the webhook event as failed with an error message
func (w *PaymentWebhookEvent) MarkFailed(errorMsg string) {
	w.Status = WebhookStatusFailed
	w.Error = &errorMsg
	w.UpdatedAt = time.Now()
}

// MarkIgnored marks the webhook event as ignored (e.g., event type not handled)
func (w *PaymentWebhookEvent) MarkIgnored() {
	w.Status = WebhookStatusIgnored
	w.UpdatedAt = time.Now()
}
