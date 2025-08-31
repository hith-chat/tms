package service

import (
	"context"
	"strings"
	"unicode"

	"github.com/bareuptime/tms/internal/config"
)

// QuestionType represents the type of question being asked
type QuestionType string

const (
	QuestionTypeHowTo          QuestionType = "how_to"
	QuestionTypeWhatIs         QuestionType = "what_is"
	QuestionTypeTroubleshooting QuestionType = "troubleshooting"
	QuestionTypePricing        QuestionType = "pricing"
	QuestionTypeGeneral        QuestionType = "general"
	QuestionTypeGreeting       QuestionType = "greeting"
	QuestionTypeRequest        QuestionType = "request"
)

// QuestionIntent represents the intent behind the question
type QuestionIntent string

const (
	IntentSeekingInfo     QuestionIntent = "seeking_info"
	IntentRequestingAction QuestionIntent = "requesting_action"
	IntentGreeting        QuestionIntent = "greeting"
	IntentComplaint       QuestionIntent = "complaint"
)

// QuestionDomain represents the domain/category of the question
type QuestionDomain string

const (
	DomainTechnical QuestionDomain = "technical"
	DomainPricing   QuestionDomain = "pricing"
	DomainSupport   QuestionDomain = "support"
	DomainGeneral   QuestionDomain = "general"
	DomainProduct   QuestionDomain = "product"
	DomainBilling   QuestionDomain = "billing"
	DomainAccount   QuestionDomain = "account"
)

// QuestionClassificationResult contains the result of question classification
type QuestionClassificationResult struct {
	IsQuestion       bool            `json:"is_question"`
	QuestionType     QuestionType    `json:"question_type"`
	Intent           QuestionIntent  `json:"intent"`
	Domain           QuestionDomain  `json:"domain"`
	Complexity       string          `json:"complexity"` // simple, moderate, complex
	Confidence       float64         `json:"confidence"`
	Keywords         []string        `json:"keywords"`
	RequiresKnowledge bool           `json:"requires_knowledge"`
	CanAutoRespond   bool            `json:"can_auto_respond"`
}

// QuestionClassificationService handles intelligent question classification
type QuestionClassificationService struct {
	config *config.AgenticConfig
	
	// Question indicators
	questionWords    []string
	howToKeywords    []string
	whatIsKeywords   []string
	troubleshootingKeywords []string
	pricingKeywords  []string
	requestKeywords  []string
	
	// Domain keywords
	technicalKeywords []string
	pricingDomainKeywords []string
	supportKeywords  []string
	productKeywords  []string
	billingKeywords  []string
	accountKeywords  []string
}

// NewQuestionClassificationService creates a new question classification service
func NewQuestionClassificationService(config *config.AgenticConfig) *QuestionClassificationService {
	return &QuestionClassificationService{
		config: config,
		
		// Question indicators
		questionWords: []string{
			"what", "how", "why", "when", "where", "who", "which", "can", "could", 
			"would", "should", "is", "are", "do", "does", "did", "will", "won't",
			"?", "help", "explain", "tell", "show",
		},
		
		// Question type keywords
		howToKeywords: []string{
			"how to", "how do", "how can", "how should", "steps", "guide", 
			"tutorial", "instructions", "process", "procedure", "way to",
		},
		
		whatIsKeywords: []string{
			"what is", "what are", "what does", "define", "definition", 
			"meaning", "explain", "describe", "tell me about",
		},
		
		troubleshootingKeywords: []string{
			"problem", "issue", "error", "bug", "broken", "not working", 
			"fix", "solve", "troubleshoot", "help", "wrong", "failed",
			"doesn't work", "can't", "unable", "trouble",
		},
		
		pricingKeywords: []string{
			"price", "cost", "pricing", "fee", "charge", "payment", "plan", 
			"subscription", "how much", "expensive", "cheap", "discount",
		},
		
		requestKeywords: []string{
			"please", "can you", "could you", "would you", "i need", "i want", 
			"i would like", "help me", "assist", "support", "do this",
		},
		
		// Domain keywords
		technicalKeywords: []string{
			"api", "code", "programming", "technical", "development", "integration", 
			"database", "server", "error", "bug", "configuration", "setup",
			"install", "deployment", "authentication", "security",
		},
		
		pricingDomainKeywords: []string{
			"price", "cost", "pricing", "plan", "subscription", "billing", 
			"payment", "fee", "charge", "discount", "upgrade", "downgrade",
		},
		
		supportKeywords: []string{
			"help", "support", "assistance", "problem", "issue", "question", 
			"contact", "service", "customer", "agent", "representative",
		},
		
		productKeywords: []string{
			"product", "feature", "functionality", "capability", "service", 
			"tool", "platform", "application", "software", "system",
		},
		
		billingKeywords: []string{
			"bill", "billing", "invoice", "payment", "charge", "subscription", 
			"refund", "credit", "debit", "account", "transaction",
		},
		
		accountKeywords: []string{
			"account", "profile", "login", "password", "username", "email", 
			"settings", "preferences", "access", "permissions", "user",
		},
	}
}

