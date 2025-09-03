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
