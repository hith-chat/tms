package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/models"
)

// KnowledgeResponseResult contains the result of knowledge-based response generation
type KnowledgeResponseResult struct {
	HasResponse        bool                           `json:"has_response"`
	Response           string                         `json:"response"`
	Confidence         float64                        `json:"confidence"`
	SourceResults      []models.KnowledgeSearchResult `json:"source_results"`
	Citations          []string                       `json:"citations"`
	ResponseQuality    string                         `json:"response_quality"` // excellent, good, adequate, poor
	IsOutOfDomain      bool                           `json:"is_out_of_domain"`
	NeedsMoreInfo      bool                           `json:"needs_more_info"`
	ShouldEscalate     bool                           `json:"should_escalate"`
	ProcessingTime     time.Duration                  `json:"processing_time"`
	SearchQuery        string                         `json:"search_query"`
	ChunksFound        int                            `json:"chunks_found"`
	TopSimilarityScore float64                        `json:"top_similarity_score"`
}

// KnowledgeResponseService handles intelligent knowledge-based response generation
type KnowledgeResponseService struct {
	config            *config.AgenticConfig
	knowledgeService  *KnowledgeService
	questionClassifier *QuestionClassificationService
	
	// Out-of-domain indicators
	outOfDomainKeywords []string
	
	// Quality thresholds
	excellentThreshold  float64
	goodThreshold      float64
	adequateThreshold  float64
	
	// Response templates
	outOfDomainTemplate string
	moreInfoTemplate    string
	escalationTemplate  string
}

// NewKnowledgeResponseService creates a new knowledge response service
func NewKnowledgeResponseService(
	config *config.AgenticConfig,
	knowledgeService *KnowledgeService,
	questionClassifier *QuestionClassificationService,
) *KnowledgeResponseService {
	return &KnowledgeResponseService{
		config:             config,
		knowledgeService:   knowledgeService,
		questionClassifier: questionClassifier,
		
		// Out-of-domain indicators
		outOfDomainKeywords: []string{
			"weather", "sports", "politics", "news", "cooking", "recipes",
			"travel", "entertainment", "movies", "music", "celebrities",
			"personal", "relationship", "health", "medical", "legal",
			"homework", "assignment", "school", "university",
		},
		
		// Quality thresholds
		excellentThreshold: 0.9,
		goodThreshold:     0.8,
		adequateThreshold: 0.7,
		
		// Response templates
		outOfDomainTemplate: "I'm sorry, but that question seems to be outside of my area of expertise. I'm here to help with questions related to our products and services. Is there something specific about our platform I can help you with?",
		moreInfoTemplate:    "I found some relevant information, but I might need a bit more context to give you the most helpful answer. Could you provide more details about what specifically you're looking for?",
		escalationTemplate:  "This seems like a complex question that would benefit from human expertise. Let me connect you with one of our support specialists who can provide more detailed assistance.",
	}
}

// GenerateKnowledgeResponse generates an intelligent response based on knowledge base
func (s *KnowledgeResponseService) GenerateKnowledgeResponse(
	ctx context.Context,
	tenantID, projectID uuid.UUID,
	question string,
) (*KnowledgeResponseResult, error) {
	startTime := time.Now()
	
	// Check if service is enabled
	if !s.config.Enabled || !s.config.KnowledgeResponses {
		return &KnowledgeResponseResult{
			HasResponse:    false,
			Confidence:     0.0,
			ResponseQuality: "poor",
			ProcessingTime: time.Since(startTime),
		}, nil
	}
	
	// First, classify the question
	classification := s.questionClassifier.ClassifyQuestion(ctx, question)
	
	// If it's not a question or doesn't require knowledge, return early
	if !classification.IsQuestion || !classification.RequiresKnowledge {
		return &KnowledgeResponseResult{
			HasResponse:    false,
			Confidence:     0.0,
			ResponseQuality: "poor",
			ProcessingTime: time.Since(startTime),
		}, nil
	}
	
	// Check for out-of-domain questions
	if s.isOutOfDomain(question, classification) {
		return &KnowledgeResponseResult{
			HasResponse:     true,
			Response:        s.outOfDomainTemplate,
			Confidence:      0.8,
			ResponseQuality: "good",
			IsOutOfDomain:   true,
			ProcessingTime:  time.Since(startTime),
			SearchQuery:     question,
		}, nil
	}
	
	// Generate search query from the question
	searchQuery := s.generateSearchQuery(question, classification)
	
	// Search knowledge base (check if service is available)
	if s.knowledgeService == nil {
		// Return a mock response for testing when no knowledge service is available
		return &KnowledgeResponseResult{
			HasResponse:     true,
			Response:        "I would search the knowledge base for information related to: " + searchQuery,
			Confidence:      0.5,
			ResponseQuality: "adequate",
			ProcessingTime:  time.Since(startTime),
			SearchQuery:     searchQuery,
			ChunksFound:     0,
		}, nil
	}
	
	searchRequest := &models.KnowledgeSearchRequest{
		Query:            searchQuery,
		MaxResults:       10,
		SimilarityScore:  0.6, // Lower threshold for initial search
		IncludeDocuments: true,
		IncludePages:     true,
	}
	
	searchResponse, err := s.knowledgeService.SearchKnowledgeBase(ctx, tenantID, projectID, searchRequest)
	if err != nil {
		return nil, fmt.Errorf("knowledge search failed: %w", err)
	}
	
	// Analyze search results
	result := s.analyzeSearchResults(searchResponse, classification, question)
	result.ProcessingTime = time.Since(startTime)
	result.SearchQuery = searchQuery
	
	return result, nil
}

