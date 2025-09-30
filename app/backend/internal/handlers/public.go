package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bareuptime/tms/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// PublicHandler handles public endpoints (magic link access)
type PublicHandler struct {
	publicService *service.PublicService
	validator     *validator.Validate
}

// PublicMessage is the external representation of a ticket message returned
// by public endpoints. It intentionally omits TenantID and ProjectID which
// must not be exposed to public clients.
type PublicMessage struct {
	ID              uuid.UUID                `json:"id"`
	TicketID        uuid.UUID                `json:"ticket_id"`
	AuthorType      string                   `json:"author_type"`
	AuthorID        *uuid.UUID               `json:"author_id,omitempty"`
	Body            string                   `json:"body"`
	IsPrivate       bool                     `json:"is_private"`
	CreatedAt       time.Time                `json:"created_at"`
	MessageUserInfo *service.MessageUserInfo `json:"user_info,omitempty"`
}

func toPublicMessages(msgs []*service.MessageWithDetails) []PublicMessage {
	out := make([]PublicMessage, 0, len(msgs))
	for _, m := range msgs {
		if m == nil {
			continue
		}
		out = append(out, PublicMessage{
			ID:              m.ID,
			TicketID:        m.TicketID,
			AuthorType:      m.AuthorType,
			AuthorID:        m.AuthorID,
			Body:            m.Body,
			IsPrivate:       m.IsPrivate,
			CreatedAt:       m.CreatedAt,
			MessageUserInfo: m.MessageUserInfo,
		})
	}
	return out
}

// NewPublicHandler creates a new public handler
func NewPublicHandler(publicService *service.PublicService) *PublicHandler {
	return &PublicHandler{
		publicService: publicService,
		validator:     validator.New(),
	}
}

// GetTicketByMagicLink handles public ticket access via magic link
// @Summary Get ticket by magic link
// @Description Retrieve ticket details and messages using a magic link token (no authentication required)
// @Tags Public API
// @Accept json
// @Produce json
// @Param token path string true "Magic link token"
// @Success 200 {object} map[string]interface{} "Ticket details with messages"
// @Failure 400 {object} map[string]interface{} "Bad request - Missing or invalid token"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid or expired token"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/public/tickets/tokens/{token} [get]
func (h *PublicHandler) GetTicketByMagicLink(c *gin.Context) {
	magicToken := c.Param("token")
	if magicToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Magic token is required"})
		return
	}

	ticket, err := h.publicService.GetTicketByMagicLink(c.Request.Context(), magicToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Also get messages for the combined response that the frontend expects
	messages, _, err := h.publicService.GetTicketMessagesByMagicLink(c.Request.Context(), magicToken, "", 50)
	if err != nil {
		// If messages fail, just return the ticket without them
		c.JSON(http.StatusOK, gin.H{
			"valid":    true,
			"ticket":   ticket,
			"messages": []interface{}{},
		})
		return
	}

	publicMsgs := toPublicMessages(messages)

	c.JSON(http.StatusOK, gin.H{
		"valid":    true,
		"ticket":   ticket,
		"messages": publicMsgs,
	})
}

// GetTicketMessagesByMagicLink handles public ticket messages access via magic link
// @Summary Get ticket messages by magic link
// @Description Retrieve paginated messages for a ticket using a magic link token
// @Tags Public API
// @Accept json
// @Produce json
// @Param token path string true "Magic link token"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Number of messages per page" minimum(1) maximum(100) default(50)
// @Success 200 {object} map[string]interface{} "Paginated list of messages"
// @Failure 400 {object} map[string]interface{} "Bad request - Missing or invalid token"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid or expired token"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/public/tickets/tokens/{token}/messages [get]
func (h *PublicHandler) GetTicketMessagesByMagicLink(c *gin.Context) {
	magicToken := c.Param("token")
	if magicToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Magic token is required"})
		return
	}

	// Parse query parameters
	cursor := c.Query("cursor")
	limit := 50 // default limit

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	messages, nextCursor, err := h.publicService.GetTicketMessagesByMagicLink(c.Request.Context(), magicToken, cursor, limit)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"messages": toPublicMessages(messages),
	}

	if nextCursor != "" {
		response["next_cursor"] = nextCursor
	}

	c.JSON(http.StatusOK, response)
}

// AddMessageByMagicLink handles adding a public message via magic link
// @Summary Add message by magic link
// @Description Add a new message to a ticket using a magic link token
// @Tags Public API
// @Accept json
// @Produce json
// @Param token path string true "Magic link token"
// @Param message body service.AddPublicMessageRequest true "Message content"
// @Success 201 {object} PublicMessage "Message created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - Invalid input data"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid or expired token"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/public/tickets/tokens/{token}/messages [post]
func (h *PublicHandler) AddMessageByMagicLink(c *gin.Context) {
	magicToken := c.Param("token")
	if magicToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Magic token is required"})
		return
	}

	var req service.AddPublicMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	message, err := h.publicService.AddMessageByMagicLink(c.Request.Context(), magicToken, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Map to public response shape
	publicMsg := PublicMessage{
		ID:         message.ID,
		TicketID:   message.TicketID,
		AuthorType: message.AuthorType,
		AuthorID:   message.AuthorID,
		Body:       message.Body,
		IsPrivate:  message.IsPrivate,
		CreatedAt:  message.CreatedAt,
	}

	c.JSON(http.StatusCreated, publicMsg)
}

