package models

import (
	"time"

	"github.com/google/uuid"
)

// Credits represents the credit balance for a tenant
type Credits struct {
	ID                int64      `db:"id" json:"id"`
	TenantID          uuid.UUID  `db:"tenant_id" json:"tenant_id"`
	Balance           int64      `db:"balance" json:"balance"`                                   // Current credit balance
	TotalEarned       int64      `db:"total_earned" json:"total_earned"`                         // Total credits earned since account creation
	TotalSpent        int64      `db:"total_spent" json:"total_spent"`                           // Total credits spent since account creation
	LastTransactionAt *time.Time `db:"last_transaction_at" json:"last_transaction_at,omitempty"` // Last transaction timestamp
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
}

// CreditTransaction represents a single credit transaction for audit trail
type CreditTransaction struct {
	ID              int64     `db:"id" json:"id"`
	TenantID        uuid.UUID `db:"tenant_id" json:"tenant_id"`
	Amount          int64     `db:"amount" json:"amount"`                               // Positive for credit, negative for debit
	TransactionType string    `db:"transaction_type" json:"transaction_type"`           // 'payment', 'usage', 'refund', 'bonus', etc.
	PaymentGateway  *string   `db:"payment_gateway" json:"payment_gateway,omitempty"`   // 'stripe', 'cashfree', null for non-payment transactions
	PaymentEventID  *string   `db:"payment_event_id" json:"payment_event_id,omitempty"` // Reference to payment webhook event
	Description     *string   `db:"description" json:"description,omitempty"`
	BalanceBefore   int64     `db:"balance_before" json:"balance_before"` // Balance before this transaction
	BalanceAfter    int64     `db:"balance_after" json:"balance_after"`   // Balance after this transaction
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}

// Transaction type constants
const (
	TransactionTypePayment    = "payment"
	TransactionTypeUsage      = "usage"
	TransactionTypeAIUsage    = "ai_usage"
	TransactionTypeRefund     = "refund"
	TransactionTypeBonus      = "bonus"
	TransactionTypeAdjustment = "adjustment"
)

// Payment gateway constants
const (
	PaymentGatewayStripe   = "stripe"
	PaymentGatewayCashfree = "cashfree"
)

// CalculateCreditsFromPayment calculates credits based on payment amount
// Default rate: $1 (100 cents) = 100 credits
func CalculateCreditsFromPayment(paymentAmountCents int64) int64 {
	return paymentAmountCents // 1:1 ratio for now
}
