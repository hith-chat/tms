package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bareuptime/tms/internal/logger"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/bareuptime/tms/internal/service"
	ws "github.com/bareuptime/tms/internal/websocket"
)

// SlackEventsHandler handles Slack event webhooks
type SlackEventsHandler struct {
	slackService       *service.SlackService
	chatSessionService *service.ChatSessionService
	connectionManager  *ws.ConnectionManager
	integrationRepo    *repo.ProjectIntegrationRepository
	chatSessionRepo    *repo.ChatSessionRepo
	aiAgentClient      *service.AiAgentClient
	aiService          *service.AIService
}

// NewSlackEventsHandler creates a new Slack events handler
func NewSlackEventsHandler(
	slackService *service.SlackService,
	chatSessionService *service.ChatSessionService,
	connectionManager *ws.ConnectionManager,
	integrationRepo *repo.ProjectIntegrationRepository,
	chatSessionRepo *repo.ChatSessionRepo,
	aiAgentClient *service.AiAgentClient,
	aiService *service.AIService,
) *SlackEventsHandler {
	return &SlackEventsHandler{
		slackService:       slackService,
		chatSessionService: chatSessionService,
		connectionManager:  connectionManager,
		integrationRepo:    integrationRepo,
		chatSessionRepo:    chatSessionRepo,
		aiAgentClient:      aiAgentClient,
		aiService:          aiService,
	}
}

// SlackEventRequest represents the incoming Slack event
type SlackEventRequest struct {
	Type      string          `json:"type"`
	Challenge string          `json:"challenge,omitempty"` // For URL verification
	Token     string          `json:"token"`
	TeamID    string          `json:"team_id"`
	Event     json.RawMessage `json:"event"`
}

// SlackMessageEvent represents a Slack message event
type SlackMessageEvent struct {
	Type     string `json:"type"`
	User     string `json:"user"`
	Text     string `json:"text"`
	TS       string `json:"ts"`
	Channel  string `json:"channel"`
	ThreadTS string `json:"thread_ts,omitempty"`
	BotID    string `json:"bot_id,omitempty"`
	Subtype  string `json:"subtype,omitempty"`
}

