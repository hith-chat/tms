package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/logger"
	"github.com/bareuptime/tms/internal/models"
	ws "github.com/bareuptime/tms/internal/websocket"
)

// AIService handles AI-powered chat assistance
type AIService struct {
	config              *config.AIConfig
	agenticConfig       *config.AgenticConfig
	chatSessionService  *ChatSessionService
	knowledgeService    *KnowledgeService
	usageService        *AIUsageService
	greetingDetection   *GreetingDetectionService
	brandGreeting       *BrandGreetingService
	connectionManager   *ws.ConnectionManager
	howlingAlarmService *HowlingAlarmService
	httpClient          *http.Client
}

// NewAIService creates a new AI service instance
func NewAIService(cfg *config.AIConfig, agenticConfig *config.AgenticConfig, chatSessionService *ChatSessionService, knowledgeService *KnowledgeService, usageService *AIUsageService, greetingDetection *GreetingDetectionService, brandGreeting *BrandGreetingService, connectionManager *ws.ConnectionManager, howlingAlarmService *HowlingAlarmService) *AIService {
	return &AIService{
		config:              cfg,
		agenticConfig:       agenticConfig,
		chatSessionService:  chatSessionService,
		knowledgeService:    knowledgeService,
		usageService:        usageService,
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
	ProviderBB        AIProvider = "bb"
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
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // Can be string or []ContentPart for multi-modal
}

// ContentPart represents a part of multi-modal content (text or image)
type ContentPart struct {
	Type     string    `json:"type"` // "text" or "image"
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// ImageURL represents an image in the content
type ImageURL struct {
	URL    string `json:"url"`              // base64 data URL or regular URL
	Detail string `json:"detail,omitempty"` // "low", "high", or "auto"
}

// ChatCompletionResponse represents the response from AI providers
type ChatCompletionResponse struct {
	Choices []struct {
		Message ChatCompletionMessage `json:"message"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
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
	go s.ProcessAiTyping(session, models.WSMessage{}, connID, true)
	resp, err := s.processComplexMessage(ctx, session, messageContent, connID)
	go s.ProcessAiTyping(session, models.WSMessage{}, connID, false)
	return resp, err
}

// processVisitorTyping handles visitor typing indicators
func (s *AIService) ProcessAiTyping(session *models.ChatSession, msg models.WSMessage, connID string, isTyping bool) {

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
	s.connectionManager.DeliverWebSocketMessage(session.ID, broadcastMsg)
	typingDataAgent, _ := json.Marshal(map[string]interface{}{
		"author_name": "AI",
		"author_type": "agent",
	})
	broadcastMsg.Data = typingDataAgent
	broadcastMsg.FromType = ws.ConnectionTypeAgent
	go s.connectionManager.SendToConnection(connID, broadcastMsg)
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
	return s.SendAIResponse(ctx, session, connID, response, map[string]interface{}{
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
	response, usageMetrics, err := s.generateResponseWithContext(ctx, session, recentMessages, messageContent)
	if err != nil {
		fmt.Println("Error generating AI response:", err.Error())
		return nil, fmt.Errorf("failed to generate AI response: %w", err)
	}

	fmt.Println("Response from ai -", response)

	if usageMetrics != nil && s.usageService != nil {
		if _, err := s.usageService.DeductUsage(ctx, UsageDeductionInput{
			TenantID:  session.TenantID,
			ProjectID: session.ProjectID,
			Model:     s.config.Model,
			SessionID: &session.ID,
			Metrics:   *usageMetrics,
		}); err != nil {
			fmt.Printf("Failed to deduct AI usage credits: %v\n", err)
		}
	}

	// Send the AI response
	return s.SendAIResponse(ctx, session, connID, response, map[string]interface{}{
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
	response, usageMetrics, err := s.generateResponseWithContext(ctx, session, recentMessages, messageContent)
	if err != nil {
		fmt.Println("Error generating AI response:", err.Error())
		return nil, fmt.Errorf("failed to generate AI response: %w", err)
	}

	fmt.Println("Response from ai -", response)

	if usageMetrics != nil && s.usageService != nil {
		if _, err := s.usageService.DeductUsage(ctx, UsageDeductionInput{
			TenantID:  session.TenantID,
			ProjectID: session.ProjectID,
			Model:     s.config.Model,
			SessionID: &session.ID,
			Metrics:   *usageMetrics,
		}); err != nil {
			fmt.Printf("Failed to deduct AI usage credits: %v\n", err)
		}
	}

	// Send the AI response
	return s.SendAIResponse(ctx, session, connID, response, map[string]interface{}{
		"ai_generated":  true,
		"response_type": "knowledge_based",
	})
}

// SendAIResponse is a helper method to send AI responses
func (s *AIService) SendAIResponse(ctx context.Context, session *models.ChatSession, connID, content string, metadata map[string]interface{}) (*models.ChatMessage, error) {
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
func (s *AIService) generateResponseWithContext(ctx context.Context, session *models.ChatSession, messages []models.ChatMessage, userMessage string) (string, *TokenUsageMetrics, error) {
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
	case ProviderBB:
		return s.callBB(ctx, req)
	default:
		return "", nil, fmt.Errorf("unsupported AI provider: %s", s.config.Provider)
	}
}

func (s *AIService) generateResponseForAIRequest(ctx context.Context, req ChatCompletionRequest) (string, *TokenUsageMetrics, error) {
	// Get relevant knowledge context if knowledge service is available
	// Make API call based on provider
	switch AIProvider(s.config.Provider) {
	case ProviderOpenAI:
		return s.callOpenAI(ctx, req)
	case ProviderAnthropic:
		return s.callAnthropic(ctx, req)
	case ProviderAzure:
		return s.callAzureOpenAI(ctx, req)
	case ProviderBB:
		return s.callBB(ctx, req)
	default:
		return "", nil, fmt.Errorf("unsupported AI provider: %s", s.config.Provider)
	}
}

// callOpenAI makes API call to OpenAI
func (s *AIService) callOpenAI(ctx context.Context, req ChatCompletionRequest) (string, *TokenUsageMetrics, error) {
	url := "https://api.openai.com/v1/chat/completions"
	if s.config.BaseURL != "" {
		url = s.config.BaseURL + "/v1/chat/completions"
	}

	return s.makeAPICall(ctx, url, req, map[string]string{
		"Authorization": "Bearer " + s.config.APIKey,
		"Content-Type":  "application/json",
	})
}

func (s *AIService) callBB(ctx context.Context, req ChatCompletionRequest) (string, *TokenUsageMetrics, error) {

	logger.Info("Calling BB AI Provider")

	url := ""
	if s.config.BaseURL != "" {
		url = s.config.BaseURL + "/v1/chat/completions"
	}
	return s.makeAPICall(ctx, url, req, map[string]string{
		"api-key":      s.config.APIKey,
		"Content-Type": "application/json",
	})
}

// callAnthropic makes API call to Anthropic Claude
func (s *AIService) callAnthropic(ctx context.Context, req ChatCompletionRequest) (string, *TokenUsageMetrics, error) {
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
		return "", nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", nil, err
	}

	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("API call failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse Anthropic response format
	var anthropicResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return "", nil, err
	}

	if len(anthropicResp.Content) == 0 {
		return "", nil, fmt.Errorf("no content in response")
	}

	return anthropicResp.Content[0].Text, nil, nil
}

// callAzureOpenAI makes API call to Azure OpenAI
func (s *AIService) callAzureOpenAI(ctx context.Context, req ChatCompletionRequest) (string, *TokenUsageMetrics, error) {
	if s.config.BaseURL == "" {
		return "", nil, fmt.Errorf("base URL required for Azure OpenAI")
	}

	url := s.config.BaseURL + "/openai/deployments/" + s.config.Model + "/chat/completions?api-version=2023-12-01-preview"

	return s.makeAPICall(ctx, url, req, map[string]string{
		"api-key":      s.config.APIKey,
		"Content-Type": "application/json",
	})
}

// makeAPICall is a helper function for making HTTP API calls
func (s *AIService) makeAPICall(ctx context.Context, url string, req ChatCompletionRequest, headers map[string]string) (string, *TokenUsageMetrics, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", nil, err
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 300*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctxWithTimeout, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", nil, err
	}

	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("API call failed with status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", nil, err
	}

	if len(chatResp.Choices) == 0 {
		return "", nil, fmt.Errorf("no choices in response")
	}

	var usageMetrics *TokenUsageMetrics
	if chatResp.Usage != nil {
		usageMetrics = &TokenUsageMetrics{
			PromptTokens:     int64(chatResp.Usage.PromptTokens),
			CompletionTokens: int64(chatResp.Usage.CompletionTokens),
			TotalTokens:      int64(chatResp.Usage.TotalTokens),
		}
	}

	// Extract content from the message (handle both string and multi-modal content)
	content := ""
	switch v := chatResp.Choices[0].Message.Content.(type) {
	case string:
		content = v
	case []interface{}:
		// For multi-modal responses, extract text parts
		for _, part := range v {
			if partMap, ok := part.(map[string]interface{}); ok {
				if partMap["type"] == "text" {
					if text, ok := partMap["text"].(string); ok {
						content += text
					}
				}
			}
		}
	}

	return content, usageMetrics, nil
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

// ThemeData interface to avoid circular imports
type ThemeData interface {
	GetColors() []string
	GetBackgroundHues() []string
	GetFontFamilies() []string
	GetBrandName() string
	GetPageTitle() string
	GetMetaDesc() string
}

// KnowledgeLinkCandidate represents a discovered link that can be evaluated for indexing
type KnowledgeLinkCandidate struct {
	URL        string
	Title      string
	Depth      int
	TokenCount int
}

// KnowledgeLinkSelection is the AI-ranked result for which links to index automatically
type KnowledgeLinkSelection struct {
	URL       string `json:"url"`
	Category  string `json:"category"`
	Rationale string `json:"rationale"`
}

// KnowledgeSectionSummary summarises scraped content for FAQ generation
type KnowledgeSectionSummary struct {
	URL     string
	Title   string
	Content string
}

// GeneratedFAQItem captures AI-produced FAQ entries prior to persistence
type GeneratedFAQItem struct {
	Question   string `json:"question"`
	Answer     string `json:"answer"`
	SourceURL  string `json:"source_url"`
	SectionRef string `json:"section_reference"`
}

// GenerateWidgetTheme generates chat widget theme based on website analysis
func (s *AIService) GenerateWidgetTheme(ctx context.Context, themeData ThemeData) (*models.CreateChatWidgetRequest, error) {
	return s.GenerateWidgetThemeWithScreenshot(ctx, themeData, nil)
}

func (s *AIService) GenerateWidgetThemeWithScreenshot(ctx context.Context, themeData ThemeData, screenshot []byte) (*models.CreateChatWidgetRequest, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("AI service is not enabled")
	}

	// Prepare the prompt using unified color extraction approach
	prompt := s.buildUnifiedThemePrompt(themeData)

	// Build the user message with screenshot if available
	var userMessage ChatCompletionMessage
	if screenshot != nil && len(screenshot) > 0 {
		// Use multi-modal content with image
		base64Image := base64.StdEncoding.EncodeToString(screenshot)
		dataURL := fmt.Sprintf("data:image/png;base64,%s", base64Image)

		userMessage = ChatCompletionMessage{
			Role: "user",
			Content: []ContentPart{
				{
					Type: "text",
					Text: prompt,
				},
				{
					Type: "image_url",
					ImageURL: &ImageURL{
						URL:    dataURL,
						Detail: "high",
					},
				},
			},
		}
	} else {
		// Text-only message
		userMessage = ChatCompletionMessage{
			Role:    "user",
			Content: prompt,
		}
	}

	// Create the request
	req := ChatCompletionRequest{
		Model: s.config.ThemeExtractionModel,
		Messages: []ChatCompletionMessage{
			{
				Role:    "system",
				Content: "You are a Color Extraction Agent specializing in analyzing website screenshots and DOM data to extract optimal color schemes for chat widgets. Return only valid JSON with no explanations.",
			},
			userMessage,
		},
		Temperature: 0.2, // Lower temperature for more consistent color extraction
		MaxTokens:   800, // Reduced since we only need 3 colors + metadata
	}

	// Make the API call
	response, _, err := s.generateResponseForAIRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to call AI: %w", err)
	}
	response = stripMarkdownCodeBlocks(response)

	// Parse the JSON response
	var themeConfig models.CreateChatWidgetRequest
	if err := json.Unmarshal([]byte(response), &themeConfig); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	themeConfig.ShowAgentAvatars = true
	themeConfig.UseAI = true

	return &themeConfig, nil
}

// stripMarkdownCodeBlocks removes markdown code block syntax from AI responses
// Handles cases where the AI wraps JSON in ```json...``` or ```...```
func stripMarkdownCodeBlocks(response string) string {
	// Trim whitespace
	response = strings.TrimSpace(response)

	// Check if response starts with markdown code block
	if strings.HasPrefix(response, "```") {
		// Find the end of the opening marker (could be ```json or just ```)
		firstNewline := strings.Index(response, "\n")
		if firstNewline == -1 {
			// No newline found, might be malformed
			return response
		}

		// Skip the opening marker line
		response = response[firstNewline+1:]

		// Remove closing ``` if present (TrimSuffix is safe even if not present)
		response = strings.TrimSuffix(response, "```")

		// Trim any remaining whitespace
		response = strings.TrimSpace(response)
	}

	return response
}

// buildUnifiedThemePrompt creates a unified prompt for theme extraction that works with or without screenshots
func (s *AIService) buildUnifiedThemePrompt(themeData ThemeData) string {
	return fmt.Sprintf(`You are a Color Extraction Agent. You will receive a screenshot of a website and/or extracted DOM data.
Your task is to analyze ONLY the colors visible and return exactly 3 colors for building a chat widget theme.

The output MUST be pure JSON with NO explanation.

----------------------------------------------------
WEBSITE CONTEXT:
----------------------------------------------------
- Brand Name: %s
- Page Title: %s
- Meta Description: %s
- Detected Colors (from DOM): %v
- Background Colors (from DOM): %v
- Font Families: %v

----------------------------------------------------
EXTRACTION RULES
----------------------------------------------------

PRIMARY COLOR
Definition:
- The main accent or highlight color on the website.
- Typically the most saturated or visually dominant color.
- Usually appears in: buttons, CTAs, gradients, brand accents, highlights, or important text.

Usage:
- User chat bubble background
- Chat title bar background
- Agent's message bubble background (white text on top)

Rules:
- If the primary accent is a gradient, identify all colors in the gradient.
- Choose the MOST DOMINANT or VISUALLY STRONGEST color from that gradient.
- Return only one HEX value.
- MUST work with WHITE (#FFFFFF) text - ensure minimum 4.5:1 contrast ratio.

----------------------------------------------------
SECONDARY COLOR
Definition:
- A lighter or softer supporting UI color that contrasts with the primary color.
- Found in: secondary accents, pastel sections, light UI surfaces, tag backgrounds, muted gradients.

Usage:
- AI/Agent message bubble background
- Black text will be placed on top (so this MUST be a light color)

Rules:
- If a secondary element uses a gradient, extract the gradient colors.
- Choose the lightest or most appropriate color that ensures contrast with black text.
- Return only one HEX value.
- MUST work with DARK text (#1F2937 or #000000) - ensure minimum 4.5:1 contrast ratio.

----------------------------------------------------
BACKGROUND COLOR
Definition:
- The dominant background color of the website.
- Found in: hero section, site background, large layout containers.

Usage:
- Overall chat background

Rules:
- Choose the most visually present background color.
- Avoid accents or highlight colors.
- Usually white (#FFFFFF) or very light gray (#F9FAFB, #F3F4F6).

----------------------------------------------------
GRADIENT HANDLING RULE
----------------------------------------------------
If ANY UI element uses a gradient:
1. Identify all colors in the gradient.
2. Collapse the gradient into ONE representative HEX by choosing:
   - The most dominant or saturated color (for PRIMARY)
   - The lightest color suitable for black text (for SECONDARY)
3. Do NOT return gradient strings. Return one HEX color only.

----------------------------------------------------
STRICT OUTPUT FORMAT
----------------------------------------------------
Return EXACTLY this JSON structure:
{
  "primary_color": "#RRGGBB",
  "secondary_color": "#RRGGBB",
  "background_color": "#RRGGBB",
  "position": "bottom-right",
  "widget_shape": "rounded",
  "widget_size": "medium",
  "chat_bubble_style": "modern",
  "welcome_message": "Hi! How can we help you today?",
  "custom_greeting": "Welcome! We're here to assist you.",
  "agent_name": "Founder's name or 'Support Agent'",
  "font_family": "One of the detected font families from the site",
}

- Return ONLY valid 6-digit HEX codes in uppercase (e.g., #FF5733).
- No explanations.
- No additional fields.
- Ensure PRIMARY works with white text, SECONDARY with dark text.
- Messages should match brand tone and be under 100 characters.
`,
		themeData.GetBrandName(),
		themeData.GetPageTitle(),
		themeData.GetMetaDesc(),
		themeData.GetColors(),
		themeData.GetBackgroundHues(),
		themeData.GetFontFamilies(),
	)
}

// SelectTopKnowledgeLinks asks the AI model to prioritise the most relevant links for knowledge ingestion
func (s *AIService) SelectTopKnowledgeLinks(ctx context.Context, rootURL string, candidates []KnowledgeLinkCandidate, maxLinks int) ([]KnowledgeLinkSelection, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("AI service is not enabled")
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no link candidates provided")
	}

	if maxLinks <= 0 {
		maxLinks = 5
	}

	var builder strings.Builder
	for i, candidate := range candidates {
		builder.WriteString(fmt.Sprintf("Candidate %d:\n", i+1))
		builder.WriteString(fmt.Sprintf("URL: %s\n", candidate.URL))
		builder.WriteString(fmt.Sprintf("Title: %s\n", candidate.Title))
		builder.WriteString(fmt.Sprintf("Depth: %d\n", candidate.Depth))
		builder.WriteString(fmt.Sprintf("ApproxTokens: %d\n\n", candidate.TokenCount))
	}

	prompt := fmt.Sprintf(`You are selecting the %d most useful pages to ingest for a customer-support knowledge base.
The root website is %s.
Pick a mix of pages that best represent documentation, pricing, support, and key marketing information while staying within the same domain.

Here are the candidate pages:
%s

Respond with STRICT JSON using this schema (no additional commentary):
{
  "selections": [
    {"url": "https://example.com/path", "category": "pricing|docs|support|homepage|blog|product", "rationale": "Short justification under 120 characters"}
  ]
}

Return exactly %d unique URLs. Always include the homepage (%s) even if it was not in the candidates. All URLs must belong to the same domain.
`,
		maxLinks,
		rootURL,
		builder.String(),
		maxLinks,
		rootURL,
	)

	req := ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []ChatCompletionMessage{
			{Role: "system", Content: "You are an expert knowledge architect who curates high-signal website content for AI assistants."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.2,
		MaxTokens:   800,
	}

	response, _, err := s.generateResponseForAIRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to rank knowledge links: %w", err)
	}

	var parsed struct {
		Selections []KnowledgeLinkSelection `json:"selections"`
	}

	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse knowledge link selections: %w", err)
	}

	normalise := func(u string) string {
		return strings.TrimRight(strings.TrimSpace(u), "/")
	}

	rootKey := normalise(rootURL)
	selectionMap := make(map[string]KnowledgeLinkSelection)
	ordered := make([]KnowledgeLinkSelection, 0, len(parsed.Selections))
	for _, sel := range parsed.Selections {
		url := normalise(sel.URL)
		if url == "" {
			continue
		}
		if _, exists := selectionMap[url]; exists {
			continue
		}
		selectionMap[url] = sel
		ordered = append(ordered, sel)
	}

	if _, exists := selectionMap[rootKey]; !exists {
		selectionMap[rootKey] = KnowledgeLinkSelection{
			URL:       rootURL,
			Category:  "homepage",
			Rationale: "Homepage baseline coverage",
		}
		ordered = append([]KnowledgeLinkSelection{selectionMap[rootKey]}, ordered...)
	}

	if len(ordered) < maxLinks {
		candidateMap := make(map[string]KnowledgeLinkCandidate)
		for _, c := range candidates {
			candidateMap[normalise(c.URL)] = c
		}
		for _, c := range candidates {
			norm := normalise(c.URL)
			if _, exists := selectionMap[norm]; exists {
				continue
			}
			selectionMap[norm] = KnowledgeLinkSelection{
				URL:       c.URL,
				Category:  "content",
				Rationale: "High-signal page discovered during crawl",
			}
			ordered = append(ordered, selectionMap[norm])
			if len(ordered) >= maxLinks {
				break
			}
		}
	}

	if len(ordered) > maxLinks {
		ordered = ordered[:maxLinks]
	}

	return ordered, nil
}

