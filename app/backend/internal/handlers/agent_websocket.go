package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/service"
	ws "github.com/bareuptime/tms/internal/websocket"
)

// AgentWebSocketRequest represents WebSocket message types that agents can send
// @Description WebSocket message format for agent communications
type AgentWebSocketRequest struct {
	Type            string      `json:"type" example:"chat_message" enums:"chat_message,typing_start,typing_stop,ping,session_subscribe,session_unsubscribe"`
	ClientSessionID *string     `json:"client_session_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	AgentSessionID  *string     `json:"agent_session_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	ProjectID       *string     `json:"project_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	Data            interface{} `json:"data,omitempty"`
	Timestamp       time.Time   `json:"timestamp" example:"2023-09-29T10:00:00Z"`
	MessageID       *string     `json:"message_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// AgentWebSocketResponse represents WebSocket message types that agents receive
// @Description WebSocket message format for agent responses
type AgentWebSocketResponse struct {
	Type      string      `json:"type" example:"agent_connected" enums:"agent_connected,pong,error,chat_message,typing_start,typing_stop,notification"`
	SessionID string      `json:"session_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Data      interface{} `json:"data"`
	FromType  string      `json:"from_type" example:"agent" enums:"agent,customer,system"`
	Timestamp time.Time   `json:"timestamp" example:"2023-09-29T10:00:00Z"`
}

// ChatMessageData represents the data structure for chat messages
// @Description Chat message data structure
type ChatMessageData struct {
	Content     string `json:"content" example:"Hello, how can I help you?"`
	MessageType string `json:"message_type" example:"text" enums:"text,file,image"`
}

