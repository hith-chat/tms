package db

import (
	"time"

	"github.com/bareuptime/tms/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Tenant represents a top-level organization
type Tenant struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Name      string    `db:"name" json:"name" validate:"required,min=1,max=255"`
	Status    string    `db:"status" json:"status" validate:"oneof=active inactive suspended"`
	Region    string    `db:"region" json:"region"`
	KMSKeyID  string    `db:"kms_key_id" json:"kms_key_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Project represents a tenant-scoped workspace
type Project struct {
	ID        uuid.UUID `db:"id" json:"id"`
	TenantID  uuid.UUID `db:"tenant_id" json:"tenant_id"`
	Key       string    `db:"key" json:"key" validate:"required,min=1,max=50"`
	Name      string    `db:"name" json:"name" validate:"required,min=1,max=255"`
	Status    string    `db:"status" json:"status" validate:"oneof=active inactive"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type AgentStatus string

const (
	AgentStatusOnline       AgentStatus = "online"
	AgentStatusAway         AgentStatus = "away"
	AgentStatusBusy         AgentStatus = "busy"
	AgentStatusOffline      AgentStatus = "offline"
	AgentStatusDoNotDisturb AgentStatus = "dnd"
)

// AgentSkill represents skills an agent possesses
type AgentSkill string

const (
	SkillGeneral   AgentSkill = "general"
	SkillTechnical AgentSkill = "technical"
	SkillBilling   AgentSkill = "billing"
	SkillSupport   AgentSkill = "support"
	SkillSales     AgentSkill = "sales"
	SkillComplaint AgentSkill = "complaint"
)

// Agent represents a user who can access the system
type Agent struct {
	ID              uuid.UUID    `db:"id" json:"id"`
	TenantID        uuid.UUID    `db:"tenant_id" json:"tenant_id"`
	Email           string       `db:"email" json:"email" validate:"required,email"`
	Name            string       `db:"name" json:"name" validate:"required,min=1,max=255"`
	Status          AgentStatus  `db:"status" json:"status" validate:"oneof=active inactive suspended"`
	PasswordHash    *string      `db:"password_hash" json:"-"`
	CreatedAt       time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time    `db:"updated_at" json:"updated_at"`
	Skills          []AgentSkill `json:"skills"`
	ActiveChats     int          `json:"active_chats"`
	MaxChats        int          `json:"max_chats"`
	AvgResponseTime float64      `json:"avg_response_time_seconds"`
	LastActivity    time.Time    `json:"last_activity"`
	LastAssignment  time.Time    `json:"last_assignment"`
	Workload        float64      `json:"workload"` // 0.0 to 1.0 representing capacity usage
}

// AgentProjectRole represents role binding between agents and projects
type AgentProjectRole struct {
	AgentID   uuid.UUID       `db:"agent_id" json:"agent_id"`
	TenantID  uuid.UUID       `db:"tenant_id" json:"tenant_id"`
	ProjectID uuid.UUID       `db:"project_id" json:"project_id"`
	Role      models.RoleType `db:"role" json:"role" validate:"required"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt time.Time       `db:"updated_at" json:"updated_at"`
}

// Role represents a system role
type Role struct {
	Role        models.RoleType `db:"role" json:"role" validate:"required"`
	Description string          `db:"description" json:"description"`
}

// RolePermission represents permissions for a role
type RolePermission struct {
	Role models.RoleType `db:"role" json:"role" validate:"required"`
	Perm string          `db:"perm" json:"perm" validate:"required"`
}

// Customer represents an external customer
type Customer struct {
	ID        uuid.UUID         `db:"id" json:"id"`
	TenantID  uuid.UUID         `db:"tenant_id" json:"tenant_id"`
	Email     string            `db:"email" json:"email" validate:"required,email"`
	Name      string            `db:"name" json:"name" validate:"required,min=1,max=255"`
	OrgID     *uuid.UUID        `db:"org_id" json:"org_id,omitempty"`
	Metadata  map[string]string `db:"metadata" json:"metadata,omitempty"`
	CreatedAt time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt time.Time         `db:"updated_at" json:"updated_at"`
}

// Organization represents a customer organization
type Organization struct {
	ID          uuid.UUID `db:"id" json:"id"`
	TenantID    uuid.UUID `db:"tenant_id" json:"tenant_id"`
	Name        string    `db:"name" json:"name" validate:"required,min=1,max=255"`
	ExternalRef *string   `db:"external_ref" json:"external_ref,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// Ticket represents a support ticket
type Ticket struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	TenantID        uuid.UUID  `db:"tenant_id" json:"tenant_id"`
	ProjectID       uuid.UUID  `db:"project_id" json:"project_id"`
	Number          int        `db:"number" json:"number"`
	Subject         string     `db:"subject" json:"subject" validate:"required,min=1,max=500"`
	Status          string     `db:"status" json:"status" validate:"oneof=new open pending resolved closed"`
	Priority        string     `db:"priority" json:"priority" validate:"oneof=low normal high urgent"`
	Type            string     `db:"type" json:"type" validate:"oneof=question incident problem task"`
	Source          string     `db:"source" json:"source" validate:"oneof=web email api phone chat"`
	CustomerID      uuid.UUID  `db:"customer_id" json:"customer_id"`
	AssigneeAgentID *uuid.UUID `db:"assignee_agent_id" json:"assignee_agent_id,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// TicketMessage represents a message on a ticket
type TicketMessage struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	TenantID   uuid.UUID  `db:"tenant_id" json:"tenant_id"`
	ProjectID  uuid.UUID  `db:"project_id" json:"project_id"`
	TicketID   uuid.UUID  `db:"ticket_id" json:"ticket_id"`
	AuthorType string     `db:"author_type" json:"author_type" validate:"oneof=agent customer system ai-agent"`
	AuthorID   *uuid.UUID `db:"author_id" json:"author_id,omitempty"`
	Body       string     `db:"body" json:"body" validate:"required"`
	IsPrivate  bool       `db:"is_private" json:"is_private"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
}

// TicketTag represents a tag on a ticket
type TicketTag struct {
	TicketID  uuid.UUID `db:"ticket_id" json:"ticket_id"`
	TenantID  uuid.UUID `db:"tenant_id" json:"tenant_id"`
	ProjectID uuid.UUID `db:"project_id" json:"project_id"`
	Tag       string    `db:"tag" json:"tag" validate:"required,min=1,max=50"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// Attachment represents a file attachment
type Attachment struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	TenantID    uuid.UUID  `db:"tenant_id" json:"tenant_id"`
	ProjectID   uuid.UUID  `db:"project_id" json:"project_id"`
	TicketID    uuid.UUID  `db:"ticket_id" json:"ticket_id"`
	MessageID   *uuid.UUID `db:"message_id" json:"message_id,omitempty"`
	BlobKey     string     `db:"blob_key" json:"blob_key"`
	Filename    string     `db:"filename" json:"filename" validate:"required,min=1,max=255"`
	ContentType string     `db:"content_type" json:"content_type"`
	SizeBytes   int64      `db:"size_bytes" json:"size_bytes"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
}

// SLAPolicy represents an SLA policy
type SLAPolicy struct {
	ID                   uuid.UUID `db:"id" json:"id"`
	TenantID             uuid.UUID `db:"tenant_id" json:"tenant_id"`
	ProjectID            uuid.UUID `db:"project_id" json:"project_id"`
	Name                 string    `db:"name" json:"name" validate:"required,min=1,max=255"`
	FirstResponseMinutes int       `db:"first_response_minutes" json:"first_response_minutes"`
	ResolutionMinutes    int       `db:"resolution_minutes" json:"resolution_minutes"`
	BusinessHoursRef     *string   `db:"business_hours_ref" json:"business_hours_ref,omitempty"`
	CreatedAt            time.Time `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time `db:"updated_at" json:"updated_at"`
}

// UnauthToken represents a token for unauthenticated access
type UnauthToken struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	TenantID   uuid.UUID  `db:"tenant_id" json:"tenant_id"`
	ProjectID  uuid.UUID  `db:"project_id" json:"project_id"`
	TicketID   uuid.UUID  `db:"ticket_id" json:"ticket_id"`
	JTI        string     `db:"jti" json:"jti"`
	Scope      string     `db:"scope" json:"scope" validate:"oneof=view reply"`
	Exp        time.Time  `db:"exp" json:"exp"`
	ConsumedAt *time.Time `db:"consumed_at" json:"consumed_at,omitempty"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
}

// Webhook represents a webhook configuration
type Webhook struct {
	ID        uuid.UUID `db:"id" json:"id"`
	TenantID  uuid.UUID `db:"tenant_id" json:"tenant_id"`
	ProjectID uuid.UUID `db:"project_id" json:"project_id"`
	URL       string    `db:"url" json:"url" validate:"required,url"`
	Secret    string    `db:"secret" json:"secret"`
	EventMask []string  `db:"event_mask" json:"event_mask"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	TenantID     uuid.UUID  `db:"tenant_id" json:"tenant_id"`
	ProjectID    *uuid.UUID `db:"project_id" json:"project_id,omitempty"`
	ActorType    string     `db:"actor_type" json:"actor_type" validate:"oneof=agent customer system"`
	ActorID      *uuid.UUID `db:"actor_id" json:"actor_id,omitempty"`
	Action       string     `db:"action" json:"action" validate:"required"`
	ResourceType string     `db:"resource_type" json:"resource_type" validate:"required"`
	ResourceID   uuid.UUID  `db:"resource_id" json:"resource_id"`
	Meta         *string    `db:"meta" json:"meta,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
}

// RoleBinding represents an agent's role binding to a project
type RoleBinding struct {
	AgentID   uuid.UUID       `db:"agent_id" json:"agent_id"`
	TenantID  uuid.UUID       `db:"tenant_id" json:"tenant_id"`
	ProjectID *uuid.UUID      `db:"project_id" json:"project_id,omitempty"`
	Role      models.RoleType `db:"role" json:"role"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt time.Time       `db:"updated_at" json:"updated_at"`
}

// ApiKey represents an API key for tenant/project access
type ApiKey struct {
	ID         uuid.UUID      `db:"id" json:"id"`
	TenantID   uuid.UUID      `db:"tenant_id" json:"tenant_id"`
	ProjectID  *uuid.UUID     `db:"project_id" json:"project_id,omitempty"`
	Name       string         `db:"name" json:"name" validate:"required,min=1,max=255"`
	KeyHash    string         `db:"key_hash" json:"-"`             // Never expose the hash
	KeyPrefix  string         `db:"key_prefix" json:"key_preview"` // For display
	Scopes     pq.StringArray `db:"scopes" json:"scopes,omitempty"`
	LastUsedAt *time.Time     `db:"last_used_at" json:"last_used,omitempty"`
	ExpiresAt  *time.Time     `db:"expires_at" json:"expires_at,omitempty"`
	IsActive   bool           `db:"is_active" json:"is_active"`
	CreatedBy  uuid.UUID      `db:"created_by" json:"created_by"`
	CreatedAt  time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time      `db:"updated_at" json:"updated_at"`
}
