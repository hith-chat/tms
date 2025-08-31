package main

import (
	"context"
	"fmt"
	"log"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/service"
)

func main() {
	fmt.Println("🔍 Testing Agent Request Detection Service")
	fmt.Println("==========================================")

	// Create test configuration
	agenticConfig := &config.AgenticConfig{
		Enabled:               true,
		AgentRequestDetection: true,
		AgentRequestThreshold: 0.7,
		AgentRequestKeywords: []string{
			"agent", "human", "speak to someone", "talk to agent", "connect me", "customer service",
		},
	}

	// Create the service
	detector := service.NewAgentRequestDetectionService(agenticConfig)

	// Test scenarios
	testScenarios := []struct {
		name     string
		message  string
		expected bool
	}{
		{
			name:     "Explicit Agent Request",
			message:  "Can I speak to a human agent please?",
			expected: true,
		},
		{
			name:     "Direct Human Request",
			message:  "I need to talk to a real person",
			expected: true,
		},
		{
			name:     "Customer Service Request",
			message:  "Connect me to customer service",
			expected: true,
		},
		{
			name:     "Escalation Request",
			message:  "This bot can't help me, I need a supervisor",
			expected: true,
		},
		{
			name:     "Urgent Issue",
			message:  "This is urgent, I need help immediately",
			expected: true,
		},
		{
			name:     "Complaint with Agent Request",
			message:  "This is terrible service, get me a manager",
			expected: true,
		},
		{
			name:     "Technical Issue Escalation",
			message:  "The API is not working and I've tried everything. Can someone help?",
			expected: true,
		},
		{
			name:     "Billing Dispute",
			message:  "You charged me twice and I demand a refund",
			expected: true,
		},
		{
			name:     "Legal Threat",
			message:  "This is unacceptable, I'm going to sue you",
			expected: true,
		},
		{
			name:     "Simple Question",
			message:  "How do I reset my password?",
			expected: false,
		},
		{
			name:     "General Inquiry",
			message:  "What are your pricing plans?",
			expected: false,
		},
		{
			name:     "Thank You",
			message:  "Thank you for your help",
			expected: false,
		},
		{
			name:     "Greeting",
			message:  "Hello there!",
			expected: false,
		},
		{
			name:     "Product Question",
			message:  "Does your software support integrations?",
			expected: false,
		},
		{
			name:     "Documentation Request",
			message:  "Where can I find the API documentation?",
			expected: false,
		},
	}

	ctx := context.Background()
	totalTests := len(testScenarios)
	correctPredictions := 0
	agentRequests := 0

	fmt.Println("\n🧠 Agent Request Detection Results:")
	fmt.Println("-----------------------------------")

	for i, scenario := range testScenarios {
		result, err := detector.DetectAgentRequest(ctx, scenario.message)
		if err != nil {
			log.Printf("Error detecting agent request: %v", err)
			continue
		}

		isCorrect := result.IsAgentRequest == scenario.expected
		if isCorrect {
			correctPredictions++
		}

		if result.IsAgentRequest {
			agentRequests++
		}

		statusIcon := "✅"
		if !isCorrect {
			statusIcon = "❌"
		}

		urgencyColor := ""
		switch result.Urgency {
		case service.UrgencyCritical:
			urgencyColor = "🚨"
		case service.UrgencyHigh:
			urgencyColor = "🔴"
		case service.UrgencyNormal:
			urgencyColor = "🟡"
		case service.UrgencyLow:
			urgencyColor = "🟢"
		}

		fmt.Printf("\n%d. %s: %s\n", i+1, scenario.name, statusIcon)
		fmt.Printf("   Message: \"%s\"\n", scenario.message)
		fmt.Printf("   🎯 Detection Results:\n")
		fmt.Printf("      Agent Request: %t (expected: %t)\n", result.IsAgentRequest, scenario.expected)
		fmt.Printf("      Confidence: %.2f\n", result.Confidence)
		fmt.Printf("      Request Type: %s\n", result.RequestType)
		fmt.Printf("      Urgency: %s %s\n", urgencyColor, result.Urgency)
		fmt.Printf("      Processing Time: %dms\n", result.ProcessingTimeMs)

		if len(result.Keywords) > 0 {
			fmt.Printf("      📝 Keywords: %v\n", result.Keywords)
		}

		if len(result.Reasoning) > 0 {
			fmt.Printf("      🧠 Reasoning:\n")
			for _, reason := range result.Reasoning {
				fmt.Printf("         • %s\n", reason)
			}
		}
	}

	// Test with disabled configuration
	fmt.Println("\n🚫 Testing with Disabled Agent Request Detection:")
	fmt.Println("------------------------------------------------")

	disabledConfig := &config.AgenticConfig{
		Enabled:               true,
		AgentRequestDetection: false,
		AgentRequestThreshold: 0.7,
	}

	disabledDetector := service.NewAgentRequestDetectionService(disabledConfig)
	disabledResult, err := disabledDetector.DetectAgentRequest(ctx, "Can I please speak to a human agent?")
	if err != nil {
		log.Printf("Error with disabled detector: %v", err)
	} else {
		fmt.Printf("Message: \"Can I please speak to a human agent?\"\n")
		fmt.Printf("   Agent Request: %t (should be false)\n", disabledResult.IsAgentRequest)
		fmt.Printf("   Reasoning: %v\n", disabledResult.Reasoning)
	}

	// Performance summary
	fmt.Println("\n📊 Performance & Capabilities Summary:")
	fmt.Println("------------------------------------")
	accuracy := float64(correctPredictions) / float64(totalTests) * 100
	fmt.Printf("   Total Scenarios Tested: %d\n", totalTests)
	fmt.Printf("   Correct Predictions: %d (%.1f%%)\n", correctPredictions, accuracy)
	fmt.Printf("   Agent Requests Detected: %d (%.1f%%)\n", agentRequests, float64(agentRequests)/float64(totalTests)*100)

	fmt.Println("\n✅ Service Capabilities:")
	fmt.Println("   • Explicit agent request detection with high confidence")
	fmt.Println("   • Complaint and escalation indicators")
	fmt.Println("   • Urgency level classification (low, normal, high, critical)")
	fmt.Println("   • Request type categorization (general, urgent, complaint, technical, billing)")
	fmt.Println("   • Contextual request detection (bot failure, tried everything)")
	fmt.Println("   • Technical and billing issue escalation")
	fmt.Println("   • Legal threat and emergency detection")
	fmt.Println("   • Configurable confidence thresholds")
	fmt.Println("   • Detailed reasoning and keyword extraction")

	if accuracy >= 80 {
		fmt.Println("\n🎉 Phase 3 Task 3.1 Complete!")
		fmt.Println("    (Agent Request Detection)")
		fmt.Println("\n🚀 Agent Request Detection Service is Ready!")
	} else {
		fmt.Printf("\n⚠️  Accuracy %.1f%% below target 80%% - needs improvement\n", accuracy)
	}
}
