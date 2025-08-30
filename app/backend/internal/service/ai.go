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
)

// AIService handles AI-powered chat assistance
type AIService struct {
	config             *config.AIConfig
	chatSessionService *ChatSessionService
	knowledgeService   *KnowledgeService
	httpClient         *http.Client
}

// NewAIService creates a new AI service instance
func NewAIService(cfg *config.AIConfig, chatSessionService *ChatSessionService, knowledgeService *KnowledgeService) *AIService {
	fmt.Println("Creating AI Service")
	fmt.Println("AI API Key:", cfg.APIKey)
	return &AIService{
		config:             cfg,
		chatSessionService: chatSessionService,
		knowledgeService:   knowledgeService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
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
		return false
	}

	// Don't handle if already assigned to a human agent
	if session.AssignedAgentID != nil {
		return false
	}

	// Check if session has been ongoing for too long (auto handoff)
	if s.config.AutoHandoffTime > 0 {
		if time.Since(session.CreatedAt) > s.config.AutoHandoffTime {
			return false
		}
	}

	return true
}

// ProcessMessage handles incoming visitor messages and generates AI responses
func (s *AIService) ProcessMessage(ctx context.Context, session *models.ChatSession, message *models.ChatMessage) (*models.ChatMessage, error) {
	if !s.ShouldHandleSession(ctx, session) {
		return nil, nil
	}
	// Check for handoff keywords
	if s.shouldHandoffToAgent(message.Content) {
		s.requestHumanAgent(ctx, session, "Customer requested human assistance")
		return nil, nil
	}

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
	response, err := s.generateResponseWithContext(ctx, session, recentMessages, message.Content)
	if err != nil {
		fmt.Println("Error generating AI response:", err.Error())
		return nil, fmt.Errorf("failed to generate AI response: %w", err)
	}

	fmt.Println("Response from ai -", response)

	// Send AI response
	aiMessageReq := &models.SendChatMessageRequest{
		Content:     response,
		MessageType: "text",
		IsPrivate:   false,
		Metadata:    map[string]interface{}{"ai_generated": true},
	}

	// Create AI agent UUID (deterministic based on session)
	aiAgentID := s.generateAIAgentID(session.ID)

	messageAI, err := s.chatSessionService.SendMessage(
		ctx,
		session.TenantID,
		session.ProjectID,
		session.ID,
		aiMessageReq,
		"ai-agent",
		&aiAgentID,
		"AI Assistant",
	)

	return messageAI, nil
}

// generateAIAgentID creates a deterministic UUID for AI agent
func (s *AIService) generateAIAgentID(sessionID uuid.UUID) uuid.UUID {
	// Create a deterministic UUID based on session ID
	namespace := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000") // AI namespace
	return uuid.NewSHA1(namespace, sessionID[:])
}

// shouldHandoffToAgent checks if the message contains handoff keywords
func (s *AIService) shouldHandoffToAgent(content string) bool {
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
func (s *AIService) requestHumanAgent(ctx context.Context, session *models.ChatSession, reason string) error {
	// Send notification message about handoff
	handoffMessage := &models.SendChatMessageRequest{
		Content:     "I'll connect you with a human agent who can better assist you. Please wait a moment.",
		MessageType: "text",
		IsPrivate:   false,
		Metadata: map[string]interface{}{
			"ai_generated":   true,
			"handoff_reason": reason,
		},
	}

	aiAgentID := s.generateAIAgentID(session.ID)

	_, err := s.chatSessionService.SendMessage(
		ctx,
		session.TenantID,
		session.ProjectID,
		session.ID,
		handoffMessage,
		"ai-agent",
		&aiAgentID,
		"AI Assistant",
	)

	// Update session status to indicate human agent needed
	// This could trigger notifications to available agents

	return err
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
