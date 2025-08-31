package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/db"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
)

type ChatSessionService struct {
	chatSessionRepo *repo.ChatSessionRepo
	chatMessageRepo *repo.ChatMessageRepo
	chatWidgetRepo  *repo.ChatWidgetRepo
	customerRepo    repo.CustomerRepository
	ticketService   *TicketService
	agentService    *AgentService
}

func NewChatSessionService(
	chatSessionRepo *repo.ChatSessionRepo,
	chatMessageRepo *repo.ChatMessageRepo,
	chatWidgetRepo *repo.ChatWidgetRepo,
	customerRepo repo.CustomerRepository,
	ticketService *TicketService,
	agentService *AgentService,
) *ChatSessionService {
	return &ChatSessionService{
		chatSessionRepo: chatSessionRepo,
		chatMessageRepo: chatMessageRepo,
		chatWidgetRepo:  chatWidgetRepo,
		customerRepo:    customerRepo,
		ticketService:   ticketService,
		agentService:    agentService,
	}
}

// InitiateChat starts a new chat session
func (s *ChatSessionService) InitiateChat(ctx context.Context, widgetID uuid.UUID, clientSessionID string, req *models.InitiateChatRequest) (*models.ChatSession, error) {
	// Get widget to validate and get tenant/project context
	widget, err := s.chatWidgetRepo.GetChatWidgetById(ctx, widgetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get widget: %w", err)
	}
	if widget == nil || !widget.IsActive {
		return nil, fmt.Errorf("widget not found or inactive")
	}

	// Find or create customer if email provided
	var customerID *uuid.UUID
	if req.VisitorEmail != "" {
		customer, err := s.customerRepo.GetByEmail(ctx, widget.TenantID, req.VisitorEmail)
		if err != nil {
			return nil, fmt.Errorf("failed to get customer: %w", err)
		}
		if customer == nil {
			// Create new customer
			newCustomer := &db.Customer{
				ID:       uuid.New(),
				TenantID: widget.TenantID,
				Email:    req.VisitorEmail,
				Name:     req.VisitorName,
			}
			err = s.customerRepo.Create(ctx, newCustomer)
			if err != nil {
				return nil, fmt.Errorf("failed to create customer: %w", err)
			}
			customerID = &newCustomer.ID
		} else {
			customerID = &customer.ID
		}
	}

	// Create chat session
	session := &models.ChatSession{
		ID:              uuid.New(),
		TenantID:        widget.TenantID,
		ProjectID:       widget.ProjectID,
		WidgetID:        widget.ID,
		CustomerID:      customerID,
		ClientSessionID: clientSessionID,
		Status:          "active",
		VisitorInfo:     req.VisitorInfo,
		StartedAt:       time.Now(),
		LastActivityAt:  time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		UseAI:           widget.UseAI,
	}

	err = s.chatSessionRepo.CreateChatSession(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat session: %w", err)
	}

	// Send initial message if provided
	if req.InitialMessage != "" {
		_, err = s.SendMessage(ctx, widget.TenantID, widget.ProjectID, session.ID, &models.SendChatMessageRequest{
			Content: req.InitialMessage,
		}, "visitor", nil, req.VisitorName)
		if err != nil {
			// Log error but don't fail session creation
			fmt.Printf("Failed to send initial message: %v\n", err)
		}
	}

	return session, nil
}

// GetChatSession gets a chat session by ID
func (s *ChatSessionService) GetChatSession(ctx context.Context, tenantID, projectID, sessionID uuid.UUID) (*models.ChatSession, error) {
	return s.chatSessionRepo.GetChatSession(ctx, tenantID, projectID, sessionID)
}

// GetChatSessionByID gets a chat session by ID for any tenant (used for global agent operations)
func (s *ChatSessionService) GetChatSessionByID(ctx context.Context, tenantID, sessionID uuid.UUID) (*models.ChatSession, error) {
	return s.chatSessionRepo.GetChatSessionByID(ctx, tenantID, sessionID)
}

func (s *ChatSessionService) GetChatSessionByClientSessionID(ctx context.Context, clientSessionID string) (*models.ChatSession, error) {
	return s.chatSessionRepo.GetChatSessionByClientSessionID(ctx, clientSessionID)
}

