package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/bareuptime/tms/internal/service"
)

type AIBuilderHandler struct {
	builder      *service.AIBuilderService
	tenantRepo   *repo.TenantRepository
	projectRepo  *repo.ProjectRepository
	agentRepo    *repo.AgentRepository
}

type aiBuildRequest struct {
	URL   string `json:"url" binding:"required,url"`
	Depth int    `json:"depth"`
	Email string `json:"email"`
}

func NewAIBuilderHandler(builder *service.AIBuilderService) *AIBuilderHandler {
	return &AIBuilderHandler{builder: builder}
}

func NewPublicAIBuilderHandler(builder *service.AIBuilderService, tenantRepo *repo.TenantRepository, projectRepo *repo.ProjectRepository, agentRepo *repo.AgentRepository) *AIBuilderHandler {
	return &AIBuilderHandler{
		builder:     builder,
		tenantRepo:  tenantRepo,
		projectRepo: projectRepo,
		agentRepo:   agentRepo,
	}
}

// StreamBuild builds AI knowledge base from website URL with streaming response
// @Summary Stream AI knowledge base build
// @Description Build AI knowledge base from website URL and stream the progress
// @Tags ai-builder
// @Accept json
// @Produce text/event-stream
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Param url query string false "Website URL to scrape"
// @Param depth query int false "Scraping depth (1-5, default: 3)"
// @Param build body aiBuildRequest false "Build request (alternative to query params)"
// @Success 200 {string} string "Server-sent events stream"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/ai-builder/stream [post]
func (h *AIBuilderHandler) StreamBuild(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	urlParam := c.Query("url")
	depth := 3

	if depthParam := c.Query("depth"); depthParam != "" {
		if value, err := strconv.Atoi(depthParam); err == nil && value >= 1 && value <= 5 {
			depth = value
		}
	}

	if urlParam == "" {
		var req aiBuildRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url parameter is required"})
			return
		}
		urlParam = req.URL
		if req.Depth >= 1 && req.Depth <= 5 {
			depth = req.Depth
		}
	}

	if urlParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url parameter is required"})
		return
	}

	events := make(chan service.AIBuilderEvent)
	ctx := c.Request.Context()

	go func() {
		defer close(events)
		if err := h.builder.Run(ctx, tenantID, projectID, urlParam, depth, events); err != nil {
			select {
			case <-ctx.Done():
			case events <- service.AIBuilderEvent{
				Type:      "error",
				Stage:     "internal",
				Message:   "AI builder terminated",
				Detail:    err.Error(),
				Timestamp: time.Now(),
			}:
			}
		}
	}()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			return false
		case event, ok := <-events:
			if !ok {
				return false
			}

			payload, err := json.Marshal(event)
			if err != nil {
				return true
			}

			if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
				return false
			}
			if flusher, ok := c.Writer.(http.Flusher); ok {
				flusher.Flush()
			}
			return true
		}
	})
}

// PublicStreamBuild builds AI widget from website URL for public/unauthenticated users with streaming response
// @Summary Public AI widget builder with streaming
// @Description Build AI widget from website URL without authentication and stream the progress
// @Tags public-ai-builder
// @Accept json
// @Produce text/event-stream
// @Param url query string false "Website URL to scrape"
// @Param email query string false "User email address"
// @Param depth query int false "Scraping depth (1-5, default: 3)"
// @Success 200 {string} string "Server-sent events stream"
// @Failure 400 {object} models.ErrorResponse
// @Failure 429 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/public/ai-widget-builder [post]
func (h *AIBuilderHandler) PublicStreamBuild(c *gin.Context) {
	urlParam := c.Query("url")
	emailParam := c.Query("email")
	depth := 3

	if depthParam := c.Query("depth"); depthParam != "" {
		if value, err := strconv.Atoi(depthParam); err == nil && value >= 1 && value <= 5 {
			depth = value
		}
	}

	// Try to get from request body if not in query params
	if urlParam == "" || emailParam == "" {
		var req aiBuildRequest
		if err := c.ShouldBindJSON(&req); err == nil {
			if urlParam == "" {
				urlParam = req.URL
			}
			if emailParam == "" {
				emailParam = req.Email
			}
			if req.Depth >= 1 && req.Depth <= 5 {
				depth = req.Depth
			}
		}
	}

	if urlParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url parameter is required"})
		return
	}

	if emailParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email parameter is required"})
		return
	}

	ctx := c.Request.Context()

	// Get or create tenant and project for this email
	tenantID, projectID, err := h.getOrCreateTenantAndProject(ctx, emailParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to setup account: %v", err)})
		return
	}

	events := make(chan service.AIBuilderEvent)

	go func() {
		defer close(events)
		if err := h.builder.Run(ctx, tenantID, projectID, urlParam, depth, events); err != nil {
			select {
			case <-ctx.Done():
			case events <- service.AIBuilderEvent{
				Type:      "error",
				Stage:     "internal",
				Message:   "AI builder terminated",
				Detail:    err.Error(),
				Timestamp: time.Now(),
			}:
			}
		}
	}()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Flush()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			return false
		case event, ok := <-events:
			if !ok {
				return false
			}

			payload, err := json.Marshal(event)
			if err != nil {
				return true
			}

			if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
				return false
			}
			if flusher, ok := c.Writer.(http.Flusher); ok {
				flusher.Flush()
			}
			return true
		}
	})
}

// getOrCreateTenantAndProject finds an existing tenant/project by email or creates new ones
func (h *AIBuilderHandler) getOrCreateTenantAndProject(ctx context.Context, email string) (uuid.UUID, uuid.UUID, error) {
	if h.tenantRepo == nil || h.projectRepo == nil || h.agentRepo == nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("repositories not initialized")
	}

	// Try to find existing agent by email
	agent, err := h.agentRepo.GetAgentByEmail(email)
	if err == nil && agent != nil {
		// Agent exists, get their tenant and first project
		tenantID := agent.TenantID

		// Get first project for this tenant
		projects, err := h.projectRepo.GetProjectsByTenantID(tenantID)
		if err == nil && len(projects) > 0 {
			return tenantID, projects[0].ID, nil
		}

		// Create a default project if none exists
		project := &models.Project{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        "Widget Builder Project",
			Description: "Auto-created project for widget builder",
		}
		if err := h.projectRepo.CreateProject(project); err != nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("failed to create project: %w", err)
		}
		return tenantID, project.ID, nil
	}

	// Agent doesn't exist, create new tenant and project
	tenantID := uuid.New()
	tenant := &models.Tenant{
		ID:   tenantID,
		Name: email, // Use email as tenant name
	}
	if err := h.tenantRepo.CreateTenant(tenant); err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Create default project
	projectID := uuid.New()
	project := &models.Project{
		ID:          projectID,
		TenantID:    tenantID,
		Name:        "Widget Builder Project",
		Description: "Auto-created project for widget builder",
	}
	if err := h.projectRepo.CreateProject(project); err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Create agent
	agent = &models.Agent{
		ID:       uuid.New(),
		TenantID: tenantID,
		Email:    email,
		Name:     email,
		Role:     "admin", // Give admin role for new signups
	}
	if err := h.agentRepo.CreateAgent(agent); err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return tenantID, projectID, nil
}
