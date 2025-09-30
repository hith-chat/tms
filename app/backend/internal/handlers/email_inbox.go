package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/bareuptime/tms/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// EmailInboxHandler handles email inbox HTTP requests
type EmailInboxHandler struct {
	emailInboxService *service.EmailInboxService
}

// NewEmailInboxHandler creates a new email inbox handler
func NewEmailInboxHandler(emailInboxService *service.EmailInboxService) *EmailInboxHandler {
	return &EmailInboxHandler{
		emailInboxService: emailInboxService,
	}
}

// ListEmailsRequest represents the request for listing emails
type ListEmailsRequest struct {
	ProjectID      *string `form:"project_id"`
	MailboxAddress *string `form:"mailbox_address"`
	Mailbox        *string `form:"mailbox"` // Add mailbox field for frontend compatibility
	IsRead         *bool   `form:"is_read"`
	IsReply        *bool   `form:"is_reply"`
	HasTicket      *bool   `form:"has_ticket"`
	ThreadID       *string `form:"thread_id"`
	FromAddress    *string `form:"from_address"`
	Subject        *string `form:"subject"`
	Search         *string `form:"search"` // Add search field
	StartDate      *string `form:"start_date"`
	EndDate        *string `form:"end_date"`
	Limit          int     `form:"limit,default=50"`
	Offset         int     `form:"offset,default=0"`
	Page           int     `form:"page,default=1"` // Add page for frontend compatibility
	OrderBy        string  `form:"order_by,default=received_at"`
	OrderDir       string  `form:"order_dir,default=DESC"`
}

// ListEmailsResponse represents the response for listing emails
type ListEmailsResponse struct {
	Emails []EmailResponse `json:"emails"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

// EmailResponse represents an email in the response
type EmailResponse struct {
	ID                  uuid.UUID      `json:"id"`
	MessageID           string         `json:"message_id"`
	ThreadID            *string        `json:"thread_id,omitempty"`
	MailboxAddress      string         `json:"mailbox_address"`
	FromAddress         string         `json:"from_address"`
	FromName            *string        `json:"from_name,omitempty"`
	ToAddresses         []string       `json:"to_addresses"`
	CcAddresses         []string       `json:"cc_addresses,omitempty"`
	Subject             string         `json:"subject"`
	BodyText            *string        `json:"body_text,omitempty"`
	BodyHTML            *string        `json:"body_html,omitempty"`
	Snippet             *string        `json:"snippet,omitempty"`
	IsRead              bool           `json:"is_read"`
	IsReply             bool           `json:"is_reply"`
	HasAttachments      bool           `json:"has_attachments"`
	AttachmentCount     int            `json:"attachment_count"`
	SentAt              *time.Time     `json:"sent_at,omitempty"`
	ReceivedAt          time.Time      `json:"received_at"`
	TicketID            *uuid.UUID     `json:"ticket_id,omitempty"`
	IsConvertedToTicket bool           `json:"is_converted_to_ticket"`
	Headers             models.JSONMap `json:"headers,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
}

// ConvertToTicketRequest represents request to convert email to ticket
type ConvertToTicketRequest struct {
	Type     string `json:"type" binding:"required,oneof=question incident problem task"`
	Priority string `json:"priority" binding:"required,oneof=low normal high urgent"`
}

// ReplyToEmailRequest represents request to reply to an email
type ReplyToEmailRequest struct {
	Body        string   `json:"body" binding:"required"`
	Subject     *string  `json:"subject,omitempty"`      // Optional custom subject, defaults to "Re: original subject"
	CCAddresses []string `json:"cc_addresses,omitempty"` // Additional CC recipients
	IsPrivate   bool     `json:"is_private"`             // Whether this is an internal reply
}

// SyncEmailsResponse represents the response for sync operation
type SyncEmailsResponse struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	StartedAt time.Time `json:"started_at"`
}