// ListChatSessions lists chat sessions for a project
func (s *ChatSessionService) ListChatSessions(ctx context.Context, tenantID, projectID uuid.UUID, filters repo.ChatSessionFilters) ([]*models.ChatSession, error) {
	return s.chatSessionRepo.ListChatSessions(ctx, tenantID, projectID, filters)
}

// AssignAgent assigns an agent to a chat session
func (s *ChatSessionService) AssignAgent(ctx context.Context, tenantID, projectID, sessionID, agentID uuid.UUID) error {
	session, err := s.chatSessionRepo.GetChatSession(ctx, tenantID, projectID, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return fmt.Errorf("session not found")
	}

	now := time.Now()
	session.AssignedAgentID = &agentID
	session.AssignedAt = &now

	err = s.chatSessionRepo.UpdateChatSession(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to assign agent: %w", err)
	}

	// Fetch agent details for the system message
	agent, agentErr := s.agentService.GetAgent(ctx, tenantID, agentID)
	agentName := "Agent"
	if agentErr == nil && agent != nil && agent.Name != "" {
		agentName = agent.Name
	}

	// Send system message about agent assignment
	_, err = s.SendMessage(ctx, tenantID, projectID, sessionID, &models.SendChatMessageRequest{
		Content: fmt.Sprintf("Our agent %s has joined the conversation", agentName),
	}, "system", nil, "System")

	return err
}

// EndSession ends a chat session
func (s *ChatSessionService) EndSession(ctx context.Context, tenantID, projectID, sessionID uuid.UUID) error {
	session, err := s.chatSessionRepo.GetChatSession(ctx, tenantID, projectID, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return fmt.Errorf("session not found")
	}

	now := time.Now()
	session.Status = "ended"
	session.EndedAt = &now

	return s.chatSessionRepo.UpdateChatSession(ctx, session)
}

