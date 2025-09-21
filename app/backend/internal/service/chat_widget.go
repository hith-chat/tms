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
	if req.AgentName == "" {
		req.AgentName = "Support Agent"
	}
	if req.ChatBubbleStyle == "" {
		req.ChatBubbleStyle = "modern"
	}
	if req.WidgetShape == "" {
		req.WidgetShape = "rounded"
	}

	if req.AgentAvatarURL == nil || *req.AgentAvatarURL == "" {
		req.AgentAvatarURL = nil
	}

	widget := &models.ChatWidget{
		ID:               uuid.New(),
		TenantID:         tenantID,
		ProjectID:        projectID,
		Name:             req.AgentName,
		AgentAvatarURL:   req.AgentAvatarURL,
		AgentName:        req.AgentName,
		CustomGreeting:   req.CustomGreeting,
		ChatBubbleStyle:  req.ChatBubbleStyle,
		WidgetShape:      req.WidgetShape,
		UseAI:            req.UseAI,
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

	err := s.chatWidgetRepo.CreateChatWidget(ctx, widget)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat widget: %w", err)
	}

	// Generate embed code
	embedCode := s.generateEmbedCode(widget.ID)
	widget.EmbedCode = &embedCode
	return widget, nil
}

// GetChatWidget gets a chat widget by ID
func (s *ChatWidgetService) GetChatWidget(ctx context.Context, tenantID, projectID, widgetID uuid.UUID) (*models.ChatWidget, error) {
	widget, err := s.chatWidgetRepo.GetChatWidget(ctx, tenantID, projectID, widgetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat widget: %w", err)
	}
	embedCode := s.generateEmbedCode(widget.ID)
	widget.EmbedCode = &embedCode
	return widget, nil
}

func (s *ChatWidgetService) GetChatWidgetById(ctx context.Context, widgetID uuid.UUID) (*models.ChatWidget, error) {
	widget, err := s.chatWidgetRepo.GetChatWidgetById(ctx, widgetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat widget: %w", err)
	}
	if widget == nil {
		return nil, nil
	}
	embedCode := s.generateEmbedCode(widget.ID)
	widget.EmbedCode = &embedCode
	return widget, nil
}

// ListChatWidgets lists all chat widgets for a project
func (s *ChatWidgetService) ListChatWidgets(ctx context.Context, tenantID, projectID uuid.UUID) ([]*models.ChatWidget, error) {
	widgetList, err := s.chatWidgetRepo.ListChatWidgets(ctx, tenantID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat widget: %w", err)
	}
	for _, widget := range widgetList {
		embedCode := s.generateEmbedCode(widget.ID)
		widget.EmbedCode = &embedCode
	}
	return widgetList, nil
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
	if req.AgentName != nil {
		widget.Name = *req.AgentName
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

	if req.AgentAvatarURL != nil && *req.AgentAvatarURL != "" {
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

	// update timestamp to reflect modification
	widget.UpdatedAt = time.Now()

	err = s.chatWidgetRepo.UpdateChatWidget(ctx, widget)
	if err != nil {
		return nil, fmt.Errorf("failed to update chat widget: %w", err)
	}

	// regenerate embed code so returned widget matches Create/Get behavior
	embedCode := s.generateEmbedCode(widget.ID)
	widget.EmbedCode = &embedCode

	return widget, nil
}

// DeleteChatWidget deletes a chat widget
func (s *ChatWidgetService) DeleteChatWidget(ctx context.Context, tenantID, projectID, widgetID uuid.UUID) error {
	return s.chatWidgetRepo.DeleteChatWidget(ctx, tenantID, projectID, widgetID)
}

// generateEmbedCode generates the JavaScript embed code for the chat widget
func (s *ChatWidgetService) generateEmbedCode(widgetID uuid.UUID) string {
	return fmt.Sprintf(`<!-- TMS Chat Widget -->
<script>
  (function() {
    window.TMSChatConfig = {
      widgetId: '%s'
    };
    var script = document.createElement('script');
    script.src = 'https://cdn.jsdelivr.net/npm/@taral/web-chat/dist/chat-widget.js';
    script.async = true;
    document.head.appendChild(script);
  })();
</script>`, widgetID.String())
}

// generateSessionToken generates a secure session token
func generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
