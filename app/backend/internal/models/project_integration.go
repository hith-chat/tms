package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ProjectIntegrationStatus represents the status of a project integration
type ProjectIntegrationStatus string

const (
	ProjectIntegrationStatusActive   ProjectIntegrationStatus = "active"
	ProjectIntegrationStatusInactive ProjectIntegrationStatus = "inactive"
	ProjectIntegrationStatusError    ProjectIntegrationStatus = "error"
)

// ProjectIntegrationType represents the type of integration
type ProjectIntegrationType string

const (
	ProjectIntegrationTypeSlack   ProjectIntegrationType = "slack"
	ProjectIntegrationTypeDiscord ProjectIntegrationType = "discord"
	ProjectIntegrationTypeTeams   ProjectIntegrationType = "microsoft_teams"
)

// ProjectIntegration represents a simplified integration for a project
type ProjectIntegration struct {
	ID              uuid.UUID                `json:"id" db:"id"`
	TenantID        uuid.UUID                `json:"tenant_id" db:"tenant_id"`
	ProjectID       uuid.UUID                `json:"project_id" db:"project_id"`
	IntegrationType ProjectIntegrationType   `json:"integration_type" db:"integration_type"`
	Meta            IntegrationMeta          `json:"meta" db:"meta"`
	Status          ProjectIntegrationStatus `json:"status" db:"status"`
	CreatedAt       time.Time                `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time                `json:"updated_at" db:"updated_at"`
}

// IntegrationMeta is a flexible JSONB field for storing integration-specific data
type IntegrationMeta map[string]interface{}

// Value implements the driver.Valuer interface for database storage
func (m IntegrationMeta) Value() (driver.Value, error) {
	if m == nil {
		return json.Marshal(map[string]interface{}{})
	}
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for database retrieval
func (m *IntegrationMeta) Scan(value interface{}) error {
	if value == nil {
		*m = make(IntegrationMeta)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, m)
}

// SlackWebhookConfig represents a Slack incoming webhook configuration
type SlackWebhookConfig struct {
	URL              string `json:"url"`
	Channel          string `json:"channel"`
	ChannelID        string `json:"channel_id"`
	ConfigurationURL string `json:"configuration_url,omitempty"`
}

// SlackIntegrationMeta represents the metadata stored for a Slack integration
type SlackIntegrationMeta struct {
	// OAuth tokens
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`

	// Bot info
	BotUserID string `json:"bot_user_id"`
	AppID     string `json:"app_id"`
	Scope     string `json:"scope"`

	// Team info
	TeamID   string `json:"team_id"`
	TeamName string `json:"team_name"`

	// User who authorized
	AuthedUserID string `json:"authed_user_id"`

	// Webhooks
	Webhooks []SlackWebhookConfig `json:"webhooks,omitempty"`

	// Installation metadata
	InstalledByAgentID string    `json:"installed_by_agent_id,omitempty"`
	InstalledAt        time.Time `json:"installed_at"`
	LastUpdatedAt      time.Time `json:"last_updated_at"`
}

// ToMeta converts SlackIntegrationMeta to IntegrationMeta
func (s *SlackIntegrationMeta) ToMeta() (IntegrationMeta, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	var meta IntegrationMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return meta, nil
}

// SlackMetaFromIntegration extracts SlackIntegrationMeta from IntegrationMeta
func SlackMetaFromIntegration(meta IntegrationMeta) (*SlackIntegrationMeta, error) {
	data, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}

	var slackMeta SlackIntegrationMeta
	if err := json.Unmarshal(data, &slackMeta); err != nil {
		return nil, err
	}

	return &slackMeta, nil
}

// OAuthStateData represents the data stored in Redis for OAuth state validation
type OAuthStateData struct {
	TenantID        uuid.UUID              `json:"tenant_id"`
	ProjectID       uuid.UUID              `json:"project_id"`
	IntegrationType ProjectIntegrationType `json:"integration_type"`
	AgentID         uuid.UUID              `json:"agent_id"`
	CreatedAt       time.Time              `json:"created_at"`
}