// GenerateKnowledgeFAQ produces curated FAQ entries from scraped sections
func (s *AIService) GenerateKnowledgeFAQ(ctx context.Context, baseURL string, sections []KnowledgeSectionSummary, count int) ([]GeneratedFAQItem, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("AI service is not enabled")
	}

	if len(sections) == 0 {
		return nil, fmt.Errorf("no knowledge sections provided")
	}

	if count <= 0 {
		count = 10
	}

	var builder strings.Builder
	for i, section := range sections {
		content := strings.TrimSpace(section.Content)
		if len(content) > 1600 {
			content = content[:1600] + "..."
		}
		builder.WriteString(fmt.Sprintf("Section %d\nURL: %s\nTitle: %s\nExcerpt: %s\n\n", i+1, section.URL, section.Title, content))
	}

	prompt := fmt.Sprintf(`You are creating high-quality FAQ content for the website %s.
Use the provided sections to generate the top %d customer questions and concise answers.
Each FAQ must cite the most relevant section URL and, if possible, include a short section reference (e.g., pricing table, onboarding docs).

Return STRICT JSON (no commentary) with the following schema:
{
  "items": [
    {"question": "...", "answer": "...", "source_url": "https://example.com/path", "section_reference": "Short reference"}
  ]
}

Constraints:
- Answers must stay under 120 words.
- Use only the provided sections; do not invent URLs.
- Prefer actionable, customer-facing questions.
- Keep language aligned with the brand tone implied by the excerpts.

Here are the sections:
%s
`,
		baseURL,
		count,
		builder.String(),
	)

	req := ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []ChatCompletionMessage{
			{Role: "system", Content: "You are a world-class product support specialist who writes crisp FAQs grounded strictly in the provided material."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.3,
		MaxTokens:   1600,
	}

	response, _, err := s.generateResponseForAIRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate FAQs: %w", err)
	}

	var parsed struct {
		Items []GeneratedFAQItem `json:"items"`
	}

	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse FAQ payload: %w", err)
	}

	root := strings.TrimRight(baseURL, "/")
	fallbackURL := root

	for i := range parsed.Items {
		item := &parsed.Items[i]
		item.Question = strings.TrimSpace(item.Question)
		item.Answer = strings.TrimSpace(item.Answer)
		item.SourceURL = strings.TrimSpace(item.SourceURL)
		item.SectionRef = strings.TrimSpace(item.SectionRef)
		if item.SourceURL == "" {
			item.SourceURL = fallbackURL
		}
		if !strings.HasPrefix(item.SourceURL, root) {
			item.SourceURL = fallbackURL
		}
	}

	if len(parsed.Items) > count {
		parsed.Items = parsed.Items[:count]
	}

	return parsed.Items, nil
}

// CallAIForRanking calls AI to rank URLs and select the most relevant ones
func (s *AIService) CallAIForRanking(ctx context.Context, prompt string) (string, error) {
	if !s.IsEnabled() {
		return "", fmt.Errorf("AI service is not enabled")
	}

	// Use GPT-4 for better reasoning about relevance
	req := ChatCompletionRequest{
		Model: s.config.UrlRankingModel,
		Messages: []ChatCompletionMessage{
			{
				Role:    "system",
				Content: "You are an expert at analyzing web content and selecting the most relevant pages for building a knowledge base. You MUST respond with ONLY a valid JSON array - no explanations, no notes, no additional text before or after the JSON. Your entire response must be parseable as JSON.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.3, // Lower temperature for consistent, focused results
		MaxTokens:   1000,
	}

	// Make the API call
	response, _, err := s.generateResponseForAIRequest(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to call AI: %w", err)
	}

	return response, nil
}
