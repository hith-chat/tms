package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStripMarkdownCodeBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "JSON wrapped in markdown with json tag",
			input: "```json\n" +
				`{"primary_color": "#075985", "secondary_color": "#1F2937"}` + "\n" +
				"```",
			expected: `{"primary_color": "#075985", "secondary_color": "#1F2937"}`,
		},
		{
			name: "JSON wrapped in markdown without tag",
			input: "```\n" +
				`{"primary_color": "#075985", "secondary_color": "#1F2937"}` + "\n" +
				"```",
			expected: `{"primary_color": "#075985", "secondary_color": "#1F2937"}`,
		},
		{
			name: "Plain JSON without markdown",
			input: `{"primary_color": "#075985", "secondary_color": "#1F2937"}`,
			expected: `{"primary_color": "#075985", "secondary_color": "#1F2937"}`,
		},
		{
			name: "JSON with extra whitespace",
			input: "  \n```json\n" +
				`{"primary_color": "#075985"}` + "\n" +
				"```  \n",
			expected: `{"primary_color": "#075985"}`,
		},
		{
			name: "Multiline JSON in markdown",
			input: "```json\n" +
				"{\n" +
				`  "primary_color": "#075985",` + "\n" +
				`  "secondary_color": "#1F2937"` + "\n" +
				"}\n" +
				"```",
			expected: "{\n" +
				`  "primary_color": "#075985",` + "\n" +
				`  "secondary_color": "#1F2937"` + "\n" +
				"}",
		},
		{
			name: "Real example from BB AI",
			input: "```json\n" +
				"{\n" +
				`  "primary_color": "#075985",` + "\n" +
				`  "secondary_color": "#1F2937",` + "\n" +
				`  "background_color": "#F3F4F6",` + "\n" +
				`  "position": "bottom-right",` + "\n" +
				`  "widget_shape": "modern",` + "\n" +
				`  "chat_bubble_style": "modern",` + "\n" +
				`  "welcome_message": "Hello! How can we help you with your documentation needs today?",` + "\n" +
				`  "custom_greeting": "Welcome to Penify.dev! Ask us anything about automated docs.",` + "\n" +
				`  "agent_name": "Penify Support"` + "\n" +
				"}\n" +
				"```",
			expected: "{\n" +
				`  "primary_color": "#075985",` + "\n" +
				`  "secondary_color": "#1F2937",` + "\n" +
				`  "background_color": "#F3F4F6",` + "\n" +
				`  "position": "bottom-right",` + "\n" +
				`  "widget_shape": "modern",` + "\n" +
				`  "chat_bubble_style": "modern",` + "\n" +
				`  "welcome_message": "Hello! How can we help you with your documentation needs today?",` + "\n" +
				`  "custom_greeting": "Welcome to Penify.dev! Ask us anything about automated docs.",` + "\n" +
				`  "agent_name": "Penify Support"` + "\n" +
				"}",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Just whitespace",
			input:    "   \n  \n  ",
			expected: "",
		},
		{
			name: "Markdown block with language tag and leading text",
			input: "Here's the JSON:\n```json\n" +
				`{"key": "value"}` + "\n" +
				"```",
			expected: "Here's the JSON:\n```json\n" +
				`{"key": "value"}` + "\n" +
				"```",
		},
		{
			name: "Incomplete markdown block (missing closing)",
			input: "```json\n" +
				`{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripMarkdownCodeBlocks(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestStripMarkdownCodeBlocks_WithUnmarshal(t *testing.T) {
	// Test that stripped JSON can actually be unmarshaled
	input := "```json\n" +
		"{\n" +
		`  "primary_color": "#075985",` + "\n" +
		`  "secondary_color": "#1F2937",` + "\n" +
		`  "background_color": "#F3F4F6",` + "\n" +
		`  "position": "bottom-right",` + "\n" +
		`  "widget_shape": "modern",` + "\n" +
		`  "chat_bubble_style": "modern",` + "\n" +
		`  "welcome_message": "Hello! How can we help you with your documentation needs today?",` + "\n" +
		`  "custom_greeting": "Welcome to Penify.dev! Ask us anything about automated docs.",` + "\n" +
		`  "agent_name": "Penify Support"` + "\n" +
		"}\n" +
		"```"

	cleaned := stripMarkdownCodeBlocks(input)

	// Try to unmarshal it
	var result map[string]interface{}
	err := json.Unmarshal([]byte(cleaned), &result)
	require.NoError(t, err, "Should be able to unmarshal cleaned JSON")

	// Verify some values
	require.Equal(t, "#075985", result["primary_color"])
	require.Equal(t, "bottom-right", result["position"])
	require.Equal(t, "Penify Support", result["agent_name"])
}
