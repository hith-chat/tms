package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/bareuptime/tms/internal/db"
	"github.com/google/uuid"
)

type creditsRepository struct {
	db *sql.DB
}

// NewCreditsRepository creates a new credits repository
func NewCreditsRepository(database *sql.DB) CreditsRepository {
	return &creditsRepository{db: database}
}

// Create creates a new credits record for a tenant
func (r *creditsRepository) Create(ctx context.Context, credits *db.Credits) error {
	query := `
		INSERT INTO credits (tenant_id, balance, total_earned, total_spent, last_transaction_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		credits.TenantID, credits.Balance, credits.TotalEarned, credits.TotalSpent, credits.LastTransactionAt).
		Scan(&credits.ID, &credits.CreatedAt, &credits.UpdatedAt)
	return err
}

// GetByTenantID retrieves credits by tenant ID
func (r *creditsRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*db.Credits, error) {
	query := `
		SELECT id, tenant_id, balance, total_earned, total_spent, last_transaction_at, created_at, updated_at
		FROM credits
		WHERE tenant_id = $1
	`

	var credits db.Credits
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&credits.ID, &credits.TenantID, &credits.Balance, &credits.TotalEarned, &credits.TotalSpent,
		&credits.LastTransactionAt, &credits.CreatedAt, &credits.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &credits, nil
}

// Update updates an existing credits record
func (r *creditsRepository) Update(ctx context.Context, credits *db.Credits) error {
	query := `
		UPDATE credits 
		SET balance = $2, total_earned = $3, total_spent = $4, last_transaction_at = $5, updated_at = NOW()
		WHERE tenant_id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		credits.TenantID, credits.Balance, credits.TotalEarned, credits.TotalSpent, credits.LastTransactionAt)
	return err
}

