package handlers

import (
	"net/http"
	"strconv"

	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CustomerHandler struct {
	customerService *service.CustomerService
}

func NewCustomerHandler(customerService *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{customerService: customerService}
}

// CreateCustomer handles POST /tenants/:tenant_id/customers
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	creatorAgentID := middleware.GetAgentID(c)
	isTenantAdmin := middleware.IsTenantAdmin(c)

	// Only tenant admins or agents with write permission should be allowed by service RBAC,
	// but protect early for UI flows where only admins create customers.
	if !isTenantAdmin {
		// allow non-admins and let service RBAC re-check permissions as well
	}

	// Accept phone either in metadata or as top-level optional field
	var body struct {
		Email    string            `json:"email" binding:"required,email"`
		Name     string            `json:"name" binding:"required,min=1,max=255"`
		Phone    *string           `json:"phone,omitempty"`
		Metadata map[string]string `json:"metadata,omitempty"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure metadata exists and include phone if provided
	metadata := body.Metadata
	if metadata == nil {
		metadata = make(map[string]string)
	}
	if body.Phone != nil && *body.Phone != "" {
		metadata["phone"] = *body.Phone
	}

	req := service.CreateCustomerRequest{
		Email:    body.Email,
		Name:     body.Name,
		Metadata: metadata,
	}

	customer, err := h.customerService.CreateCustomer(c.Request.Context(), tenantID, creatorAgentID, req)
	if err != nil {
		// simple error classification
		if err.Error() == "customer with email "+body.Email+" already exists" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create customer"})
		return
	}

	c.JSON(http.StatusCreated, customer)
}

// UpdateCustomer handles PUT /tenants/:tenant_id/customers/:customer_id
func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	requestorAgentID := middleware.GetAgentID(c)

	customerIDStr := c.Param("customer_id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer_id"})
		return
	}

	var body struct {
		Name     *string            `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
		Phone    *string            `json:"phone,omitempty"`
		Metadata *map[string]string `json:"metadata,omitempty"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Merge phone into metadata if provided
	var metadataPtr *map[string]string
	if body.Metadata != nil {
		metadataPtr = body.Metadata
	} else if body.Phone != nil {
		m := make(map[string]string)
		if *body.Phone != "" {
			m["phone"] = *body.Phone
		}
		metadataPtr = &m
	}

	req := service.UpdateCustomerRequest{
		Name:     body.Name,
		Metadata: metadataPtr,
	}

	customer, err := h.customerService.UpdateCustomer(c.Request.Context(), tenantID, customerID, requestorAgentID, req)
	if err != nil {
		if err.Error() == "customer not found" || err.Error() == "customer not found: sql: no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update customer"})
		return
	}

	c.JSON(http.StatusOK, customer)
}

// DeleteCustomer handles DELETE /tenants/:tenant_id/customers/:customer_id
func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	requestorAgentID := middleware.GetAgentID(c)

	customerIDStr := c.Param("customer_id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer_id"})
		return
	}

	err = h.customerService.DeleteCustomer(c.Request.Context(), tenantID, customerID, requestorAgentID)
	if err != nil {
		if err.Error() == "insufficient permissions to delete customer" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to delete customer"})
			return
		}
		if err.Error() == "customer not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete customer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully"})
}

// ListCustomers handles GET /tenants/:tenant_id/customers
func (h *CustomerHandler) ListCustomers(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	agentID := middleware.GetAgentID(c)

	// Parse query parameters
	email := c.Query("email")
	search := c.Query("search")
	cursor := c.Query("cursor")
	limitStr := c.DefaultQuery("limit", "50")

	limit := 50
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	req := service.ListCustomersRequest{
		Email:  email,
		Search: search,
		Cursor: cursor,
		Limit:  limit,
	}

	customers, nextCursor, err := h.customerService.ListCustomers(c.Request.Context(), tenantID, agentID, req)
	if err != nil {
		if err.Error() == "insufficient permissions" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to list customers"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list customers"})
		return
	}

	response := gin.H{
		"customers": customers,
	}
	if nextCursor != "" {
		response["next_cursor"] = nextCursor
	}

	c.JSON(http.StatusOK, response)
}