// SendMessage sends a message in a chat session
func (s *ChatSessionService) SendMessage(ctx context.Context, tenantID, projectID, sessionID uuid.UUID, req *models.SendChatMessageRequest, authorType string, authorID *uuid.UUID, authorName string) (*models.ChatMessage, error) {
	// Set defaults
	if req.MessageType == "" {
		req.MessageType = "text"
	}

	message := &models.ChatMessage{
		ID:            uuid.New(),
		TenantID:      tenantID,
		ProjectID:     projectID,
		SessionID:     sessionID,
		MessageType:   req.MessageType,
		Content:       req.Content,
		AuthorType:    authorType,
		AuthorID:      authorID,
		AuthorName:    authorName,
		Metadata:      req.Metadata,
		IsPrivate:     req.IsPrivate,
		ReadByVisitor: authorType == "visitor", // Auto-mark as read by sender
		ReadByAgent:   authorType == "agent",   // Auto-mark as read by sender
		CreatedAt:     time.Now(),
	}

	if message.Metadata == nil {
		message.Metadata = make(models.JSONMap)
	}

	err := s.chatMessageRepo.CreateChatMessage(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Update session last activity
	err = s.chatSessionRepo.UpdateLastActivity(ctx, sessionID)
	if err != nil {
		// Log error but don't fail message creation
		fmt.Printf("Failed to update session last activity: %v\n", err)
	}

	return message, nil
}

// SendMessageByID sends a message in a chat session using sessionID (for contexts where session object isn't available)
func (s *ChatSessionService) SendMessageByID(ctx context.Context, sessionID uuid.UUID, req *models.SendChatMessageRequest, authorType string, authorID *uuid.UUID, authorName string) (*models.ChatMessage, error) {
	// This is a temporary workaround for callers that only have sessionID
	// For now, we'll create the message directly without validating the session exists
	// This is not ideal but allows the system to work while we fix the architecture

	// Set defaults
	if req.MessageType == "" {
		req.MessageType = "text"
	}

	message := &models.ChatMessage{
		ID: uuid.New(),
		// Note: TenantID and ProjectID will be nil here, which may cause issues
		// This is a temporary fix and should be addressed properly
		SessionID:     sessionID,
		MessageType:   req.MessageType,
		Content:       req.Content,
		AuthorType:    authorType,
		AuthorID:      authorID,
		AuthorName:    authorName,
		Metadata:      req.Metadata,
		IsPrivate:     req.IsPrivate,
		ReadByVisitor: authorType == "visitor", // Auto-mark as read by sender
		ReadByAgent:   authorType == "agent",   // Auto-mark as read by sender
		CreatedAt:     time.Now(),
	}

	if message.Metadata == nil {
		message.Metadata = make(models.JSONMap)
	}

	err := s.chatMessageRepo.CreateChatMessage(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Update session last activity
	err = s.chatSessionRepo.UpdateLastActivity(ctx, sessionID)
	if err != nil {
		// Log error but don't fail message creation
		fmt.Printf("Failed to update session last activity: %v\n", err)
	}

	return message, nil
}

// GetChatMessages gets messages for a chat session
func (s *ChatSessionService) GetChatMessages(ctx context.Context, tenantID, projectID, sessionID uuid.UUID, includePrivate bool) ([]*models.ChatMessage, error) {
	return s.chatMessageRepo.ListChatMessages(ctx, tenantID, projectID, sessionID, includePrivate)
}

// GetChatMessagesForSession gets messages for a session (public access)
func (s *ChatSessionService) GetChatMessagesForSession(ctx context.Context, sessionID uuid.UUID) ([]*models.ChatMessage, error) {
	return s.chatMessageRepo.ListChatMessagesForSession(ctx, sessionID)
}

// MarkMessagesAsRead marks messages as read
func (s *ChatSessionService) MarkAgentMessagesAsRead(ctx context.Context, tenantID, projectID, sessionID, messageID uuid.UUID, readerType string) error {
	return s.chatMessageRepo.MarkAgentMessagesAsRead(ctx, tenantID, projectID, sessionID, messageID, readerType)
}

// MarkVisitorMessagesAsRead marks messages as read unauthenticated
func (s *ChatSessionService) MarkVisitorMessagesAsRead(ctx context.Context, sessionID, messageID uuid.UUID, readerType string) error {
	if messageID == uuid.Nil {
		return fmt.Errorf("invalid message ID")
	}
	return s.chatMessageRepo.MarkVisitorMessagesAsRead(ctx, sessionID, messageID, readerType)
}

// CreateTicketFromChat creates a ticket from a chat session
func (s *ChatSessionService) CreateTicketFromChat(ctx context.Context, tenantID, projectID, sessionID, agentID uuid.UUID, subject string) (*db.Ticket, error) {
	session, err := s.chatSessionRepo.GetChatSession(ctx, tenantID, projectID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	if session.TicketID != nil {
		return nil, fmt.Errorf("session already has an associated ticket")
	}

	// Get messages to create ticket body
	messages, err := s.chatMessageRepo.ListChatMessages(ctx, tenantID, projectID, sessionID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// Build ticket body from chat history
	body := "Chat conversation:\n\n"
	for _, msg := range messages {
		if !msg.IsPrivate {
			body += fmt.Sprintf("[%s] %s: %s\n", msg.CreatedAt.Format("15:04"), msg.AuthorName, msg.Content)
		}
	}

	// Create ticket request
	requesterEmail := ""
	if session.CustomerEmail != nil {
		requesterEmail = *session.CustomerEmail
	}

	requesterName := "Chat Visitor"
	if session.CustomerName != nil && *session.CustomerName != "" {
		requesterName = *session.CustomerName
	}

	ticketReq := CreateTicketRequest{
		Subject:        subject,
		Priority:       "normal",
		Type:           "question",
		Source:         "chat",
		RequesterEmail: requesterEmail,
		RequesterName:  requesterName,
		InitialMessage: body,
	}

	if session.AssignedAgentID != nil {
		agentIDStr := session.AssignedAgentID.String()
		ticketReq.AssigneeAgentID = &agentIDStr
	}

	// Create ticket
	ticket, err := s.ticketService.CreateTicket(ctx, tenantID, projectID, agentID, ticketReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	// Update session with ticket ID
	session.TicketID = &ticket.ID
	err = s.chatSessionRepo.UpdateChatSession(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to update session with ticket ID: %w", err)
	}

	return ticket, nil
}