// ListEmails handles GET /emails
// @Summary List emails
// @Description Retrieve a filtered list of emails from the inbox
// @Tags email-inbox
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id query string false "Filter by project ID"
// @Param mailbox_address query string false "Filter by mailbox address"
// @Param is_read query boolean false "Filter by read status"
// @Param has_ticket query boolean false "Filter by ticket association"
// @Param search query string false "Search in subject and content"
// @Param limit query int false "Number of emails to return (default: 50)"
// @Param offset query int false "Number of emails to skip (default: 0)"
// @Param order_by query string false "Order by field (default: received_at)"
// @Param order_dir query string false "Order direction (default: DESC)"
// @Success 200 {object} object{emails=[]EmailResponse,total=int,page=int}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/emails [get]
func (h *EmailInboxHandler) ListEmails(c *gin.Context) {

	tenantUUID := middleware.GetTenantID(c)

	var req ListEmailsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert request to filter
	filter := h.convertToEmailFilter(req)

	emails, total, err := h.emailInboxService.ListEmails(c.Request.Context(), tenantUUID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list emails"})
		return
	}

	// Convert to response format
	emailResponses := make([]EmailResponse, len(emails))
	for i, email := range emails {
		emailResponses[i] = h.convertToEmailResponse(email)
	}

	response := ListEmailsResponse{
		Emails: emailResponses,
		Total:  total,
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	c.JSON(http.StatusOK, response)
}

// GetEmail handles GET /emails/:id
// @Summary Get email
// @Description Retrieve a specific email by its ID
// @Tags email-inbox
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param id path string true "Email ID"
// @Success 200 {object} EmailResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/emails/{id} [get]
func (h *EmailInboxHandler) GetEmail(c *gin.Context) {

	tenantUUID := middleware.GetTenantID(c)
	projectUUID := middleware.GetProjectID(c)

	emailID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	email, attachments, err := h.emailInboxService.GetEmailWithAttachments(c.Request.Context(), tenantUUID, projectUUID, emailID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "email not found"})
		return
	}

	// Mark email as read
	if !email.IsRead {
		err = h.emailInboxService.MarkEmailAsRead(c.Request.Context(), tenantUUID, projectUUID, emailID)
		if err != nil {
			// Log error but don't fail the request
		}
	}

	response := h.convertToEmailResponse(email)

	// Add attachments to response
	if len(attachments) > 0 {
		// You can extend the response structure to include attachments
	}

	c.JSON(http.StatusOK, response)
}

// MarkAsRead handles PUT /emails/:id/read
// MarkAsRead marks an email as read
// @Summary Mark email as read
// @Description Mark a specific email as read
// @Tags email-inbox
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param id path string true "Email ID"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/emails/{id}/mark-read [post]
func (h *EmailInboxHandler) MarkAsRead(c *gin.Context) {

	tenantUUID := middleware.GetTenantID(c)
	projectUUID := middleware.GetProjectID(c)

	emailID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	err = h.emailInboxService.MarkEmailAsRead(c.Request.Context(), tenantUUID, projectUUID, emailID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark email as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "email marked as read"})
}

// ConvertToTicket handles POST /emails/:id/convert-to-ticket
// ConvertToTicket converts an email to a support ticket
// @Summary Convert email to ticket
// @Description Convert an email message to a support ticket
// @Tags email-inbox
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param id path string true "Email ID"
// @Success 200 {object} models.Ticket
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/emails/{id}/convert-to-ticket [post]
func (h *EmailInboxHandler) ConvertToTicket(c *gin.Context) {

	tenantUUID := middleware.GetTenantID(c)
	projectUUID := middleware.GetProjectID(c)

	emailID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req ConvertToTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticket, err := h.emailInboxService.ConvertEmailToTicket(c.Request.Context(), tenantUUID, emailID, projectUUID, req.Type, req.Priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "email converted to ticket successfully",
		"ticket_id":     ticket.ID,
		"ticket_number": ticket.Number,
	})
}

// ReplyToEmail handles POST /emails/:id/reply
func (h *EmailInboxHandler) ReplyToEmail(c *gin.Context) {

	tenantUUID := middleware.GetTenantID(c)
	projectUUID := middleware.GetProjectID(c)

	emailID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req ReplyToEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the request
	if strings.TrimSpace(req.Body) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reply body cannot be empty"})
		return
	}

	// Verify the email exists and belongs to this tenant/project
	originalEmail, _, err := h.emailInboxService.GetEmailWithAttachments(c.Request.Context(), tenantUUID, projectUUID, emailID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "email not found"})
		return
	}

	err = h.emailInboxService.ReplyToEmail(c.Request.Context(), tenantUUID, emailID, projectUUID, originalEmail, req.Body, req.Subject, req.CCAddresses, req.IsPrivate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "reply sent successfully",
		"email_id":   emailID,
		"is_private": req.IsPrivate,
	})
}

