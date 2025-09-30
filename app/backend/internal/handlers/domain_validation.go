package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/service"
)

type DomainNameHandler struct {
	domainService *service.DomainNameService
}

func NewDomainValidationHandler(domainService *service.DomainNameService) *DomainNameHandler {
	return &DomainNameHandler{
		domainService: domainService,
	}
}

// CreateDomainValidationRequest represents request to create domain validation
type CreateDomainValidationRequest struct {
	Domain string `json:"domain" binding:"required"`
}

// SendValidationEmailRequest represents request to send validation email
type SendValidationEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// VerifyDomainRequest represents request to verify domain
type VerifyDomainRequest struct {
	Proof string `json:"proof" binding:"required"`
}

// CreateDomainName creates a new domain validation
// @Summary Create domain validation
// @Description Create a new domain validation request for domain ownership verification
// @Tags domain-validation
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Param domain body CreateDomainValidationRequest true "Domain validation request"
// @Success 201 {object} models.EmailDomain
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/domains [post]
func (h *DomainNameHandler) CreateDomainName(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	var req CreateDomainValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validation, err := h.domainService.CreateDomainValidation(c.Request.Context(), tenantID, projectID, req.Domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create domain validation: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, validation)
}

// VerifyDomain verifies domain ownership
// @Summary Verify domain ownership
// @Description Verify domain ownership using provided verification proof
// @Tags domain-validation
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param domain_id path string true "Domain ID"
// @Param proof body VerifyDomainRequest true "Domain verification proof"
// @Success 200 {object} object{message=string,verified_at=string}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/domains/{domain_id}/verify [post]
func (h *DomainNameHandler) VerifyDomain(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)

	domainIDStr := c.Param("domain_id")
	domainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain ID format"})
		return
	}

	var req VerifyDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.domainService.VerifyDomainName(c.Request.Context(), tenantID, domainID, req.Proof)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Domain verified successfully",
		"verified_at": time.Now(),
	})
}

// ListDomainNames lists all domain validations for a project
// @Summary List domain validations
// @Description Retrieve all domain validations for a project
// @Tags domain-validation
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Success 200 {object} object{domains=[]models.EmailDomain}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/domains [get]
func (h *DomainNameHandler) ListDomainNames(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)
	domains, err := h.domainService.GetDomainNames(c.Request.Context(), tenantID, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list domain validations"})
		return
	}

	// Ensure domains is an empty array if nil
	if domains == nil {
		domains = []*models.EmailDomain{}
	}

	// Remove sensitive tokens from response
	for _, domain := range domains {
		domain.ValidationToken = "***"
	}

	c.JSON(http.StatusOK, gin.H{"domains": domains})
}

// DeleteDomainName deletes a domain validation
// @Summary Delete domain validation
// @Description Delete a domain validation by its ID
// @Tags domain-validation
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Param domain_id path string true "Domain ID"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/domains/{domain_id} [delete]
func (h *DomainNameHandler) DeleteDomainName(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)
	domainIDStr, _ := c.Params.Get("domain_id")
	fmt.Println("je;efk -> ", domainIDStr)
	domainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain ID format"})
		return
	}

	err = h.domainService.DeleteDomainName(c.Request.Context(), tenantID, projectID, domainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete domain validation"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
