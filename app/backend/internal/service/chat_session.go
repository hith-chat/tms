package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/db"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/bareuptime/tms/internal/websocket"
)

type ChatSessionService struct {
	chatSessionRepo   *repo.ChatSessionRepo
	chatMessageRepo   *repo.ChatMessageRepo
	chatWidgetRepo    *repo.ChatWidgetRepo
	customerRepo      repo.CustomerRepository
	ticketService     *TicketService
	agentService      *AgentService
	connectionManager *websocket.ConnectionManager
}

func NewChatSessionService(
	chatSessionRepo *repo.ChatSessionRepo,
	chatMessageRepo *repo.ChatMessageRepo,
	chatWidgetRepo *repo.ChatWidgetRepo,
	customerRepo repo.CustomerRepository,
	ticketService *TicketService,
	agentService *AgentService,
	connectionManager *websocket.ConnectionManager,
) *ChatSessionService {
	return &ChatSessionService{
		chatSessionRepo:   chatSessionRepo,
		chatMessageRepo:   chatMessageRepo,
		chatWidgetRepo:    chatWidgetRepo,
		customerRepo:      customerRepo,
		ticketService:     ticketService,
		agentService:      agentService,
		connectionManager: connectionManager,
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
		_, err = s.SendMessage(ctx, session, &models.SendChatMessageRequest{
			Content: req.InitialMessage,
		}, "visitor", nil, req.VisitorName, "")
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
	return s.AssignAgentWithSessionObj(ctx, tenantID, projectID, agentID, session)
}

func (s *ChatSessionService) AssignAgentWithSessionObj(ctx context.Context, tenantID, projectID, agentID uuid.UUID, session *models.ChatSession) error {

	now := time.Now()
	session.AssignedAgentID = &agentID
	session.AssignedAt = &now

	err := s.chatSessionRepo.UpdateChatSession(ctx, session)
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
	_, err = s.SendMessage(ctx, session, &models.SendChatMessageRequest{
		Content: fmt.Sprintf("Our agent %s has joined the conversation", agentName),
	}, "system", nil, "System", "")

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
func (s *ChatSessionService) SendMessage(ctx context.Context, session *models.ChatSession, req *models.SendChatMessageRequest, authorType string, authorID *uuid.UUID, authorName, connID string) (*models.ChatMessage, error) {
	// Set defaults
	if req.MessageType == "" {
		req.MessageType = "text"
	}

	message, err := s.SendMessageWithUuidDetails(ctx, session.TenantID, session.ProjectID, session.ID, req, authorType, authorID, authorName, connID)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// SendMessage sends a message in a chat session
func (s *ChatSessionService) SendMessageWithUuidDetails(ctx context.Context, tenantID, projectID, sessionID uuid.UUID, req *models.SendChatMessageRequest, authorType string, authorID *uuid.UUID, authorName, connID string) (*models.ChatMessage, error) {
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

	// Avoid dereferencing authorID when it's nil. Use uuid.Nil as a sentinel
	// for "no author ID" (e.g., system or visitor messages).
	var assignedAgentID uuid.UUID
	if authorID != nil {
		assignedAgentID = *authorID
	} else {
		assignedAgentID = uuid.Nil
	}

	go s.broadcastChatMessage(tenantID, projectID, assignedAgentID, sessionID, message, authorType, connID)

	go s.chatMessageRepo.CreateChatMessage(ctx, message)

	// Update session last activity
	go s.chatSessionRepo.UpdateLastActivity(ctx, sessionID)

	return message, nil
}

// broadcastChatMessage builds and delivers a websocket.Message for a chat message.
func (s *ChatSessionService) broadcastChatMessage(tenantID, projectID, assignedAgentID, sessionID uuid.UUID, message *models.ChatMessage, authorType, connID string) {
	if s.connectionManager == nil {
		return
	}

	messageData, _ := json.Marshal(message)

	// Determine the FromType based on authorType
	var fromType websocket.ConnectionType
	switch authorType {
	case "visitor":
		fromType = websocket.ConnectionTypeVisitor
	case "agent":
		fromType = websocket.ConnectionTypeAgent
	case "ai-agent":
		fromType = websocket.ConnectionTypeAiAgent
	default:
		fromType = websocket.ConnectionTypeVisitor // Default to visitor for system messages
	}

	broadcastMsg := &websocket.Message{
		Type:         "chat_message",
		SessionID:    sessionID,
		Data:         messageData,
		FromType:     fromType,
		ProjectID:    &projectID,
		TenantID:     &tenantID,
		AgentID:      &assignedAgentID,
		DeliveryType: websocket.Direct,
		Timestamp:    time.Now(),
	}

	// Deliver the message to all session connections
	switch fromType {
	case websocket.ConnectionTypeVisitor, websocket.ConnectionTypeAgent:
		go s.connectionManager.DeliverWebSocketMessage(sessionID, broadcastMsg)
	case websocket.ConnectionTypeAiAgent:
		fmt.Println("AI Agent message - sending to connection:", connID)
		s.connectionManager.SendToConnection(connID, broadcastMsg)
		broadcastMsg.FromType = websocket.ConnectionTypeVisitor
		go s.connectionManager.DeliverWebSocketMessage(sessionID, broadcastMsg)
	}

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
