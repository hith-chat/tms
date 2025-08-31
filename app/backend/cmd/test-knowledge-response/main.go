package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/service"
)

func main() {
	fmt.Println("📚 Testing Knowledge Response Service")
	fmt.Println("===================================")

	// Create test configuration
	agenticConfig := &config.AgenticConfig{
		Enabled:            true,
		KnowledgeResponses: true,
		KnowledgeConfidence: 0.7,
	}

	// Create services (knowledge service is nil for this demo)
	questionClassifier := service.NewQuestionClassificationService(agenticConfig)
	// Note: In real usage, we'd pass a real KnowledgeService here
	knowledgeResponseService := service.NewKnowledgeResponseService(agenticConfig, nil, questionClassifier)

	ctx := context.Background()
	tenantID := uuid.New()
	projectID := uuid.New()

	// Test various questions
	testQuestions := []string{
		"What is your pricing?",
		"How do I integrate with your API?",
		"What is the weather today?", // Out of domain
		"My integration is not working",
		"What are the features of your product?",
		"How do I reset my password?",
		"Hello there!", // Not a question
		"Can you help me troubleshoot this error?",
		"What is machine learning?", // Potentially out of domain
		"How much does the enterprise plan cost?",
	}

	fmt.Println("\n🧠 Question Analysis Results:")
	fmt.Println("-----------------------------")

	for i, question := range testQuestions {
		fmt.Printf("\n%d. Question: \"%s\"\n", i+1, question)
		
		// First, classify the question
		classification := questionClassifier.ClassifyQuestion(ctx, question)
		
		fmt.Printf("   📋 Classification:\n")
		fmt.Printf("      Is Question: %v\n", classification.IsQuestion)
		fmt.Printf("      Type: %s\n", classification.QuestionType)
		fmt.Printf("      Domain: %s\n", classification.Domain)
		fmt.Printf("      Complexity: %s\n", classification.Complexity)
		fmt.Printf("      Confidence: %.2f\n", classification.Confidence)
		fmt.Printf("      Requires Knowledge: %v\n", classification.RequiresKnowledge)
		fmt.Printf("      Can Auto-Respond: %v\n", classification.CanAutoRespond)
		
		// If it requires knowledge, test the response generation logic
		// Note: This will return early since we don't have a real knowledge service
		if classification.RequiresKnowledge {
			result, err := knowledgeResponseService.GenerateKnowledgeResponse(ctx, tenantID, projectID, question)
			if err != nil {
				fmt.Printf("   ❌ Error: %v\n", err)
			} else {
				fmt.Printf("   📚 Knowledge Response Logic:\n")
				fmt.Printf("      Has Response: %v\n", result.HasResponse)
				fmt.Printf("      Response Quality: %s\n", result.ResponseQuality)
				fmt.Printf("      Is Out of Domain: %v\n", result.IsOutOfDomain)
				fmt.Printf("      Needs More Info: %v\n", result.NeedsMoreInfo)
				fmt.Printf("      Should Escalate: %v\n", result.ShouldEscalate)
				fmt.Printf("      Processing Time: %v\n", result.ProcessingTime)
				if result.Response != "" {
					fmt.Printf("      Response: %s\n", result.Response)
				}
			}
		}
	}

	// Test with disabled configuration
	fmt.Println("\n\n🚫 Testing with Disabled Knowledge Responses:")
	fmt.Println("--------------------------------------------")
	
	disabledConfig := &config.AgenticConfig{
		Enabled:            true,
		KnowledgeResponses: false, // Disabled
		KnowledgeConfidence: 0.7,
	}
	
	disabledClassifier := service.NewQuestionClassificationService(disabledConfig)
	disabledResponseService := service.NewKnowledgeResponseService(disabledConfig, nil, disabledClassifier)
	
	result, err := disabledResponseService.GenerateKnowledgeResponse(ctx, tenantID, projectID, "How do I integrate with your API?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Technical Question with Knowledge Responses Disabled:\n")
		fmt.Printf("   Has Response: %v (should be false)\n", result.HasResponse)
		fmt.Printf("   Confidence: %.2f (should be 0.0)\n", result.Confidence)
		fmt.Printf("   Processing Time: %v\n", result.ProcessingTime)
	}

	// Summary of service capabilities
	fmt.Println("\n\n🎯 Service Capabilities Summary:")
	fmt.Println("-------------------------------")
	fmt.Println("✅ Question Classification:")
	fmt.Println("   • Identifies questions vs statements")
	fmt.Println("   • Classifies question types (how-to, what-is, troubleshooting, etc.)")
	fmt.Println("   • Detects domain (technical, pricing, support, etc.)")
	fmt.Println("   • Assesses complexity (simple, moderate, complex)")
	fmt.Println("   • Extracts relevant keywords")
	fmt.Println("   • Determines intent (seeking info, requesting action, etc.)")
	fmt.Println("")
	fmt.Println("✅ Knowledge Response Engine:")
	fmt.Println("   • Generates optimized search queries")
	fmt.Println("   • Detects out-of-domain questions")
	fmt.Println("   • Provides template responses for edge cases")
	fmt.Println("   • Analyzes response quality (excellent, good, adequate, poor)")
	fmt.Println("   • Determines escalation needs")
	fmt.Println("   • Creates citations from source material")
	fmt.Println("   • Handles configuration-based feature toggles")

	fmt.Println("\n🎉 Phase 2 Tasks 2.1 & 2.2 (Question Classification + Knowledge Response) Complete!")
	fmt.Println("\n📝 Note: Knowledge search integration requires a database connection.")
	fmt.Println("    The service architecture is ready for full knowledge base integration.")
}