// ClassifyQuestion analyzes a message to determine if it's a question and classify it
func (s *QuestionClassificationService) ClassifyQuestion(ctx context.Context, message string) *QuestionClassificationResult {
	// Check if agentic behavior is enabled
	if !s.config.Enabled || !s.config.KnowledgeResponses {
		return &QuestionClassificationResult{
			IsQuestion:       false,
			QuestionType:     QuestionTypeGeneral,
			Intent:           IntentSeekingInfo,
			Domain:           DomainGeneral,
			Complexity:       "simple",
			Confidence:       0.0,
			RequiresKnowledge: false,
			CanAutoRespond:   false,
		}
	}
	
	normalizedMessage := s.normalizeMessage(message)
	
	// Determine if this is a question
	isQuestion := s.isQuestion(normalizedMessage)
	
	// If not a question, return early
	if !isQuestion {
		return &QuestionClassificationResult{
			IsQuestion:       false,
			QuestionType:     QuestionTypeGeneral,
			Intent:           s.detectIntent(normalizedMessage),
			Domain:           s.detectDomain(normalizedMessage),
			Complexity:       "simple",
			Confidence:       0.1,
			RequiresKnowledge: false,
			CanAutoRespond:   false,
		}
	}
	
	// Classify the question
	questionType := s.classifyQuestionType(normalizedMessage)
	intent := s.detectIntent(normalizedMessage)
	domain := s.detectDomain(normalizedMessage)
	complexity := s.assessComplexity(normalizedMessage)
	keywords := s.extractKeywords(normalizedMessage)
	
	// Calculate confidence based on various factors
	confidence := s.calculateConfidence(normalizedMessage, questionType, intent, domain)
	
	// Determine if this requires knowledge base lookup
	requiresKnowledge := s.requiresKnowledgeBase(questionType, domain, complexity)
	
	// Determine if we can auto-respond
	canAutoRespond := confidence >= s.config.KnowledgeConfidence && requiresKnowledge
	
	return &QuestionClassificationResult{
		IsQuestion:       true,
		QuestionType:     questionType,
		Intent:           intent,
		Domain:           domain,
		Complexity:       complexity,
		Confidence:       confidence,
		Keywords:         keywords,
		RequiresKnowledge: requiresKnowledge,
		CanAutoRespond:   canAutoRespond,
	}
}

// normalizeMessage cleans and normalizes the input message
func (s *QuestionClassificationService) normalizeMessage(message string) string {
	// Convert to lowercase
	normalized := strings.ToLower(message)
	
	// Remove extra whitespace
	normalized = strings.TrimSpace(normalized)
	
	// Remove punctuation except question marks
	var cleaned strings.Builder
	for _, r := range normalized {
		if unicode.IsLetter(r) || unicode.IsSpace(r) || r == '?' || r == '\'' {
			cleaned.WriteRune(r)
		}
	}
	
	return cleaned.String()
}

