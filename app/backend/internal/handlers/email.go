package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bareuptime/tms/internal/crypto"
	"github.com/bareuptime/tms/internal/mail"
	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/redis"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// EmailHandler handles email-related HTTP requests
type EmailHandler struct {
	emailRepo    *repo.EmailRepo
	redisService *redis.Service
	mailService  *mail.Service
	encryption   *crypto.PasswordEncryption
}

// NewEmailHandler creates a new email handler
func NewEmailHandler(emailRepo *repo.EmailRepo, redisService *redis.Service, mailService *mail.Service) *EmailHandler {
	encryption, err := crypto.NewPasswordEncryption()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize password encryption: %v", err))
	}

	return &EmailHandler{
		emailRepo:    emailRepo,
		redisService: redisService,
		mailService:  mailService,
		encryption:   encryption,
	}
}

// CreateConnectorRequest represents a request to create an email connector
type CreateConnectorRequest struct {
	Type             models.EmailConnectorType `json:"type" binding:"required"`
	Name             string                    `json:"name" binding:"required"`
	IMAPHost         *string                   `json:"imap_host,omitempty"`
	IMAPPort         *int                      `json:"imap_port,omitempty"`
	IMAPUseTLS       *bool                     `json:"imap_use_tls,omitempty"`
	IMAPUsername     *string                   `json:"imap_username,omitempty"`
	IMAPPassword     *string                   `json:"imap_password,omitempty"`
	IMAPFolder       string                    `json:"imap_folder,omitempty"`
	IMAPSeenStrategy *models.IMAPSeenStrategy  `json:"imap_seen_strategy,omitempty"`
	SMTPHost         *string                   `json:"smtp_host,omitempty"`
	SMTPPort         *int                      `json:"smtp_port,omitempty"`
	SMTPUseTLS       *bool                     `json:"smtp_use_tls,omitempty"`
	SMTPUsername     *string                   `json:"smtp_username,omitempty"`
	SMTPPassword     *string                   `json:"smtp_password,omitempty"`
}

// ValidateConnectorRequest represents a request to validate email connector
type ValidateConnectorRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// VerifyOTPRequest represents a request to verify OTP
type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

