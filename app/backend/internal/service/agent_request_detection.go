package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/bareuptime/tms/internal/config"
)

// AgentRequestType represents different types of agent requests
type AgentRequestType string

const (
	AgentRequestTypeGeneral    AgentRequestType = "general"
	AgentRequestTypeUrgent     AgentRequestType = "urgent"
	AgentRequestTypeComplaint  AgentRequestType = "complaint"
	AgentRequestTypeTechnical  AgentRequestType = "technical"
	AgentRequestTypeBilling    AgentRequestType = "billing"
	AgentRequestTypeSupport    AgentRequestType = "support"
)

// AgentRequestUrgency represents the urgency level of an agent request
type AgentRequestUrgency string

const (
	UrgencyLow      AgentRequestUrgency = "low"
	UrgencyNormal   AgentRequestUrgency = "normal"
	UrgencyHigh     AgentRequestUrgency = "high"
	UrgencyCritical AgentRequestUrgency = "critical"
)

// AgentRequestResult contains the analysis result of agent request detection
type AgentRequestResult struct {
	IsAgentRequest   bool                `json:"is_agent_request"`
	Confidence       float64             `json:"confidence"`
	RequestType      AgentRequestType    `json:"request_type"`
	Urgency          AgentRequestUrgency `json:"urgency"`
	Keywords         []string            `json:"keywords"`
	Reasoning        []string            `json:"reasoning"`
	ProcessingTimeMs int64               `json:"processing_time_ms"`
}

// AgentRequestDetectionService handles detection of customer requests for human agents
type AgentRequestDetectionService struct {
	config *config.AgenticConfig
}

// NewAgentRequestDetectionService creates a new agent request detection service
func NewAgentRequestDetectionService(agenticConfig *config.AgenticConfig) *AgentRequestDetectionService {
	return &AgentRequestDetectionService{
		config: agenticConfig,
	}
}

// DetectAgentRequest analyzes a message to determine if the customer is requesting a human agent
func (s *AgentRequestDetectionService) DetectAgentRequest(ctx context.Context, message string) (*AgentRequestResult, error) {
	startTime := time.Now()
	
	// Check if agent request detection is enabled
	if !s.config.AgentRequestDetection {
		return &AgentRequestResult{
			IsAgentRequest:   false,
			Confidence:       0.0,
			RequestType:      AgentRequestTypeGeneral,
			Urgency:          UrgencyLow,
			Keywords:         []string{},
			Reasoning:        []string{"Agent request detection is disabled"},
			ProcessingTimeMs: time.Since(startTime).Milliseconds(),
		}, nil
	}

	message = strings.ToLower(strings.TrimSpace(message))
	if message == "" {
		return &AgentRequestResult{
			IsAgentRequest:   false,
			Confidence:       0.0,
			RequestType:      AgentRequestTypeGeneral,
			Urgency:          UrgencyLow,
			Keywords:         []string{},
			Reasoning:        []string{"Empty message"},
			ProcessingTimeMs: time.Since(startTime).Milliseconds(),
		}, nil
	}

	result := &AgentRequestResult{
		Keywords: []string{},
		Reasoning: []string{},
	}

	// Analyze different aspects of the message
	s.analyzeExplicitAgentRequests(message, result)
	s.analyzeComplaintIndicators(message, result)
	s.analyzeUrgencyIndicators(message, result)
	s.analyzeContextualRequests(message, result)
	s.analyzeTechnicalEscalation(message, result)
	s.analyzeBillingEscalation(message, result)

	// Calculate final confidence and determine if it's an agent request
	s.calculateFinalScore(result)

	// Determine request type based on keywords and context
	s.determineRequestType(message, result)

	// Determine urgency level
	s.determineUrgencyLevel(message, result)

	result.ProcessingTimeMs = time.Since(startTime).Milliseconds()

	log.Debug().
		Str("message", message).
		Bool("is_agent_request", result.IsAgentRequest).
		Float64("confidence", result.Confidence).
		Str("type", string(result.RequestType)).
		Str("urgency", string(result.Urgency)).
		Strs("keywords", result.Keywords).
		Msg("Agent request detection completed")

	return result, nil
}

