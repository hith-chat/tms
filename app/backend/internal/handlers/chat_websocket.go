package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/bareuptime/tms/internal/auth"
	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/service"
	ws "github.com/bareuptime/tms/internal/websocket"
)

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

// HandleWebSocket handles WebSocket connections for real-time chat from visitors
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
			fmt.Println("should respond with AI: overall ->", shouldRespondWithAI)

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
func (h *ChatWebSocketHandler) processVisitorResponseUsingAi(ctx context.Context, session *models.ChatSession, content string, connectionID string) {
	// Create agent request
	request := service.ChatRequest{
		Message:   content,
		TenantID:  session.ProjectID.String(), // Using ProjectID as TenantID for now
		ProjectID: session.ProjectID.String(),
		SessionID: session.ID.String(),
		UserID:    getStringValue(session.CustomerEmail), // Use customer email as user ID if available
		Metadata: map[string]string{
			"connection_id": connectionID,
			"widget_id":     session.WidgetID.String(),
		},
	}

	// Start streaming response from agent service
	responseChan, errorChan := h.aiAgentClient.ProcessMessageStream(ctx, request)

	// Handle the streaming response
	var responseContent strings.Builder
	isFirstChunk := true

	for {
		select {
		case response, ok := <-responseChan:
			if !ok {
				// Channel closed, finish processing
				finalResponse := responseContent.String()
				if finalResponse != "" {
					// Save the complete AI response as a message
					h.saveAiAgentMessage(ctx, session, finalResponse, connectionID)
				}
				return
			}

			// Handle different response types
			switch response.Type {
			case "message":
				if response.Content != "" {
					responseContent.WriteString(response.Content)

					// Send streaming chunk to WebSocket
					h.sendStreamingResponse(session, response.Content, isFirstChunk, connectionID)
					isFirstChunk = false
				}
			case "thinking":
				// Send typing indicator
				h.sendTypingIndicator(session, true, connectionID)
			case "done", "metadata":
				// Processing complete
				h.sendTypingIndicator(session, false, connectionID)
			case "error":
				log.Printf("AI agent processing error for session %s: %s", session.ID, response.Content)
				h.sendTypingIndicator(session, false, connectionID)
				return
			}

		case err, ok := <-errorChan:
			if !ok {
				return
			}
			log.Printf("AI Agent client error for session %s: %v", session.ID, err)
			h.sendTypingIndicator(session, false, connectionID)
			return

		case <-ctx.Done():
			log.Printf("Context cancelled for ai-agent processing in session %s", session.ID)
			return
		}
	}
}

// sendStreamingResponse sends a streaming chunk to the WebSocket
func (h *ChatWebSocketHandler) sendStreamingResponse(session *models.ChatSession, content string, isFirst bool, connectionID string) {
	message := models.WSMessage{
		Type: models.WSMsgTypeChatMessage,
		Data: map[string]interface{}{
			"content":       content,
			"streaming":     true,
			"is_first":      isFirst,
			"session_id":    session.ID.String(),
			"connection_id": connectionID,
			"sender":        "agent",
			"sender_name":   "AI Agent",
		},
		ClientSessionID: &connectionID,
		ProjectID:       &session.ProjectID,
	}

	// Send to all connections for this session
	sessionConnections, _ := h.connectionManager.GetSessionConnections(session.ID.String())
	for _, conn := range sessionConnections {
		if err := conn.WsConnection.WriteJSON(message); err != nil {
			log.Printf("Failed to send streaming chunk: %v", err)
		}
	}
}

// sendTypingIndicator sends typing indicator to WebSocket
func (h *ChatWebSocketHandler) sendTypingIndicator(session *models.ChatSession, isTyping bool, connectionID string) {
	var msgType models.WSMessageType
	if isTyping {
		msgType = models.WSMsgTypeTypingStart
	} else {
		msgType = models.WSMsgTypeTypingStop
	}

	message := models.WSMessage{
		Type: msgType,
		Data: map[string]interface{}{
			"session_id":    session.ID.String(),
			"connection_id": connectionID,
			"sender":        "agent",
			"sender_name":   "AI Agent",
		},
		ClientSessionID: &connectionID,
		ProjectID:       &session.ProjectID,
	}

	// Send to all connections for this session
	sessionConnections, _ := h.connectionManager.GetSessionConnections(session.ID.String())
	for _, conn := range sessionConnections {
		if err := conn.WsConnection.WriteJSON(message); err != nil {
			log.Printf("Failed to send typing indicator: %v", err)
		}
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