// AgentConnectedData represents the data for agent connection confirmation
// @Description Agent connection confirmation data
type AgentConnectedData struct {
	Type    string `json:"type" example:"connected"`
	Message string `json:"message" example:"Connected to console"`
	AgentID string `json:"agent_id" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// ErrorData represents error message data
// @Description Error message data structure
type ErrorData struct {
	Error   string `json:"error" example:"Failed to send message"`
	Details string `json:"details,omitempty" example:"Session not found"`
}

// TypingIndicatorData represents typing indicator data
// @Description Typing indicator data structure
type TypingIndicatorData struct {
	AuthorType string `json:"author_type" example:"agent" enums:"agent,customer"`
	AuthorName string `json:"author_name" example:"John Doe"`
	IsTyping   bool   `json:"is_typing" example:"true"`
}

type AgentWebSocketHandler struct {
	chatSessionService *service.ChatSessionService
	connectionManager  *ws.ConnectionManager
	agentService       *service.AgentService
	chatWSHandler      *ChatWebSocketHandler // Reference to main WebSocket handler
}

func NewAgentWebSocketHandler(chatSessionService *service.ChatSessionService, connectionManager *ws.ConnectionManager, agentService *service.AgentService) *AgentWebSocketHandler {
	handler := &AgentWebSocketHandler{
		chatSessionService: chatSessionService,
		connectionManager:  connectionManager,
		agentService:       agentService,
		chatWSHandler:      nil, // Will be set later
	}

	return handler
}

// HandleAgentWebSocket handles a single WebSocket connection per agent for all sessions
// @Summary Establish WebSocket connection for agent
// @Description Establishes a WebSocket connection for an agent to handle real-time chat across all sessions they have access to. The agent can send messages, typing indicators, and subscribe to specific chat sessions.
// @Tags Agent WebSocket
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID" format(uuid)
// @Security BearerAuth
// @Success 101 {object} AgentWebSocketResponse "WebSocket connection established"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid or missing JWT token"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /v1/tenants/{tenant_id}/chat/agent/ws [get]
func (h *AgentWebSocketHandler) HandleAgentWebSocket(c *gin.Context) {
	agentID := middleware.GetAgentID(c)
	tenantID := middleware.GetTenantID(c)
	// Note: No project_id needed for global agent connection

	// Get agent information
	agent, err := h.agentService.GetAgent(c.Request.Context(), tenantID, agentID)
	if err != nil {
		log.Printf("Failed to get agent information: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Agent not found"})
		return
	}

	projects, err := h.agentService.GetAgentProjectsList(c.Request.Context(), tenantID, agentID)
	if err != nil {
		log.Printf("Failed to get agent projects: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agent projects"})
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// Register connection with the enterprise connection manager (with WebSocket)
	connectionID, err := h.connectionManager.AddConnection(
		ws.ConnectionTypeAgent,
		agentID,
		projects,
		&agentID,
		conn, // Pass WebSocket connection to connection manager
	)
	if err != nil {
		log.Printf("Failed to register agent global connection: %v", err)
		return
	}

	defer func() {
		h.connectionManager.RemoveConnection(connectionID)
	}()

	// Send welcome message to agent
	welcomeMsg := &ws.Message{
		Type:      "agent_connected",
		SessionID: agentID,
		Data: json.RawMessage(`{
			"type": "connected",
			"message": "Connected to console",
			"agent_id": "` + agentID.String() + `"
		}`),
		FromType:     ws.ConnectionTypeAgent,
		DeliveryType: ws.Self,
	}
	h.connectionManager.SendToConnection(connectionID, welcomeMsg)

	// Set up ping handler for connection health
	conn.SetPongHandler(func(string) error {
		h.connectionManager.UpdateConnectionPing(connectionID)
		return nil
	})

	// Handle messages from agent
	for {
		var msg models.WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		h.handleAgentGlobalMessage(c.Request.Context(), tenantID, agentID, agent.Name, msg, connectionID)
	}
}

func (h *AgentWebSocketHandler) handleAgentGlobalMessage(ctx context.Context, tenantID, agentUUID uuid.UUID, agentName string, msg models.WSMessage, connectionID string) {
	switch msg.Type {
	case models.WSMsgTypeChatMessage:
		if msg.Data != nil {
			projectID := *msg.ProjectID
			h.handleChatMessage(ctx, tenantID, projectID, agentUUID, *msg.AgentSessionID, agentName, msg, connectionID)
		}
	case models.WSMsgTypeTypingStart, models.WSMsgTypeTypingStop:
		h.broadcastTypingIndicator(ctx, *msg.AgentSessionID, string(msg.Type), agentName)

	case "ping":
		// Respond to ping
		pongMsg := &ws.Message{
			Type:      "pong",
			SessionID: agentUUID,
			Data:      json.RawMessage(`{"timestamp":"` + time.Now().Format(time.RFC3339) + `"}`),
			FromType:  ws.ConnectionTypeAgent,
		}
		h.connectionManager.SendToConnection(connectionID, pongMsg)
	case "session_subscribe":
		// Agent wants to receive updates for a specific session
		if msg.AgentSessionID != nil {
			h.subscribeAgentToSession(ctx, agentUUID, *msg.AgentSessionID, connectionID)
		}
	case "session_unsubscribe":
		// Agent no longer wants updates for a specific session
		if msg.AgentSessionID != nil {
			h.unsubscribeAgentFromSession(ctx, agentUUID, *msg.AgentSessionID, connectionID)
		}
	}
}

func (h *AgentWebSocketHandler) handleChatMessage(ctx context.Context, tenantID, projectID, agentUUID, sessionID uuid.UUID, agentName string, msg models.WSMessage, connectionID string) {
	// Parse message data
	var messageData struct {
		Content     string `json:"content"`
		MessageType string `json:"message_type"`
	}

	// Type assert the interface{} to []byte
	dataBytes, ok := msg.Data.([]byte)
	if !ok {
		// Try marshaling if it's not already bytes
		var err error
		dataBytes, err = json.Marshal(msg.Data)
		if err != nil {
			log.Printf("Failed to marshal message data: %v", err)
			return
		}
	}

	if err := json.Unmarshal(dataBytes, &messageData); err != nil {
		log.Printf("Failed to parse message data: %v", err)
		return
	}

	// Create message request using the correct SendChatMessageRequest structure
	request := &models.SendChatMessageRequest{
		Content:     messageData.Content,
		MessageType: messageData.MessageType,
		SenderName:  "Agent", // This will be updated with proper agent name
		IsPrivate:   false,
		Metadata:    make(models.JSONMap),
	}

	// Send message via service layer with correct parameters
	_, err := h.chatSessionService.SendMessageWithUuidDetails(ctx, tenantID, projectID, sessionID, request, "agent", &agentUUID, "Agent", connectionID)
	if err != nil {
		log.Printf("Failed to send chat message: %v", err)
		// Send error back to agent
		errorMsg := &ws.Message{
			Type:      "error",
			SessionID: sessionID,
			Data:      json.RawMessage(`{"error":"Failed to send message","details":"` + err.Error() + `"}`),
			FromType:  ws.ConnectionTypeAgent,
		}
		h.connectionManager.SendToConnection(connectionID, errorMsg)
		return
	}
}

func (h *AgentWebSocketHandler) broadcastTypingIndicator(ctx context.Context, sessionID uuid.UUID, typingType, agentName string) {
	typingData := map[string]interface{}{
		"author_type": "agent",
		"author_name": agentName,
		"is_typing":   typingType == "typing_start",
	}
	typingBytes, _ := json.Marshal(typingData)

	typingMsg := &ws.Message{
		Type:      typingType,
		SessionID: sessionID,
		Data:      json.RawMessage(typingBytes),
		FromType:  ws.ConnectionTypeAgent,
	}
	// Note: This broadcasts to session which includes customers and other agents
	// The agent sending the typing indicator should not receive their own typing event (handled on frontend)
	h.connectionManager.DeliverWebSocketMessage(sessionID, typingMsg)
}

// SetChatWSHandler sets the reference to the main ChatWebSocketHandler
func (h *AgentWebSocketHandler) SetChatWSHandler(chatWSHandler *ChatWebSocketHandler) {
	h.chatWSHandler = chatWSHandler
}

func (h *AgentWebSocketHandler) subscribeAgentToSession(ctx context.Context, agentUUID, sessionID uuid.UUID, connectionID string) {
	// Store the subscription in Redis so the agent receives messages for this session
	subscriptionKey := fmt.Sprintf("agent_subscription:%s:%s", agentUUID.String(), sessionID.String())

	// Store the connection ID for this subscription
	err := h.connectionManager.GetRedisClient().Set(ctx, subscriptionKey, connectionID, time.Hour).Err()
	if err != nil {
		log.Printf("Failed to store agent subscription: %v", err)
		return
	}

	// Also add to a reverse lookup (session -> agents)
	sessionAgentsKey := fmt.Sprintf("session_agents:%s", sessionID)
	err = h.connectionManager.GetRedisClient().SAdd(ctx, sessionAgentsKey, agentUUID.String()).Err()
	if err != nil {
		log.Printf("Failed to add agent to session agents list: %v", err)
	}
	h.connectionManager.GetRedisClient().Expire(ctx, sessionAgentsKey, time.Hour)

	log.Printf("Agent %s subscribed to session %s (connection: %s)", agentUUID.String(), sessionID, connectionID)
}

func (h *AgentWebSocketHandler) unsubscribeAgentFromSession(ctx context.Context, agentUUID, sessionID uuid.UUID, connectionID string) {
	// Remove the subscription from Redis
	subscriptionKey := fmt.Sprintf("agent_subscription:%s:%s", agentUUID.String(), sessionID.String())
	h.connectionManager.GetRedisClient().Del(ctx, subscriptionKey)

	// Remove from reverse lookup
	sessionAgentsKey := fmt.Sprintf("session_agents:%s", sessionID.String())
	h.connectionManager.GetRedisClient().SRem(ctx, sessionAgentsKey, agentUUID.String())

	log.Printf("Agent %s unsubscribed from session %s", agentUUID.String(), sessionID.String())
}
