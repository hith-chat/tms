package service

import (
	"context"
	"strings"
	"unicode"

	"github.com/bareuptime/tms/internal/config"
)

// GreetingDetectionService handles detection and classification of greeting messages
type GreetingDetectionService struct {
	greetingKeywords []string
	negativeKeywords []string // Words that indicate this is NOT a greeting
	minConfidence    float64
	config           *config.AgenticConfig
}

// NewGreetingDetectionService creates a new greeting detection service
func NewGreetingDetectionService(agenticConfig *config.AgenticConfig) *GreetingDetectionService {
	// Use config values if provided, otherwise use defaults
	greetingKeywords := agenticConfig.GreetingKeywords
	if len(greetingKeywords) == 0 {
		greetingKeywords = []string{
			// English greetings
			"hello", "hi", "hey", "greetings", "good morning", "good afternoon", 
			"good evening", "howdy", "hiya", "welcome", "salutations",
			
			// Common variations
			"helo", "hallo", "helllo", "heyyy", "hii", "hiiii",
			
			// International greetings
			"hola", "bonjour", "guten tag", "ciao", "konnichiwa", "namaste",
			
			// Casual starts
			"yo", "sup", "what's up", "whats up", "wassup",
		}
	}

	negativeKeywords := agenticConfig.NegativeKeywords
	if len(negativeKeywords) == 0 {
		negativeKeywords = []string{
			// Words that strongly indicate this is NOT a greeting
			"help", "support", "problem", "issue", "error", "bug", "question",
			"pricing", "cost", "payment", "refund", "return", "policy",
			"technical", "api", "integration", "setup", "configuration",
			"account", "login", "password", "billing", "invoice",
		}
	}

	minConfidence := agenticConfig.GreetingConfidence
	if minConfidence == 0 {
		minConfidence = 0.4 // 40% default confidence threshold
	}

	return &GreetingDetectionService{
		greetingKeywords: greetingKeywords,
		negativeKeywords: negativeKeywords,
		minConfidence:    minConfidence,
		config:           agenticConfig,
	}
}

// GreetingResult represents the result of greeting detection
type GreetingResult struct {
	IsGreeting   bool    `json:"is_greeting"`
	Confidence   float64 `json:"confidence"`
	MatchedTerms []string `json:"matched_terms"`
	MessageType  string  `json:"message_type"` // "simple_greeting", "question_greeting", "complex"
}

// DetectGreeting analyzes a message to determine if it's a greeting
func (g *GreetingDetectionService) DetectGreeting(ctx context.Context, message string) *GreetingResult {
	if strings.TrimSpace(message) == "" {
		return &GreetingResult{
			IsGreeting:  false,
			Confidence:  0.0,
			MessageType: "empty",
		}
	}

	// Normalize the message
	normalizedMessage := g.normalizeMessage(message)
	words := strings.Fields(normalizedMessage)
	
	// Check for greeting keywords first
	matchedTerms := g.findGreetingMatches(normalizedMessage, words)
	
	// If no greeting matches found, return early
	if len(matchedTerms) == 0 {
		return &GreetingResult{
			IsGreeting:  false,
			Confidence:  0.0,
			MessageType: "complex",
		}
	}
	
	// Calculate initial confidence based on greeting matches
	confidence := g.calculateConfidence(normalizedMessage, words, matchedTerms)
	
	// Reduce confidence if negative keywords are present, but don't completely reject
	for _, negativeKeyword := range g.negativeKeywords {
		if strings.Contains(normalizedMessage, negativeKeyword) {
			confidence *= 0.5 // Reduce confidence by half
			break
		}
	}
	
	// If message is very short (1-3 words) and has greeting, boost confidence despite negative keywords
	if len(words) <= 3 && len(matchedTerms) > 0 {
		confidence += 0.2
	}
	
	// Cap confidence at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	// Determine message type
	messageType := g.classifyMessageType(normalizedMessage, words, matchedTerms, confidence)
	
	return &GreetingResult{
		IsGreeting:   confidence >= g.minConfidence,
		Confidence:   confidence,
		MatchedTerms: matchedTerms,
		MessageType:  messageType,
	}
}

// normalizeMessage converts message to lowercase and removes punctuation
func (g *GreetingDetectionService) normalizeMessage(message string) string {
	// Convert to lowercase
	message = strings.ToLower(message)
	
	// Remove extra whitespace
	message = strings.TrimSpace(message)
	
	// Remove punctuation but keep spaces
	var result strings.Builder
	for _, r := range message {
		if unicode.IsLetter(r) || unicode.IsSpace(r) {
			result.WriteRune(r)
		} else if unicode.IsPunct(r) {
			result.WriteRune(' ')
		}
	}
	
	// Clean up multiple spaces using a better approach
	cleaned := result.String()
	// Use a simple approach to handle multiple spaces
	words := strings.Fields(cleaned) // This splits on any whitespace and removes empty elements
	return strings.Join(words, " ")  // Join back with single spaces
}

// findGreetingMatches identifies greeting keywords in the message
func (g *GreetingDetectionService) findGreetingMatches(normalizedMessage string, words []string) []string {
	var matches []string
	matchMap := make(map[string]bool)
	
	// Check for exact keyword matches only
	for _, keyword := range g.greetingKeywords {
		if strings.Contains(normalizedMessage, keyword) {
			if !matchMap[keyword] {
				matches = append(matches, keyword)
				matchMap[keyword] = true
			}
		}
	}
	
	return matches
}

