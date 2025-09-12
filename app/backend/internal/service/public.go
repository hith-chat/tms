package service

import (
	"context"
	"fmt"
	"time"

	"github.com/bareuptime/tms/internal/auth"
	"github.com/bareuptime/tms/internal/db"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/google/uuid"
)

// PublicService handles public operations (magic link access, etc.)
type PublicService struct {
	ticketRepo     repo.TicketRepository
	messageRepo    repo.TicketMessageRepository
	jwtAuth        *auth.Service
	messageService *MessageService
}

// NewPublicService creates a new public service
func NewPublicService(
	ticketRepo repo.TicketRepository,
	messageRepo repo.TicketMessageRepository,
	jwtAuth *auth.Service,
	messageService *MessageService,
) *PublicService {
	return &PublicService{
		ticketRepo:     ticketRepo,
		messageRepo:    messageRepo,
		jwtAuth:        jwtAuth,
		messageService: messageService,
	}
}

// AddPublicMessageRequest represents a request to add a public message to a ticket
type AddPublicMessageRequest struct {
	Body string `json:"body" validate:"required"`
}

// GetTicketByMagicLink retrieves a ticket using a magic link token
func (s *PublicService) GetTicketByMagicLink(ctx context.Context, magicToken string) (*db.Ticket, error) {
	// Validate public ticket token
	claims, err := s.jwtAuth.ValidatePublicToken(magicToken)
	if err != nil {
		return nil, fmt.Errorf("invalid magic link: %w", err)
	}

	// Check token type
	if claims.Sub != "public-ticket" {
		return nil, fmt.Errorf("invalid token type")
	}

	// Check expiration
	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("magic link has expired")
	}

	// Extract ticket information from claims
	ticketID := claims.TicketID

	// Get ticket
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}

	return ticket, nil
}

// GetTicketMessagesByMagicLink retrieves public messages for a ticket using a magic link token
func (s *PublicService) GetTicketMessagesByMagicLink(ctx context.Context, magicToken string, cursor string, limit int) ([]*MessageWithDetails, string, error) {
	// Validate public ticket token
	claims, err := s.jwtAuth.ValidatePublicToken(magicToken)
	if err != nil {
		return nil, "", fmt.Errorf("invalid magic link: %w", err)
	}

	// Check token type
	if claims.Sub != "public-ticket" {
		return nil, "", fmt.Errorf("invalid token type")
	}

	// Check expiration
	if time.Now().Unix() > claims.Exp {
		return nil, "", fmt.Errorf("magic link has expired")
	}

	ticketID := claims.TicketID

	// Verify ticket exists
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, "", fmt.Errorf("ticket not found: %w", err)
	}

	pagination := repo.PaginationParams{
		Cursor: cursor,
		Limit:  limit,
	}

	// Get public messages only (includePrivate = false)
	messages, nextCursor, err := s.messageRepo.GetByTicketID(ctx, ticketID, false, pagination)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get messages: %w", err)
	}

	messagesWithDetails, err := s.messageService.UpdateMessageWithDetails(ctx, ticket.TenantID, messages)
	if err != nil {
		return nil, "", fmt.Errorf("failed to update message details: %w", err)
	}

	return messagesWithDetails, nextCursor, nil
}

// AddMessageByMagicLink adds a public message to a ticket using a magic link token
func (s *PublicService) AddMessageByMagicLink(ctx context.Context, magicToken string, req AddPublicMessageRequest) (*db.TicketMessage, error) {
	// Validate public ticket token
	claims, err := s.jwtAuth.ValidatePublicToken(magicToken)
	if err != nil {
		return nil, fmt.Errorf("invalid magic link: %w", err)
	}

	// Check token type
	if claims.Sub != "public-ticket" {
		return nil, fmt.Errorf("invalid token type")
	}

	// Check expiration
	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("magic link has expired")
	}

	// Extract ticket information from claims
	ticketID := claims.TicketID

	// Verify ticket exists and get the requester (customer) ID
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}

	// Use the ticket's requester as the customer ID
	customerID := ticket.CustomerID

	// Create message from customer
	message := &db.TicketMessage{
		ID:         uuid.New(),
		TenantID:   ticket.TenantID,
		ProjectID:  ticket.ProjectID,
		TicketID:   ticketID,
		AuthorType: "customer",
		AuthorID:   &customerID,
		Body:       req.Body,
		IsPrivate:  false, // Customer messages are always public
		CreatedAt:  time.Now(),
	}

	err = s.messageRepo.Create(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return message, nil
}

// GenerateMagicLinkToken generates a magic link token for a ticket
func (s *PublicService) GenerateMagicLinkToken(ticketID, customerID uuid.UUID) (string, error) {

	// For public ticket access, we don't need to include customer ID in the token
	// The customer ownership is verified when the ticket is accessed
	scope := []string{"read", "write"}
	return s.jwtAuth.GeneratePublicToken(ticketID, customerID, scope)
}

// GetTicketByID retrieves a ticket using ticket ID
func (s *PublicService) GetTicketByID(ctx context.Context, ticketID uuid.UUID) (*db.Ticket, error) {
	// Get ticket directly by ID
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}

	return ticket, nil
}

// GetTicketMessagesByID retrieves public messages for a ticket using ticket ID
func (s *PublicService) GetTicketMessagesByID(ctx context.Context, ticketID uuid.UUID, cursor string, limit int) ([]*MessageWithDetails, string, error) {
	// Verify ticket exists
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, "", fmt.Errorf("ticket not found: %w", err)
	}

	pagination := repo.PaginationParams{
		Cursor: cursor,
		Limit:  limit,
	}

	// Get public messages only (includePrivate = false)
	messages, nextCursor, err := s.messageRepo.GetByTicketID(ctx, ticketID, false, pagination)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get messages: %w", err)
	}

	messagesWithDetails, err := s.messageService.UpdateMessageWithDetails(ctx, ticket.TenantID, messages)
	if err != nil {
		return nil, "", fmt.Errorf("failed to update message details: %w", err)
	}

	return messagesWithDetails, nextCursor, nil
}

// AddMessageByID adds a public message to a ticket using ticket ID
func (s *PublicService) AddMessageByID(ctx context.Context, ticketID uuid.UUID, req AddPublicMessageRequest) (*db.TicketMessage, error) {
	// Verify ticket exists and get the requester (customer) ID
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}

	// Use the ticket's requester as the customer ID
	customerID := ticket.CustomerID

	// Create message from customer
	message := &db.TicketMessage{
		ID:         uuid.New(),
		TenantID:   ticket.TenantID,
		ProjectID:  ticket.ProjectID,
		TicketID:   ticketID,
		AuthorType: "customer",
		AuthorID:   &customerID,
		Body:       req.Body,
		IsPrivate:  false, // Customer messages are always public
		CreatedAt:  time.Now(),
	}

	err = s.messageRepo.Create(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return message, nil
}
