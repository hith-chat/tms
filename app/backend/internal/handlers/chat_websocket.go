package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/bareuptime/tms/internal/auth"
	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/service"
	ws "github.com/bareuptime/tms/internal/websocket"
)

// ChatWebSocketRequest represents incoming WebSocket messages from visitors
// @Description WebSocket message structure for visitor chat communication
type ChatWebSocketRequest struct {
	Type    string      `json:"type" example:"chat_message"`
	Payload interface{} `json:"payload"`
}

// ChatWebSocketResponse represents outgoing WebSocket messages to visitors
// @Description WebSocket response structure for visitor chat communication
type ChatWebSocketResponse struct {
	Type      string      `json:"type" example:"chat_message"`
	Payload   interface{} `json:"payload"`
	MessageID string      `json:"message_id,omitempty"`
	Timestamp string      `json:"timestamp,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
}

type ChatWebSocketHandler struct {
	chatSessionService  *service.ChatSessionService
	connectionManager   *ws.ConnectionManager
	notificationService *service.NotificationService
	aiService           *service.AIService
	aiAgentClient       *service.AiAgentClient
	authService         *auth.Service
}

func NewChatWebSocketHandler(chatSessionService *service.ChatSessionService, connectionManager *ws.ConnectionManager, notificationService *service.NotificationService, aiService *service.AIService, agentClient *service.AiAgentClient, authService *auth.Service) *ChatWebSocketHandler {
	return &ChatWebSocketHandler{
		chatSessionService:  chatSessionService,
		connectionManager:   connectionManager,
		notificationService: notificationService,
		aiService:           aiService,
		aiAgentClient:       agentClient,
		authService:         authService,
	}
}

// HandleWebSocketPublic handles WebSocket connections for real-time chat from visitors
// @Summary Public chat WebSocket connection
// @Description Establish WebSocket connection for real-time chat communication from visitors
// @Tags chat-websocket
// @Accept json
// @Produce json
// @Param session_token header string true "Session token for authentication"
// @Success 101 "Switching Protocols - WebSocket connection established"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /public/chat/ws [get]
func (h *ChatWebSocketHandler) HandleWebSocketPublic(c *gin.Context) {
	sessionToken := middleware.GetSessionToken(c)
	widgetID := middleware.GetWidgetID(c)

	// Validate chat token and extract claims
	claims, err := h.authService.ValidateChatToken(sessionToken, widgetID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session token"})
		return
	}
	// Extract session ID from claims
	clientSessionID := claims.SessionID
	// Validate session
	session, err := h.chatSessionService.GetChatSessionByClientSessionID(c.Request.Context(), clientSessionID)
	if err != nil || session == nil {
		// Create a minimal session or use InitiateChat with required data
		initReq := &models.InitiateChatRequest{
			VisitorName:  "",
			VisitorEmail: "",
			VisitorInfo:  claims.VisitorInfo,
		}
		if claims.VisitorName != nil {
			initReq.VisitorName = *claims.VisitorName
			initReq.VisitorInfo["visitor_name"] = *claims.VisitorName
		}
		if claims.VisitorEmail != nil {
			initReq.VisitorEmail = *claims.VisitorEmail
			initReq.VisitorInfo["visitor_email"] = *claims.VisitorEmail
		}

		fmt.Println("Initiating chat session:", initReq)

		session, err = h.chatSessionService.InitiateChat(c.Request.Context(), claims.WidgetID, clientSessionID, initReq)
		if err != nil {
			fmt.Println("Error creating chat session:", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chat session: " + err.Error()})
			return
		}

		// Update the ClientSessionID to match the token
		session.ClientSessionID = clientSessionID
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// Register connection with the enterprise connection manager
	connectionID, err := h.connectionManager.AddConnection(
		ws.ConnectionTypeVisitor,
		session.ID,
		[]uuid.UUID{session.ProjectID},
		nil, // No user ID for visitors
		conn,
	)
	if err != nil {
		log.Printf("Failed to register connection: %v", err)
		return
	}

	// Clean up on disconnect
	defer func() {
		h.connectionManager.RemoveConnection(connectionID)
	}()

	// Send welcome message
	welcomeMsg := &ws.Message{
		Type:      "session_update",
		SessionID: session.ID,
		Data: json.RawMessage(`{
			"type": "connected",
			"message": "Connected to chat session"
		}`),
		FromType:     ws.ConnectionTypeVisitor,
		DeliveryType: ws.Self,
	}
	h.connectionManager.SendToConnection(connectionID, welcomeMsg)

	// Set up ping handler for connection health
	conn.SetPongHandler(func(string) error {
		h.connectionManager.UpdateConnectionPing(connectionID)
		return nil
	})

	// Handle messages
	for {
		var msg models.WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Error reading WebSocket message:", err.Error())
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		h.handleVisitorMessage(c.Request.Context(), session, msg, connectionID)
	}
}

// handleVisitorMessage handles incoming WebSocket messages from visitors
func (h *ChatWebSocketHandler) handleVisitorMessage(ctx context.Context, session *models.ChatSession, msg models.WSMessage, connID string) {
	switch msg.Type {
	case models.WSMsgTypeChatMessage:
		h.processVisitorChatMessage(ctx, session, msg, connID)
	case models.WSMsgTypeTypingStart:
		h.processVisitorTyping(session, msg, true)
	case models.WSMsgTypeTypingStop:
		h.processVisitorTyping(session, msg, false)
	case models.WSMsgTypeReadReceipt:
		h.processReadReceipt(ctx, session, msg, "visitor")
	}
}

// // handleAgentMessage handles incoming WebSocket messages from agents
// func (h *ChatWebSocketHandler) handleAgentMessage(ctx context.Context, tenantID, projectID, sessionID, agentID uuid.UUID, msg models.WSMessage, connID string) {
// 	switch msg.Type {
// 	case models.WSMsgTypeChatMessage:
// 		h.processAgentChatMessage(ctx, tenantID, projectID, sessionID, agentID, msg, connID)
// 	case models.WSMsgTypeTypingStart:
// 		h.processAgentTyping(ctx, tenantID, projectID, sessionID, agentID, msg, true)
// 	case models.WSMsgTypeTypingStop:
// 		h.processAgentTyping(ctx, tenantID, projectID, sessionID, agentID, msg, false)
// 	case models.WSMsgTypeReadReceipt:
// 		h.processReadReceipt(ctx, sessionID, "agent", connID)
// 	}
// }

// processVisitorChatMessage processes chat messages from visitors
func (h *ChatWebSocketHandler) processVisitorChatMessage(ctx context.Context, session *models.ChatSession, msg models.WSMessage, connID string) {
	if data, ok := msg.Data.(map[string]interface{}); ok {
		content, _ := data["content"].(string)
		if content != "" {
			req := &models.SendChatMessageRequest{
				Content: content,
			}

			visitorName := "Visitor"
			if session.CustomerName != nil && *session.CustomerName != "" {
				visitorName = *session.CustomerName
			}

			go h.chatSessionService.SendMessage(
				ctx,
				session,
				req,
				"visitor",
				nil,
				visitorName,
				connID,
			)

			// Process AI response using agent client SSE
			shouldRespondWithAI := h.aiAgentClient != nil && session.UseAI && session.AssignedAgentID == nil
			fmt.Println("should respond with AI: h.agentClient != nil ->", h.aiAgentClient != nil)
			fmt.Println("should respond with AI: session.UseAI ->", session.UseAI)
			fmt.Println("should respond with AI: session.AssignedAgentID == nil ->", session.AssignedAgentID == nil)
			fmt.Println("should respond with AI: overall okay ->", shouldRespondWithAI)

			if shouldRespondWithAI {
				go func() {
					// _, err := h.aiService.ProcessMessage(ctx, session, content, connID)
					// if err != nil {
					// 	log.Printf("AI processing error for session %s: %v", session.ID, err)
					// 	return
					// }
					fmt.Println("Processing visitor message with AI agent client... ", content)
					h.processVisitorResponseUsingAi(ctx, session, content, connID)
				}()
			}
		}
	}
}

// getStringValue safely extracts string value from string pointer
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// processVisitorResponseUsingAi handles AI agent response using SSE streaming
func (h *ChatWebSocketHandler) processVisitorResponseUsingAi(ctx context.Context, session *models.ChatSession, content string, connID string) {
	// Create agent request
	request := service.ChatRequest{
		Message:   content,
		TenantID:  session.TenantID.String(), // Using ProjectID as TenantID for now
		ProjectID: session.ProjectID.String(),
		SessionID: session.ID.String(),
		UserID:    getStringValue(session.CustomerEmail), // Use customer email as user ID if available
		Metadata: map[string]string{
			"connection_id": connID,
			"widget_id":     session.WidgetID.String(),
		},
	}

	// Start streaming response from agent service
	responseChan, errorChan := h.aiAgentClient.ProcessMessageStream(ctx, request)

	// Handle the streaming response
	for {
		select {
		case response, ok := <-responseChan:
			fmt.Println("\nReceived from AI agent response channel:", response, ok)
			if !ok {
				// Channel closed, finish processing
				return
			}

			// Handle different response types
			switch response.Type {
			case "message":
				fmt.Println("\nReceived AI agent message chunk:", response.Content)
				if response.Content != "" {
					h.aiService.SendAIResponse(ctx, session, connID, response.Content, map[string]interface{}{
						"ai_generated":  true,
						"response_type": "knowledge_based",
					})
				}
				h.aiService.ProcessAiTyping(session, models.WSMessage{}, connID, false)
			case "thinking":
				// Send typing indicator
				h.aiService.ProcessAiTyping(session, models.WSMessage{}, connID, true)
			case "done", "metadata":
				// Processing complete
				h.aiService.ProcessAiTyping(session, models.WSMessage{}, connID, false)
			case "error":
				log.Printf("AI agent processing error for session %s: %s", session.ID, response.Content)
				h.aiService.ProcessAiTyping(session, models.WSMessage{}, connID, false)
				return
			}

		case err, ok := <-errorChan:
			fmt.Println("\n\nReceived from AI agent error channel:", err, ok)
			if !ok {
				return
			}
			log.Printf("AI Agent client error for session %s: %v", session.ID, err)
			h.aiService.ProcessAiTyping(session, models.WSMessage{}, connID, false)
			return

		case <-ctx.Done():
			log.Printf("Context cancelled for ai-agent processing in session %s", session.ID)
			return
		}
	}
}

// sendStreamingResponse sends a streaming chunk to the WebSocket
func (h *ChatWebSocketHandler) sendStreamingResponse(session *models.ChatSession, content string, isFirst bool, connectionID string) {
	message := &ws.Message{
		Type:      "chat_message",
		SessionID: session.ID,
		Data: json.RawMessage(fmt.Sprintf(`{
			"content": %q,
			"streaming": true,
			"is_first": %t,
			"session_id": %q,
			"connection_id": %q,
			"sender": "agent",
			"sender_name": "AI Agent"
		}`, content, isFirst, session.ID.String(), connectionID)),
		FromType:  ws.ConnectionTypeAiAgent,
		ProjectID: &session.ProjectID,
	}

	// Use the connection manager's proper messaging system instead of direct WebSocket writes
	if err := h.connectionManager.DeliverWebSocketMessage(session.ID, message); err != nil {
		log.Printf("Failed to send streaming chunk via connection manager: %v", err)
	}
}

// saveAiAgentMessage saves the complete agent response as a chat message
func (h *ChatWebSocketHandler) saveAiAgentMessage(ctx context.Context, session *models.ChatSession, content string, connectionID string) {
	request := &models.SendChatMessageRequest{
		Content: content,
	}

	// Save message as agent response
	_, err := h.chatSessionService.SendMessageWithUuidDetails(
		ctx,
		session.TenantID,
		session.ProjectID,
		session.ID,
		request,
		"ai-agent",
		nil, // No specific agent UUID for AI agent
		"AI Agent",
		connectionID,
	)

	if err != nil {
		log.Printf("Failed to save agent message for session %s: %v", session.ID, err)
	}
}

// processVisitorTyping handles visitor typing indicators
func (h *ChatWebSocketHandler) processVisitorTyping(session *models.ChatSession, msg models.WSMessage, isTyping bool) {
	visitorName := "Visitor"
	if session.CustomerName != nil && *session.CustomerName != "" {
		visitorName = *session.CustomerName
	}

	msgType := "typing_stop"
	if isTyping {
		msgType = "typing_start"
	}

	typingData, _ := json.Marshal(map[string]interface{}{
		"author_name": visitorName,
		"author_type": "visitor",
	})

	broadcastMsg := &ws.Message{
		Type:      msgType,
		SessionID: session.ID,
		Data:      typingData,
		FromType:  ws.ConnectionTypeVisitor,
		ProjectID: &session.ProjectID,
		TenantID:  &session.TenantID,
		AgentID:   session.AssignedAgentID,
	}
	h.connectionManager.DeliverWebSocketMessage(session.ID, broadcastMsg)
}

// processReadReceipt handles read receipt processing
func (h *ChatWebSocketHandler) processReadReceipt(ctx context.Context, session *models.ChatSession, msg models.WSMessage, readerType string) {
	broadcastMsg := &ws.Message{
		Type:      "read_receipt_confirmed",
		SessionID: session.ID,
		Data:      json.RawMessage(`{"status":"acknowledged"}`),
		FromType:  ws.ConnectionTypeVisitor,
		ProjectID: &session.ProjectID,
		TenantID:  &session.TenantID,
		AgentID:   session.AssignedAgentID,
	}
	go h.connectionManager.DeliverWebSocketMessage(session.ID, broadcastMsg)
	go h.chatSessionService.MarkVisitorMessagesAsRead(ctx, session.ID, *msg.MessageID, readerType)
}

// sendError sends an error message to a specific connection
func (h *ChatWebSocketHandler) sendError(connID string, errorMsg string) {
	errorData, _ := json.Marshal(map[string]interface{}{
		"error": errorMsg,
	})

	msg := &ws.Message{
		Type: "error",
		Data: errorData,
	}
	h.connectionManager.SendToConnection(connID, msg)
}
