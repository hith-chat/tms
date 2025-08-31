package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/config"
)

// AutoResponseDecision represents the final decision on how to handle a message
type AutoResponseDecision struct {
	ShouldRespond       bool                          `json:"should_respond"`
	ResponseType        string                        `json:"response_type"` // greeting, knowledge, escalation, out_of_domain
	Response            string                        `json:"response"`
	Confidence          float64                       `json:"confidence"`
	ReasoningSteps      []string                      `json:"reasoning_steps"`
	RequiresEscalation  bool                          `json:"requires_escalation"`
	EscalationReason    string                        `json:"escalation_reason"`
	ProcessingTime      time.Duration                 `json:"processing_time"`
	Classification      *QuestionClassificationResult `json:"classification"`
	KnowledgeResponse   *KnowledgeResponseResult      `json:"knowledge_response,omitempty"`
	GreetingDetection   *GreetingResult               `json:"greeting_detection,omitempty"`
	BrandResponse       string                        `json:"brand_response,omitempty"`
	Citations           []string                      `json:"citations,omitempty"`
	Metadata            map[string]interface{}        `json:"metadata"`
}

// AutoResponseDecisionService coordinates all agentic services to make intelligent response decisions
type AutoResponseDecisionService struct {
	config                *config.AgenticConfig
	greetingDetection     *GreetingDetectionService
	brandGreeting         *BrandGreetingService
	questionClassifier    *QuestionClassificationService
	knowledgeResponse     *KnowledgeResponseService
	
	// Decision thresholds
	highConfidenceThreshold    float64
	mediumConfidenceThreshold  float64
	lowConfidenceThreshold     float64
	
	// Fallback responses
	fallbackResponses map[string]string
}

// NewAutoResponseDecisionService creates a new auto response decision service
func NewAutoResponseDecisionService(
	config *config.AgenticConfig,
	greetingDetection *GreetingDetectionService,
	brandGreeting *BrandGreetingService,
	questionClassifier *QuestionClassificationService,
	knowledgeResponse *KnowledgeResponseService,
) *AutoResponseDecisionService {
	
	fallbackResponses := map[string]string{
		"generic": "Thank you for your message. I'm here to help! Could you please provide a bit more detail about what you're looking for?",
		"technical": "I understand you have a technical question. Let me connect you with one of our technical support specialists who can provide detailed assistance.",
		"billing": "I see you have a billing-related inquiry. For billing matters, I'll connect you with our billing support team who can help resolve this for you.",
		"complex": "This seems like a complex question that would benefit from human expertise. Let me connect you with one of our support specialists.",
		"unclear": "I want to make sure I understand your question correctly. Could you provide a bit more context about what specifically you're looking for help with?",
		"out_of_domain": "I'm here to help with questions about our products and services. Is there something specific about our platform I can assist you with?",
	}
	
	return &AutoResponseDecisionService{
		config:             config,
		greetingDetection:  greetingDetection,
		brandGreeting:      brandGreeting,
		questionClassifier: questionClassifier,
		knowledgeResponse:  knowledgeResponse,
		
		// Confidence thresholds
		highConfidenceThreshold:   0.8,
		mediumConfidenceThreshold: 0.6,
		lowConfidenceThreshold:    0.4,
		
		fallbackResponses: fallbackResponses,
	}
}