// analyzeExplicitAgentRequests looks for direct requests for human agents
func (s *AgentRequestDetectionService) analyzeExplicitAgentRequests(message string, result *AgentRequestResult) {
	explicitPatterns := []struct {
		pattern    string
		confidence float64
		keywords   []string
	}{
		{`\b(speak|talk|chat)\s+(?:to|with)\s+(?:a\s+)?(?:human|person|agent|representative|someone)\b`, 0.9, []string{"speak to human", "agent"}},
		{`\b(?:can\s+i|i\s+(?:want|need|would\s+like)\s+to)\s+(?:speak|talk|chat)\s+(?:to|with)\s+(?:a\s+)?(?:human|person|agent|representative|someone)\b`, 0.9, []string{"want to speak", "agent"}},
		{`\b(?:connect|transfer)\s+me\s+(?:to|with)\s+(?:a\s+)?(?:human|person|agent|representative|someone|customer\s+service|support)\b`, 0.85, []string{"connect me", "agent"}},
		{`\b(?:i\s+need|get\s+me)\s+(?:a\s+)?(?:human|person|agent|representative|someone)\b`, 0.8, []string{"need human", "agent"}},
		{`\breal\s+(?:human|person|agent)\b`, 0.8, []string{"real human", "agent"}},
		{`\b(?:human|live)\s+(?:agent|support|person|representative)\b`, 0.85, []string{"human agent", "live support"}},
		{`\b(?:customer\s+)?(?:service|support)\s+(?:agent|representative|person)\b`, 0.75, []string{"customer service", "agent"}},
		{`\boperator\b`, 0.7, []string{"operator"}},
		{`\b(?:escalate|supervisor|manager)\b`, 0.8, []string{"escalate", "supervisor"}},
	}

	for _, pattern := range explicitPatterns {
		if matched, _ := regexp.MatchString(pattern.pattern, message); matched {
			result.Confidence += pattern.confidence
			result.Keywords = append(result.Keywords, pattern.keywords...)
			result.Reasoning = append(result.Reasoning, fmt.Sprintf("Explicit agent request detected: %s", pattern.keywords[0]))
		}
	}
}

// analyzeComplaintIndicators looks for complaint language that might need agent escalation
func (s *AgentRequestDetectionService) analyzeComplaintIndicators(message string, result *AgentRequestResult) {
	complaintPatterns := []struct {
		pattern    string
		confidence float64
		keywords   []string
	}{
		{`\b(?:frustrated|angry|upset|disappointed|unsatisfied)\b`, 0.4, []string{"frustrated", "complaint"}},
		{`\b(?:terrible|awful|horrible|worst)\s+(?:service|experience|support)\b`, 0.6, []string{"terrible service", "complaint"}},
		{`\b(?:not\s+working|broken|doesn't\s+work|isn't\s+working)\b`, 0.3, []string{"not working", "issue"}},
		{`\b(?:charged|billed)\s+(?:twice|wrong|incorrectly|error)\b`, 0.5, []string{"billing error", "complaint"}},
		{`\bthis\s+is\s+(?:ridiculous|unacceptable|outrageous)\b`, 0.7, []string{"unacceptable", "complaint"}},
		{`\bi\s+(?:demand|want)\s+(?:a\s+)?(?:refund|compensation)\b`, 0.7, []string{"demand refund", "complaint"}},
		{`\b(?:sue|lawsuit|legal\s+action|attorney|lawyer)\b`, 0.8, []string{"legal threat", "urgent"}},
	}

	for _, pattern := range complaintPatterns {
		if matched, _ := regexp.MatchString(pattern.pattern, message); matched {
			result.Confidence += pattern.confidence
			result.Keywords = append(result.Keywords, pattern.keywords...)
			result.Reasoning = append(result.Reasoning, fmt.Sprintf("Complaint indicator detected: %s", pattern.keywords[0]))
		}
	}
}

