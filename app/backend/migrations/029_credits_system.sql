-- +goose Up
-- Create credits table for tenant credit management

CREATE TABLE IF NOT EXISTS credits (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    balance BIGINT NOT NULL DEFAULT 0, -- Credit balance in smallest unit (e.g., cents equivalent)
    total_earned BIGINT NOT NULL DEFAULT 0, -- Total credits earned since account creation
    total_spent BIGINT NOT NULL DEFAULT 0, -- Total credits spent since account creation
    last_transaction_at TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    
    -- Constraints
    CONSTRAINT credits_balance_non_negative CHECK (balance >= 0),
    CONSTRAINT credits_totals_non_negative CHECK (total_earned >= 0 AND total_spent >= 0)
);

-- Create credit transactions table for audit trail
CREATE TABLE IF NOT EXISTS credit_transactions (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    amount BIGINT NOT NULL, -- Positive for credit, negative for debit
    transaction_type VARCHAR(50) NOT NULL, -- 'payment', 'usage', 'refund', 'bonus', etc.
    payment_gateway VARCHAR(20), -- 'stripe', 'cashfree', null for non-payment transactions
    payment_event_id VARCHAR(255), -- Reference to payment webhook event
    description TEXT,
    balance_before BIGINT NOT NULL,
    balance_after BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    
    -- Constraints
    CONSTRAINT credit_transactions_amount_not_zero CHECK (amount != 0)
);

-- Create payment webhook events table for tracking payment gateway webhooks
CREATE TABLE IF NOT EXISTS payment_webhook_events (
    id BIGSERIAL PRIMARY KEY,
    event_id VARCHAR(255) NOT NULL UNIQUE, -- Gateway's event ID
    event_type VARCHAR(100) NOT NULL,
    object_id VARCHAR(255) NOT NULL,
    tenant_email VARCHAR(255) NOT NULL,
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
    amount BIGINT, -- Amount in cents
    currency VARCHAR(10),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payload_data JSONB NOT NULL,
    processed_at TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    error TEXT,
    payment_gateway VARCHAR(20) NOT NULL, -- 'stripe', 'cashfree'
    
    -- Constraints
    CONSTRAINT payment_webhook_events_status_check CHECK (status IN ('pending', 'processed', 'failed', 'ignored'))
);

-- Create indexes for performance optimization
CREATE INDEX IF NOT EXISTS idx_credits_tenant_id ON credits(tenant_id);
CREATE INDEX IF NOT EXISTS idx_credit_transactions_tenant_id ON credit_transactions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_credit_transactions_payment_event_id ON credit_transactions(payment_event_id);
CREATE INDEX IF NOT EXISTS idx_credit_transactions_created_at ON credit_transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_credit_transactions_type ON credit_transactions(transaction_type);
CREATE INDEX IF NOT EXISTS idx_payment_webhook_events_event_id ON payment_webhook_events(event_id);
CREATE INDEX IF NOT EXISTS idx_payment_webhook_events_tenant_id ON payment_webhook_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_payment_webhook_events_tenant_email ON payment_webhook_events(tenant_email);
CREATE INDEX IF NOT EXISTS idx_payment_webhook_events_status ON payment_webhook_events(status);
CREATE INDEX IF NOT EXISTS idx_payment_webhook_events_gateway ON payment_webhook_events(payment_gateway);

-- +goose Down
-- Remove credits system

DROP INDEX IF EXISTS idx_credits_tenant_id;
DROP INDEX IF EXISTS idx_credit_transactions_tenant_id;
DROP INDEX IF EXISTS idx_credit_transactions_payment_event_id;
DROP INDEX IF EXISTS idx_credit_transactions_created_at;
DROP INDEX IF EXISTS idx_credit_transactions_type;
DROP INDEX IF EXISTS idx_payment_webhook_events_event_id;
DROP INDEX IF EXISTS idx_payment_webhook_events_tenant_id;
DROP INDEX IF EXISTS idx_payment_webhook_events_tenant_email;
DROP INDEX IF EXISTS idx_payment_webhook_events_status;
DROP INDEX IF EXISTS idx_payment_webhook_events_gateway;
DROP TABLE IF EXISTS payment_webhook_events;
DROP TABLE IF EXISTS credit_transactions;
DROP TABLE IF EXISTS credits;
