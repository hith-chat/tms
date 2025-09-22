package repo

import (
	"context"

	"github.com/bareuptime/tms/internal/db"
	"github.com/google/uuid"
)

// TicketFilters represents filters for ticket queries
type TicketFilters struct {
	Status      []string
	Priority    []string
	AssigneeID  *uuid.UUID
	RequesterID *uuid.UUID
	Tags        []string
	Search      string
	Source      []string
	Type        []string
}

// AgentFilters represents filters for agent queries
type AgentFilters struct {
	Email       string
	IsActive    *bool
	Search      string
	AgentID     uuid.UUID
	ProjectID   *uuid.UUID // Filter agents by project access
	Status      []string   // Filter by agent status (online, available, busy, offline)
	Skills      []string   // Filter by agent skills
	MaxWorkload *int       // Filter by maximum current workload
}

// CustomerFilters represents filters for customer queries
type CustomerFilters struct {
	Email  string
	Search string
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Cursor string
	Limit  int
}

// TicketRepository interface
type TicketRepository interface {
	Create(ctx context.Context, ticket *db.Ticket) error
	GetByTenantAndProjectID(ctx context.Context, tenantID, projectID, ticketID uuid.UUID) (*db.Ticket, error)
	GetByID(ctx context.Context, ticketID uuid.UUID) (*db.Ticket, error)
	Update(ctx context.Context, ticket *db.Ticket) error
	Delete(ctx context.Context, tenantID, projectID, ticketID uuid.UUID) error
	List(ctx context.Context, tenantID, projectID uuid.UUID, filters TicketFilters, pagination PaginationParams) ([]*db.Ticket, string, error)
	GetByNumber(ctx context.Context, tenantID uuid.UUID, number int) (*db.Ticket, error)
}

// AgentRepository interface
type AgentRepository interface {
	Create(ctx context.Context, agent *db.Agent) error
	GetByID(ctx context.Context, tenantID, agentID uuid.UUID) (*db.Agent, error)
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*db.Agent, error)
	GetByEmailWithoutTenantID(ctx context.Context, email string) (*db.Agent, error)
	Update(ctx context.Context, agent *db.Agent) error
	Delete(ctx context.Context, tenantID, agentID uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filters AgentFilters, pagination PaginationParams) ([]*db.Agent, string, error)
	GetTenantAdmins(ctx context.Context, tenantID uuid.UUID) ([]*db.Agent, error)
}

// CustomerRepository interface
type CustomerRepository interface {
	Create(ctx context.Context, customer *db.Customer) error
	GetByID(ctx context.Context, tenantID, customerID uuid.UUID) (*db.Customer, error)
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*db.Customer, error)
	Update(ctx context.Context, customer *db.Customer) error
	Delete(ctx context.Context, tenantID, customerID uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filters CustomerFilters, pagination PaginationParams) ([]*db.Customer, string, error)
}

// ProjectRepository interface
type ProjectRepository interface {
	Create(ctx context.Context, project *db.Project) error
	GetByID(ctx context.Context, tenantID, projectID uuid.UUID) (*db.Project, error)
	GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (*db.Project, error)
	Update(ctx context.Context, project *db.Project) error
	Delete(ctx context.Context, tenantID, projectID uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID) ([]*db.Project, error)
	ListForAgent(ctx context.Context, tenantID, agentID uuid.UUID) ([]*db.Project, error)
	// Count returns number of projects for the tenant
	Count(ctx context.Context, tenantID uuid.UUID) (int, error)
}

// TenantRepository interface
type TenantRepository interface {
	Create(ctx context.Context, tenant *db.Tenant) error
	GetByID(ctx context.Context, tenantID uuid.UUID) (*db.Tenant, error)
	Update(ctx context.Context, tenant *db.Tenant) error
	Delete(ctx context.Context, tenantID uuid.UUID) error
	List(ctx context.Context) ([]*db.Tenant, error)
}

// TicketMessageRepository interface
type TicketMessageRepository interface {
	Create(ctx context.Context, message *db.TicketMessage) error
	GetByTicketID(ctx context.Context, ticketID uuid.UUID, includePrivate bool, pagination PaginationParams) ([]*db.TicketMessage, string, error)
	GetByTenantProjectTicketAndMessageID(ctx context.Context, tenantID, projectID, ticketID, messageID uuid.UUID) (*db.TicketMessage, error)
	GetByTenantProjectAndTicketID(ctx context.Context, tenantID, projectID, ticketID uuid.UUID, includePrivate bool, pagination PaginationParams) ([]*db.TicketMessage, string, error)
	Update(ctx context.Context, message *db.TicketMessage) error
	Delete(ctx context.Context, tenantID, projectID, ticketID, messageID uuid.UUID) error
}

// AttachmentRepository interface
type AttachmentRepository interface {
	Create(ctx context.Context, attachment *db.Attachment) error
	GetByID(ctx context.Context, tenantID, projectID, attachmentID uuid.UUID) (*db.Attachment, error)
	ListByTicket(ctx context.Context, tenantID, projectID, ticketID uuid.UUID) ([]*db.Attachment, error)
	ListByMessage(ctx context.Context, tenantID, projectID, messageID uuid.UUID) ([]*db.Attachment, error)
	Delete(ctx context.Context, tenantID, projectID, attachmentID uuid.UUID) error
}

// UnauthTokenRepository interface
type UnauthTokenRepository interface {
	Create(ctx context.Context, token *db.UnauthToken) error
	GetByJTI(ctx context.Context, jti string) (*db.UnauthToken, error)
	MarkConsumed(ctx context.Context, jti string) error
	DeleteExpired(ctx context.Context) error
	RevokeByTicket(ctx context.Context, tenantID, projectID, ticketID uuid.UUID) error
}

// ApiKeyRepository interface
type ApiKeyRepository interface {
	Create(ctx context.Context, apiKey *db.ApiKey) error
	GetByID(ctx context.Context, tenantID uuid.UUID, keyID uuid.UUID) (*db.ApiKey, error)
	GetByHash(ctx context.Context, keyHash string) (*db.ApiKey, error)
	List(ctx context.Context, tenantID uuid.UUID, projectID uuid.UUID) ([]*db.ApiKey, error)
	Update(ctx context.Context, apiKey *db.ApiKey) error
	Delete(ctx context.Context, tenantID uuid.UUID, keyID uuid.UUID) error
	UpdateLastUsed(ctx context.Context, keyID uuid.UUID) error
}

// CreditsRepository interface
type CreditsRepository interface {
	Create(ctx context.Context, credits *db.Credits) error
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*db.Credits, error)
	Update(ctx context.Context, credits *db.Credits) error
	AddCredits(ctx context.Context, tenantID uuid.UUID, amount int64, transactionType, paymentGateway, paymentEventID, description string) (*db.CreditTransaction, error)
	DeductCredits(ctx context.Context, tenantID uuid.UUID, amount int64, transactionType, description string) (*db.CreditTransaction, error)
	CreateTransaction(ctx context.Context, transaction *db.CreditTransaction) error
	GetTransactionsByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*db.CreditTransaction, string, error)
	GetTransactionByPaymentEventID(ctx context.Context, paymentEventID string) (*db.CreditTransaction, error)
}

// PaymentWebhookRepository interface
type PaymentWebhookRepository interface {
	Create(event *db.PaymentWebhookEvent) error
	FindByPaymentEventID(eventID string) (*db.PaymentWebhookEvent, error)
	Update(event *db.PaymentWebhookEvent) error
	FindTenantByEmail(email string) (*db.Tenant, error)
}