// analyzeUrgencyIndicators looks for urgent language
func (s *AgentRequestDetectionService) analyzeUrgencyIndicators(message string, result *AgentRequestResult) {
	urgentPatterns := []struct {
		pattern    string
		confidence float64
		keywords   []string
	}{
		{`\b(?:urgent|emergency|asap|immediately|right\s+now|urgent)\b`, 0.6, []string{"urgent", "emergency"}},
		{`\b(?:critical|mission\s+critical|production\s+down|outage)\b`, 0.7, []string{"critical", "outage"}},
		{`\b(?:losing\s+money|revenue\s+impact|business\s+impact)\b`, 0.6, []string{"business impact", "urgent"}},
		{`\b(?:deadline|time\s+sensitive|running\s+out\s+of\s+time)\b`, 0.5, []string{"deadline", "time sensitive"}},
		{`\bhelp\s+(?:me\s+)?(?:asap|now|immediately|urgently)\b`, 0.6, []string{"help now", "urgent"}},
	}

	for _, pattern := range urgentPatterns {
		if matched, _ := regexp.MatchString(pattern.pattern, message); matched {
			result.Confidence += pattern.confidence
			result.Keywords = append(result.Keywords, pattern.keywords...)
			result.Reasoning = append(result.Reasoning, fmt.Sprintf("Urgency indicator detected: %s", pattern.keywords[0]))
		}
	}
}

// analyzeContextualRequests looks for contextual clues that suggest agent needed
func (s *AgentRequestDetectionService) analyzeContextualRequests(message string, result *AgentRequestResult) {
	contextualPatterns := []struct {
		pattern    string
		confidence float64
		keywords   []string
	}{
		{`\b(?:can\s+someone|could\s+someone|is\s+there\s+someone)\s+(?:help|assist)\b`, 0.6, []string{"someone help", "assistance"}},
		{`\bi\s+(?:need|require)\s+(?:help|assistance|support)\s+(?:with|from)\b`, 0.4, []string{"need help", "assistance"}},
		{`\b(?:this\s+(?:bot|chatbot|system)|you)\s+(?:can't|cannot|isn't|doesn't)\s+(?:help|understand|solve)\b`, 0.7, []string{"bot cannot help", "escalation"}},
		{`\bi\s+(?:already\s+)?tried\s+(?:that|everything|this)\s+(?:and\s+it\s+(?:doesn't|didn't)\s+work|but\s+it\s+(?:doesn't|didn't)\s+work)\b`, 0.5, []string{"tried everything", "escalation"}},
		{`\b(?:still\s+not\s+working|still\s+having\s+(?:issues|problems)|doesn't\s+solve\s+my\s+problem)\b`, 0.4, []string{"still not working", "escalation"}},
		{`\b(?:complex|complicated)\s+(?:issue|problem|situation)\b`, 0.3, []string{"complex issue", "escalation"}},
	}

	for _, pattern := range contextualPatterns {
		if matched, _ := regexp.MatchString(pattern.pattern, message); matched {
			result.Confidence += pattern.confidence
			result.Keywords = append(result.Keywords, pattern.keywords...)
			result.Reasoning = append(result.Reasoning, fmt.Sprintf("Contextual request detected: %s", pattern.keywords[0]))
		}
	}
}

// analyzeTechnicalEscalation looks for technical issues that might need escalation
func (s *AgentRequestDetectionService) analyzeTechnicalEscalation(message string, result *AgentRequestResult) {
	technicalPatterns := []struct {
		pattern    string
		confidence float64
		keywords   []string
	}{
		{`\b(?:api|integration|webhook|ssl|certificate|database|server)\s+(?:error|issue|problem|not\s+working)\b`, 0.5, []string{"technical issue", "api error"}},
		{`\b(?:deployment|production|staging|environment)\s+(?:issue|problem|error)\b`, 0.6, []string{"deployment issue", "technical"}},
		{`\b(?:authentication|authorization|login|access)\s+(?:issue|problem|error|denied)\b`, 0.4, []string{"auth issue", "technical"}},
		{`\b(?:configuration|setup|installation)\s+(?:help|issue|problem)\b`, 0.4, []string{"configuration help", "technical"}},
	}

	for _, pattern := range technicalPatterns {
		if matched, _ := regexp.MatchString(pattern.pattern, message); matched {
			result.Confidence += pattern.confidence
			result.Keywords = append(result.Keywords, pattern.keywords...)
			result.Reasoning = append(result.Reasoning, fmt.Sprintf("Technical escalation detected: %s", pattern.keywords[0]))
		}
	}
}