// MakeResponseDecision analyzes a message and decides how to respond
func (s *AutoResponseDecisionService) MakeResponseDecision(
	ctx context.Context,
	tenantID, projectID uuid.UUID,
	message, companyName string,
) (*AutoResponseDecision, error) {
	startTime := time.Now()
	
	decision := &AutoResponseDecision{
		Metadata: make(map[string]interface{}),
	}
	
	var reasoningSteps []string
	
	// Check if agentic behavior is enabled
	if !s.config.Enabled {
		reasoningSteps = append(reasoningSteps, "Agentic behavior is disabled")
		decision.ShouldRespond = false
		decision.ResponseType = "disabled"
		decision.Confidence = 0.0
		decision.ReasoningSteps = reasoningSteps
		decision.ProcessingTime = time.Since(startTime)
		return decision, nil
	}
	
	reasoningSteps = append(reasoningSteps, "Agentic behavior is enabled")
	
	// Step 1: Check for greetings first
	if s.config.GreetingDetection {
		greetingResult := s.greetingDetection.DetectGreeting(ctx, message)
		decision.GreetingDetection = greetingResult
		
		if greetingResult.IsGreeting && greetingResult.Confidence >= s.config.GreetingConfidence {
			reasoningSteps = append(reasoningSteps, fmt.Sprintf("Detected greeting with confidence %.2f", greetingResult.Confidence))
			
			// Generate brand-specific greeting response
			var responseText string
			if s.brandGreeting != nil {
				brandResponse, err := s.brandGreeting.GenerateGreetingResponse(ctx, tenantID, projectID, message)
				if err != nil || brandResponse == nil {
					responseText = fmt.Sprintf("Hello! Welcome to %s. How can I help you today?", companyName)
				} else {
					responseText = brandResponse.Message
				}
			} else {
				responseText = fmt.Sprintf("Hello! Welcome to %s. How can I help you today?", companyName)
			}
			
			decision.ShouldRespond = true
			decision.ResponseType = "greeting"
			decision.Response = responseText
			decision.BrandResponse = responseText
			decision.Confidence = greetingResult.Confidence
			decision.ReasoningSteps = reasoningSteps
			decision.ProcessingTime = time.Since(startTime)
			decision.Metadata["message_type"] = greetingResult.MessageType
			
			reasoningSteps = append(reasoningSteps, "Generated branded greeting response")
			return decision, nil
		} else if greetingResult.IsGreeting {
			reasoningSteps = append(reasoningSteps, fmt.Sprintf("Detected low-confidence greeting (%.2f < %.2f)", greetingResult.Confidence, s.config.GreetingConfidence))
		}
	}
	
	// Step 2: Classify the message as a question
	if s.config.KnowledgeResponses {
		classification := s.questionClassifier.ClassifyQuestion(ctx, message)
		decision.Classification = classification
		
		reasoningSteps = append(reasoningSteps, fmt.Sprintf("Question classification: IsQuestion=%v, Type=%s, Domain=%s", 
			classification.IsQuestion, classification.QuestionType, classification.Domain))
		
		// If it's a question that requires knowledge
		if classification.IsQuestion && classification.RequiresKnowledge {
			reasoningSteps = append(reasoningSteps, "Question requires knowledge base lookup")
			
			// Generate knowledge-based response
			knowledgeResult, err := s.knowledgeResponse.GenerateKnowledgeResponse(ctx, tenantID, projectID, message)
			if err != nil {
				reasoningSteps = append(reasoningSteps, fmt.Sprintf("Knowledge lookup failed: %v", err))
				
				// Fallback to escalation
				decision.ShouldRespond = true
				decision.ResponseType = "escalation"
				decision.Response = s.fallbackResponses["technical"]
				decision.RequiresEscalation = true
				decision.EscalationReason = "Knowledge lookup error"
				decision.Confidence = 0.3
				decision.ReasoningSteps = reasoningSteps
				decision.ProcessingTime = time.Since(startTime)
				return decision, nil
			}
			
			decision.KnowledgeResponse = knowledgeResult
			reasoningSteps = append(reasoningSteps, fmt.Sprintf("Knowledge lookup completed: HasResponse=%v, Confidence=%.2f", 
				knowledgeResult.HasResponse, knowledgeResult.Confidence))
			
			// Analyze knowledge response and make decision
			return s.analyzeKnowledgeResponse(decision, knowledgeResult, classification, reasoningSteps, startTime)
		}
		
		// If it's a question but doesn't require knowledge (e.g., simple greeting question)
		if classification.IsQuestion {
			reasoningSteps = append(reasoningSteps, "Question detected but doesn't require knowledge base")
			
			// Check if we can provide a simple response
			if classification.Domain == DomainGeneral && classification.Complexity == "simple" {
				decision.ShouldRespond = true
				decision.ResponseType = "generic"
				decision.Response = s.fallbackResponses["generic"]
				decision.Confidence = 0.5
				decision.ReasoningSteps = reasoningSteps
				decision.ProcessingTime = time.Since(startTime)
				return decision, nil
			}
		}
	}
	
	// Step 3: Handle non-questions or unclear messages
	reasoningSteps = append(reasoningSteps, "Message is not a clear question or greeting")
	
	// Check if it looks like a support request
	if s.containsSupportIndicators(message) {
		reasoningSteps = append(reasoningSteps, "Detected support request indicators")
		
		decision.ShouldRespond = true
		decision.ResponseType = "escalation"
		decision.Response = s.fallbackResponses["generic"]
		decision.RequiresEscalation = true
		decision.EscalationReason = "Support request detected"
		decision.Confidence = 0.4
		decision.ReasoningSteps = reasoningSteps
		decision.ProcessingTime = time.Since(startTime)
		return decision, nil
	}
	
	// Default: No automatic response
	reasoningSteps = append(reasoningSteps, "No automatic response criteria met")
	decision.ShouldRespond = false
	decision.ResponseType = "none"
	decision.Confidence = 0.0
	decision.ReasoningSteps = reasoningSteps
	decision.ProcessingTime = time.Since(startTime)
	
	return decision, nil
}

