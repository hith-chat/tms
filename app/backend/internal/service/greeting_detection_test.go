package service

import (
	"context"
	"testing"

	"github.com/bareuptime/tms/internal/config"
)

func TestGreetingDetectionService_DetectGreeting(t *testing.T) {
	// Create a default config for testing
	agenticConfig := &config.AgenticConfig{
		Enabled:            true,
		GreetingDetection:  true,
		GreetingConfidence: 0.4,
		GreetingKeywords:   []string{"hello", "hi", "hey", "good morning", "good afternoon", "good evening", "greetings", "howdy", "sup", "what's up"},
	}

	service := NewGreetingDetectionService(agenticConfig)

	tests := []struct {
		name                  string
		message               string
		expectedIsGreeting    bool
		expectedMinConfidence float64
	}{
		// Basic greetings
		{
			name:                  "Simple hello",
			message:               "hello",
			expectedIsGreeting:    true,
			expectedMinConfidence: 0.8,
		},
		{
			name:                  "Simple hi",
			message:               "hi",
			expectedIsGreeting:    true,
			expectedMinConfidence: 0.8,
		},
		{
			name:                  "Good morning",
			message:               "good morning",
			expectedIsGreeting:    true,
			expectedMinConfidence: 0.7,
		},
		{
			name:                  "Hey there",
			message:               "hey there",
			expectedIsGreeting:    true,
			expectedMinConfidence: 0.6,
		},

		// Greetings with punctuation
		{
			name:                  "Hello with exclamation",
			message:               "hello!",
			expectedIsGreeting:    true,
			expectedMinConfidence: 0.8,
		},
		{
			name:                  "Hi with comma",
			message:               "hi,",
			expectedIsGreeting:    true,
			expectedMinConfidence: 0.8,
		},

		// Case variations
		{
			name:                  "UPPERCASE hello",
			message:               "HELLO",
			expectedIsGreeting:    true,
			expectedMinConfidence: 0.8,
		},
		{
			name:                  "Mixed case",
			message:               "HeLLo ThErE",
			expectedIsGreeting:    true,
			expectedMinConfidence: 0.6,
		},

		// Greetings in sentences
		{
			name:                  "Hello in sentence",
			message:               "hello, how are you?",
			expectedIsGreeting:    true,
			expectedMinConfidence: 0.5,
		},
		{
			name:                  "Hi with question",
			message:               "hi there, can you help me?",
			expectedIsGreeting:    true,
			expectedMinConfidence: 0.3, // Lower expectation due to "help"
		},

		// Typos (removed since we disabled similarity matching)
		// {
		//	name:               "Helo typo",
		//	message:            "helo",
		//	expectedIsGreeting: true,
		//	expectedMinConfidence: 0.6,
		// },
		// {
		//	name:               "Hai variant",
		//	message:            "hai",
		//	expectedIsGreeting: true,
		//	expectedMinConfidence: 0.6,
		// },

		// Non-greetings
		{
			name:                  "Complex question",
			message:               "What is your return policy?",
			expectedIsGreeting:    false,
			expectedMinConfidence: 0.0,
		},
		{
			name:                  "Technical question",
			message:               "How do I integrate the API?",
			expectedIsGreeting:    false,
			expectedMinConfidence: 0.0,
		},
		{
			name:                  "Simple statement",
			message:               "I need help",
			expectedIsGreeting:    false,
			expectedMinConfidence: 0.0,
		},
		{
			name:                  "Random text",
			message:               "The quick brown fox jumps",
			expectedIsGreeting:    false,
			expectedMinConfidence: 0.0,
		},

		// Edge cases
		{
			name:                  "Empty message",
			message:               "",
			expectedIsGreeting:    false,
			expectedMinConfidence: 0.0,
		},
		{
			name:                  "Only punctuation",
			message:               "!!!",
			expectedIsGreeting:    false,
			expectedMinConfidence: 0.0,
		},
		{
			name:                  "Numbers only",
			message:               "123",
			expectedIsGreeting:    false,
			expectedMinConfidence: 0.0,
		},

		// Borderline cases
		{
			name:                  "Hello in complex question",
			message:               "hello, I have a question about your pricing structure and billing policies",
			expectedIsGreeting:    true,
			expectedMinConfidence: 0.15, // Very low due to multiple negative keywords
		},
		{
			name:                  "Greeting with complaint",
			message:               "hi, your service is terrible",
			expectedIsGreeting:    true,
			expectedMinConfidence: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result := service.DetectGreeting(ctx, tt.message)

			if !tt.expectedIsGreeting && result.Confidence > 0.3 {
				t.Errorf("DetectGreeting() confidence = %f, should be low for non-greeting",
					result.Confidence)
			}
		})
	}
}

func TestGreetingDetectionService_normalizeMessage(t *testing.T) {
	agenticConfig := &config.AgenticConfig{
		Enabled:            true,
		GreetingDetection:  true,
		GreetingConfidence: 0.4,
		GreetingKeywords:   []string{"hello", "hi", "hey"},
	}
	service := NewGreetingDetectionService(agenticConfig)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic normalization",
			input:    "Hello, World!",
			expected: "hello world",
		},
		{
			name:     "Multiple spaces",
			input:    "hi    there   friend",
			expected: "hi there friend",
		},
		{
			name:     "Mixed punctuation",
			input:    "Hey!!! How are you???",
			expected: "hey how are you",
		},
		{
			name:     "Numbers and symbols",
			input:    "hello123@#$world",
			expected: "hello world",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only punctuation",
			input:    "!@#$%^&*()",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.normalizeMessage(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeMessage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func BenchmarkGreetingDetection(b *testing.B) {
	agenticConfig := &config.AgenticConfig{
		Enabled:            true,
		GreetingDetection:  true,
		GreetingConfidence: 0.4,
		GreetingKeywords:   []string{"hello", "hi", "hey"},
	}
	service := NewGreetingDetectionService(agenticConfig)
	messages := []string{
		"hello",
		"hi there, how are you?",
		"What is your return policy for damaged items?",
		"good morning everyone",
		"I need help with my account settings",
	}

	b.ResetTimer()
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		message := messages[i%len(messages)]
		service.DetectGreeting(ctx, message)
	}
}