// CreateConnector creates a new email connector
// @Summary Create email connector
// @Description Create a new email connector for incoming emails
// @Tags email
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param connector body CreateConnectorRequest true "Email connector configuration"
// @Success 201 {object} models.EmailConnector
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/email/connectors [post]
func (h *EmailHandler) CreateConnector(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)

	// Get project ID from URL params
	projectID := middleware.GetProjectID(c)

	var req CreateConnectorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create connector model
	connector := &models.EmailConnector{
		ID:               uuid.New(),
		TenantID:         tenantID,
		ProjectID:        &projectID,
		Type:             req.Type,
		Name:             req.Name,
		IsActive:         true,
		IsValidated:      false,
		ValidationStatus: models.ValidationStatusPending,
		IMAPHost:         req.IMAPHost,
		IMAPPort:         req.IMAPPort,
		IMAPUseTLS:       req.IMAPUseTLS,
		IMAPUsername:     req.IMAPUsername,
		IMAPFolder:       req.IMAPFolder,
		SMTPHost:         req.SMTPHost,
		SMTPPort:         req.SMTPPort,
		SMTPUseTLS:       req.SMTPUseTLS,
		SMTPUsername:     req.SMTPUsername,
		LastHealth:       make(models.JSONMap),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Set default IMAP folder if not provided
	if connector.IMAPFolder == "" {
		connector.IMAPFolder = "INBOX"
	}

	// Set IMAP seen strategy - use request value or default
	if req.IMAPSeenStrategy != nil {
		connector.IMAPSeenStrategy = *req.IMAPSeenStrategy
	} else {
		connector.IMAPSeenStrategy = models.SeenStrategyMarkAfterParse
	}

	// Encrypt passwords using AES encryption
	if req.IMAPPassword != nil {
		encryptedPassword, err := h.encryption.Encrypt(*req.IMAPPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt IMAP password"})
			return
		}
		connector.IMAPPasswordEnc = encryptedPassword
	}
	if req.SMTPPassword != nil {
		encryptedPassword, err := h.encryption.Encrypt(*req.SMTPPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt SMTP password"})
			return
		}
		connector.SMTPPasswordEnc = encryptedPassword
	}

	// Save to database
	if statusCode, err := h.emailRepo.CreateConnector(c.Request.Context(), connector); err != nil {
		fmt.Println("Failed to create connector:", err)
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	// Remove sensitive data from response
	connector.IMAPPasswordEnc = nil
	connector.SMTPPasswordEnc = nil

	c.JSON(http.StatusCreated, connector)
}

// ListConnectors lists all email connectors for a tenant
func (h *EmailHandler) ListConnectors(c *gin.Context) {

	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	var connectorType *models.EmailConnectorType
	if typeParam := c.Query("type"); typeParam != "" {
		t := models.EmailConnectorType(typeParam)
		connectorType = &t
	}

	connectors, err := h.emailRepo.ListConnectors(c.Request.Context(), tenantID, projectID, connectorType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list connectors"})
		return
	}

	// Remove sensitive data from response
	for _, connector := range connectors {
		connector.IMAPPasswordEnc = nil
		connector.SMTPPasswordEnc = nil
		connector.DKIMPrivateKeyEnc = nil
	}

	c.JSON(http.StatusOK, gin.H{"connectors": connectors})
}

// GetConnector gets a specific email connector
func (h *EmailHandler) GetConnector(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	connectorIDParam := c.Param("connector_id")
	connectorID, err := uuid.Parse(connectorIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connector ID"})
		return
	}

	connector, err := h.emailRepo.GetConnector(c.Request.Context(), tenantID, projectID, connectorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get connector"})
		return
	}

	if connector == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Connector not found"})
		return
	}

	// Remove sensitive data from response
	connector.IMAPPasswordEnc = nil
	connector.SMTPPasswordEnc = nil
	connector.DKIMPrivateKeyEnc = nil

	c.JSON(http.StatusOK, connector)
}

// UpdateConnector updates an email connector
func (h *EmailHandler) UpdateConnector(c *gin.Context) {
	tenantIDStr := c.MustGet("tenant_id").(string)
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID format"})
		return
	}

	projectID := middleware.GetProjectID(c)

	connectorIDParam := c.Param("connector_id")
	connectorID, err := uuid.Parse(connectorIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connector ID"})
		return
	}

	// Get existing connector
	connector, err := h.emailRepo.GetConnector(c.Request.Context(), tenantID, projectID, connectorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get connector"})
		return
	}

	if connector == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Connector not found"})
		return
	}

	var req CreateConnectorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	connector.Name = req.Name
	connector.IMAPHost = req.IMAPHost
	connector.IMAPPort = req.IMAPPort
	connector.IMAPUseTLS = req.IMAPUseTLS
	connector.IMAPUsername = req.IMAPUsername
	if req.IMAPFolder != "" {
		connector.IMAPFolder = req.IMAPFolder
	}
	if req.IMAPSeenStrategy != nil {
		connector.IMAPSeenStrategy = *req.IMAPSeenStrategy
	}
	connector.SMTPHost = req.SMTPHost
	connector.SMTPPort = req.SMTPPort
	connector.SMTPUseTLS = req.SMTPUseTLS
	connector.SMTPUsername = req.SMTPUsername
	connector.UpdatedAt = time.Now()

	// Update passwords if provided (with proper encryption)
	if req.IMAPPassword != nil {
		encryptedPassword, err := h.encryption.Encrypt(*req.IMAPPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt IMAP password"})
			return
		}
		connector.IMAPPasswordEnc = encryptedPassword
	}
	if req.SMTPPassword != nil {
		encryptedPassword, err := h.encryption.Encrypt(*req.SMTPPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt SMTP password"})
			return
		}
		connector.SMTPPasswordEnc = encryptedPassword
	}

	// Save to database
	if err := h.emailRepo.UpdateConnector(c.Request.Context(), connector); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update connector"})
		return
	}

	// Remove sensitive data from response
	connector.IMAPPasswordEnc = nil
	connector.SMTPPasswordEnc = nil

	c.JSON(http.StatusOK, connector)
}

