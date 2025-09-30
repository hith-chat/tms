package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/redis"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/bareuptime/tms/internal/service"
)

type ChatSessionHandler struct {
	chatSessionService *service.ChatSessionService
	chatWidgetService  *service.ChatWidgetService
	redisClient        *redis.Service
}

func NewChatSessionHandler(chatSessionService *service.ChatSessionService, chatWidgetService *service.ChatWidgetService, redisClient *redis.Service) *ChatSessionHandler {
	return &ChatSessionHandler{
		chatSessionService: chatSessionService,
		chatWidgetService:  chatWidgetService,
		redisClient:        redisClient,
	}
}

// GetChatSession gets a chat session by ID
// @Summary Get chat session
// @Description Retrieve a specific chat session by its ID
// @Tags chat-sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Param session_id path string true "Chat Session ID"
// @Success 200 {object} models.ChatSession
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/chat-sessions/{session_id} [get]
func (h *ChatSessionHandler) GetChatSession(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID format"})
		return
	}

	session, err := h.chatSessionService.GetChatSession(c.Request.Context(), tenantID, projectID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get chat session"})
		return
	}
	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// ListChatSessions lists chat sessions for a project
// @Summary List chat sessions
// @Description Retrieve a list of chat sessions for a project with optional filtering
// @Tags chat-sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Param status query string false "Filter by session status"
// @Param agent_id query string false "Filter by assigned agent ID"
// @Param limit query int false "Number of sessions to return (default: 50)"
// @Param offset query int false "Number of sessions to skip (default: 0)"
// @Success 200 {object} object{sessions=[]models.ChatSession,total=int}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/chat-sessions [get]
func (h *ChatSessionHandler) ListChatSessions(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	// Parse query parameters
	filters := repo.ChatSessionFilters{}

	if status := c.Query("status"); status != "" {
		filters.Status = status
	}

	if agentIDStr := c.Query("assigned_agent_id"); agentIDStr != "" {
		if agentID, err := uuid.Parse(agentIDStr); err == nil {
			filters.AssignedAgentID = &agentID
		}
	}

	if widgetIDStr := c.Query("widget_id"); widgetIDStr != "" {
		if widgetID, err := uuid.Parse(widgetIDStr); err == nil {
			filters.WidgetID = &widgetID
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	}

	sessions, err := h.chatSessionService.ListChatSessions(c.Request.Context(), tenantID, projectID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list chat sessions: " + err.Error()})
		return
	}

	if sessions == nil {
		sessions = []*models.ChatSession{}
	}

	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

// AssignAgent assigns an agent to a chat session
// @Summary Assign agent to chat session
// @Description Assign a specific agent to handle a chat session
// @Tags chat-sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Param session_id path string true "Chat Session ID"
// @Param assignment body object{agent_id=string} true "Agent assignment request"
// @Success 200 {object} models.ChatSession
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/chat-sessions/{session_id}/assign [post]
func (h *ChatSessionHandler) AssignAgent(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID format"})
		return
	}

	var req models.AssignChatSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.chatSessionService.AssignAgent(c.Request.Context(), tenantID, projectID, sessionID, req.AgentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign agent: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agent assigned successfully"})
}

// GetChatMessages gets messages for a chat session
// @Summary Get chat messages
// @Description Retrieve all messages for a specific chat session
// @Tags chat-sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Param session_id path string true "Chat Session ID"
// @Success 200 {object} object{messages=[]models.ChatMessage}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/chat-sessions/{session_id}/messages [get]
func (h *ChatSessionHandler) GetChatMessages(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID format"})
		return
	}

	includePrivate := c.Query("include_private") == "true"

	messages, err := h.chatSessionService.GetChatMessages(c.Request.Context(), tenantID, projectID, sessionID, includePrivate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	if messages == nil {
		messages = []*models.ChatMessage{}
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// MarkAgentMessagesAsRead marks agent messages as read
// @Summary Mark agent messages as read
// @Description Mark all agent messages in a chat session as read by the customer
// @Tags chat-sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Param session_id path string true "Chat Session ID"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/chat-sessions/{session_id}/mark-agent-messages-read [post]
func (h *ChatSessionHandler) MarkAgentMessagesAsRead(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)
	sessionID := middleware.GetSessionID(c)
	messageIDStr := c.Param("message_id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID format"})
		return
	}

	err = h.chatSessionService.MarkAgentMessagesAsRead(c.Request.Context(), tenantID, projectID, sessionID, messageID, "agent")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark messages as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Messages marked as read"})
}

// IsCustomerOnline checks if customer is online in chat session
// @Summary Check customer online status
// @Description Check if a customer is currently online in a chat session
// @Tags chat-sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param session_id header string true "Session ID"
// @Success 200 {object} object{online=boolean}
// @Failure 500 {object} models.ErrorResponse
// @Router /api/chat-sessions/customer-online [get]
func (h *ChatSessionHandler) IsCustomerOnline(c *gin.Context) {
	sessionID := middleware.GetSessionID(c)
	sessionKey := fmt.Sprintf("livechat:session:%s", sessionID)

	connIDs, err := h.redisClient.GetClient().SMembers(c.Request.Context(), sessionKey).Result()
	if err != nil {
		log.Error().Err(err).Str("session_id", sessionID.String()).Msg("Failed to get session connections for delivery")
		return
	}
	status := "offline"
	if len(connIDs) > 0 {
		status = "online"
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}

// MarkVisitorMessagesAsRead marks messages as read (visitor endpoint)
// @Summary Mark visitor messages as read
// @Description Mark messages as read from the visitor/customer perspective
// @Tags chat-sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param session_id header string true "Session ID"
// @Param message_id path string true "Message ID"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/chat-sessions/visitor/mark-messages-read/{message_id} [post]
func (h *ChatSessionHandler) MarkVisitorMessagesAsRead(c *gin.Context) {
	sessionID := middleware.GetSessionID(c)
	messageID := c.Param("message_id")
	messageIDParsed, err := uuid.Parse(messageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID format"})
		return
	}

	err = h.chatSessionService.MarkVisitorMessagesAsRead(c.Request.Context(), sessionID, messageIDParsed, "visitor")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark messages as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Messages marked as read"})
}

// EscalateSession escalates a chat session to human agents with alarm notification
// @Summary Escalate chat session
// @Description Escalate a chat session to human agents with alarm notification
// @Tags chat-sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param project_id header string true "Project ID"
// @Param agent_id header string true "Agent ID (who is escalating)"
// @Param session_id path string true "Chat Session ID"
// @Param escalation body object{reason=string} true "Escalation reason"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/chat-sessions/{session_id}/escalate [post]
func (h *ChatSessionHandler) EscalateSession(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)
	agentID := middleware.GetAgentID(c) // Who is escalating

	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID format"})
		return
	}

	var req models.EscalateChatSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.chatSessionService.EscalateSession(
		c.Request.Context(),
		tenantID,
		projectID,
		sessionID,
		&req,
		agentID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to escalate session: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
