package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/service"
)

type ChatWidgetHandler struct {
	chatWidgetService  *service.ChatWidgetService
	webScrapingService *service.WebScrapingService
	aiService          *service.AIService
}

func NewChatWidgetHandler(chatWidgetService *service.ChatWidgetService, webScrapingService *service.WebScrapingService, aiService *service.AIService) *ChatWidgetHandler {
	return &ChatWidgetHandler{
		chatWidgetService:  chatWidgetService,
		webScrapingService: webScrapingService,
		aiService:          aiService,
	}
}

// CreateChatWidget creates a new chat widget
// @Summary Create chat widget
// @Description Create a new chat widget for a project
// @Tags chat-widget
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Param widget body models.CreateChatWidgetRequest true "Chat widget creation request"
// @Success 201 {object} models.ChatWidget
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/chat-widgets [post]
func (h *ChatWidgetHandler) CreateChatWidget(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	var req models.CreateChatWidgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	widget, err := h.chatWidgetService.CreateChatWidget(c.Request.Context(), tenantID, projectID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chat widget: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, widget)
}

// GetChatWidget gets a chat widget by ID
// @Summary Get chat widget
// @Description Retrieve a chat widget by its ID
// @Tags chat-widget
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Param widget_id path string true "Chat Widget ID"
// @Success 200 {object} models.ChatWidget
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/chat-widgets/{widget_id} [get]
func (h *ChatWidgetHandler) GetChatWidget(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	widgetIDStr := c.Param("widget_id")
	widgetID, err := uuid.Parse(widgetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid widget ID format"})
		return
	}

	widget, err := h.chatWidgetService.GetChatWidget(c.Request.Context(), tenantID, projectID, widgetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get chat widget"})
		return
	}
	if widget == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat widget not found"})
		return
	}

	c.JSON(http.StatusOK, widget)
}

// ListChatWidgets lists all chat widgets for a project
// @Summary List chat widgets
// @Description Retrieve all chat widgets for a project
// @Tags chat-widget
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Success 200 {object} object{widgets=[]models.ChatWidget}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/chat-widgets [get]
func (h *ChatWidgetHandler) ListChatWidgets(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	widgets, err := h.chatWidgetService.ListChatWidgets(c.Request.Context(), tenantID, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list chat widgets: " + err.Error()})
		return
	}

	if widgets == nil {
		widgets = []*models.ChatWidget{}
	}

	c.JSON(http.StatusOK, gin.H{"widgets": widgets})
}

// UpdateChatWidget updates a chat widget
// @Summary Update chat widget
// @Description Update an existing chat widget
// @Tags chat-widget
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Param widget_id path string true "Chat Widget ID"
// @Param widget body models.UpdateChatWidgetRequest true "Chat widget update request"
// @Success 200 {object} models.ChatWidget
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/chat-widgets/{widget_id} [put]
func (h *ChatWidgetHandler) UpdateChatWidget(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	widgetIDStr := c.Param("widget_id")
	widgetID, err := uuid.Parse(widgetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid widget ID format"})
		return
	}

	var req models.UpdateChatWidgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	widget, err := h.chatWidgetService.UpdateChatWidget(c.Request.Context(), tenantID, projectID, widgetID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update chat widget: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, widget)
}

// DeleteChatWidget deletes a chat widget
// @Summary Delete chat widget
// @Description Delete a chat widget by its ID
// @Tags chat-widget
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Param widget_id path string true "Chat Widget ID"
// @Success 204 "No Content"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/chat-widgets/{widget_id} [delete]
func (h *ChatWidgetHandler) DeleteChatWidget(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	widgetIDStr := c.Param("widget_id")
	widgetID, err := uuid.Parse(widgetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid widget ID format"})
		return
	}

	err = h.chatWidgetService.DeleteChatWidget(c.Request.Context(), tenantID, projectID, widgetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete chat widget"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetChatWidgetByPublicId gets a chat widget by public ID
// @Summary Get chat widget by public ID
// @Description Retrieve a chat widget by its public ID (for public access)
// @Tags chat-widget
// @Accept json
// @Produce json
// @Param widget_id path string true "Chat Widget Public ID"
// @Success 200 {object} models.ChatWidget
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /public/chat-widgets/{widget_id} [get]
func (h *ChatWidgetHandler) GetChatWidgetByPublicId(c *gin.Context) {
	widgetIDStr := c.Param("widget_id")
	widgetID, err := uuid.Parse(widgetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid widget ID format"})
		return
	}

	widget, err := h.chatWidgetService.GetChatWidgetById(c.Request.Context(), widgetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get chat widget"})
		return
	}
	if widget == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat widget not found for domain"})
		return
	}

	//typecast widget to widgetPublic
	widgetPublic := models.ChatWidgetPublic(*widget)

	c.JSON(http.StatusOK, widgetPublic)
}

// ScrapeWebsiteTheme scrapes a website and generates theme configuration using AI
// @Summary Scrape website theme
// @Description Scrape a website and generate theme configuration using AI
// @Tags chat-widget
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Param url body object{url=string} true "Website URL to scrape"
// @Success 200 {object} object{theme=object}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/chat-widgets/scrape-theme [post]
func (h *ChatWidgetHandler) ScrapeWebsiteTheme(c *gin.Context) {
	// Get URL from query parameter
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL parameter is required"})
		return
	}

	// Scrape website theme data
	themeData, err := h.webScrapingService.ScrapeWebsiteTheme(c.Request.Context(), url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scrape website: " + err.Error()})
		return
	}

	// Generate theme configuration using AI
	themeConfig, err := h.aiService.GenerateWidgetTheme(c.Request.Context(), themeData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate theme: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, themeConfig)
}

// GetEmbedSnippet returns a lightweight JavaScript loader for a widget
// @Summary Get widget embed snippet
// @Description Returns a tiny JavaScript loader (~100 bytes) that sets widget config and loads the main widget from jsDelivr CDN for optimal performance. The loader is cached for 1 hour and loads the main widget asynchronously.
// @Tags public-chat-widget
// @Produce javascript
// @Param widget_id path string true "Widget ID"
// @Success 200 {string} string "JavaScript loader snippet"
// @Failure 400 {object} map[string]string
// @Router /api/public/chat/widgets/{widget_id}/embed.js [get]
func (h *ChatWidgetHandler) GetEmbedSnippet(c *gin.Context) {
	widgetIDStr := c.Param("widget_id")
	widgetID, err := uuid.Parse(widgetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid widget ID format"})
		return
	}

	// JavaScript snippet that sets the widget config and appends the remote script.
	// This is a tiny loader (~100 bytes) that loads the main widget from jsDelivr CDN.
	// The main widget script is shared across all widgets for excellent caching.
	js := fmt.Sprintf(`(function() {
  window.TMSChatConfig = { widgetId: '%s' };
  var s = document.createElement('script');
  s.src = 'https://cdn.jsdelivr.net/npm/@hith/web-chat/dist/chat-widget.js';
  s.async = true;
  document.head.appendChild(s);
})();`, widgetID)

	// Set caching headers for optimal performance
	// Cache for 1 hour - balances performance with config update flexibility
	c.Header("Content-Type", "application/javascript; charset=utf-8")
	c.Header("Cache-Control", "public, max-age=3600")
	c.Header("X-Content-Type-Options", "nosniff")
	c.String(http.StatusOK, js)
}