// GetTicketByID handles public ticket access via ticket ID
// @Summary Get ticket by ID
// @Description Retrieve ticket details and messages using ticket ID (no authentication required)
// @Tags Public API
// @Accept json
// @Produce json
// @Param ticketId path string true "Ticket ID" format(uuid)
// @Success 200 {object} map[string]interface{} "Ticket details with messages"
// @Failure 400 {object} map[string]interface{} "Bad request - Invalid ticket ID format"
// @Failure 404 {object} map[string]interface{} "Ticket not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/public/tickets/{ticketId} [get]
func (h *PublicHandler) GetTicketByID(c *gin.Context) {
	ticketIdStr := c.Param("ticketId")
	if ticketIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ticket ID is required"})
		return
	}

	ticketID, err := uuid.Parse(ticketIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID format"})
		return
	}

	ticket, err := h.publicService.GetTicketByID(c.Request.Context(), ticketID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Also get messages for the combined response that the frontend expects
	messages, _, err := h.publicService.GetTicketMessagesByID(c.Request.Context(), ticketID, "", 50)
	if err != nil {
		// If messages fail, just return the ticket without them
		c.JSON(http.StatusOK, gin.H{
			"valid":    true,
			"ticket":   ticket,
			"messages": []interface{}{},
		})
		return
	}

	publicMsgs := toPublicMessages(messages)

	c.JSON(http.StatusOK, gin.H{
		"valid":    true,
		"ticket":   ticket,
		"messages": publicMsgs,
	})
}

// GetTicketMessagesByID handles public ticket messages access via ticket ID
// @Summary Get ticket messages by ID
// @Description Retrieve paginated messages for a ticket using ticket ID
// @Tags Public API
// @Accept json
// @Produce json
// @Param ticketId path string true "Ticket ID" format(uuid)
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Number of messages per page" minimum(1) maximum(100) default(50)
// @Success 200 {object} map[string]interface{} "Paginated list of messages"
// @Failure 400 {object} map[string]interface{} "Bad request - Invalid ticket ID format"
// @Failure 404 {object} map[string]interface{} "Ticket not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/public/tickets/{ticketId}/messages [get]
func (h *PublicHandler) GetTicketMessagesByID(c *gin.Context) {
	ticketIdStr := c.Param("ticketId")
	if ticketIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ticket ID is required"})
		return
	}

	ticketID, err := uuid.Parse(ticketIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID format"})
		return
	}

	// Parse query parameters
	cursor := c.Query("cursor")
	limit := 50 // default limit

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	messages, nextCursor, err := h.publicService.GetTicketMessagesByID(c.Request.Context(), ticketID, cursor, limit)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"messages": toPublicMessages(messages),
	}

	if nextCursor != "" {
		response["next_cursor"] = nextCursor
	}

	c.JSON(http.StatusOK, response)
}

// AddMessageByID handles adding a public message via ticket ID
// @Summary Add message by ticket ID
// @Description Add a new message to a ticket using ticket ID
// @Tags Public API
// @Accept json
// @Produce json
// @Param ticketId path string true "Ticket ID" format(uuid)
// @Param message body service.AddPublicMessageRequest true "Message content"
// @Success 201 {object} PublicMessage "Message created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - Invalid input data"
// @Failure 404 {object} map[string]interface{} "Ticket not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/public/tickets/{ticketId}/messages [post]
func (h *PublicHandler) AddMessageByID(c *gin.Context) {
	ticketIdStr := c.Param("ticketId")
	if ticketIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ticket ID is required"})
		return
	}

	ticketID, err := uuid.Parse(ticketIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID format"})
		return
	}

	var req service.AddPublicMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	message, err := h.publicService.AddMessageByID(c.Request.Context(), ticketID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Map to public response shape
	publicMsg := PublicMessage{
		ID:         message.ID,
		TicketID:   message.TicketID,
		AuthorType: message.AuthorType,
		AuthorID:   message.AuthorID,
		Body:       message.Body,
		IsPrivate:  message.IsPrivate,
		CreatedAt:  message.CreatedAt,
	}

	c.JSON(http.StatusCreated, publicMsg)
}

// GenerateMagicLink generates a magic link for testing purposes
// This endpoint should be removed in production
// @Summary Generate magic link (Testing only)
// @Description Generate a magic link token for testing purposes - should be removed in production
// @Tags Public API
// @Accept json
// @Produce json
// @Param request body map[string]string true "Ticket ID and Customer ID"
// @Success 200 {object} map[string]interface{} "Magic link token and URL"
// @Failure 400 {object} map[string]interface{} "Bad request - Invalid input data"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/public/generate-magic-link [post]
func (h *PublicHandler) GenerateMagicLink(c *gin.Context) {
	type GenerateMagicLinkRequest struct {
		TicketID   string `json:"ticket_id" binding:"required"`
		CustomerID string `json:"customer_id" binding:"required"`
	}

	var req GenerateMagicLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticketID, err := uuid.Parse(req.TicketID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket_id"})
		return
	}

	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer_id"})
		return
	}

	// Generate magic link token
	token, err := h.publicService.GenerateMagicLinkToken(ticketID, customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate magic link"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"magic_token": token,
		"public_url":  fmt.Sprintf("http://localhost:8081/index.html?token=%s", token),
	})
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Time    string `json:"time"`
}

// Health handles health check endpoint
// @Summary Health check
// @Description Check the health status of the API server
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse "Service is healthy"
// @Router /health [get]
// @Router /api/public/health [get]
func (h *PublicHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:  "healthy",
		Version: "1.0.0",
		Time:    "2024-01-01T00:00:00Z", // This would be set at build time
	})
}