// DeleteConnector deletes an email connector
func (h *EmailHandler) DeleteConnector(c *gin.Context) {
	tenantIDStr := c.MustGet("tenant_id").(string)
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID format"})
		return
	}

	projectID := middleware.GetProjectID(c)

	connectorIDParam := c.Param("connector_id")
	connectorID, err := uuid.Parse(connectorIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connector ID"})
		return
	}

	if err := h.emailRepo.DeleteConnector(c.Request.Context(), tenantID, projectID, connectorID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete connector"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// TestConnector tests an email connector connection
func (h *EmailHandler) TestConnector(c *gin.Context) {
	tenantIDStr := c.MustGet("tenant_id").(string)
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID format"})
		return
	}

	projectID := middleware.GetProjectID(c)

	connectorIDParam := c.Param("connector_id")
	connectorID, err := uuid.Parse(connectorIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connector ID"})
		return
	}

	connector, err := h.emailRepo.GetConnector(c.Request.Context(), tenantID, projectID, connectorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get connector"})
		return
	}

	if connector == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Connector not found"})
		return
	}

	// TODO: Implement actual connection testing
	// For now, just return success
	result := map[string]interface{}{
		"status":    "success",
		"message":   "Connection test successful",
		"tested_at": time.Now(),
	}

	c.JSON(http.StatusOK, result)
}

// CreateMailboxRequest represents a request to create an email mailbox
type CreateMailboxRequest struct {
	Address            string               `json:"address" binding:"required,email"`
	DisplayName        *string              `json:"display_name,omitempty"`
	InboundConnectorID uuid.UUID            `json:"inbound_connector_id" binding:"required"`
	RoutingRules       []models.RoutingRule `json:"routing_rules,omitempty"`
	AllowNewTicket     bool                 `json:"allow_new_ticket"`
}

// CreateMailbox creates a new email mailbox
func (h *EmailHandler) CreateMailbox(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	var req CreateMailboxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert routing rules to JSONMap
	routingRules := make(models.JSONMap)
	if len(req.RoutingRules) > 0 {
		rules := make([]map[string]interface{}, len(req.RoutingRules))
		for i, rule := range req.RoutingRules {
			rules[i] = map[string]interface{}{
				"match":      rule.Match,
				"project_id": projectID.String(),
			}
		}
		routingRules["rules"] = rules
	}

	mailbox := &models.EmailMailbox{
		ID:                 uuid.New(),
		TenantID:           tenantID,
		Address:            req.Address,
		DisplayName:        req.DisplayName,
		InboundConnectorID: req.InboundConnectorID,
		RoutingRules:       routingRules,
		AllowNewTicket:     req.AllowNewTicket,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := h.emailRepo.CreateMailbox(c.Request.Context(), mailbox); err != nil {
		fmt.Println("Failed to create mailbox:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create mailbox"})
		return
	}

	c.JSON(http.StatusCreated, mailbox)
}

// ListMailboxes lists all email mailboxes for a tenant
func (h *EmailHandler) ListMailboxes(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	// Get all mailboxes for the tenant (not project-specific)
	mailboxes, err := h.emailRepo.ListMailboxes(c.Request.Context(), tenantID, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list mailboxes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mailboxes": mailboxes})
}

// GetMailbox gets a specific email mailbox
func (h *EmailHandler) GetMailbox(c *gin.Context) {
	tenantIDStr := c.MustGet("tenant_id").(string)
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID format"})
		return
	}

	mailboxIDStr := c.Param("mailbox_id")
	mailboxID, err := uuid.Parse(mailboxIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mailbox ID format"})
		return
	}

	mailbox, err := h.emailRepo.GetMailboxByID(c.Request.Context(), tenantID, mailboxID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get mailbox"})
		return
	}

	if mailbox == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mailbox not found"})
		return
	}

	c.JSON(http.StatusOK, mailbox)
}

