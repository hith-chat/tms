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
	authService         *auth.Service
}

func NewChatWebSocketHandler(chatSessionService *service.ChatSessionService, connectionManager *ws.ConnectionManager, notificationService *service.NotificationService, aiService *service.AIService, authService *auth.Service) *ChatWebSocketHandler {
	return &ChatWebSocketHandler{
		chatSessionService:  chatSessionService,
		connectionManager:   connectionManager,
		notificationService: notificationService,
		aiService:           aiService,
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

			// Process AI response if enabled and applicable
			shouldRespondWithAI := h.aiService != nil && session.UseAI && session.AssignedAgentID == nil
			fmt.Println("should respond with AI: h.aiService != nil ->", h.aiService != nil)
			fmt.Println("should respond with AI: session.UseAI ->", session.UseAI)
			fmt.Println("should respond with AI: session.AssignedAgentID == nil ->", session.AssignedAgentID == nil)
			fmt.Println("should respond with AI: overall ->", shouldRespondWithAI)

			if shouldRespondWithAI {
				go func() {
					_, err := h.aiService.ProcessMessage(ctx, session, content, connID)
					if err != nil {
						log.Printf("AI processing error for session %s: %v", session.ID, err)
						return
					}
				}()
			}
		}
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
