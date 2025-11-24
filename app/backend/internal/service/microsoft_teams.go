package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/logger"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
)

// MicrosoftTeamsService handles Microsoft Teams API interactions
type MicrosoftTeamsService struct {
	integrationRepo *repo.ProjectIntegrationRepository
	httpClient      *http.Client
}

// NewMicrosoftTeamsService creates a new Microsoft Teams service
func NewMicrosoftTeamsService(
	integrationRepo *repo.ProjectIntegrationRepository,
) *MicrosoftTeamsService {
	return &MicrosoftTeamsService{
		integrationRepo: integrationRepo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// TeamsMessageCard represents a simple Microsoft Teams MessageCard (legacy but widely supported)
// For more complex layouts, Adaptive Cards are recommended, but MessageCard is easier for simple text.
type TeamsMessageCard struct {
	Type       string `json:"@type"`
	Context    string `json:"@context"`
	ThemeColor string `json:"themeColor,omitempty"`
	Summary    string `json:"summary,omitempty"`
	Title      string `json:"title,omitempty"`
	Text       string `json:"text,omitempty"`
	Sections   []TeamsSection `json:"sections,omitempty"`
}

type TeamsSection struct {
	ActivityTitle    string `json:"activityTitle,omitempty"`
	ActivitySubtitle string `json:"activitySubtitle,omitempty"`
	ActivityImage    string `json:"activityImage,omitempty"`
	Text             string `json:"text,omitempty"`
	Facts            []TeamsFact `json:"facts,omitempty"`
}

type TeamsFact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// GetTeamsIntegration retrieves the Microsoft Teams integration for a project
func (s *MicrosoftTeamsService) GetTeamsIntegration(ctx context.Context, tenantID, projectID uuid.UUID) (*models.ProjectIntegration, error) {
	integration, err := s.integrationRepo.GetByProjectAndType(ctx, tenantID, projectID, models.ProjectIntegrationTypeTeams)
	if err != nil {
		return nil, fmt.Errorf("failed to get Microsoft Teams integration: %w", err)
	}
	if integration == nil {
		return nil, nil // No Teams integration configured
	}

	if integration.Status != models.ProjectIntegrationStatusActive {
		return nil, fmt.Errorf("Microsoft Teams integration is not active")
	}

	return integration, nil
}

// PostMessageToTeams posts a message to Microsoft Teams using Incoming Webhook
func (s *MicrosoftTeamsService) PostMessageToTeams(
	ctx context.Context,
	tenantID, projectID uuid.UUID,
	session *models.ChatSession,
	messageContent string,
	senderName string,
) error {
	// Get Teams integration for this project
	integration, err := s.GetTeamsIntegration(ctx, tenantID, projectID)
	if err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to get Microsoft Teams integration")
		return err
	}
	if integration == nil {
		// No Teams integration configured, skip
		return nil
	}

	// Get the webhook URL from meta
	webhookURL := ""
	if url, ok := integration.Meta["webhook_url"].(string); ok && url != "" {
		webhookURL = url
	}

	if webhookURL == "" {
		return fmt.Errorf("no Microsoft Teams webhook URL configured")
	}

	// Prepare the message card
	card := TeamsMessageCard{
		Type:       "MessageCard",
		Context:    "http://schema.org/extensions",
		ThemeColor: "0076D7",
		Summary:    "New Chat Message",
		Sections: []TeamsSection{
			{
				ActivityTitle:    senderName,
				ActivitySubtitle: fmt.Sprintf("Session: %s", session.ID),
				Text:             messageContent,
			},
		},
	}

	// Marshal payload
	jsonData, err := json.Marshal(card)
	if err != nil {
		return fmt.Errorf("failed to marshal Teams message: %w", err)
	}

	// Send request
	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message to Teams: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("teams API returned status: %d", resp.StatusCode)
	}

	return nil
}