// analyzeBillingEscalation looks for billing issues that might need escalation
func (s *AgentRequestDetectionService) analyzeBillingEscalation(message string, result *AgentRequestResult) {
	billingPatterns := []struct {
		pattern    string
		confidence float64
		keywords   []string
	}{
		{`\b(?:billing|invoice|payment|charge|subscription)\s+(?:issue|problem|error|dispute)\b`, 0.5, []string{"billing issue", "payment"}},
		{`\b(?:refund|chargeback|dispute|cancel)\s+(?:request|charge|payment|subscription)\b`, 0.6, []string{"refund request", "billing"}},
		{`\b(?:overcharged|double\s+charged|incorrect\s+amount|wrong\s+charge)\b`, 0.6, []string{"billing error", "overcharged"}},
		{`\b(?:upgrade|downgrade|change\s+plan|billing\s+cycle)\b`, 0.3, []string{"plan change", "billing"}},
	}

	for _, pattern := range billingPatterns {
		if matched, _ := regexp.MatchString(pattern.pattern, message); matched {
			result.Confidence += pattern.confidence
			result.Keywords = append(result.Keywords, pattern.keywords...)
			result.Reasoning = append(result.Reasoning, fmt.Sprintf("Billing escalation detected: %s", pattern.keywords[0]))
		}
	}
}

// calculateFinalScore determines if this is an agent request based on accumulated confidence
func (s *AgentRequestDetectionService) calculateFinalScore(result *AgentRequestResult) {
	// Normalize confidence to 0-1 range
	if result.Confidence > 1.0 {
		result.Confidence = 1.0
	}

	// Apply threshold from configuration
	threshold := s.config.AgentRequestThreshold
	if threshold <= 0 {
		threshold = 0.7 // Default threshold
	}

	result.IsAgentRequest = result.Confidence >= threshold

	if result.IsAgentRequest {
		result.Reasoning = append(result.Reasoning, fmt.Sprintf("Confidence %.2f meets threshold %.2f", result.Confidence, threshold))
	} else {
		result.Reasoning = append(result.Reasoning, fmt.Sprintf("Confidence %.2f below threshold %.2f", result.Confidence, threshold))
	}
}

// determineRequestType categorizes the type of agent request
func (s *AgentRequestDetectionService) determineRequestType(message string, result *AgentRequestResult) {
	if !result.IsAgentRequest {
		result.RequestType = AgentRequestTypeGeneral
		return
	}

	// Check for specific request types based on keywords
	keywordString := strings.Join(result.Keywords, " ")

	if strings.Contains(keywordString, "billing") || strings.Contains(keywordString, "payment") || strings.Contains(keywordString, "refund") {
		result.RequestType = AgentRequestTypeBilling
	} else if strings.Contains(keywordString, "technical") || strings.Contains(keywordString, "api") || strings.Contains(keywordString, "configuration") {
		result.RequestType = AgentRequestTypeTechnical
	} else if strings.Contains(keywordString, "complaint") || strings.Contains(keywordString, "frustrated") || strings.Contains(keywordString, "unacceptable") {
		result.RequestType = AgentRequestTypeComplaint
	} else if strings.Contains(keywordString, "urgent") || strings.Contains(keywordString, "emergency") || strings.Contains(keywordString, "critical") {
		result.RequestType = AgentRequestTypeUrgent
	} else if strings.Contains(keywordString, "help") || strings.Contains(keywordString, "support") || strings.Contains(keywordString, "assistance") {
		result.RequestType = AgentRequestTypeSupport
	} else {
		result.RequestType = AgentRequestTypeGeneral
	}
}

// determineUrgencyLevel determines the urgency of the agent request
func (s *AgentRequestDetectionService) determineUrgencyLevel(message string, result *AgentRequestResult) {
	if !result.IsAgentRequest {
		result.Urgency = UrgencyLow
		return
	}

	keywordString := strings.Join(result.Keywords, " ")

	// Critical urgency
	if strings.Contains(keywordString, "legal threat") || strings.Contains(keywordString, "critical") || strings.Contains(keywordString, "outage") {
		result.Urgency = UrgencyCritical
	} else if strings.Contains(keywordString, "urgent") || strings.Contains(keywordString, "emergency") || strings.Contains(keywordString, "business impact") {
		result.Urgency = UrgencyHigh
	} else if strings.Contains(keywordString, "complaint") || strings.Contains(keywordString, "billing error") || strings.Contains(keywordString, "escalation") {
		result.Urgency = UrgencyNormal
	} else {
		result.Urgency = UrgencyLow
	}
}
