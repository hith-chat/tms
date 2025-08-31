package service

import (
	"context"
	"testing"

	"github.com/bareuptime/tms/internal/config"
)

func TestQuestionClassificationService_ClassifyQuestion(t *testing.T) {
	// Create test configuration
	agenticConfig := &config.AgenticConfig{
		Enabled:            true,
		KnowledgeResponses: true,
		KnowledgeConfidence: 0.7,
	}

	service := NewQuestionClassificationService(agenticConfig)
	ctx := context.Background()

	tests := []struct {
		name                 string
		message              string
		expectedIsQuestion   bool
		expectedQuestionType QuestionType
		expectedIntent       QuestionIntent
		expectedDomain       QuestionDomain
		expectedComplexity   string
		minConfidence        float64
	}{
		// Basic questions
		{
			name:                 "Simple what-is question",
			message:              "What is your pricing?",
			expectedIsQuestion:   true,
			expectedQuestionType: QuestionTypeWhatIs,
			expectedIntent:       IntentSeekingInfo,
			expectedDomain:       DomainPricing,
			expectedComplexity:   "simple",
			minConfidence:        0.6,
		},
		{
			name:                 "How-to question",
			message:              "How do I integrate with your API?",
			expectedIsQuestion:   true,
			expectedQuestionType: QuestionTypeHowTo,
			expectedIntent:       IntentSeekingInfo,
			expectedDomain:       DomainTechnical,
			expectedComplexity:   "moderate",
			minConfidence:        0.7,
		},
		{
			name:                 "Troubleshooting question",
			message:              "My integration is not working, can you help?",
			expectedIsQuestion:   true,
			expectedQuestionType: QuestionTypeTroubleshooting,
			expectedIntent:       IntentRequestingAction,
			expectedDomain:       DomainTechnical,
			expectedComplexity:   "moderate",
			minConfidence:        0.6,
		},
		{
			name:                 "Pricing question",
			message:              "How much does the enterprise plan cost?",
			expectedIsQuestion:   true,
			expectedQuestionType: QuestionTypePricing,
			expectedIntent:       IntentSeekingInfo,
			expectedDomain:       DomainPricing,
			expectedComplexity:   "simple",
			minConfidence:        0.7,
		},
		// Non-questions
		{
			name:                 "Statement not question",
			message:              "Thank you for your help.",
			expectedIsQuestion:   false,
			expectedQuestionType: QuestionTypeGeneral,
			expectedIntent:       IntentSeekingInfo,
			expectedDomain:       DomainGeneral,
			expectedComplexity:   "simple",
			minConfidence:        0.0,
		},
		{
			name:                 "Greeting not question",
			message:              "Hello there!",
			expectedIsQuestion:   false,
			expectedQuestionType: QuestionTypeGeneral,
			expectedIntent:       IntentGreeting,
			expectedDomain:       DomainGeneral,
			expectedComplexity:   "simple",
			minConfidence:        0.0,
		},
		// Complex questions
		{
			name:                 "Complex technical question",
			message:              "How do I configure advanced authentication with custom headers for enterprise API integration?",
			expectedIsQuestion:   true,
			expectedQuestionType: QuestionTypeHowTo,
			expectedIntent:       IntentSeekingInfo,
			expectedDomain:       DomainTechnical,
			expectedComplexity:   "complex",
			minConfidence:        0.7,
		},
		// Domain detection
		{
			name:                 "Billing domain question",
			message:              "Why was I charged twice for my subscription?",
			expectedIsQuestion:   true,
			expectedQuestionType: QuestionTypeTroubleshooting,
			expectedIntent:       IntentComplaint,
			expectedDomain:       DomainBilling,
			expectedComplexity:   "moderate",
			minConfidence:        0.6,
		},
		{
			name:                 "Account domain question",
			message:              "How do I reset my password?",
			expectedIsQuestion:   true,
			expectedQuestionType: QuestionTypeHowTo,
			expectedIntent:       IntentSeekingInfo,
			expectedDomain:       DomainAccount,
			expectedComplexity:   "simple",
			minConfidence:        0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ClassifyQuestion(ctx, tt.message)

			// Test basic classification
			if result.IsQuestion != tt.expectedIsQuestion {
				t.Errorf("IsQuestion = %v, want %v", result.IsQuestion, tt.expectedIsQuestion)
			}

			if result.QuestionType != tt.expectedQuestionType {
				t.Errorf("QuestionType = %v, want %v", result.QuestionType, tt.expectedQuestionType)
			}

			if result.Intent != tt.expectedIntent {
				t.Errorf("Intent = %v, want %v", result.Intent, tt.expectedIntent)
			}

			if result.Domain != tt.expectedDomain {
				t.Errorf("Domain = %v, want %v", result.Domain, tt.expectedDomain)
			}

			if result.Complexity != tt.expectedComplexity {
				t.Errorf("Complexity = %v, want %v", result.Complexity, tt.expectedComplexity)
			}

			if result.Confidence < tt.minConfidence {
				t.Errorf("Confidence = %v, want >= %v", result.Confidence, tt.minConfidence)
			}

			// Test knowledge requirements
			if tt.expectedIsQuestion {
				if result.RequiresKnowledge && tt.expectedDomain == DomainGeneral && tt.expectedComplexity == "simple" {
					t.Errorf("Simple general questions should not require knowledge base")
				}
			}

			// Log results for debugging
			t.Logf("Message: '%s'", tt.message)
			t.Logf("  IsQuestion: %v", result.IsQuestion)
			t.Logf("  Type: %v", result.QuestionType)
			t.Logf("  Intent: %v", result.Intent)
			t.Logf("  Domain: %v", result.Domain)
			t.Logf("  Complexity: %v", result.Complexity)
			t.Logf("  Confidence: %.2f", result.Confidence)
			t.Logf("  RequiresKnowledge: %v", result.RequiresKnowledge)
			t.Logf("  CanAutoRespond: %v", result.CanAutoRespond)
			t.Logf("  Keywords: %v", result.Keywords)
		})
	}
}

