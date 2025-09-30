package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// TicketHandler handles ticket-related endpoints
type TicketHandler struct {
	ticketService  *service.TicketService
	messageService *service.MessageService
	validator      *validator.Validate
}

// NewTicketHandler creates a new ticket handler
func NewTicketHandler(ticketService *service.TicketService, messageService *service.MessageService) *TicketHandler {
	return &TicketHandler{
		ticketService:  ticketService,
		messageService: messageService,
		validator:      validator.New(),
	}
}

// CreateTicket handles ticket creation
// @Summary Create a new ticket
// @Description Create a new support ticket with customer information and initial message
// @Tags Tickets
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID" format(uuid)
// @Param project_id path string true "Project ID" format(uuid)
// @Param ticket body service.CreateTicketRequest true "Ticket creation data"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Success 201 {object} models.Ticket "Ticket created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - Invalid input data"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid or missing authentication"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /v1/tenants/{tenant_id}/projects/{project_id}/tickets [post]
func (h *TicketHandler) CreateTicket(c *gin.Context) {
	var req service.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)
	agentID := middleware.GetAgentID(c)

	ticket, err := h.ticketService.CreateTicket(c.Request.Context(), tenantID, projectID, agentID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

// UpdateTicket handles ticket updates
// @Summary Update an existing ticket
// @Description Update ticket properties such as status, priority, assignee, etc.
// @Tags Tickets
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID" format(uuid)
// @Param project_id path string true "Project ID" format(uuid)
// @Param ticket_id path string true "Ticket ID" format(uuid)
// @Param ticket body service.UpdateTicketRequest true "Ticket update data"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Success 200 {object} models.Ticket "Ticket updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - Invalid input data"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid or missing authentication"
// @Failure 403 {object} map[string]interface{} "Forbidden - Insufficient permissions"
// @Failure 404 {object} map[string]interface{} "Ticket not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /v1/tenants/{tenant_id}/projects/{project_id}/tickets/{ticket_id} [patch]
func (h *TicketHandler) UpdateTicket(c *gin.Context) {
	var req service.UpdateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	// Check if reassignment is being attempted
	if req.AssigneeAgentID != nil {
		allowReassignment, exists := c.Get("allow_reassignment")
		if !exists || !allowReassignment.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to reassign tickets"})
			return
		}
	}

	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)
	ticketID := c.Param("ticket_id")
	agentID := middleware.GetAgentID(c)

	ticketUUID, _ := uuid.Parse(ticketID)

	ticket, err := h.ticketService.UpdateTicket(c.Request.Context(), tenantID, projectID, ticketUUID, agentID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

// GetTicket handles ticket retrieval
// @Summary Get ticket details
// @Description Retrieve detailed information about a specific ticket
// @Tags Tickets
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID" format(uuid)
// @Param project_id path string true "Project ID" format(uuid)
// @Param ticket_id path string true "Ticket ID" format(uuid)
// @Security BearerAuth
// @Security ApiKeyAuth
// @Success 200 {object} models.Ticket "Ticket details"
// @Failure 400 {object} map[string]interface{} "Bad request - Invalid ticket ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid or missing authentication"
// @Failure 404 {object} map[string]interface{} "Ticket not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /v1/tenants/{tenant_id}/projects/{project_id}/tickets/{ticket_id} [get]
func (h *TicketHandler) GetTicket(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)
	ticketID := c.Param("ticket_id")
	agentID := middleware.GetAgentID(c)

	ticketUUID, _ := uuid.Parse(ticketID)

	ticket, err := h.ticketService.GetTicket(c.Request.Context(), tenantID, projectID, ticketUUID, agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

// ListTickets handles ticket listing
// @Summary List tickets
// @Description Get a paginated list of tickets with optional filtering
// @Tags Tickets
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID" format(uuid)
// @Param project_id path string true "Project ID" format(uuid)
// @Param status query []string false "Filter by status" collectionFormat(multi)
// @Param priority query []string false "Filter by priority" collectionFormat(multi)
// @Param assignee_id query string false "Filter by assignee ID" format(uuid)
// @Param customer_id query string false "Filter by customer ID" format(uuid)
// @Param tags query []string false "Filter by tags" collectionFormat(multi)
// @Param search query string false "Search in ticket content"
// @Param source query []string false "Filter by source" collectionFormat(multi)
// @Param type query []string false "Filter by type" collectionFormat(multi)
// @Param limit query int false "Number of tickets per page" minimum(1) maximum(100) default(20)
// @Param cursor query string false "Pagination cursor"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "List of tickets with pagination info"
// @Failure 400 {object} map[string]interface{} "Bad request - Invalid query parameters"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid or missing authentication"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /v1/tenants/{tenant_id}/projects/{project_id}/tickets [get]
func (h *TicketHandler) ListTickets(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	fmt.Println("Tenant ID:", tenantID)
	projectID := middleware.GetProjectID(c)
	fmt.Println("Project ID:", projectID)
	agentID := middleware.GetAgentID(c)
	fmt.Println("Agent ID:", agentID)

	// Parse query parameters
	req := service.ListTicketsRequest{
		Status:   c.QueryArray("status"),
		Priority: c.QueryArray("priority"),
		Tags:     c.QueryArray("tags"),
		Search:   c.Query("search"),
		Source:   c.QueryArray("source"),
		Type:     c.QueryArray("type"),
		Cursor:   c.Query("cursor"),
	}

	if assigneeID := c.Query("assignee_id"); assigneeID != "" {
		req.AssigneeID = &assigneeID
	}

	if requesterID := c.Query("customer_id"); requesterID != "" {
		req.CustomerID = &requesterID
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = limit
		}
	}

	tickets, nextCursor, err := h.ticketService.ListTickets(c.Request.Context(), tenantID, projectID, agentID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"tickets": tickets,
	}

	if nextCursor != "" {
		response["next_cursor"] = nextCursor
	}

	c.JSON(http.StatusOK, response)
}

// AddMessage handles adding a message to a ticket
func (h *TicketHandler) AddMessage(c *gin.Context) {
	var req service.AddMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)
	ticketID := c.Param("ticket_id")
	agentID := middleware.GetAgentID(c)

	ticketUUID, _ := uuid.Parse(ticketID)

	message, err := h.ticketService.AddMessage(c.Request.Context(), tenantID, projectID, ticketUUID, agentID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// GetTicketMessages handles retrieving messages for a ticket
func (h *TicketHandler) GetTicketMessages(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)
	ticketID := c.Param("ticket_id")
	agentID := middleware.GetAgentID(c)

	// Parse query parameters
	includePrivate := c.Query("include_private") == "true"
	cursor := c.Query("cursor")
	limit := 50 // default limit

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	ticketUUID, _ := uuid.Parse(ticketID)

	messages, nextCursor, err := h.messageService.GetTicketMessages(c.Request.Context(), tenantID, projectID, ticketUUID, agentID, includePrivate, cursor, limit)

	// get CustomerId and AgentId from messages in set and than fetch their details to be injected in tickets

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"messages": messages,
	}

	if nextCursor != "" {
		response["next_cursor"] = nextCursor
	}

	c.JSON(http.StatusOK, response)
}