// AddCredits adds credits to a tenant account and creates a transaction record
func (r *creditsRepository) AddCredits(ctx context.Context, tenantID uuid.UUID, amount int64, transactionType, paymentGateway, paymentEventID, description string) (*db.CreditTransaction, error) {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get current credits (with row lock)
	var credits db.Credits
	query := `
		SELECT id, tenant_id, balance, total_earned, total_spent, last_transaction_at, created_at, updated_at
		FROM credits
		WHERE tenant_id = $1
		FOR UPDATE
	`
	err = tx.QueryRowContext(ctx, query, tenantID).Scan(
		&credits.ID, &credits.TenantID, &credits.Balance, &credits.TotalEarned, &credits.TotalSpent,
		&credits.LastTransactionAt, &credits.CreatedAt, &credits.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// Create new credits record if it doesn't exist
			credits = db.Credits{
				TenantID:    tenantID,
				Balance:     0,
				TotalEarned: 0,
				TotalSpent:  0,
			}

			createQuery := `
				INSERT INTO credits (tenant_id, balance, total_earned, total_spent, created_at, updated_at)
				VALUES ($1, $2, $3, $4, NOW(), NOW())
				RETURNING id, created_at, updated_at
			`
			err = tx.QueryRowContext(ctx, createQuery,
				credits.TenantID, credits.Balance, credits.TotalEarned, credits.TotalSpent).
				Scan(&credits.ID, &credits.CreatedAt, &credits.UpdatedAt)
			if err != nil {
				return nil, fmt.Errorf("failed to create credits record: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get credits: %w", err)
		}
	}

	// Calculate new balance
	balanceBefore := credits.Balance
	balanceAfter := credits.Balance + amount
	now := time.Now()

	// Update credits
	updateQuery := `
		UPDATE credits 
		SET balance = $2, total_earned = $3, last_transaction_at = $4, updated_at = NOW()
		WHERE tenant_id = $1
	`
	_, err = tx.ExecContext(ctx, updateQuery, tenantID, balanceAfter, credits.TotalEarned+amount, now)
	if err != nil {
		return nil, fmt.Errorf("failed to update credits: %w", err)
	}

	// Create transaction record
	transaction := &db.CreditTransaction{
		TenantID:        tenantID,
		Amount:          amount,
		TransactionType: transactionType,
		BalanceBefore:   balanceBefore,
		BalanceAfter:    balanceAfter,
		CreatedAt:       now,
	}

	if paymentGateway != "" {
		transaction.PaymentGateway = &paymentGateway
	}
	if paymentEventID != "" {
		transaction.PaymentEventID = &paymentEventID
	}
	if description != "" {
		transaction.Description = &description
	}

	insertTxQuery := `
		INSERT INTO credit_transactions (tenant_id, amount, transaction_type, payment_gateway, payment_event_id, description, balance_before, balance_after, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	err = tx.QueryRowContext(ctx, insertTxQuery,
		transaction.TenantID, transaction.Amount, transaction.TransactionType,
		transaction.PaymentGateway, transaction.PaymentEventID, transaction.Description,
		transaction.BalanceBefore, transaction.BalanceAfter, transaction.CreatedAt).
		Scan(&transaction.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return transaction, nil
}

// DeductCredits deducts credits from a tenant account and creates a transaction record
func (r *creditsRepository) DeductCredits(ctx context.Context, tenantID uuid.UUID, amount int64, transactionType, description string) (*db.CreditTransaction, error) {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get current credits (with row lock)
	var credits db.Credits
	query := `
		SELECT id, tenant_id, balance, total_earned, total_spent, last_transaction_at, created_at, updated_at
		FROM credits
		WHERE tenant_id = $1
		FOR UPDATE
	`
	err = tx.QueryRowContext(ctx, query, tenantID).Scan(
		&credits.ID, &credits.TenantID, &credits.Balance, &credits.TotalEarned, &credits.TotalSpent,
		&credits.LastTransactionAt, &credits.CreatedAt, &credits.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("credits account not found for tenant %s", tenantID)
		}
		return nil, fmt.Errorf("failed to get credits: %w", err)
	}

	// Check if sufficient credits available
	if credits.Balance < amount {
		return nil, fmt.Errorf("insufficient credits: balance %d, required %d", credits.Balance, amount)
	}

	// Calculate new balance
	balanceBefore := credits.Balance
	balanceAfter := credits.Balance - amount
	now := time.Now()

	// Update credits
	updateQuery := `
		UPDATE credits 
		SET balance = $2, total_spent = $3, last_transaction_at = $4, updated_at = NOW()
		WHERE tenant_id = $1
	`
	_, err = tx.ExecContext(ctx, updateQuery, tenantID, balanceAfter, credits.TotalSpent+amount, now)
	if err != nil {
		return nil, fmt.Errorf("failed to update credits: %w", err)
	}

	// Create transaction record
	transaction := &db.CreditTransaction{
		TenantID:        tenantID,
		Amount:          -amount, // Negative for deduction
		TransactionType: transactionType,
		BalanceBefore:   balanceBefore,
		BalanceAfter:    balanceAfter,
		CreatedAt:       now,
	}

	if description != "" {
		transaction.Description = &description
	}

	insertTxQuery := `
		INSERT INTO credit_transactions (tenant_id, amount, transaction_type, payment_gateway, payment_event_id, description, balance_before, balance_after, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	err = tx.QueryRowContext(ctx, insertTxQuery,
		transaction.TenantID, transaction.Amount, transaction.TransactionType,
		transaction.PaymentGateway, transaction.PaymentEventID, transaction.Description,
		transaction.BalanceBefore, transaction.BalanceAfter, transaction.CreatedAt).
		Scan(&transaction.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return transaction, nil
}

// CreateTransaction creates a new credit transaction record
func (r *creditsRepository) CreateTransaction(ctx context.Context, transaction *db.CreditTransaction) error {
	query := `
		INSERT INTO credit_transactions (tenant_id, amount, transaction_type, payment_gateway, payment_event_id, description, balance_before, balance_after, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query,
		transaction.TenantID, transaction.Amount, transaction.TransactionType,
		transaction.PaymentGateway, transaction.PaymentEventID, transaction.Description,
		transaction.BalanceBefore, transaction.BalanceAfter, transaction.CreatedAt).
		Scan(&transaction.ID)
	return err
}

// GetTransactionsByTenantID retrieves credit transactions for a tenant with pagination
func (r *creditsRepository) GetTransactionsByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*db.CreditTransaction, string, error) {
	query := `
		SELECT id, tenant_id, amount, transaction_type, payment_gateway, payment_event_id, description, balance_before, balance_after, created_at
		FROM credit_transactions
		WHERE tenant_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2
	`

	// For simplicity, implementing basic limit-based pagination
	// In production, you might want cursor-based pagination
	limit := pagination.Limit
	if limit <= 0 {
		limit = 50 // Default limit
	}

	rows, err := r.db.QueryContext(ctx, query, tenantID, limit)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var transactions []*db.CreditTransaction
	for rows.Next() {
		var transaction db.CreditTransaction
		err := rows.Scan(
			&transaction.ID, &transaction.TenantID, &transaction.Amount, &transaction.TransactionType,
			&transaction.PaymentGateway, &transaction.PaymentEventID, &transaction.Description,
			&transaction.BalanceBefore, &transaction.BalanceAfter, &transaction.CreatedAt)
		if err != nil {
			return nil, "", err
		}
		transactions = append(transactions, &transaction)
	}

	// Simple next cursor implementation
	nextCursor := ""
	if len(transactions) == limit {
		nextCursor = fmt.Sprintf("%d", transactions[len(transactions)-1].ID)
	}

	return transactions, nextCursor, nil
}

// GetTransactionByPaymentEventID retrieves a transaction by payment event ID for idempotency
func (r *creditsRepository) GetTransactionByPaymentEventID(ctx context.Context, paymentEventID string) (*db.CreditTransaction, error) {
	query := `
		SELECT id, tenant_id, amount, transaction_type, payment_gateway, payment_event_id, description, balance_before, balance_after, created_at
		FROM credit_transactions
		WHERE payment_event_id = $1
	`

	var transaction db.CreditTransaction
	err := r.db.QueryRowContext(ctx, query, paymentEventID).Scan(
		&transaction.ID, &transaction.TenantID, &transaction.Amount, &transaction.TransactionType,
		&transaction.PaymentGateway, &transaction.PaymentEventID, &transaction.Description,
		&transaction.BalanceBefore, &transaction.BalanceAfter, &transaction.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &transaction, nil
}