// generateSearchQuery creates an optimized search query from the question
func (s *KnowledgeResponseService) generateSearchQuery(question string, classification *QuestionClassificationResult) string {
	// Use the extracted keywords as base
	query := strings.Join(classification.Keywords, " ")
	
	// Add domain-specific terms
	switch classification.Domain {
	case DomainTechnical:
		if !strings.Contains(query, "api") && !strings.Contains(query, "technical") {
			query += " technical integration"
		}
	case DomainPricing:
		if !strings.Contains(query, "price") && !strings.Contains(query, "cost") {
			query += " pricing cost"
		}
	case DomainAccount:
		if !strings.Contains(query, "account") && !strings.Contains(query, "user") {
			query += " account user"
		}
	case DomainBilling:
		if !strings.Contains(query, "billing") && !strings.Contains(query, "payment") {
			query += " billing payment"
		}
	}
	
	// Add question type context
	switch classification.QuestionType {
	case QuestionTypeHowTo:
		query += " guide tutorial steps"
	case QuestionTypeTroubleshooting:
		query += " troubleshoot fix solution"
	case QuestionTypeWhatIs:
		query += " overview explanation"
	}
	
	return strings.TrimSpace(query)
}

// analyzeSearchResults analyzes knowledge search results and generates response
func (s *KnowledgeResponseService) analyzeSearchResults(
	searchResponse *models.KnowledgeSearchResponse,
	classification *QuestionClassificationResult,
	originalQuestion string,
) *KnowledgeResponseResult {
	
	result := &KnowledgeResponseResult{
		ChunksFound: len(searchResponse.Results),
	}
	
	// If no results found
	if len(searchResponse.Results) == 0 {
		result.HasResponse = true
		result.Response = s.moreInfoTemplate
		result.Confidence = 0.3
		result.ResponseQuality = "poor"
		result.NeedsMoreInfo = true
		return result
	}
	
	// Calculate similarity scores and confidence
	topScore := 0.0
	relevantChunks := []models.KnowledgeSearchResult{}
	
	for _, chunk := range searchResponse.Results {
		if chunk.Score > topScore {
			topScore = chunk.Score
		}
		
		// Only use chunks with decent similarity
		if chunk.Score >= 0.7 {
			relevantChunks = append(relevantChunks, chunk)
		}
	}
	
	result.TopSimilarityScore = topScore
	
	// If no relevant chunks found
	if len(relevantChunks) == 0 {
		if topScore < 0.5 {
			// Very low similarity - might be out of domain
			result.HasResponse = true
			result.Response = s.outOfDomainTemplate
			result.Confidence = 0.6
			result.ResponseQuality = "adequate"
			result.IsOutOfDomain = true
		} else {
			// Some similarity but not enough - need more info
			result.HasResponse = true
			result.Response = s.moreInfoTemplate
			result.Confidence = 0.4
			result.ResponseQuality = "poor"
			result.NeedsMoreInfo = true
		}
		return result
	}
	
	// Generate response from relevant chunks
	response, confidence := s.synthesizeResponse(relevantChunks, classification, originalQuestion)
	
	result.HasResponse = true
	result.Response = response
	result.Confidence = confidence
	result.ResponseQuality = s.assessResponseQuality(confidence, len(relevantChunks), topScore)
	
	// Add source results and citations
	result.SourceResults = relevantChunks
	result.Citations = s.generateCitations(relevantChunks)
	
	// Determine if escalation is needed
	result.ShouldEscalate = s.shouldEscalate(classification, confidence, topScore)
	
	return result
}