// isQuestion determines if the message is a question
func (s *QuestionClassificationService) isQuestion(message string) bool {
	// Check for question mark
	if strings.Contains(message, "?") {
		return true
	}
	
	// Check for question words at the beginning
	words := strings.Fields(message)
	if len(words) == 0 {
		return false
	}
	
	firstWord := words[0]
	for _, qWord := range s.questionWords {
		if firstWord == qWord {
			return true
		}
	}
	
	// Check for question patterns
	questionPatterns := []string{
		"can you", "could you", "would you", "will you",
		"how do", "how to", "what is", "what are",
		"where is", "when is", "why is", "who is",
		"tell me", "show me", "explain", "help me",
	}
	
	for _, pattern := range questionPatterns {
		if strings.Contains(message, pattern) {
			return true
		}
	}
	
	return false
}

// classifyQuestionType determines the type of question
func (s *QuestionClassificationService) classifyQuestionType(message string) QuestionType {
	// Check for how-to questions
	for _, keyword := range s.howToKeywords {
		if strings.Contains(message, keyword) {
			return QuestionTypeHowTo
		}
	}
	
	// Check for what-is questions
	for _, keyword := range s.whatIsKeywords {
		if strings.Contains(message, keyword) {
			return QuestionTypeWhatIs
		}
	}
	
	// Check for troubleshooting questions
	for _, keyword := range s.troubleshootingKeywords {
		if strings.Contains(message, keyword) {
			return QuestionTypeTroubleshooting
		}
	}
	
	// Check for pricing questions
	for _, keyword := range s.pricingKeywords {
		if strings.Contains(message, keyword) {
			return QuestionTypePricing
		}
	}
	
	// Check for request-type questions
	for _, keyword := range s.requestKeywords {
		if strings.Contains(message, keyword) {
			return QuestionTypeRequest
		}
	}
	
	return QuestionTypeGeneral
}

// detectIntent determines the intent behind the question
func (s *QuestionClassificationService) detectIntent(message string) QuestionIntent {
	// Action request patterns
	actionPatterns := []string{
		"please", "can you", "could you", "would you", "help me",
		"i need", "i want", "i would like", "assist", "do this",
		"create", "make", "setup", "configure", "fix", "solve",
	}
	
	for _, pattern := range actionPatterns {
		if strings.Contains(message, pattern) {
			return IntentRequestingAction
		}
	}
	
	// Complaint patterns
	complaintPatterns := []string{
		"problem", "issue", "broken", "not working", "doesn't work",
		"frustrated", "angry", "disappointed", "terrible", "awful",
	}
	
	for _, pattern := range complaintPatterns {
		if strings.Contains(message, pattern) {
			return IntentComplaint
		}
	}
	
	return IntentSeekingInfo
}

// detectDomain determines the domain/category of the question
func (s *QuestionClassificationService) detectDomain(message string) QuestionDomain {
	domainScores := make(map[QuestionDomain]int)
	
	// Score technical domain
	for _, keyword := range s.technicalKeywords {
		if strings.Contains(message, keyword) {
			domainScores[DomainTechnical]++
		}
	}
	
	// Score pricing domain
	for _, keyword := range s.pricingDomainKeywords {
		if strings.Contains(message, keyword) {
			domainScores[DomainPricing]++
		}
	}
	
	// Score support domain
	for _, keyword := range s.supportKeywords {
		if strings.Contains(message, keyword) {
			domainScores[DomainSupport]++
		}
	}
	
	// Score product domain
	for _, keyword := range s.productKeywords {
		if strings.Contains(message, keyword) {
			domainScores[DomainProduct]++
		}
	}
	
	// Score billing domain
	for _, keyword := range s.billingKeywords {
		if strings.Contains(message, keyword) {
			domainScores[DomainBilling]++
		}
	}
	
	// Score account domain
	for _, keyword := range s.accountKeywords {
		if strings.Contains(message, keyword) {
			domainScores[DomainAccount]++
		}
	}
	
	// Find highest scoring domain
	maxScore := 0
	bestDomain := DomainGeneral
	
	for domain, score := range domainScores {
		if score > maxScore {
			maxScore = score
			bestDomain = domain
		}
	}
	
	return bestDomain
}

