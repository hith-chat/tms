package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/models"
	ws "github.com/bareuptime/tms/internal/websocket"
)

// AIService handles AI-powered chat assistance
type AIService struct {
	config              *config.AIConfig
	agenticConfig       *config.AgenticConfig
	chatSessionService  *ChatSessionService
	knowledgeService    *KnowledgeService
	greetingDetection   *GreetingDetectionService
	brandGreeting       *BrandGreetingService
	connectionManager   *ws.ConnectionManager
	howlingAlarmService *HowlingAlarmService
	httpClient          *http.Client
}

// NewAIService creates a new AI service instance
func NewAIService(cfg *config.AIConfig, agenticConfig *config.AgenticConfig, chatSessionService *ChatSessionService, knowledgeService *KnowledgeService, greetingDetection *GreetingDetectionService, brandGreeting *BrandGreetingService, connectionManager *ws.ConnectionManager, howlingAlarmService *HowlingAlarmService) *AIService {
	return &AIService{
		config:              cfg,
		agenticConfig:       agenticConfig,
		chatSessionService:  chatSessionService,
		knowledgeService:    knowledgeService,
		greetingDetection:   greetingDetection,
		brandGreeting:       brandGreeting,
		connectionManager:   connectionManager,
		howlingAlarmService: howlingAlarmService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsAgenticBehaviorEnabled checks if agentic behavior is enabled
func (ai *AIService) IsAgenticBehaviorEnabled() bool {
	return ai.agenticConfig != nil && ai.agenticConfig.Enabled
}

// IsGreetingDetectionEnabled checks if greeting detection is enabled
func (ai *AIService) IsGreetingDetectionEnabled() bool {
	return ai.IsAgenticBehaviorEnabled() && ai.agenticConfig.GreetingDetection
}

// IsKnowledgeResponsesEnabled checks if knowledge-based responses are enabled
func (ai *AIService) IsKnowledgeResponsesEnabled() bool {
	return ai.IsAgenticBehaviorEnabled() && ai.agenticConfig.KnowledgeResponses
}

// AIProvider represents different AI providers
type AIProvider string

const (
	ProviderOpenAI    AIProvider = "openai"
	ProviderAnthropic AIProvider = "anthropic"
	ProviderAzure     AIProvider = "azure"
)

// ChatCompletionRequest represents a request to AI providers
type ChatCompletionRequest struct {
	Model       string                  `json:"model"`
	Messages    []ChatCompletionMessage `json:"messages"`
	MaxTokens   int                     `json:"max_tokens"`
	Temperature float64                 `json:"temperature"`
}

// ChatCompletionMessage represents a message in the chat completion
type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse represents the response from AI providers
type ChatCompletionResponse struct {
	Choices []struct {
		Message ChatCompletionMessage `json:"message"`
	} `json:"choices"`
}

// IsEnabled returns whether AI assistance is enabled
func (s *AIService) IsEnabled() bool {
	return s.config.Enabled && s.config.APIKey != ""
}

// ShouldHandleSession determines if AI should handle this session
func (s *AIService) ShouldHandleSession(ctx context.Context, session *models.ChatSession) bool {
	if !s.IsEnabled() {
		fmt.Println("AI Service is not enabled for session:", session.ID)
		return false
	}

	// Don't handle if already assigned to a human agent
	if session.AssignedAgentID != nil {
		fmt.Println("Session already assigned to human agent:", *session.AssignedAgentID)
		return false
	}

	return true
}

// ProcessMessage handles incoming visitor messages and generates AI responses
func (s *AIService) ProcessMessage(ctx context.Context, session *models.ChatSession, messageContent, connID string) (*models.ChatMessage, error) {
	if !s.ShouldHandleSession(ctx, session) {
		fmt.Println("AI Service not handling session:", session.ID)
		return nil, nil
	}

	// Check for handoff keywords first
	if s.shouldHandoffToAgent(session, messageContent) {
		s.requestHumanAgent(ctx, session, "Customer requested human assistance", connID)
		return nil, nil
	}

	// Check if this is a greeting message (only if agentic behavior is enabled)
	if s.IsGreetingDetectionEnabled() && s.greetingDetection != nil {
		greetingResult := s.greetingDetection.DetectGreeting(ctx, messageContent)

		// If it's a simple greeting, respond with brand-aware greeting
		if greetingResult.IsGreeting && greetingResult.MessageType == "simple_greeting" {
			return s.handleGreetingMessage(ctx, session, messageContent, connID)
		}

		// Log greeting detection for debugging
		fmt.Printf("Greeting detection result: IsGreeting=%t, Type=%s, Confidence=%.2f\n",
			greetingResult.IsGreeting, greetingResult.MessageType, greetingResult.Confidence)
	}

	// For complex messages, use the existing AI processing flow
	go s.processAiTyping(session, models.WSMessage{}, connID, true)
	resp, err := s.processComplexMessage(ctx, session, messageContent, connID)
	go s.processAiTyping(session, models.WSMessage{}, connID, false)
	return resp, err
}

// processVisitorTyping handles visitor typing indicators
func (h *AIService) processAiTyping(session *models.ChatSession, msg models.WSMessage, connID string, isTyping bool) {

	msgType := "typing_stop"
	if isTyping {
		msgType = "typing_start"
	}

	typingData, _ := json.Marshal(map[string]interface{}{
		"author_name": "AI",
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
	typingDataAgent, _ := json.Marshal(map[string]interface{}{
		"author_name": "AI",
		"author_type": "agent",
	})
	broadcastMsg.Data = typingDataAgent
	broadcastMsg.FromType = ws.ConnectionTypeAgent
	go h.connectionManager.SendToConnection(connID, broadcastMsg)
}

// handleGreetingMessage processes simple greeting messages with brand-aware responses
func (s *AIService) handleGreetingMessage(ctx context.Context, session *models.ChatSession, messageContent, connID string) (*models.ChatMessage, error) {
	var response string

	// Try to generate brand-aware greeting
	if s.brandGreeting != nil {
		greetingResponse, err := s.brandGreeting.GenerateGreetingResponse(ctx, session.TenantID, session.ProjectID, messageContent)
		if err != nil {
			fmt.Printf("Error generating brand greeting: %v, falling back to default\n", err)
			response = "Hello! Thanks for reaching out. How can we help you today?"
		} else {
			response = greetingResponse.Message
			fmt.Printf("Generated brand-aware greeting: %s\n", response)
		}
	} else {
		// Fallback to simple greeting
		response = "Hello! Thanks for reaching out. How can we help you today?"
	}

	// Send the greeting response
	return s.sendAIResponse(ctx, session, connID, response, map[string]interface{}{
		"ai_generated":  true,
		"response_type": "greeting",
		"brand_aware":   s.brandGreeting != nil,
	})
}

func (s *AIService) processComplexMessageThroughAgent(ctx context.Context, session *models.ChatSession, messageContent, connID string) (*models.ChatMessage, error) {
	// Get conversation history
	messages, err := s.chatSessionService.GetChatMessages(ctx, session.TenantID, session.ProjectID, session.ID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation history: %w", err)
	}

	// Convert to slice and limit to recent messages
	var recentMessages []models.ChatMessage
	messageCount := len(messages)
	startIdx := 0
	if messageCount > 20 {
		startIdx = messageCount - 20
	}

	for i := startIdx; i < messageCount; i++ {
		if messages[i] != nil {
			recentMessages = append(recentMessages, *messages[i])
		}
	}

	// Generate AI response with knowledge context
	response, err := s.generateResponseWithContext(ctx, session, recentMessages, messageContent)
	if err != nil {
		fmt.Println("Error generating AI response:", err.Error())
		return nil, fmt.Errorf("failed to generate AI response: %w", err)
	}

	fmt.Println("Response from ai -", response)

	// Send the AI response
	return s.sendAIResponse(ctx, session, connID, response, map[string]interface{}{
		"ai_generated":  true,
		"response_type": "knowledge_based",
	})
}

// processComplexMessage handles non-greeting messages using AI and knowledge base
func (s *AIService) processComplexMessage(ctx context.Context, session *models.ChatSession, messageContent, connID string) (*models.ChatMessage, error) {
	// Get conversation history
	messages, err := s.chatSessionService.GetChatMessages(ctx, session.TenantID, session.ProjectID, session.ID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation history: %w", err)
	}

	// Convert to slice and limit to recent messages
	var recentMessages []models.ChatMessage
	messageCount := len(messages)
	startIdx := 0
	if messageCount > 20 {
		startIdx = messageCount - 20
	}

	for i := startIdx; i < messageCount; i++ {
		if messages[i] != nil {
			recentMessages = append(recentMessages, *messages[i])
		}
	}

	// Generate AI response with knowledge context
	response, err := s.generateResponseWithContext(ctx, session, recentMessages, messageContent)
	if err != nil {
		fmt.Println("Error generating AI response:", err.Error())
		return nil, fmt.Errorf("failed to generate AI response: %w", err)
	}

	fmt.Println("Response from ai -", response)

	// Send the AI response
	return s.sendAIResponse(ctx, session, connID, response, map[string]interface{}{
		"ai_generated":  true,
		"response_type": "knowledge_based",
	})
}

// sendAIResponse is a helper method to send AI responses
func (s *AIService) sendAIResponse(ctx context.Context, session *models.ChatSession, connID, content string, metadata map[string]interface{}) (*models.ChatMessage, error) {
	// Send AI response
	aiMessageReq := &models.SendChatMessageRequest{
		Content:     content,
		MessageType: "text",
		IsPrivate:   false,
		Metadata:    metadata,
	}

	// Create AI agent UUID (deterministic based on session)

	messageAI, err := s.chatSessionService.SendMessage(
		ctx,
		session,
		aiMessageReq,
		"ai-agent",
		nil,
		"AI Assistant",
		connID,
	)

	return messageAI, err
}

// generateAIAgentID creates a deterministic UUID for AI agent
func (s *AIService) generateAIAgentID(sessionID uuid.UUID) uuid.UUID {
	// Create a deterministic UUID based on session ID
	namespace := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000") // AI namespace
	return uuid.NewSHA1(namespace, sessionID[:])
}

// shouldHandoffToAgent checks if the message contains handoff keywords
func (s *AIService) shouldHandoffToAgent(session *models.ChatSession, content string) bool {

	// Check if session has been ongoing for too long (auto handoff)
	if s.config.AutoHandoffTime > 0 {
		if time.Since(session.CreatedAt) > s.config.AutoHandoffTime {
			fmt.Println("Session exceeded auto handoff time:", session.ID)
			return true
		}
	}
	content = strings.ToLower(content)

	defaultKeywords := []string{
		"speak to human", "human agent", "real person", "live agent",
		"escalate", "supervisor", "manager", "human help",
		"not helpful", "doesn't work", "frustrated", "angry",
	}

	keywords := s.config.HandoffKeywords
	if len(keywords) == 0 {
		keywords = defaultKeywords
	}

	for _, keyword := range keywords {
		if strings.Contains(content, strings.ToLower(keyword)) {
			return true
		}
	}

	return false
}

// requestHumanAgent triggers handoff to human agent
func (s *AIService) requestHumanAgent(ctx context.Context, session *models.ChatSession, reason, connID string) error {
	// Send notification message about handoff
	messageContent := "I'll connect you with a human agent who can better assist you. Please wait a moment."
	if s.config.AutoHandoffTime > 0 {
		if time.Since(session.CreatedAt) > s.config.AutoHandoffTime {
			fmt.Println("Session exceeded auto handoff time:", session.ID)
			messageContent = "It seems we've been chatting for a while. I'll connect you with a human agent who can better assist you. Please wait a moment."
		}
	}

	handoffMessage := &models.SendChatMessageRequest{
		Content:     messageContent,
		MessageType: "text",
		IsPrivate:   false,
		Metadata: map[string]interface{}{
			"ai_generated":   true,
			"handoff_reason": reason,
		},
	}

	aiAgentID := s.generateAIAgentID(session.ID)

	go s.chatSessionService.SendMessage(
		ctx,
		session,
		handoffMessage,
		"ai-agent",
		&aiAgentID,
		"AI Assistant",
		connID,
	)

	// Send real-time handoff notification to all agents in the project
	fmt.Printf("🤝 Sending handoff notification for session %s to project %s\n", session.ID, session.ProjectID)

	// Use HowlingAlarmService to alert all agents in the project
	if s.howlingAlarmService != nil {
		fmt.Printf("🚨 Triggering handoff alarm for session %s\n", session.ID)

		// Create metadata for the handoff alarm
		metadata := models.JSONMap{
			"handoff_reason":     reason,
			"session_id":         session.ID,
			"customer_id":        session.CustomerID,
			"widget_id":          session.WidgetID,
			"session_created_at": session.CreatedAt,
			"handoff_type":       "ai_to_human",
		}

		// Trigger the alarm with high priority for human agent requests
		alarm, alarmErr := s.howlingAlarmService.TriggerAlarm(
			ctx,
			uuid.New(), // assignment ID - generate a new one for this handoff
			uuid.Nil,   // agent ID - no specific agent yet
			session.TenantID,
			session.ProjectID,
			"Human Agent Requested",
			fmt.Sprintf("Customer in session %s is requesting human assistance: %s", session.ID, reason),
			models.NotificationPriorityHigh, // priority
			metadata,
		)

		if alarmErr != nil {
			fmt.Printf("❌ Failed to trigger handoff alarm: %v\n", alarmErr)
		} else {
			fmt.Printf("✅ Handoff alarm triggered successfully: %s\n", alarm.ID)
		}
	} else {
		fmt.Printf("⚠️ HowlingAlarmService not available for handoff notification\n")
	}

	return nil
}

// generateResponseWithContext generates AI response with knowledge context
func (s *AIService) generateResponseWithContext(ctx context.Context, session *models.ChatSession, messages []models.ChatMessage, userMessage string) (string, error) {
	// Get relevant knowledge context if knowledge service is available
	var knowledgeContext string
	if s.knowledgeService != nil {
		contextResults, err := s.knowledgeService.GetRelevantContext(ctx, session.TenantID, session.ProjectID, userMessage)
		if err != nil {
			// Log error but don't fail - continue without knowledge context
			fmt.Printf("Error getting knowledge context: %v\n", err)
		} else if len(contextResults) > 0 {
			knowledgeContext = s.knowledgeService.FormatContextForAI(contextResults)
		}
	}

	// Build conversation context with knowledge
	chatMessages := []ChatCompletionMessage{}

	// Enhanced system prompt with knowledge context
	systemPrompt := s.config.SystemPrompt
	if knowledgeContext != "" {
		systemPrompt = fmt.Sprintf("%s\n\n%s", systemPrompt, knowledgeContext)
	}

	chatMessages = append(chatMessages, ChatCompletionMessage{
		Role:    "system",
		Content: systemPrompt,
	})

	// Add recent conversation history
	for _, msg := range messages {
		role := "user"
		if msg.AuthorType == "ai-agent" || msg.AuthorType == "agent" {
			role = "assistant"
		}

		chatMessages = append(chatMessages, ChatCompletionMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	// Create request
	req := ChatCompletionRequest{
		Model:       s.config.Model,
		Messages:    chatMessages,
		MaxTokens:   s.config.MaxTokens,
		Temperature: s.config.Temperature,
	}

	// Make API call based on provider
	switch AIProvider(s.config.Provider) {
	case ProviderOpenAI:
		return s.callOpenAI(ctx, req)
	case ProviderAnthropic:
		return s.callAnthropic(ctx, req)
	case ProviderAzure:
		return s.callAzureOpenAI(ctx, req)
	default:
		return "", fmt.Errorf("unsupported AI provider: %s", s.config.Provider)
	}
}

// callOpenAI makes API call to OpenAI
func (s *AIService) callOpenAI(ctx context.Context, req ChatCompletionRequest) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"
	if s.config.BaseURL != "" {
		url = s.config.BaseURL + "/v1/chat/completions"
	}

	return s.makeAPICall(ctx, url, req, map[string]string{
		"Authorization": "Bearer " + s.config.APIKey,
		"Content-Type":  "application/json",
	})
}

// callAnthropic makes API call to Anthropic Claude
func (s *AIService) callAnthropic(ctx context.Context, req ChatCompletionRequest) (string, error) {
	// Convert to Anthropic format
	anthropicReq := map[string]interface{}{
		"model":      req.Model,
		"max_tokens": req.MaxTokens,
		"messages":   req.Messages[1:],        // Skip system message
		"system":     req.Messages[0].Content, // System message separately
	}

	url := "https://api.anthropic.com/v1/messages"
	if s.config.BaseURL != "" {
		url = s.config.BaseURL + "/v1/messages"
	}

	headers := map[string]string{
		"x-api-key":         s.config.APIKey,
		"Content-Type":      "application/json",
		"anthropic-version": "2023-06-01",
	}

	jsonData, err := json.Marshal(anthropicReq)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API call failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse Anthropic response format
	var anthropicResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return "", err
	}

	if len(anthropicResp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return anthropicResp.Content[0].Text, nil
}

// callAzureOpenAI makes API call to Azure OpenAI
func (s *AIService) callAzureOpenAI(ctx context.Context, req ChatCompletionRequest) (string, error) {
	if s.config.BaseURL == "" {
		return "", fmt.Errorf("base URL required for Azure OpenAI")
	}

	url := s.config.BaseURL + "/openai/deployments/" + s.config.Model + "/chat/completions?api-version=2023-12-01-preview"

	return s.makeAPICall(ctx, url, req, map[string]string{
		"api-key":      s.config.APIKey,
		"Content-Type": "application/json",
	})
}

// makeAPICall is a helper function for making HTTP API calls
func (s *AIService) makeAPICall(ctx context.Context, url string, req ChatCompletionRequest, headers map[string]string) (string, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API call failed with status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// AcceptHandoff handles agent accepting a handoff request
func (s *AIService) AcceptHandoff(ctx context.Context, tenantID, projectID, sessionID, agentID uuid.UUID) error {
	// Assign the agent to the session
	err := s.chatSessionService.AssignAgent(ctx, tenantID, projectID, sessionID, agentID)
	if err != nil {
		return fmt.Errorf("failed to assign agent to session: %w", err)
	}

	// TODO: Send notification to other agents that this handoff was accepted
	// TODO: Update any handoff tracking state
	// TODO: Notify the customer that an agent has joined

	return nil
}

// DeclineHandoff handles agent declining a handoff request
func (s *AIService) DeclineHandoff(ctx context.Context, tenantID, projectID, sessionID, agentID uuid.UUID) error {
	// TODO: Mark that this agent declined the handoff
	// TODO: Potentially notify other agents or escalate
	// TODO: Log the decline for metrics

	// For now, just return success since there's no explicit tracking yet
	return nil
}
