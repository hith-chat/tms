package handlers

import (
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
