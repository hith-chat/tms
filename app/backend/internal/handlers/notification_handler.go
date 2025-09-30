package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/service"
)

type NotificationHandler struct {
	notificationService *service.NotificationService
}

func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// GetNotifications retrieves all notifications for the authenticated user
// @Summary Get notifications
// @Description Retrieve notifications for the authenticated agent with pagination
// @Tags notifications
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param agent_id header string true "Agent ID"
// @Param limit query int false "Number of notifications to return (default: 20, max: 100)"
// @Param offset query int false "Number of notifications to skip (default: 0)"
// @Success 200 {object} object{notifications=[]models.Notification,total=int,unread_count=int}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/notifications [get]
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	tenantIDStr := c.MustGet("tenant_id").(string)
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID format"})
		return
	}
	agentIDStr := c.GetString("agent_id")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	notifications, err := h.notificationService.GetNotifications(c.Request.Context(), tenantID, agentID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve notifications: " + err.Error()})
		return
	}
	//convert null to array
	if notifications == nil {
		notifications = []models.Notification{}
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"limit":         limit,
		"offset":        offset,
	})
}

// GetNotificationCount retrieves notification count for the authenticated agent
func (h *NotificationHandler) GetNotificationCount(c *gin.Context) {
	tenantIDStr := c.MustGet("tenant_id").(string)
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID format"})
		return
	}
	agentIDStr := c.GetString("agent_id")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	count, err := h.notificationService.GetNotificationCount(c.Request.Context(), tenantID, agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve notification count"})
		return
	}

	c.JSON(http.StatusOK, count)
}

// MarkNotificationAsRead marks a specific notification as read
// MarkNotificationAsRead marks a specific notification as read
// @Summary Mark notification as read
// @Description Mark a specific notification as read by its ID
// @Tags notifications
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param agent_id header string true "Agent ID"
// @Param id path string true "Notification ID"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/notifications/{id}/read [put]
func (h *NotificationHandler) MarkNotificationAsRead(c *gin.Context) {
	tenantIDStr := c.MustGet("tenant_id").(string)
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID format"})
		return
	}
	agentIDStr := c.GetString("agent_id")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	notificationIDStr := c.Param("notification_id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	err = h.notificationService.MarkNotificationAsRead(c.Request.Context(), tenantID, agentID, notificationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// MarkAllNotificationsAsRead marks all notifications as read for the authenticated agent
func (h *NotificationHandler) MarkAllNotificationsAsRead(c *gin.Context) {
	tenantIDStr := c.MustGet("tenant_id").(string)
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID format"})
		return
	}
	agentIDStr := c.GetString("agent_id")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	err = h.notificationService.MarkAllNotificationsAsRead(c.Request.Context(), tenantID, agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark all notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
