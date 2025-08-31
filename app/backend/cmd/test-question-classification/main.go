package main

import (
	"context"
	"fmt"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/service"
)

func main() {
	fmt.Println("🧠 Testing Question Classification Service")
	fmt.Println("========================================")

	// Create test configuration
	agenticConfig := &config.AgenticConfig{
		Enabled:            true,
		KnowledgeResponses: true,
		KnowledgeConfidence: 0.7,
	}

	classificationService := service.NewQuestionClassificationService(agenticConfig)
	ctx := context.Background()

	// Test various types of messages
	testMessages := []string{
		"What is your pricing?",
		"How do I integrate with your API?",
		"My integration is not working, can you help?",
		"Hello there!",
		"Thank you for your help.",
		"How much does the enterprise plan cost?",
		"Can you help me troubleshoot this error?",
		"What are the features of your product?",
		"How do I configure advanced authentication with custom headers for enterprise API integration?",
		"Why was I charged twice for my subscription?",
		"How do I reset my password?",
		"I need help with my account settings",
		"What is machine learning?",
		"Please create a new account for me",
		"The system is broken and not responding",
	}

	fmt.Println("\n📝 Question Classification Results:")
	fmt.Println("-----------------------------------")

	for i, message := range testMessages {
		result := classificationService.ClassifyQuestion(ctx, message)
		
		fmt.Printf("\n%d. Message: \"%s\"\n", i+1, message)
		fmt.Printf("   📋 Is Question: %v\n", result.IsQuestion)
		fmt.Printf("   🎯 Type: %s\n", result.QuestionType)
		fmt.Printf("   💭 Intent: %s\n", result.Intent)
		fmt.Printf("   🏷️  Domain: %s\n", result.Domain)
		fmt.Printf("   📊 Complexity: %s\n", result.Complexity)
		fmt.Printf("   🎲 Confidence: %.2f\n", result.Confidence)
		fmt.Printf("   📚 Requires Knowledge: %v\n", result.RequiresKnowledge)
		fmt.Printf("   🤖 Can Auto-Respond: %v\n", result.CanAutoRespond)
		if len(result.Keywords) > 0 {
			fmt.Printf("   🔑 Keywords: %v\n", result.Keywords)
		}
	}

	// Test with disabled configuration
	fmt.Println("\n\n🚫 Testing with Disabled Configuration:")
	fmt.Println("--------------------------------------")
	
	disabledConfig := &config.AgenticConfig{
		Enabled:            false,
		KnowledgeResponses: false,
		KnowledgeConfidence: 0.7,
	}
	
	disabledService := service.NewQuestionClassificationService(disabledConfig)
	result := disabledService.ClassifyQuestion(ctx, "What is your pricing?")
	
	fmt.Printf("Message: \"What is your pricing?\"\n")
	fmt.Printf("   📋 Is Question: %v (should be false)\n", result.IsQuestion)
	fmt.Printf("   🤖 Can Auto-Respond: %v (should be false)\n", result.CanAutoRespond)
	fmt.Printf("   🎲 Confidence: %.2f (should be 0.0)\n", result.Confidence)

	// Summary statistics
	fmt.Println("\n\n📈 Classification Summary:")
	fmt.Println("-------------------------")
	
	questionCount := 0
	knowledgeRequiredCount := 0
	autoRespondCount := 0
	
	for _, message := range testMessages {
		result := classificationService.ClassifyQuestion(ctx, message)
		if result.IsQuestion {
			questionCount++
		}
		if result.RequiresKnowledge {
			knowledgeRequiredCount++
		}
		if result.CanAutoRespond {
			autoRespondCount++
		}
	}
	
	fmt.Printf("   Total Messages: %d\n", len(testMessages))
	fmt.Printf("   Detected Questions: %d\n", questionCount)
	fmt.Printf("   Require Knowledge Base: %d\n", knowledgeRequiredCount)
	fmt.Printf("   Can Auto-Respond: %d\n", autoRespondCount)
	
	fmt.Println("\n✅ Question Classification Service Test Complete!")
	fmt.Println("\n🎉 Phase 2 Task 2.1 (Question Classification) is Working!")
}
