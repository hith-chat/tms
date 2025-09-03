package service

import (
	"context"
	"testing"

	"github.com/bareuptime/tms/internal/config"
)

func TestQuestionClassificationService_DisabledConfig(t *testing.T) {
	// Test with disabled configuration
	agenticConfig := &config.AgenticConfig{
		Enabled:             false,
		KnowledgeResponses:  false,
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
		Enabled:             true,
		KnowledgeResponses:  true,
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
		Enabled:             true,
		KnowledgeResponses:  true,
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
