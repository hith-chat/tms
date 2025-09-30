package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bareuptime/tms/internal/service"
)

type PublicAIBuilderHandler struct {
	publicBuilder *service.PublicAIBuilderService
}

type publicAIBuildRequest struct {
	URL   string `json:"url" binding:"required,url"`
	Depth int    `json:"depth"`
}

func NewPublicAIBuilderHandler(publicBuilder *service.PublicAIBuilderService) *PublicAIBuilderHandler {
	return &PublicAIBuilderHandler{
		publicBuilder: publicBuilder,
	}
}

// StreamBuild builds AI widget for public users with automatic project creation
// @Summary Build public AI widget from website URL
// @Description Automatically creates a public project and builds an AI-powered chat widget by scraping the provided website URL. Streams progress events. Limited to 2 requests per 6 hours per IP address.
// @Tags public-ai-builder
// @Accept json
// @Produce text/event-stream
// @Param url query string false "Website URL to scrape"
// @Param depth query int false "Scraping depth (1-5, default: 3)"
// @Param build body publicAIBuildRequest false "Build request (alternative to query params)"
// @Success 200 {string} string "Server-sent events stream"
// @Failure 400 {object} map[string]string
// @Failure 429 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/public/ai-widget-builder [post]
func (h *PublicAIBuilderHandler) StreamBuild(c *gin.Context) {
	// Parse URL from query param or request body
	urlParam := c.Query("url")
	depth := 3

	if depthParam := c.Query("depth"); depthParam != "" {
		if value, err := strconv.Atoi(depthParam); err == nil && value >= 1 && value <= 5 {
			depth = value
		}
	}

	// If URL not in query, try to parse from request body
	if urlParam == "" {
		var req publicAIBuildRequest
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

	// Run builder in background
	go func() {
		defer close(events)
		projectID, err := h.publicBuilder.BuildPublicWidget(ctx, urlParam, depth, events)
		if err != nil {
			// If there's an error but no events were sent, send error event
			select {
			case <-ctx.Done():
			case events <- service.AIBuilderEvent{
				Type:    "error",
				Stage:   "internal",
				Message: "Widget builder failed",
				Detail:  err.Error(),
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