func TestQuestionClassificationService_DisabledConfig(t *testing.T) {
	// Test with disabled configuration
	agenticConfig := &config.AgenticConfig{
		Enabled:            false,
		KnowledgeResponses: false,
		KnowledgeConfidence: 0.7,
	}

	service := NewQuestionClassificationService(agenticConfig)
	ctx := context.Background()

	result := service.ClassifyQuestion(ctx, "What is your pricing?")

	if result.IsQuestion {
		t.Errorf("Should not classify questions when disabled")
	}

	if result.CanAutoRespond {
		t.Errorf("Should not auto-respond when disabled")
	}

	if result.Confidence != 0.0 {
		t.Errorf("Confidence should be 0.0 when disabled, got %v", result.Confidence)
	}
}

func TestQuestionClassificationService_KeywordExtraction(t *testing.T) {
	agenticConfig := &config.AgenticConfig{
		Enabled:            true,
		KnowledgeResponses: true,
		KnowledgeConfidence: 0.7,
	}

	service := NewQuestionClassificationService(agenticConfig)
	ctx := context.Background()

	result := service.ClassifyQuestion(ctx, "How do I configure the advanced API authentication system for enterprise users?")

	expectedKeywords := []string{"configure", "advanced", "api", "authentication", "system", "enterprise", "users"}
	
	// Check that we extracted meaningful keywords
	if len(result.Keywords) == 0 {
		t.Errorf("Should extract keywords from the message")
	}

	// Check for some expected keywords
	keywordMap := make(map[string]bool)
	for _, keyword := range result.Keywords {
		keywordMap[keyword] = true
	}

	foundKeywords := 0
	for _, expected := range expectedKeywords {
		if keywordMap[expected] {
			foundKeywords++
		}
	}

	if foundKeywords < 3 {
		t.Errorf("Should find at least 3 expected keywords, found %d. Keywords: %v", foundKeywords, result.Keywords)
	}
}

func BenchmarkQuestionClassification(b *testing.B) {
	agenticConfig := &config.AgenticConfig{
		Enabled:            true,
		KnowledgeResponses: true,
		KnowledgeConfidence: 0.7,
	}

	service := NewQuestionClassificationService(agenticConfig)
	ctx := context.Background()

	messages := []string{
		"What is your pricing?",
		"How do I integrate with your API?",
		"My account is not working",
		"Can you help me troubleshoot this error?",
		"Hello there!",
		"How much does the enterprise plan cost?",
		"What are the features of your product?",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		message := messages[i%len(messages)]
		service.ClassifyQuestion(ctx, message)
	}
}
