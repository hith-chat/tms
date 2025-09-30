package handlers

import (
	"log"
	"net/http"

	"github.com/bareuptime/tms/internal/service"
	"github.com/gin-gonic/gin"
)

type TenantHandler struct {
	tenantService *service.TenantService
}

func NewTenantHandler(tenantService *service.TenantService) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
	}
}

// ListTenants handles GET /tenants - Admin only
// @Summary List tenants
// @Description Get a list of all tenants (requires super admin permissions)
// @Tags Tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "List of tenants"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid or missing authentication"
// @Failure 403 {object} map[string]interface{} "Forbidden - Insufficient permissions"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /v1/enterprise/tenants [get]
func (h *TenantHandler) ListTenants(c *gin.Context) {
	// Get requestor agent ID from JWT claims
	agentIDInterface, exists := c.Get("agent_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Agent ID not found in token"})
		return
	}

	requestorAgentID, ok := agentIDInterface.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid agent ID format"})
		return
	}

	tenants, err := h.tenantService.ListTenants(c.Request.Context(), requestorAgentID)
	if err != nil {
		log.Printf("Failed to list tenants: %v", err)
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tenants": tenants})
}