// UpdateMailbox updates an email mailbox
func (h *EmailHandler) UpdateMailbox(c *gin.Context) {
	tenantIDStr := c.MustGet("tenant_id").(string)
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID format"})
		return
	}

	mailboxIDStr := c.Param("mailbox_id")
	mailboxID, err := uuid.Parse(mailboxIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mailbox ID format"})
		return
	}

	var req CreateMailboxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing mailbox
	mailbox, err := h.emailRepo.GetMailboxByID(c.Request.Context(), tenantID, mailboxID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get mailbox"})
		return
	}

	if mailbox == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mailbox not found"})
		return
	}

	// Update fields
	mailbox.Address = req.Address
	mailbox.DisplayName = req.DisplayName
	mailbox.InboundConnectorID = req.InboundConnectorID
	mailbox.AllowNewTicket = req.AllowNewTicket
	mailbox.UpdatedAt = time.Now()

	// Convert routing rules to JSONMap
	routingRules := make(models.JSONMap)
	if len(req.RoutingRules) > 0 {
		rules := make([]map[string]interface{}, len(req.RoutingRules))
		for i, rule := range req.RoutingRules {
			rules[i] = map[string]interface{}{
				"match":      rule.Match,
				"project_id": rule.ProjectID.String(),
			}
		}
		routingRules["rules"] = rules
	}
	mailbox.RoutingRules = routingRules

	if err := h.emailRepo.UpdateMailbox(c.Request.Context(), mailbox); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update mailbox"})
		return
	}

	c.JSON(http.StatusOK, mailbox)
}

// DeleteMailbox deletes an email mailbox
func (h *EmailHandler) DeleteMailbox(c *gin.Context) {
	tenantIDStr := c.MustGet("tenant_id").(string)
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID format"})
		return
	}

	mailboxIDStr := c.Param("mailbox_id")
	mailboxID, err := uuid.Parse(mailboxIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mailbox ID format"})
		return
	}

	if err := h.emailRepo.DeleteMailbox(c.Request.Context(), tenantID, mailboxID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete mailbox"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ValidateConnector initiates domain validation for a connector
func (h *EmailHandler) ValidateConnector(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)

	projectID := middleware.GetProjectID(c)

	connectorIDStr := c.Param("connector_id")
	connectorID, err := uuid.Parse(connectorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connector ID format"})
		return
	}

	var req ValidateConnectorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the connector
	connector, err := h.emailRepo.GetConnector(c.Request.Context(), tenantID, projectID, connectorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get connector"})
		return
	}

	if connector == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Connector not found"})
		return
	}

	// Generate and store OTP
	otpKey := fmt.Sprintf("email_validation:%s:%s", tenantID.String(), connectorID.String())

	// Store OTP in Redis with 10 minute expiration
	otp, err := h.redisService.GenerateAndStoreOTP(c.Request.Context(), otpKey, 10*time.Minute)
	fmt.Println("Generated OTP:", otp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate validation code"})
		return
	}

	// Update connector status to validating
	connector.ValidationStatus = models.ValidationStatusValidating
	connector.UpdatedAt = time.Now()

	if err := h.emailRepo.UpdateConnector(c.Request.Context(), connector); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update connector"})
		return
	}

	// Send validation email using the connector's SMTP settings
	err = h.sendValidationEmail(c.Request.Context(), connector, req.Email, otp)
	if err != nil {
		// Log the error but don't fail the request - user can retry
		fmt.Printf("Failed to send validation email: %v\n", err)
		if strings.Contains(err.Error(), "SMTP authentication failed") {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "validation_started",
				"message": "Validation initiated. Please check your email for the verification code.",
				"warning": "Email sending failed, please try again or check your SMTP settings",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send validation email"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "validation_started",
		"message": "Validation email sent to " + req.Email,
	})
}