// assessComplexity determines the complexity of the question
func (s *QuestionClassificationService) assessComplexity(message string) string {
	words := strings.Fields(message)
	wordCount := len(words)
	
	// Simple indicators
	simpleIndicators := []string{
		"what is", "how much", "when", "where", "who", "yes", "no",
	}
	
	// Complex indicators
	complexIndicators := []string{
		"integration", "configuration", "troubleshoot", "multiple", 
		"complex", "advanced", "enterprise", "custom", "api", "development",
	}
	
	// Check for complex indicators
	for _, indicator := range complexIndicators {
		if strings.Contains(message, indicator) {
			return "complex"
		}
	}
	
	// Check for simple indicators
	for _, indicator := range simpleIndicators {
		if strings.Contains(message, indicator) {
			return "simple"
		}
	}
	
	// Base on word count
	if wordCount < 5 {
		return "simple"
	} else if wordCount > 15 {
		return "complex"
	}
	
	return "moderate"
}

// extractKeywords extracts relevant keywords from the message
func (s *QuestionClassificationService) extractKeywords(message string) []string {
	words := strings.Fields(message)
	var keywords []string
	
	// Filter out common stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, 
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "is": true,
		"are": true, "was": true, "were": true, "be": true, "been": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
		"can": true, "i": true, "you": true, "he": true, "she": true,
		"it": true, "we": true, "they": true, "this": true, "that": true,
		"these": true, "those": true, "my": true, "your": true, "his": true,
		"her": true, "its": true, "our": true, "their": true,
	}
	
	for _, word := range words {
		// Clean the word
		cleaned := strings.ToLower(strings.Trim(word, ".,!?;:"))
		
		// Skip if it's a stop word or too short
		if stopWords[cleaned] || len(cleaned) < 3 {
			continue
		}
		
		keywords = append(keywords, cleaned)
	}
	
	// Limit to 10 keywords
	if len(keywords) > 10 {
		keywords = keywords[:10]
	}
	
	return keywords
}

// calculateConfidence calculates the confidence score for the classification
func (s *QuestionClassificationService) calculateConfidence(message string, qType QuestionType, intent QuestionIntent, domain QuestionDomain) float64 {
	confidence := 0.0
	
	// Base confidence for being a question
	confidence += 0.3
	
	// Boost for clear question indicators
	if strings.Contains(message, "?") {
		confidence += 0.2
	}
	
	// Boost for specific question types
	if qType != QuestionTypeGeneral {
		confidence += 0.2
	}
	
	// Boost for specific domains
	if domain != DomainGeneral {
		confidence += 0.15
	}
	
	// Boost for clear intent
	if intent == IntentSeekingInfo || intent == IntentRequestingAction {
		confidence += 0.1
	}
	
	// Penalty for very short messages
	words := strings.Fields(message)
	if len(words) < 3 {
		confidence -= 0.2
	}
	
	// Ensure confidence is within bounds
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}
	
	return confidence
}

// requiresKnowledgeBase determines if the question requires knowledge base lookup
func (s *QuestionClassificationService) requiresKnowledgeBase(qType QuestionType, domain QuestionDomain, complexity string) bool {
	// Greeting-type questions don't need knowledge base
	if qType == QuestionTypeGreeting {
		return false
	}
	
	// Technical, product, and how-to questions typically need knowledge base
	if qType == QuestionTypeHowTo || qType == QuestionTypeTroubleshooting ||
		domain == DomainTechnical || domain == DomainProduct {
		return true
	}
	
	// Complex questions typically need knowledge base
	if complexity == "complex" {
		return true
	}
	
	// Moderate complexity questions in specific domains
	if complexity == "moderate" && (domain == DomainSupport || domain == DomainBilling) {
		return true
	}
	
	return false
}
