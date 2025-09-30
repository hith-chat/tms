package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AgentHandler struct {
	agentService *service.AgentService
}

func NewAgentHandler(agentService *service.AgentService) *AgentHandler {
	return &AgentHandler{
		agentService: agentService,
	}
}

// ListAgents handles GET /agents
// @Summary List agents
// @Description Get a paginated list of agents in the tenant
// @Tags Agents
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID" format(uuid)
// @Param search query string false "Search agents by name or email"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Number of agents per page" minimum(1) maximum(100) default(50)
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "List of agents with pagination info"
// @Failure 400 {object} map[string]interface{} "Bad request - Invalid query parameters"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid or missing authentication"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /v1/tenants/{tenant_id}/agents [get]
func (h *AgentHandler) ListAgents(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	agentID := middleware.GetAgentID(c)
	isTenantAdmin := middleware.IsTenantAdmin(c)

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}

	if isTenantAdmin {
		agentID = uuid.Nil
	}

	req := service.ListAgentsRequest{
		Search: c.Query("search"),
		Cursor: c.Query("cursor"),
		Limit:  limit,
	}

	agents, nextCursor, err := h.agentService.ListAgents(c.Request.Context(), tenantID, agentID, req)
	if err != nil {
		log.Printf("Failed to list agents: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agents"})
		return
	}

	response := gin.H{
		"agents": agents,
	}
	if nextCursor != "" {
		response["next_cursor"] = nextCursor
	}

	c.JSON(http.StatusOK, response)
}

// DeleteAgent handles DELETE /agents/:agent_id
// @Summary Delete agent
// @Description Delete an agent from the tenant (requires tenant admin permissions)
// @Tags Agents
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID" format(uuid)
// @Param agent_id path string true "Agent ID" format(uuid)
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Agent deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - Invalid agent ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid or missing authentication"
// @Failure 403 {object} map[string]interface{} "Forbidden - Insufficient permissions"
// @Failure 404 {object} map[string]interface{} "Agent not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /v1/tenants/{tenant_id}/agents/{agent_id} [delete]
func (h *AgentHandler) DeleteAgent(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	agentIDStr := c.Param("agent_id")

	// Get requestor agent ID from JWT claims
	requestorAgentID, exists := c.Get("agent_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Agent ID not found in token"})
		return
	}

	requestorAgentIDStr, ok := requestorAgentID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid agent ID format"})
		return
	}

	err := h.agentService.DeleteAgent(c.Request.Context(), tenantIDStr, agentIDStr, requestorAgentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agent deleted successfully"})
}

// CreateAgent handles POST /agents
// @Summary Create a new agent
// @Description Create a new agent in the tenant (requires tenant admin permissions)
// @Tags Agents
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID" format(uuid)
// @Param agent body service.CreateAgentRequest true "Agent creation data"
// @Security BearerAuth
// @Success 201 {object} models.Agent "Agent created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - Invalid input data"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid or missing authentication"
// @Failure 403 {object} map[string]interface{} "Forbidden - Insufficient permissions"
// @Failure 409 {object} map[string]interface{} "Conflict - Agent already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /v1/tenants/{tenant_id}/agents [post]
func (h *AgentHandler) CreateAgent(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	creatorAgentID := middleware.GetAgentID(c)
	isTenantAdmin := middleware.IsTenantAdmin(c)
	if !isTenantAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	var req service.CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if req.Name == "" || req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name and email are required"})
		return
	}

	agent, err := h.agentService.CreateAgent(c.Request.Context(), tenantID, creatorAgentID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create agent"})
		return
	}

	c.JSON(http.StatusCreated, agent)
}

// GetAgent handles GET /agents/:agent_id
// @Summary Get agent details
// @Description Retrieve detailed information about a specific agent
// @Tags Agents
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID" format(uuid)
// @Param agent_id path string true "Agent ID" format(uuid)
// @Security BearerAuth
// @Success 200 {object} models.Agent "Agent details"
// @Failure 400 {object} map[string]interface{} "Bad request - Invalid agent ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid or missing authentication"
// @Failure 403 {object} map[string]interface{} "Forbidden - Insufficient permissions"
// @Failure 404 {object} map[string]interface{} "Agent not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /v1/tenants/{tenant_id}/agents/{agent_id} [get]
func (h *AgentHandler) GetAgent(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	paramAgentIDStr := c.Param("agent_id")
	jwtAgentID := middleware.GetAgentID(c)
	isTenantAdmin := middleware.IsTenantAdmin(c)

	paramAgentID, err := uuid.Parse(paramAgentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
		return
	}

	if !isTenantAdmin && paramAgentID != jwtAgentID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	agent, err := h.agentService.GetAgent(c.Request.Context(), tenantID, paramAgentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	c.JSON(http.StatusOK, agent)
}

// UpdateAgent handles PATCH /agents/:agent_id
// @Summary Update agent information
// @Description Update an existing agent's information
// @Tags Agents
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID" format(uuid)
// @Param agent_id path string true "Agent ID" format(uuid)
// @Param agent body service.UpdateAgentRequest true "Agent update data"
// @Security BearerAuth
// @Success 200 {object} models.Agent "Agent updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - Invalid input data"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid or missing authentication"
// @Failure 403 {object} map[string]interface{} "Forbidden - Insufficient permissions"
// @Failure 404 {object} map[string]interface{} "Agent not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /v1/tenants/{tenant_id}/agents/{agent_id} [patch]
func (h *AgentHandler) UpdateAgent(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	paramsAgentIDStr := c.Param("agent_id")
	isTenantAdmin := middleware.IsTenantAdmin(c)
	paramsAgentID, err := uuid.Parse(paramsAgentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
		return
	}
	requestAgentID := middleware.GetAgentID(c)

	if !isTenantAdmin && paramsAgentID != requestAgentID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	var req service.UpdateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	agent, statusCode, err := h.agentService.UpdateAgent(c.Request.Context(), tenantID, requestAgentID, paramsAgentID, req)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": "Failed to update agent"})
		return
	}

	c.JSON(http.StatusOK, agent)
}