// UpdateMessage handles updating a message
func (h *TicketHandler) UpdateMessage(c *gin.Context) {
	var req service.UpdateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)
	ticketID := c.Param("ticket_id")
	messageID := c.Param("message_id")
	agentID := middleware.GetAgentID(c)
	ticketUUID, _ := uuid.Parse(ticketID)
	messageUUID, _ := uuid.Parse(messageID)
	message, err := h.messageService.UpdateMessage(c.Request.Context(), tenantID, projectID, ticketUUID, messageUUID, agentID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, message)
}

// DeleteMessage handles deleting a message
func (h *TicketHandler) DeleteMessage(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)
	ticketID := c.Param("ticket_id")
	messageID := c.Param("message_id")
	agentID := middleware.GetAgentID(c)

	ticketUUID, _ := uuid.Parse(ticketID)
	messageUUID, _ := uuid.Parse(messageID)

	err := h.messageService.DeleteMessage(c.Request.Context(), tenantID, projectID, ticketUUID, messageUUID, agentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ReassignTicket handles ticket reassignment
func (h *TicketHandler) ReassignTicket(c *gin.Context) {
	var req service.ReassignTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)
	ticketID := c.Param("ticket_id")
	agentID := middleware.GetAgentID(c)

	ticketUUID, _ := uuid.Parse(ticketID)

	ticket, err := h.ticketService.ReassignTicket(c.Request.Context(), tenantID, projectID, ticketUUID, agentID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

// ValidateCustomer handles customer validation via OTP
func (h *TicketHandler) ValidateCustomer(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)
	ticketID := c.Param("ticket_id")

	ticketUUID, err := uuid.Parse(ticketID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket_id"})
		return
	}

	result, err := h.ticketService.SendCustomerValidationOTP(c.Request.Context(), tenantID, projectID, ticketUUID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// SendMagicLink handles sending magic link to customer
func (h *TicketHandler) SendMagicLink(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)
	ticketID := c.Param("ticket_id")

	ticketUUID, err := uuid.Parse(ticketID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket_id"})
		return
	}

	result, err := h.ticketService.SendMagicLinkToCustomer(c.Request.Context(), tenantID, projectID, ticketUUID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteTicket handles ticket deletion
func (h *TicketHandler) DeleteTicket(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)
	ticketIDStr := c.Param("ticket_id")
	agentID := middleware.GetAgentID(c)

	ticketID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket_id"})
		return
	}

	err = h.ticketService.DeleteTicket(c.Request.Context(), tenantID, projectID, ticketID, agentID)
	if err != nil {
		if err.Error() == "insufficient permissions to delete ticket" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to delete ticket"})
			return
		}
		if err.Error() == "ticket not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete ticket"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ticket deleted successfully"})
}
