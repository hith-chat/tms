package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
)

type ChatWidgetService struct {
	chatWidgetRepo *repo.ChatWidgetRepo
	domainRepo     *repo.DomainValidationRepo
}

func NewChatWidgetService(chatWidgetRepo *repo.ChatWidgetRepo, domainRepo *repo.DomainValidationRepo) *ChatWidgetService {
	return &ChatWidgetService{
		chatWidgetRepo: chatWidgetRepo,
		domainRepo:     domainRepo,
	}
}

// CreateChatWidget creates a new chat widget
func (s *ChatWidgetService) CreateChatWidget(ctx context.Context, tenantID, projectID uuid.UUID, req *models.CreateChatWidgetRequest) (*models.ChatWidget, error) {
	// Verify domain exists and is verified
	domain, err := s.domainRepo.GetDomainByID(ctx, tenantID, req.DomainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}
	if domain == nil {
		return nil, fmt.Errorf("domain not found")
	}
	if domain.Status != models.DomainValidationStatusVerified {
		return nil, fmt.Errorf("domain must be verified before creating chat widget")
	}

	// Set defaults
	if req.PrimaryColor == "" {
		req.PrimaryColor = "#2563eb"
	}
	if req.SecondaryColor == "" {
		req.SecondaryColor = "#f3f4f6"
	}
	if req.BackgroundColor == "" {
		req.BackgroundColor = "#ffffff"
	}
	if req.Position == "" {
		req.Position = "bottom-right"
	}
	if req.WelcomeMessage == "" {
		req.WelcomeMessage = "Hello! How can we help you?"
	}
	if req.OfflineMessage == "" {
		req.OfflineMessage = "We are currently offline. Please leave a message."
	}
	if req.BusinessHours == nil {
		req.BusinessHours = models.JSONMap{"enabled": false}
	}

	widget := &models.ChatWidget{
		ID:               uuid.New(),
		TenantID:         tenantID,
		ProjectID:        projectID,
		DomainID:         req.DomainID,
		Name:             req.Name,
		IsActive:         true,
		PrimaryColor:     req.PrimaryColor,
		SecondaryColor:   req.SecondaryColor,
		BackgroundColor:  req.BackgroundColor,
		Position:         req.Position,
		WelcomeMessage:   req.WelcomeMessage,
		OfflineMessage:   req.OfflineMessage,
		AutoOpenDelay:    req.AutoOpenDelay,
		ShowAgentAvatars: req.ShowAgentAvatars,
		AllowFileUploads: req.AllowFileUploads,
		RequireEmail:     req.RequireEmail,
		BusinessHours:    req.BusinessHours,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Generate embed code
	embedCode := s.generateEmbedCode(widget.ID, domain.Domain)
	widget.EmbedCode = &embedCode

	err = s.chatWidgetRepo.CreateChatWidget(ctx, widget)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat widget: %w", err)
	}

	return widget, nil
}

// GetChatWidget gets a chat widget by ID
func (s *ChatWidgetService) GetChatWidget(ctx context.Context, tenantID, projectID, widgetID uuid.UUID) (*models.ChatWidget, error) {
	return s.chatWidgetRepo.GetChatWidget(ctx, tenantID, projectID, widgetID)
}

// GetChatWidgetByDomain gets a chat widget by domain (for public access)
func (s *ChatWidgetService) GetChatWidgetByDomain(ctx context.Context, domain string) (*models.ChatWidget, error) {
	return s.chatWidgetRepo.GetChatWidgetByDomain(ctx, domain)
}

// ListChatWidgets lists all chat widgets for a project
func (s *ChatWidgetService) ListChatWidgets(ctx context.Context, tenantID, projectID uuid.UUID) ([]*models.ChatWidget, error) {
	return s.chatWidgetRepo.ListChatWidgets(ctx, tenantID, projectID)
}

// UpdateChatWidget updates a chat widget
func (s *ChatWidgetService) UpdateChatWidget(ctx context.Context, tenantID, projectID, widgetID uuid.UUID, req *models.UpdateChatWidgetRequest) (*models.ChatWidget, error) {
	widget, err := s.chatWidgetRepo.GetChatWidget(ctx, tenantID, projectID, widgetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat widget: %w", err)
	}
	if widget == nil {
		return nil, fmt.Errorf("chat widget not found")
	}

	// Update fields
	if req.Name != nil {
		widget.Name = *req.Name
	}
	if req.IsActive != nil {
		widget.IsActive = *req.IsActive
	}
	if req.PrimaryColor != nil {
		widget.PrimaryColor = *req.PrimaryColor
	}
	if req.SecondaryColor != nil {
		widget.SecondaryColor = *req.SecondaryColor
	}
	if req.BackgroundColor != nil {
		widget.BackgroundColor = *req.BackgroundColor
	}
	if req.Position != nil {
		widget.Position = *req.Position
	}
	if req.WelcomeMessage != nil {
		widget.WelcomeMessage = *req.WelcomeMessage
	}
	if req.OfflineMessage != nil {
		widget.OfflineMessage = *req.OfflineMessage
	}
	if req.AutoOpenDelay != nil {
		widget.AutoOpenDelay = *req.AutoOpenDelay
	}
	if req.ShowAgentAvatars != nil {
		widget.ShowAgentAvatars = *req.ShowAgentAvatars
	}

	if req.RequireEmail != nil {
		widget.RequireEmail = *req.RequireEmail
	}

	if req.RequireName != nil {
		widget.RequireName = *req.RequireName
	}

	if req.AgentAvatarURL != nil {
		widget.AgentAvatarURL = req.AgentAvatarURL
	}

	if req.AgentName != nil {
		widget.AgentName = *req.AgentName
	}

	if req.CustomGreeting != nil {
		widget.CustomGreeting = req.CustomGreeting
	}

	if req.AllowFileUploads != nil {
		widget.AllowFileUploads = *req.AllowFileUploads
	}
	if req.RequireEmail != nil {
		widget.RequireEmail = *req.RequireEmail
	}
	if req.BusinessHours != nil {
		widget.BusinessHours = *req.BusinessHours
	}
	if req.ChatBubbleStyle != nil {
		widget.ChatBubbleStyle = *req.ChatBubbleStyle
	}
	if req.WidgetShape != nil {
		widget.WidgetShape = *req.WidgetShape
	}
	if req.UseAI != nil {
		widget.UseAI = *req.UseAI
	}

	err = s.chatWidgetRepo.UpdateChatWidget(ctx, widget)
	if err != nil {
		return nil, fmt.Errorf("failed to update chat widget: %w", err)
	}

	return widget, nil
}

// DeleteChatWidget deletes a chat widget
func (s *ChatWidgetService) DeleteChatWidget(ctx context.Context, tenantID, projectID, widgetID uuid.UUID) error {
	return s.chatWidgetRepo.DeleteChatWidget(ctx, tenantID, projectID, widgetID)
}

// generateEmbedCode generates the JavaScript embed code for the chat widget
func (s *ChatWidgetService) generateEmbedCode(widgetID uuid.UUID, domain string) string {
	return fmt.Sprintf(`<!-- TMS Chat Widget -->
<script>
  (function() {
    window.TMSChatConfig = {
      widgetId: '%s',
      domain: '%s'
    };
    var script = document.createElement('script');
    script.src = 'https://cdn.jsdelivr.net/npm/@taral/web-chat@1.0.0/dist/chat-widget.js';
    script.async = true;
    document.head.appendChild(script);
  })();
</script>`, widgetID.String(), domain)
}

// generateSessionToken generates a secure session token
func generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