// calculateConfidence determines how confident we are that this is a greeting
func (g *GreetingDetectionService) calculateConfidence(normalizedMessage string, words []string, matchedTerms []string) float64 {
	if len(matchedTerms) == 0 {
		return 0.0
	}
	
	confidence := 0.0
	
	// Base score for having greeting keywords
	confidence += float64(len(matchedTerms)) * 0.3
	
	// Bonus for short messages (greetings are typically brief)
	wordCount := len(words)
	if wordCount <= 3 {
		confidence += 0.4
	} else if wordCount <= 5 {
		confidence += 0.2
	} else if wordCount > 10 {
		confidence -= 0.1 // Long messages are less likely to be simple greetings
	}
	
	// Bonus for messages that start with greetings
	if len(words) > 0 {
		firstWord := words[0]
		for _, keyword := range g.greetingKeywords {
			if strings.HasPrefix(keyword, firstWord) || strings.HasPrefix(firstWord, keyword) {
				confidence += 0.2
				break
			}
		}
	}
	
	// Penalty for question words (makes it more complex)
	questionWords := []string{"what", "how", "when", "where", "why", "who", "can", "could", "would", "should"}
	for _, word := range words {
		for _, qWord := range questionWords {
			if word == qWord {
				confidence -= 0.1
				break
			}
		}
	}
	
	// Bonus for exclamation or enthusiasm
	if strings.Contains(normalizedMessage, "!") || strings.Contains(normalizedMessage, ":)") {
		confidence += 0.1
	}
	
	// Cap confidence at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	return confidence
}

// classifyMessageType determines the type of message based on analysis
func (g *GreetingDetectionService) classifyMessageType(normalizedMessage string, words []string, matchedTerms []string, confidence float64) string {
	if confidence < g.minConfidence {
		return "complex"
	}
	
	// Simple greeting: short message with greeting keywords, no questions
	if len(words) <= 3 && len(matchedTerms) > 0 {
		hasQuestion := g.containsQuestionWords(words)
		if !hasQuestion {
			return "simple_greeting"
		}
	}
	
	// Question greeting: greeting + question
	if len(matchedTerms) > 0 && g.containsQuestionWords(words) {
		return "question_greeting"
	}
	
	// If we have greeting terms but it's longer, still classify as greeting
	if len(matchedTerms) > 0 && confidence >= g.minConfidence {
		return "simple_greeting"
	}
	
	return "complex"
}

// containsQuestionWords checks if the message contains question indicators
func (g *GreetingDetectionService) containsQuestionWords(words []string) bool {
	questionWords := []string{"what", "how", "when", "where", "why", "who", "can", "could", "would", "should", "is", "are", "do", "does", "did"}
	questionMarkers := []string{"?"}
	
	for _, word := range words {
		for _, qWord := range questionWords {
			if word == qWord {
				return true
			}
		}
		for _, marker := range questionMarkers {
			if strings.Contains(word, marker) {
				return true
			}
		}
	}
	return false
}

// isSimilar checks if two words are similar (strict similarity check for greetings)
func (g *GreetingDetectionService) isSimilar(word1, word2 string) bool {
	if word1 == word2 {
		return true
	}
	
	// Only check for variations if words are close in length
	lengthDiff := len(word1) - len(word2)
	if lengthDiff < -2 || lengthDiff > 2 {
		return false
	}
	
	// Check if one is contained in the other only for very similar lengths
	// and only if the longer word is just repetition of letters (like "hi" vs "hiii")
	if len(word1) >= 2 && len(word2) >= 2 && abs(len(word1)-len(word2)) <= 2 {
		shorter, longer := word1, word2
		if len(word1) > len(word2) {
			shorter, longer = word2, word1
		}
		
		// Check if longer word starts with shorter and the rest are just repeated chars
		if strings.HasPrefix(longer, shorter) {
			remaining := longer[len(shorter):]
			if len(remaining) > 0 && len(shorter) > 0 {
				lastChar := shorter[len(shorter)-1]
				allSame := true
				for _, r := range remaining {
					if byte(r) != lastChar {
						allSame = false
						break
					}
				}
				if allSame {
					return true
				}
			}
		}
	}
	
	// Simple edit distance check for typos (very strict)
	if len(word1) == len(word2) && len(word1) >= 3 {
		differences := 0
		for i := 0; i < len(word1); i++ {
			if word1[i] != word2[i] {
				differences++
			}
		}
		// Only allow 1 character difference for exact same length words
		return differences == 1
	}
	
	return false
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// IsSimpleGreeting is a convenience method to check if a message is a simple greeting
func (g *GreetingDetectionService) IsSimpleGreeting(ctx context.Context, message string) bool {
	result := g.DetectGreeting(ctx, message)
	return result.IsGreeting && result.MessageType == "simple_greeting"
}

// GetGreetingKeywords returns the list of configured greeting keywords
func (g *GreetingDetectionService) GetGreetingKeywords() []string {
	return g.greetingKeywords
}

// UpdateGreetingKeywords allows updating the greeting keywords list
func (g *GreetingDetectionService) UpdateGreetingKeywords(keywords []string) {
	g.greetingKeywords = keywords
}

// SetConfidenceThreshold allows updating the confidence threshold
func (g *GreetingDetectionService) SetConfidenceThreshold(threshold float64) {
	if threshold >= 0.0 && threshold <= 1.0 {
		g.minConfidence = threshold
	}
}