// analyzeKnowledgeResponse analyzes the knowledge response and makes a decision
func (s *AutoResponseDecisionService) analyzeKnowledgeResponse(
	decision *AutoResponseDecision,
	knowledgeResult *KnowledgeResponseResult,
	classification *QuestionClassificationResult,
	reasoningSteps []string,
	startTime time.Time,
) (*AutoResponseDecision, error) {
	
	// Handle out-of-domain responses
	if knowledgeResult.IsOutOfDomain {
		reasoningSteps = append(reasoningSteps, "Question is out of domain")
		decision.ShouldRespond = true
		decision.ResponseType = "out_of_domain"
		decision.Response = knowledgeResult.Response
		decision.Confidence = knowledgeResult.Confidence
		decision.ReasoningSteps = reasoningSteps
		decision.ProcessingTime = time.Since(startTime)
		return decision, nil
	}
	
	// Handle cases where more information is needed
	if knowledgeResult.NeedsMoreInfo {
		reasoningSteps = append(reasoningSteps, "Knowledge response indicates more information needed")
		decision.ShouldRespond = true
		decision.ResponseType = "clarification"
		decision.Response = knowledgeResult.Response
		decision.Confidence = knowledgeResult.Confidence
		decision.ReasoningSteps = reasoningSteps
		decision.ProcessingTime = time.Since(startTime)
		return decision, nil
	}
	
	// Handle escalation cases
	if knowledgeResult.ShouldEscalate {
		reasoningSteps = append(reasoningSteps, "Knowledge response recommends escalation")
		decision.ShouldRespond = true
		decision.ResponseType = "escalation"
		decision.RequiresEscalation = true
		decision.EscalationReason = s.determineEscalationReason(classification, knowledgeResult)
		decision.Response = s.selectEscalationResponse(classification)
		decision.Confidence = 0.7
		decision.ReasoningSteps = reasoningSteps
		decision.ProcessingTime = time.Since(startTime)
		return decision, nil
	}
	
	// Handle successful knowledge responses
	if knowledgeResult.HasResponse && knowledgeResult.Confidence >= s.config.KnowledgeConfidence {
		reasoningSteps = append(reasoningSteps, fmt.Sprintf("High-confidence knowledge response (%.2f >= %.2f)", 
			knowledgeResult.Confidence, s.config.KnowledgeConfidence))
		
		decision.ShouldRespond = true
		decision.ResponseType = "knowledge"
		decision.Response = knowledgeResult.Response
		decision.Confidence = knowledgeResult.Confidence
		decision.Citations = knowledgeResult.Citations
		decision.ReasoningSteps = reasoningSteps
		decision.ProcessingTime = time.Since(startTime)
		decision.Metadata["response_quality"] = knowledgeResult.ResponseQuality
		decision.Metadata["chunks_found"] = knowledgeResult.ChunksFound
		
		return decision, nil
	}
	
	// Handle medium-confidence responses
	if knowledgeResult.HasResponse && knowledgeResult.Confidence >= s.mediumConfidenceThreshold {
		reasoningSteps = append(reasoningSteps, fmt.Sprintf("Medium-confidence knowledge response (%.2f >= %.2f)", 
			knowledgeResult.Confidence, s.mediumConfidenceThreshold))
		
		// Provide response but suggest escalation for complex cases
		decision.ShouldRespond = true
		decision.ResponseType = "knowledge"
		decision.Response = knowledgeResult.Response + "\n\nIf this doesn't fully answer your question, I can connect you with a specialist for more detailed assistance."
		decision.Confidence = knowledgeResult.Confidence
		decision.Citations = knowledgeResult.Citations
		decision.ReasoningSteps = reasoningSteps
		decision.ProcessingTime = time.Since(startTime)
		
		return decision, nil
	}
	
	// Low confidence or no response - escalate
	reasoningSteps = append(reasoningSteps, fmt.Sprintf("Low-confidence or no knowledge response (confidence: %.2f)", knowledgeResult.Confidence))
	
	decision.ShouldRespond = true
	decision.ResponseType = "escalation"
	decision.RequiresEscalation = true
	decision.EscalationReason = "Low confidence knowledge response"
	decision.Response = s.selectEscalationResponse(classification)
	decision.Confidence = 0.4
	decision.ReasoningSteps = reasoningSteps
	decision.ProcessingTime = time.Since(startTime)
	
	return decision, nil
}