// SyncEmails handles POST /emails/sync
func (h *EmailInboxHandler) SyncEmails(c *gin.Context) {
	tenantUUID := middleware.GetTenantID(c)
	projectUUID := middleware.GetProjectID(c)

	// Run sync in background
	go func() {
		// Use background context instead of request context to avoid cancellation
		ctx := context.Background()
		err := h.emailInboxService.SyncEmails(ctx, tenantUUID, projectUUID)
		if err != nil {
			// Log error - you might want to use a proper logger here
			fmt.Printf("Email sync error for tenant %s: %v\n", tenantUUID, err)
		}
	}()

	response := SyncEmailsResponse{
		Status:    "started",
		Message:   "Email synchronization started",
		StartedAt: time.Now(),
	}

	c.JSON(http.StatusAccepted, response)
}

// GetSyncStatus handles GET /emails/sync-status
func (h *EmailInboxHandler) GetSyncStatus(c *gin.Context) {
	tenantUUID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	statuses, err := h.emailInboxService.GetSyncStatus(c.Request.Context(), tenantUUID, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sync status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sync_statuses": statuses})
}

// Helper functions

func (h *EmailInboxHandler) convertToEmailFilter(req ListEmailsRequest) repo.EmailFilter {
	filter := repo.EmailFilter{
		Limit:    req.Limit,
		Offset:   req.Offset,
		OrderBy:  req.OrderBy,
		OrderDir: req.OrderDir,
	}

	// Convert page to offset if provided
	if req.Page > 0 {
		filter.Offset = (req.Page - 1) * req.Limit
	}

	if req.ProjectID != nil {
		if projectUUID, err := uuid.Parse(*req.ProjectID); err == nil {
			filter.ProjectID = &projectUUID
		}
	}

	// Handle both mailbox_address and mailbox parameters for frontend compatibility
	if req.MailboxAddress != nil {
		filter.MailboxAddress = req.MailboxAddress
	} else if req.Mailbox != nil {
		filter.MailboxAddress = req.Mailbox
	}

	if req.IsRead != nil {
		filter.IsRead = req.IsRead
	}

	if req.IsReply != nil {
		filter.IsReply = req.IsReply
	}

	if req.HasTicket != nil {
		filter.HasTicket = req.HasTicket
	}

	if req.ThreadID != nil {
		filter.ThreadID = req.ThreadID
	}

	if req.FromAddress != nil {
		filter.FromAddress = req.FromAddress
	}

	if req.Subject != nil {
		filter.Subject = req.Subject
	}

	// Handle search parameter - search across subject, from_address, and body
	if req.Search != nil {
		filter.Search = req.Search
	}

	if req.StartDate != nil {
		if startTime, err := time.Parse(time.RFC3339, *req.StartDate); err == nil {
			filter.StartDate = &startTime
		}
	}

	if req.EndDate != nil {
		if endTime, err := time.Parse(time.RFC3339, *req.EndDate); err == nil {
			filter.EndDate = &endTime
		}
	}

	return filter
}

func (h *EmailInboxHandler) convertToEmailResponse(email *models.EmailInbox) EmailResponse {
	return EmailResponse{
		ID:                  email.ID,
		MessageID:           email.MessageID,
		ThreadID:            email.ThreadID,
		MailboxAddress:      email.MailboxAddress,
		FromAddress:         email.FromAddress,
		FromName:            email.FromName,
		ToAddresses:         []string(email.ToAddresses),
		CcAddresses:         []string(email.CcAddresses),
		Subject:             email.Subject,
		BodyText:            email.BodyText,
		BodyHTML:            email.BodyHTML,
		Snippet:             email.Snippet,
		IsRead:              email.IsRead,
		IsReply:             email.IsReply,
		HasAttachments:      email.HasAttachments,
		AttachmentCount:     email.AttachmentCount,
		SentAt:              email.SentAt,
		ReceivedAt:          email.ReceivedAt,
		TicketID:            email.TicketID,
		IsConvertedToTicket: email.IsConvertedToTicket,
		Headers:             email.Headers,
		CreatedAt:           email.CreatedAt,
	}
}