// HandleSlackEvents handles incoming Slack event webhooks
// @Summary Slack events webhook
// @Description Receives Slack events for bidirectional chat sync
// @Tags integrations
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/public/integrations/slack/events [post]
func (h *SlackEventsHandler) HandleSlackEvents(c *gin.Context) {
	ctx := c.Request.Context()

	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	var eventReq SlackEventRequest
	if err := json.Unmarshal(body, &eventReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	// Handle URL verification challenge
	if eventReq.Type == "url_verification" {
		c.JSON(http.StatusOK, gin.H{"challenge": eventReq.Challenge})
		return
	}

	// Handle event callback
	if eventReq.Type == "event_callback" {
		var messageEvent SlackMessageEvent
		if err := json.Unmarshal(eventReq.Event, &messageEvent); err != nil {
			logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to parse Slack message event")
			c.JSON(http.StatusOK, gin.H{"ok": true}) // Always return 200 to Slack
			return
		}

		fmt.Println("messageEvent --> ", messageEvent)

		// Only process message events
		if messageEvent.Type == "message" {
			// Ignore bot messages and message changes
			if messageEvent.BotID != "" || messageEvent.Subtype != "" {
				c.JSON(http.StatusOK, gin.H{"ok": true})
				return
			}

			// Only process messages in threads
			if messageEvent.ThreadTS == "" {
				c.JSON(http.StatusOK, gin.H{"ok": true})
				return
			}

			// IMPORTANT: Respond to Slack immediately (early exit)
			// Process the message asynchronously with a background context
			// to avoid "context canceled" errors when the HTTP request completes
			c.JSON(http.StatusOK, gin.H{"ok": true})

			// Use background context for async processing
			// Preserve transaction ID from request context for tracing
			bgCtx := context.Background()
			if txID := logger.GetTransactionID(ctx); txID != "" {
				bgCtx = logger.WithTransaction(bgCtx)
			}
			go h.handleSlackThreadReply(bgCtx, eventReq.TeamID, &messageEvent)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// handleSlackThreadReply processes a Slack thread reply and syncs to chat
func (h *SlackEventsHandler) handleSlackThreadReply(ctx context.Context, teamID string, event *SlackMessageEvent) {
	// Find chat session by Slack thread_ts
	fmt.Println("Handling Slack thread reply for thread_ts:", event.ThreadTS, "channel:", event.Channel)
	session, err := h.findSessionBySlackThread(ctx, teamID, event.ThreadTS, event.Channel)
	if err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).
			Str("thread_ts", event.ThreadTS).
			Msg("Failed to find session by Slack thread")
		return
	}
	if session == nil {
		logger.GetTxLogger(ctx).Warn().
			Str("thread_ts", event.ThreadTS).
			Msg("No session found for Slack thread")
		return
	}

	// Get Slack integration to fetch access token
	slackMeta, err := h.slackService.GetSlackIntegration(ctx, session.TenantID, session.ProjectID)
	if err != nil || slackMeta == nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to get Slack integration")
		return
	}

	// Get Slack user's display name
	displayName, err := h.slackService.GetUserDisplayName(ctx, slackMeta.AccessToken, event.User)
	if err != nil {
		logger.GetTxLogger(ctx).Warn().Err(err).Msg("Failed to get Slack user display name")
		displayName = "Slack Agent"
	}

	// Check if bot is mentioned (@Hith or <@U09TY8HM0D8>)
	isBotMentioned := h.isBotMentioned(event.Text, slackMeta.BotUserID)

	if isBotMentioned && h.aiAgentClient != nil {
		// Bot is mentioned - trigger AI response
		logger.GetTxLogger(ctx).Info().
			Str("session_id", session.ID.String()).
			Str("message", event.Text).
			Msg("Bot mentioned in Slack - triggering AI response")

		// Remove bot mention from text for cleaner AI processing
		cleanedText := h.removeBotMention(event.Text, slackMeta.BotUserID)

		// Save the user's message first
		messageReq := &models.SendChatMessageRequest{
			Content:     cleanedText,
			MessageType: "text",
		}

		_, err := h.chatSessionService.SendMessageWithUuidDetails(
			ctx,
			session.TenantID,
			session.ProjectID,
			session.ID,
			messageReq,
			"visitor", // Mark as visitor to preserve message
			nil,
			fmt.Sprintf("Slack: %s", displayName),
			"",
		)
		if err != nil {
			logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to save Slack message")
			return
		}

		// Process with AI
		go h.processSlackMessageWithAI(ctx, session, cleanedText, displayName)
		return
	}

	// Regular message (not a bot mention) - save as agent message
	messageReq := &models.SendChatMessageRequest{
		Content:     event.Text,
		MessageType: "text",
	}

	// Create a pseudo agent ID for the Slack user (we'll use a deterministic UUID based on Slack user ID)
	// For now, we'll just use nil and the name
	message, err := h.chatSessionService.SendMessageWithUuidDetails(
		ctx,
		session.TenantID,
		session.ProjectID,
		session.ID,
		messageReq,
		"agent", // Mark as agent to disable AI
		nil,     // No specific agent UUID
		fmt.Sprintf("Slack: %s", displayName),
		"",
	)
	if err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to save Slack message")
		return
	}

	// Disable AI for this session (assign to a slack agent)
	if session.AssignedAgentID == nil {
		// Mark session as having a human agent (disable AI)
		session.UseAI = false
		if err := h.chatSessionRepo.UpdateChatSession(ctx, session); err != nil {
			logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to update session AI status")
		}
	}

	// Broadcast message to customer websocket
	h.broadcastToCustomer(ctx, session, message, displayName)

	// Broadcast to agent websocket (so other agents see it)
	h.broadcastToAgents(ctx, session, message, displayName)

	logger.GetTxLogger(ctx).Info().
		Str("session_id", session.ID.String()).
		Str("slack_user", displayName).
		Msg("Synced Slack reply to chat session")
}

// findSessionBySlackThread finds a chat session by Slack thread_ts and channel_id
func (h *SlackEventsHandler) findSessionBySlackThread(ctx context.Context, _, threadTS, channelID string) (*models.ChatSession, error) {
	// Use dedicated repository method with indexed columns for fast lookup
	// teamID parameter is not needed as we query by thread_ts and channel_id which are unique
	return h.chatSessionRepo.GetChatSessionBySlackThread(ctx, threadTS, channelID)
}

// broadcastToCustomer sends the message to the customer's websocket
func (h *SlackEventsHandler) broadcastToCustomer(ctx context.Context, session *models.ChatSession, message *models.ChatMessage, senderName string) {
	payload := map[string]interface{}{
		"id":          message.ID.String(),
		"content":     message.Content,
		"sender":      "agent",
		"sender_name": senderName,
		"timestamp":   message.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"session_id":  session.ID.String(),
	}

	payloadJSON, _ := json.Marshal(payload)

	wsMessage := &ws.Message{
		Type:      "chat_message",
		SessionID: session.ID,
		Data:      json.RawMessage(payloadJSON),
		Timestamp: time.Now(),
	}

	// Deliver to customer websocket
	if err := h.connectionManager.DeliverWebSocketMessage(session.ID, wsMessage); err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to deliver message to customer")
	}
}

// broadcastToAgents sends the message to agent websockets
func (h *SlackEventsHandler) broadcastToAgents(ctx context.Context, session *models.ChatSession, message *models.ChatMessage, senderName string) {
	payload := map[string]interface{}{
		"id":          message.ID.String(),
		"session_id":  session.ID.String(),
		"content":     message.Content,
		"sender":      "agent",
		"sender_name": senderName,
		"timestamp":   message.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"widget_id":   session.WidgetID.String(),
	}

	payloadJSON, _ := json.Marshal(payload)

	wsMessage := &ws.Message{
		Type:      "chat_message",
		SessionID: session.ID,
		Data:      json.RawMessage(payloadJSON),
		ProjectID: &session.ProjectID,
		Timestamp: time.Now(),
	}

	// Send to project agents
	if err := h.connectionManager.SendToProjectAgents(session.ProjectID, wsMessage); err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to send message to agents")
	}
}