// AssignRole handles POST /agents/:agent_id/roles
// @Summary Assign role to agent
// @Description Assign a role to an agent (requires tenant admin permissions)
// @Tags Agents
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID" format(uuid)
// @Param agent_id path string true "Agent ID" format(uuid)
// @Param role body service.AssignRoleRequest true "Role assignment data"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Role assigned successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - Invalid input data"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid or missing authentication"
// @Failure 403 {object} map[string]interface{} "Forbidden - Insufficient permissions"
// @Failure 404 {object} map[string]interface{} "Agent not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /v1/tenants/{tenant_id}/agents/{agent_id}/roles [post]
func (h *AgentHandler) AssignRole(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	isTenantAdmin := middleware.IsTenantAdmin(c)

	if !isTenantAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	agentIDStr := c.Param("agent_id")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
		return
	}
	assignerAgentID := middleware.GetAgentID(c)

	var req service.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err = h.agentService.AssignRole(c.Request.Context(), tenantID, agentID, assignerAgentID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role assigned successfully"})
}

// RemoveRole handles DELETE /agents/:agent_id/roles
// @Summary Remove role from agent
// @Description Remove a role from an agent (requires tenant admin permissions)
// @Tags Agents
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID" format(uuid)
// @Param agent_id path string true "Agent ID" format(uuid)
// @Param role body service.RemoveRoleRequest true "Role removal data"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Role removed successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - Invalid input data"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid or missing authentication"
// @Failure 403 {object} map[string]interface{} "Forbidden - Insufficient permissions"
// @Failure 404 {object} map[string]interface{} "Agent not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /v1/tenants/{tenant_id}/agents/{agent_id}/roles [delete]
func (h *AgentHandler) RemoveRole(c *gin.Context) {

	tenantID := middleware.GetTenantID(c)
	agentIDStr := c.Param("agent_id")

	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
		return
	}

	var req service.RemoveRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err = h.agentService.RemoveRole(c.Request.Context(), tenantID, agentID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role removed successfully"})
}

// GetAgentRoles handles GET /agents/:agent_id/roles
// @Summary Get agent roles
// @Description Retrieve all roles assigned to an agent
// @Tags Agents
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID" format(uuid)
// @Param agent_id path string true "Agent ID" format(uuid)
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "List of agent roles"
// @Failure 400 {object} map[string]interface{} "Bad request - Invalid agent ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid or missing authentication"
// @Failure 403 {object} map[string]interface{} "Forbidden - Insufficient permissions"
// @Failure 404 {object} map[string]interface{} "Agent not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /v1/tenants/{tenant_id}/agents/{agent_id}/roles [get]
func (h *AgentHandler) GetAgentRoles(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	agentIDStr := c.Param("agent_id")
	isTenantAdmin := middleware.IsTenantAdmin(c)

	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
		return
	}

	if isTenantAdmin {
		agentID = uuid.Nil
	}

	roles, err := h.agentService.GetAgentRoles(c.Request.Context(), tenantID, agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agent roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// AssignToProject handles POST /agents/:agent_id/projects/:project_id
func (h *AgentHandler) AssignToProject(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	agentIDStr := c.Param("agent_id")
	projectIDStr := c.Param("project_id")

	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
		return
	}
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID format"})
		return
	}
	var req service.AssignToProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Set the project ID from URL parameter
	req.ProjectID = projectID

	err = h.agentService.AssignToProject(c.Request.Context(), tenantID, agentID, req)
	if err != nil {
		log.Printf("Failed to assign agent to project: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign agent to project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agent assigned to project successfully"})
}

// RemoveFromProject handles DELETE /agents/:agent_id/projects/:project_id
func (h *AgentHandler) RemoveFromProject(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	agentIDStr := c.Param("agent_id")
	projectIDStr := c.Param("project_id")

	// Get remover agent ID from JWT claims
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
		return
	}
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID format"})
		return
	}

	err = h.agentService.RemoveFromProject(c.Request.Context(), tenantID, agentID, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove agent from project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agent removed from project successfully"})
}

// GetAgentProjects handles GET /agents/:agent_id/projects
func (h *AgentHandler) GetAgentProjects(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	agentIDStr := c.Param("agent_id")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
		return
	}

	isTenantAdmin := middleware.IsTenantAdmin(c)
	requestAgent := middleware.GetAgentID(c)

	if !isTenantAdmin {
		if agentID != requestAgent {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}
	}
	projects, err := h.agentService.GetAgentProjects(c.Request.Context(), tenantID, agentID)
	if err != nil {
		// Log the error for debugging purposes
		log.Printf("Failed to get agent projects: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agent projects"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"projects": projects})
}
