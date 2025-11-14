package service

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestJSONExtractionFromResponse tests the JSON extraction logic used in rankURLsWithAI
func TestJSONExtractionFromResponse(t *testing.T) {
	tests := []struct {
		name          string
		response      string
		expectedURLs  []string
		shouldError   bool
	}{
		{
			name:     "Valid JSON only",
			response: `["https://example.com/page1", "https://example.com/page2"]`,
			expectedURLs: []string{
				"https://example.com/page1",
				"https://example.com/page2",
			},
			shouldError: false,
		},
		{
			name: "JSON with note before",
			response: `Note: Only 2 unique pages were provided.

["https://example.com/page1", "https://example.com/page2"]`,
			expectedURLs: []string{
				"https://example.com/page1",
				"https://example.com/page2",
			},
			shouldError: false,
		},
		{
			name: "JSON with explanation after",
			response: `["https://example.com/page1", "https://example.com/page2"]

Note: I selected these pages because they contain the most relevant content.`,
			expectedURLs: []string{
				"https://example.com/page1",
				"https://example.com/page2",
			},
			shouldError: false,
		},
		{
			name: "JSON with markdown code block",
			response: "```json\n[\"https://example.com/page1\", \"https://example.com/page2\"]\n```",
			expectedURLs: []string{
				"https://example.com/page1",
				"https://example.com/page2",
			},
			shouldError: false,
		},
		{
			name: "JSON with whitespace",
			response: `

["https://example.com/page1", "https://example.com/page2"]

`,
			expectedURLs: []string{
				"https://example.com/page1",
				"https://example.com/page2",
			},
			shouldError: false,
		},
		{
			name:        "No JSON array",
			response:    "This is just text without JSON",
			shouldError: true,
		},
		{
			name:        "Empty response",
			response:    "",
			shouldError: true,
		},
		{
			name:        "Malformed JSON",
			response:    `["https://example.com/page1", "https://example.com/page2"`,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the cleaning logic from rankURLsWithAI
			response := stripMarkdownCodeBlocks(tt.response)
			response = strings.TrimSpace(response)

			// Extract JSON array
			startIdx := strings.Index(response, "[")
			endIdx := strings.LastIndex(response, "]")

			if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
				if !tt.shouldError {
					t.Errorf("Expected to find JSON array but didn't")
				}
				return
			}

			jsonStr := response[startIdx : endIdx+1]

			var rankedURLs []string
			err := json.Unmarshal([]byte(jsonStr), &rankedURLs)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(rankedURLs) != len(tt.expectedURLs) {
				t.Errorf("Expected %d URLs but got %d", len(tt.expectedURLs), len(rankedURLs))
				return
			}

			for i, url := range rankedURLs {
				if url != tt.expectedURLs[i] {
					t.Errorf("URL[%d]: expected %s, got %s", i, tt.expectedURLs[i], url)
				}
			}
		})
	}
}