// synthesizeResponse creates a coherent response from multiple knowledge chunks
func (s *KnowledgeResponseService) synthesizeResponse(
	chunks []models.KnowledgeSearchResult,
	classification *QuestionClassificationResult,
	originalQuestion string,
) (string, float64) {
	
	if len(chunks) == 0 {
		return "", 0.0
	}
	
	// For now, use the highest scoring chunk as primary response
	// In a more advanced implementation, we would use AI to synthesize multiple chunks
	primaryChunk := chunks[0]
	response := primaryChunk.Content
	
	// Calculate confidence based on similarity and number of supporting chunks
	confidence := primaryChunk.Score
	
	// Boost confidence if multiple chunks support the answer
	if len(chunks) > 1 {
		confidence += 0.1 * float64(len(chunks)-1)
		if confidence > 1.0 {
			confidence = 1.0
		}
	}
	
	// Add context based on question type
	switch classification.QuestionType {
	case QuestionTypeHowTo:
		if !strings.Contains(strings.ToLower(response), "step") &&
			!strings.Contains(strings.ToLower(response), "follow") {
			response = "Here's how to " + strings.ToLower(originalQuestion[strings.Index(originalQuestion, " "):]) + ":\n\n" + response
		}
	case QuestionTypeWhatIs:
		if !strings.Contains(strings.ToLower(response), "is") &&
			!strings.Contains(strings.ToLower(response), "definition") {
			response = response
		}
	case QuestionTypeTroubleshooting:
		if !strings.Contains(strings.ToLower(response), "solution") &&
			!strings.Contains(strings.ToLower(response), "fix") {
			response = "To resolve this issue:\n\n" + response
		}
	}
	
	return response, confidence
}

// generateCitations creates citation strings for the sources
func (s *KnowledgeResponseService) generateCitations(chunks []models.KnowledgeSearchResult) []string {
	citations := []string{}
	citationMap := make(map[string]bool)
	
	for _, chunk := range chunks {
		var citation string
		
		if chunk.Type == "document" && chunk.Title != nil {
			citation = fmt.Sprintf("Document: %s", *chunk.Title)
		} else if chunk.Type == "webpage" && chunk.Title != nil {
			citation = fmt.Sprintf("Page: %s", *chunk.Title)
			if chunk.Source != "" {
				citation += fmt.Sprintf(" (%s)", chunk.Source)
			}
		} else {
			citation = fmt.Sprintf("Source: %s", chunk.Source)
		}
		
		// Avoid duplicate citations
		if !citationMap[citation] {
			citations = append(citations, citation)
			citationMap[citation] = true
		}
		
		// Limit citations
		if len(citations) >= 3 {
			break
		}
	}
	
	return citations
}

// isOutOfDomain checks if the question is outside the expected domain
func (s *KnowledgeResponseService) isOutOfDomain(question string, classification *QuestionClassificationResult) bool {
	questionLower := strings.ToLower(question)
	
	// Check for out-of-domain keywords
	for _, keyword := range s.outOfDomainKeywords {
		if strings.Contains(questionLower, keyword) {
			return true
		}
	}
	
	// Check if it's a very general question with no specific domain
	if classification.Domain == DomainGeneral && 
		classification.Confidence < 0.5 &&
		len(classification.Keywords) < 2 {
		return true
	}
	
	return false
}

// assessResponseQuality determines the quality rating of the response
func (s *KnowledgeResponseService) assessResponseQuality(confidence float64, chunkCount int, topScore float64) string {
	if confidence >= s.excellentThreshold && chunkCount >= 2 && topScore >= 0.9 {
		return "excellent"
	} else if confidence >= s.goodThreshold && topScore >= 0.8 {
		return "good"
	} else if confidence >= s.adequateThreshold {
		return "adequate"
	}
	return "poor"
}

// shouldEscalate determines if the question should be escalated to human agents
func (s *KnowledgeResponseService) shouldEscalate(classification *QuestionClassificationResult, confidence float64, topScore float64) bool {
	// Escalate complex technical questions with low confidence
	if classification.Domain == DomainTechnical && 
		classification.Complexity == "complex" && 
		confidence < 0.7 {
		return true
	}
	
	// Escalate billing issues
	if classification.Domain == DomainBilling && classification.Intent == IntentComplaint {
		return true
	}
	
	// Escalate if confidence is very low
	if confidence < 0.5 && topScore < 0.7 {
		return true
	}
	
	// Escalate troubleshooting with low confidence
	if classification.QuestionType == QuestionTypeTroubleshooting && confidence < 0.6 {
		return true
	}
	
	return false
}
