package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/service"
)

type AIBuilderHandler struct {
	builder       *service.AIBuilderService
	publicBuilder *service.PublicAIBuilderService
}

type aiBuildRequest struct {
	URL   string `json:"url" binding:"required,url"`
	Depth int    `json:"depth"`
}

func NewAIBuilderHandler(builder *service.AIBuilderService, publicBuilder *service.PublicAIBuilderService) *AIBuilderHandler {
	return &AIBuilderHandler{
		builder:       builder,
		publicBuilder: publicBuilder,
	}
}

// StreamBuild builds AI knowledge base from website URL with streaming response
// @Summary Stream AI knowledge base build
// @Description Build AI knowledge base from website URL and stream the progress. Automatically generates widget, scrapes content, and creates FAQ.
// @Tags ai-builder
// @Accept json
// @Produce text/event-stream
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Param url query string false "Website URL to scrape"
// @Param max_depth query int false "Scraping depth (1-5, default: 3)"
// @Param build body aiBuildRequest false "Build request (alternative to query params)"
// @Success 200 {string} string "Server-sent events stream"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tenants/{tenant_id}/projects/{project_id}/ai/build [post]
func (h *AIBuilderHandler) StreamBuild(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	// Parse URL and depth from query params or request body
	urlParam := c.Query("url")
	depth := 3

	// Try max_depth first (matching public API), fallback to depth
	if depthParam := c.Query("max_depth"); depthParam != "" {
		if value, err := strconv.Atoi(depthParam); err == nil && value >= 1 && value <= 5 {
			depth = value
		}
	} else if depthParam := c.Query("depth"); depthParam != "" {
		if value, err := strconv.Atoi(depthParam); err == nil && value >= 1 && value <= 5 {
			depth = value
		}
	}

	// If URL not in query, try to parse from request body
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

	// Create event channel
	events := make(chan service.AIBuilderEvent)
	ctx := c.Request.Context()

	// Run builder in background using the stable public builder workflow
	go func() {
		defer close(events)
		// Use BuildWidgetForProject with FAQ generation disabled
		_, err := h.publicBuilder.BuildWidgetForProject(ctx, tenantID, projectID, urlParam, depth, false, events)
		if err != nil {
			// If there's an error but no events were sent, send error event
			select {
			case <-ctx.Done():
			case events <- service.AIBuilderEvent{
				Type:      "error",
				Stage:     "internal",
				Message:   "AI builder failed",
				Detail:    err.Error(),
				Data: map[string]any{
					"project_id": projectID.String(),
				},
				Timestamp: time.Now(),
			}:
			}
		}
	}()

	// Set up SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering
	c.Writer.Flush()

	// Stream events to client
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