// isBotMentioned checks if the bot is mentioned in the message
// Slack mentions appear as <@USER_ID> in the message text
func (h *SlackEventsHandler) isBotMentioned(text, botUserID string) bool {
	if botUserID == "" {
		// Fallback: check for common bot mention patterns
		return strings.Contains(text, "@Hith") ||
			strings.Contains(text, "@hith") ||
			strings.Contains(text, "<@U09TY8HM0D8>") // Your specific bot user ID
	}
	// Check for Slack mention format: <@USER_ID>
	mentionPattern := fmt.Sprintf("<@%s>", botUserID)
	return strings.Contains(text, mentionPattern)
}

// removeBotMention removes bot mentions from the message text
func (h *SlackEventsHandler) removeBotMention(text, botUserID string) string {
	cleaned := text

	// Remove Slack mention format <@USER_ID>
	if botUserID != "" {
		mentionPattern := fmt.Sprintf("<@%s>", botUserID)
		cleaned = strings.ReplaceAll(cleaned, mentionPattern, "")
	}

	// Remove common @Hith patterns
	cleaned = strings.ReplaceAll(cleaned, "@Hith", "")
	cleaned = strings.ReplaceAll(cleaned, "@hith", "")
	cleaned = strings.ReplaceAll(cleaned, "<@U09TY8HM0D8>", "")

	// Trim whitespace
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}

// processSlackMessageWithAI handles AI agent response for Slack messages
func (h *SlackEventsHandler) processSlackMessageWithAI(ctx context.Context, session *models.ChatSession, content string, userDisplayName string) {
	logger.GetTxLogger(ctx).Info().
		Str("session_id", session.ID.String()).
		Str("message", content).
		Msg("Processing Slack message with AI")

	// Create agent request
	request := service.ChatRequest{
		Message:   content,
		TenantID:  session.TenantID.String(),
		ProjectID: session.ProjectID.String(),
		SessionID: session.ID.String(),
		UserID:    fmt.Sprintf("slack:%s", userDisplayName),
		Metadata: map[string]string{
			"source":    "slack",
			"widget_id": session.WidgetID.String(),
			"user_name": userDisplayName,
		},
	}

	// Start streaming response from agent service
	responseChan, errorChan := h.aiAgentClient.ProcessMessageStream(ctx, request)

	var aiResponseContent strings.Builder
	var hasError bool

	// Handle the streaming response
	for {
		select {
		case response, ok := <-responseChan:
			if !ok {
				// Channel closed, finish processing
				if !hasError && aiResponseContent.Len() > 0 {
					// Send final AI response to Slack
					h.sendAIResponseToSlack(ctx, session, aiResponseContent.String())
				}
				return
			}

			// Handle different response types
			switch response.Type {
			case "message":
				logger.GetTxLogger(ctx).Debug().
					Str("chunk", response.Content).
					Msg("Received AI response chunk")
				if response.Content != "" {
					aiResponseContent.WriteString(response.Content)
				}
			case "thinking":
				// AI is thinking - we could optionally send typing indicator to Slack
				logger.GetTxLogger(ctx).Debug().Msg("AI is thinking")
			case "done", "metadata":
				// Processing complete
				logger.GetTxLogger(ctx).Info().Msg("AI processing complete")
			case "error":
				hasError = true
				log.Printf("AI agent processing error for session %s: %s", session.ID, response.Content)
				return
			}

		case err, ok := <-errorChan:
			if !ok {
				return
			}
			hasError = true
			log.Printf("AI Agent client error for session %s: %v", session.ID, err)
			return

		case <-ctx.Done():
			log.Printf("Context cancelled for ai-agent processing in session %s", session.ID)
			return
		}
	}
}

// sendAIResponseToSlack sends the AI-generated response back to Slack
func (h *SlackEventsHandler) sendAIResponseToSlack(ctx context.Context, session *models.ChatSession, aiResponse string) {
	logger.GetTxLogger(ctx).Info().
		Str("session_id", session.ID.String()).
		Str("response_length", fmt.Sprintf("%d", len(aiResponse))).
		Msg("Sending AI response to Slack")

	// Send message to Slack using the Slack service
	err := h.slackService.PostMessageToSlack(
		ctx,
		session.TenantID,
		session.ProjectID,
		session,
		aiResponse,
		"Hith (AI Assistant)",
	)
	if err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to send AI response to Slack")
		return
	}

	// Also save the AI response to the database
	messageReq := &models.SendChatMessageRequest{
		Content:     aiResponse,
		MessageType: "text",
	}

	_, err = h.chatSessionService.SendMessageWithUuidDetails(
		ctx,
		session.TenantID,
		session.ProjectID,
		session.ID,
		messageReq,
		"agent",
		nil,
		"Hith (AI Assistant)",
		"",
	)
	if err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to save AI response to database")
	}

	logger.GetTxLogger(ctx).Info().Msg("AI response sent to Slack successfully")
}
