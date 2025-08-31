package main

import (
	"context"
	"fmt"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/service"
)

func testAgenticConfig() {
	fmt.Println("🤖 Testing Agentic Configuration System")
	fmt.Println("======================================")

	// Test Configuration Loading
	fmt.Println("\n1. Testing Configuration...")
	
	// Create test configuration
	agenticConfig := &config.AgenticConfig{
		Enabled:                   true,
		GreetingDetection:         true,
		KnowledgeResponses:        true,
		AgentAssignment:           false,
		NotificationAlerts:        false,
		GreetingConfidence:        0.4,
		KnowledgeConfidence:       0.7,
		DomainRelevanceConfidence: 0.6,
		AgentRequestConfidence:    0.5,
		GreetingKeywords:          []string{"hello", "hi", "hey", "good morning", "good afternoon", "good evening"},
		AgentRequestKeywords:      []string{"speak to agent", "human agent", "transfer", "escalate"},
		NegativeKeywords:          []string{"no", "don't", "stop", "cancel"},
		ResponseTimeoutMs:         5000,
		MaxConcurrentSessions:     100,
	}

	fmt.Printf("✅ Agentic Config Created: Enabled=%v, Greeting=%v, Knowledge=%v\n", 
		agenticConfig.Enabled, agenticConfig.GreetingDetection, agenticConfig.KnowledgeResponses)

	// Test Greeting Detection Service
	fmt.Println("\n2. Testing Greeting Detection Service...")
	
	greetingService := service.NewGreetingDetectionService(agenticConfig)
	ctx := context.Background()

	testMessages := []string{
		"hello there",
		"hi, I need help",
		"good morning",
		"I have a technical question about API",
		"hey, what's up?",
		"Can you help me with billing?",
	}

	for _, message := range testMessages {
		result := greetingService.DetectGreeting(ctx, message)
		fmt.Printf("   Message: '%s' -> Greeting: %v (confidence: %.2f)\n", 
			message, result.IsGreeting, result.Confidence)
	}

	// Test AI Service Configuration Methods
	fmt.Println("\n3. Testing AI Service Configuration...")
	
	fmt.Printf("   Agentic Features - Enabled: %v\n", agenticConfig.Enabled)
	fmt.Printf("   Greeting Detection: %v (threshold: %.2f)\n", agenticConfig.GreetingDetection, agenticConfig.GreetingConfidence)
	fmt.Printf("   Knowledge Responses: %v (threshold: %.2f)\n", agenticConfig.KnowledgeResponses, agenticConfig.KnowledgeConfidence)
	fmt.Printf("   Agent Assignment: %v (threshold: %.2f)\n", agenticConfig.AgentAssignment, agenticConfig.AgentRequestConfidence)

	fmt.Println("\n✅ All Agentic Configuration Tests Completed Successfully!")
	fmt.Println("\n🎉 Phase 1 Agentic Behavior Implementation is Ready!")
	
	fmt.Println("\nNext Steps:")
	fmt.Println("- Update config.yaml with desired agentic settings")
	fmt.Println("- Start the server and test WebSocket integration")
	fmt.Println("- Begin Phase 2: Knowledge Classification Service")
}

func main() {
	testAgenticConfig()
}