// containsSupportIndicators checks if the message contains support request indicators
func (s *AutoResponseDecisionService) containsSupportIndicators(message string) bool {
	lowerMessage := strings.ToLower(message)
	
	supportIndicators := []string{
		"help", "support", "issue", "problem", "error", "bug",
		"not working", "broken", "fix", "solve", "assistance",
		"need", "want", "please", "can you", "could you",
		"trouble", "difficulty", "stuck", "confused",
	}
	
	for _, indicator := range supportIndicators {
		if strings.Contains(lowerMessage, indicator) {
			return true
		}
	}
	
	return false
}

// determineEscalationReason determines the reason for escalation
func (s *AutoResponseDecisionService) determineEscalationReason(
	classification *QuestionClassificationResult,
	knowledgeResult *KnowledgeResponseResult,
) string {
	
	if classification.Domain == DomainBilling && classification.Intent == IntentComplaint {
		return "Billing complaint requires human attention"
	}
	
	if classification.QuestionType == QuestionTypeTroubleshooting && classification.Complexity == "complex" {
		return "Complex troubleshooting requires specialist expertise"
	}
	
	if classification.Domain == DomainTechnical && classification.Complexity == "complex" {
		return "Complex technical question requires specialist assistance"
	}
	
	if knowledgeResult.Confidence < s.lowConfidenceThreshold {
		return "Low confidence in available knowledge"
	}
	
	return "Human expertise recommended for better assistance"
}

// selectEscalationResponse selects appropriate escalation response based on classification
func (s *AutoResponseDecisionService) selectEscalationResponse(classification *QuestionClassificationResult) string {
	switch classification.Domain {
	case DomainTechnical:
		return s.fallbackResponses["technical"]
	case DomainBilling:
		return s.fallbackResponses["billing"]
	default:
		if classification.Complexity == "complex" {
			return s.fallbackResponses["complex"]
		}
		return s.fallbackResponses["generic"]
	}
}

// GetDecisionSummary provides a human-readable summary of the decision process
func (s *AutoResponseDecisionService) GetDecisionSummary(decision *AutoResponseDecision) string {
	summary := fmt.Sprintf("Decision: %s (Confidence: %.2f)\n", decision.ResponseType, decision.Confidence)
	summary += fmt.Sprintf("Should Respond: %v\n", decision.ShouldRespond)
	summary += fmt.Sprintf("Processing Time: %v\n", decision.ProcessingTime)
	
	if decision.RequiresEscalation {
		summary += fmt.Sprintf("Escalation Required: %s\n", decision.EscalationReason)
	}
	
	if len(decision.ReasoningSteps) > 0 {
		summary += "\nReasoning Steps:\n"
		for i, step := range decision.ReasoningSteps {
			summary += fmt.Sprintf("%d. %s\n", i+1, step)
		}
	}
	
	return summary
}