// VerifyConnectorOTP verifies the OTP for connector validation
func (h *EmailHandler) VerifyConnectorOTP(c *gin.Context) {
	tenantIDStr := c.MustGet("tenant_id").(string)
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID format"})
		return
	}

	projectIDStr := c.Param("project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID format"})
		return
	}

	connectorIDStr := c.Param("connector_id")
	connectorID, err := uuid.Parse(connectorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connector ID format"})
		return
	}

	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the connector
	connector, err := h.emailRepo.GetConnector(c.Request.Context(), tenantID, projectID, connectorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get connector"})
		return
	}

	if connector == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Connector not found"})
		return
	}

	// Check if connector belongs to the project
	if connector.ProjectID == nil || *connector.ProjectID != projectID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Connector does not belong to this project"})
		return
	}

	// Verify OTP using Redis
	otpKey := fmt.Sprintf("email_validation:%s:%s", tenantID.String(), connectorID.String())
	isValid, err := h.redisService.VerifyOTP(c.Request.Context(), otpKey, req.OTP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify OTP"})
		return
	}

	if isValid {
		now := time.Now()
		connector.IsValidated = true
		connector.ValidationStatus = models.ValidationStatusValidated
		connector.LastValidationAt = &now
		connector.ValidationError = nil
		connector.UpdatedAt = now

		if err := h.emailRepo.UpdateConnector(c.Request.Context(), connector); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update connector"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "verified",
			"message": "Email connector has been successfully validated",
		})
	} else {
		// Update connector with failed status
		errorMsg := "Invalid or expired OTP"
		connector.ValidationStatus = models.ValidationStatusFailed
		connector.ValidationError = &errorMsg
		connector.UpdatedAt = time.Now()
		h.emailRepo.UpdateConnector(c.Request.Context(), connector)

		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "invalid_otp",
			"message": "Invalid or expired OTP provided",
		})
	}
}

// sendValidationEmail sends a validation email using the connector's SMTP settings
func (h *EmailHandler) sendValidationEmail(ctx context.Context, connector *models.EmailConnector, recipientEmail, otp string) error {
	// Validate SMTP configuration
	if connector.SMTPHost == nil || connector.SMTPPort == nil {
		return fmt.Errorf("SMTP configuration incomplete: missing host or port")
	}

	if connector.SMTPUsername == nil || connector.SMTPPasswordEnc == nil {
		return fmt.Errorf("SMTP authentication incomplete: missing username or password")
	}

	// Use SMTP username as the from address (typically the authenticated email)
	if connector.SMTPUsername == nil {
		return fmt.Errorf("no valid from address found: SMTP username required")
	}
	fromAddress := connector.SMTPUsername

	// Ensure TLS is enabled for common secure ports
	if connector.SMTPUseTLS == nil {
		useTLS := true
		// Enable TLS for common secure SMTP ports
		if *connector.SMTPPort == 587 || *connector.SMTPPort == 465 || *connector.SMTPPort == 993 {
			connector.SMTPUseTLS = &useTLS
		}
	}

	// Create the validation email message using templates
	templateData := TemplateData{
		OTP: otp,
	}

	// Load HTML template
	htmlBody, err := loadHTMLTemplate("validation_email.html", templateData)
	if err != nil {
		return fmt.Errorf("failed to load HTML template: %w", err)
	}

	// Load text template
	textBody, err := loadTextTemplate("validation_email.txt", templateData)
	if err != nil {
		return fmt.Errorf("failed to load text template: %w", err)
	}

	message := &mail.Message{
		From:     *fromAddress,
		To:       []string{recipientEmail},
		Subject:  "Email Connector Validation - Action Required",
		TextBody: textBody,
		HTMLBody: htmlBody,
		Headers: map[string]string{
			"X-Mailer":   "Hith-EmailConnector-Validator/1.0",
			"X-Priority": "1",
			"Importance": "high",
		},
	}

	// Send the email using mail service
	err = h.mailService.SendValidationEmail(ctx, connector, message)
	if err != nil {
		return fmt.Errorf("failed to send validation email via SMTP: %w", err)
	}

	return nil
}
