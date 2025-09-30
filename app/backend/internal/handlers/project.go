package handlers

import (
	"net/http"

	"github.com/bareuptime/tms/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	projectService *service.ProjectService
}

func NewProjectHandler(projectService *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

// ListProjects godoc
// @Summary List projects for agent
// @Description Get list of projects accessible to the authenticated agent
// @Tags projects
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Success 200 {array} db.Project
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tenants/{tenant_id}/projects [get]
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	tenantID, _ := uuid.Parse(c.Param("tenant_id"))
	// Get validated agent ID from middleware
	agentID, _ := uuid.Parse(c.GetString("agent_id")) // Already validated in middleware
	isTenantAdmin := c.GetBool("is_tenant_admin")     // Check if agent is tenant admin

	// Get validated claims from middleware

	// Simple check: if agent is tenant_admin for this tenant, return all projects
	if isTenantAdmin {
		projects, err := h.projectService.ListProjects(c.Request.Context(), tenantID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list projects"})
			return
		}
		c.JSON(http.StatusOK, projects)
		return
	}

	// Otherwise, return projects for the agent
	projects, err := h.projectService.ListProjectsForAgent(c.Request.Context(), tenantID, agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list projects"})
		return
	}

	c.JSON(http.StatusOK, projects)
}

// hasAccess checks if the agent has access to the given tenant based on role bindings.
func hasAccess(roleBindings []string, tenantID uuid.UUID) bool {
	tenantIDStr := tenantID.String()
	for _, binding := range roleBindings {
		if binding == tenantIDStr {
			return true
		}
	}
	return false
}

// GetProject godoc
// @Summary Get project details
// @Description Get details of a specific project
// @Tags projects
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param project_id path string true "Project ID"
// @Success 200 {object} db.Project
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// GetProject gets a specific project by ID
// @Summary Get project
// @Description Retrieve a specific project by its ID
// @Tags projects
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id path string true "Tenant ID"
// @Param project_id path string true "Project ID"
// @Success 200 {object} models.Project
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tenants/{tenant_id}/projects/{project_id} [get]
func (h *ProjectHandler) GetProject(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id"})
		return
	}

	projectIDStr := c.Param("project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	project, err := h.projectService.GetProject(c.Request.Context(), tenantID, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get project"})
		return
	}

	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

type CreateProjectRequest struct {
	Key  string `json:"key" binding:"required,min=1,max=50"`
	Name string `json:"name" binding:"required,min=1,max=255"`
}

// CreateProject godoc
// @Summary Create a new project
// @Description Create a new project within the tenant
// @Tags projects
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param request body CreateProjectRequest true "Project creation request"
// @Success 201 {object} db.Project
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tenants/{tenant_id}/projects [post]
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id"})
		return
	}

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.CreateProject(c.Request.Context(), tenantID, req.Key, req.Name)
	if err != nil {
		if err == service.ErrProjectLimitReached {
			c.JSON(http.StatusForbidden, gin.H{"error": "Maximum of 5 projects allowed. To add more projects, please contact support@hith.chat"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, project)
}

type UpdateProjectRequest struct {
	Key    string `json:"key" binding:"required,min=1,max=50"`
	Name   string `json:"name" binding:"required,min=1,max=255"`
	Status string `json:"status" binding:"required,oneof=active inactive"`
}

// UpdateProject godoc
// @Summary Update project
// @Description Update an existing project
// @Tags projects
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param project_id path string true "Project ID"
// @Param request body UpdateProjectRequest true "Project update request"
// @Success 200 {object} db.Project
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tenants/{tenant_id}/projects/{project_id} [put]
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id"})
		return
	}

	projectIDStr := c.Param("project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.UpdateProject(c.Request.Context(), tenantID, projectID, req.Key, req.Name, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update project"})
		return
	}

	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// DeleteProject godoc
// @Summary Delete project
// @Description Delete an existing project
// @Tags projects
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param project_id path string true "Project ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tenants/{tenant_id}/projects/{project_id} [delete]
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id"})
		return
	}

	projectIDStr := c.Param("project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	err = h.projectService.DeleteProject(c.Request.Context(), tenantID, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete project"})
		return
	}

	c.Status(http.StatusNoContent)
}
