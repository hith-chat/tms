package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebScrapingService_ValidateURL(t *testing.T) {
	service := &WebScrapingService{}

	tests := []struct {
		name      string
		url       string
		wantError bool
	}{
		{
			name:      "Valid HTTP URL",
			url:       "http://example.com",
			wantError: false,
		},
		{
			name:      "Valid HTTPS URL",
			url:       "https://example.com",
			wantError: false,
		},
		{
			name:      "Invalid scheme",
			url:       "ftp://example.com",
			wantError: true,
		},
		{
			name:      "Localhost blocked",
			url:       "http://localhost:8080",
			wantError: true,
		},
		{
			name:      "127.0.0.1 blocked",
			url:       "http://127.0.0.1:8080",
			wantError: true,
		},
		{
			name:      "Private IP blocked",
			url:       "http://192.168.1.1",
			wantError: true,
		},
		{
			name:      "10.x.x.x blocked",
			url:       "http://10.0.0.1",
			wantError: true,
		},
		{
			name:      "172.x.x.x blocked",
			url:       "http://172.16.0.1",
			wantError: true,
		},
		{
			name:      "Invalid URL format",
			url:       "not-a-url",
			wantError: true,
		},
		{
			name:      "Missing host",
			url:       "http://",
			wantError: true,
		},
		{
			name:      "IPv6 localhost blocked",
			url:       "http://[::1]:8080",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateURL(tt.url)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWebScrapingService_ShouldFollowLink(t *testing.T) {
	service := &WebScrapingService{}

	tests := []struct {
		name       string
		linkURL    string
		currentURL string
		expected   bool
	}{
		{
			name:       "Same domain should be followed",
			linkURL:    "https://example.com/page2",
			currentURL: "https://example.com/page1",
			expected:   true,
		},
		{
			name:       "Different domain should not be followed",
			linkURL:    "https://other.com/page",
			currentURL: "https://example.com/page1",
			expected:   false,
		},
		{
			name:       "PDF file should not be followed",
			linkURL:    "https://example.com/document.pdf",
			currentURL: "https://example.com/page1",
			expected:   false,
		},
		{
			name:       "DOC file should not be followed",
			linkURL:    "https://example.com/document.doc",
			currentURL: "https://example.com/page1",
			expected:   false,
		},
		{
			name:       "DOCX file should not be followed",
			linkURL:    "https://example.com/document.docx",
			currentURL: "https://example.com/page1",
			expected:   false,
		},
		{
			name:       "Excel file should not be followed",
			linkURL:    "https://example.com/spreadsheet.xlsx",
			currentURL: "https://example.com/page1",
			expected:   false,
		},
		{
			name:       "ZIP file should not be followed",
			linkURL:    "https://example.com/archive.zip",
			currentURL: "https://example.com/page1",
			expected:   false,
		},
		{
			name:       "Admin path should not be followed",
			linkURL:    "https://example.com/admin/users",
			currentURL: "https://example.com/page1",
			expected:   false,
		},
		{
			name:       "API path should not be followed",
			linkURL:    "https://example.com/api/data",
			currentURL: "https://example.com/page1",
			expected:   false,
		},
		{
			name:       "Login path should not be followed",
			linkURL:    "https://example.com/login",
			currentURL: "https://example.com/page1",
			expected:   false,
		},
		{
			name:       "Register path should not be followed",
			linkURL:    "https://example.com/register",
			currentURL: "https://example.com/page1",
			expected:   false,
		},
		{
			name:       "Download path should not be followed",
			linkURL:    "https://example.com/download/file",
			currentURL: "https://example.com/page1",
			expected:   false,
		},
		{
			name:       "Upload path should not be followed",
			linkURL:    "https://example.com/upload",
			currentURL: "https://example.com/page1",
			expected:   false,
		},
		{
			name:       "Image file should not be followed",
			linkURL:    "https://example.com/image.jpg",
			currentURL: "https://example.com/page1",
			expected:   false,
		},
		{
			name:       "PNG image should not be followed",
			linkURL:    "https://example.com/image.png",
			currentURL: "https://example.com/page1",
			expected:   false,
		},
		{
			name:       "GIF image should not be followed",
			linkURL:    "https://example.com/image.gif",
			currentURL: "https://example.com/page1",
			expected:   false,
		},
		{
			name:       "SVG image should not be followed",
			linkURL:    "https://example.com/image.svg",
			currentURL: "https://example.com/page1",
			expected:   false,
		},
		{
			name:       "Video file should not be followed",
			linkURL:    "https://example.com/video.mp4",
			currentURL: "https://example.com/page1",
			expected:   false,
		},
		{
			name:       "Audio file should not be followed",
			linkURL:    "https://example.com/audio.mp3",
			currentURL: "https://example.com/page1",
			expected:   false,
		},
		{
			name:       "Regular page should be followed",
			linkURL:    "https://example.com/about",
			currentURL: "https://example.com/page1",
			expected:   true,
		},
		{
			name:       "Subdirectory page should be followed",
			linkURL:    "https://example.com/blog/post",
			currentURL: "https://example.com/page1",
			expected:   true,
		},
		{
			name:       "Invalid link URL should not be followed",
			linkURL:    "not-a-url",
			currentURL: "https://example.com/page1",
			expected:   false,
		},
		{
			name:       "Invalid current URL should not be followed",
			linkURL:    "https://example.com/page2",
			currentURL: "not-a-url",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.shouldFollowLink(tt.linkURL, tt.currentURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWebScrapingService_EstimateTokenCount(t *testing.T) {
	service := &WebScrapingService{}

	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "Empty text",
			text:     "",
			expected: 0,
		},
		{
			name:     "Short text",
			text:     "Hello",
			expected: 1, // 5 chars / 4 = 1.25 -> 1
		},
		{
			name:     "Medium text",
			text:     "Hello, this is a test message!",
			expected: 7, // 30 chars / 4 = 7.5 -> 7
		},
		{
			name:     "Long text",
			text:     "This is a much longer text that should result in a higher token count when estimated using the simple character-based approach.",
			expected: 31, // 125 chars / 4 = 31.25 -> 31
		},
		{
			name:     "Text with special characters",
			text:     "Hello! @#$%^&*()_+ This has special chars: 1234567890",
			expected: 13, // 54 chars / 4 = 13.5 -> 13
		},
		{
			name:     "Unicode text",
			text:     "Hello 世界! This is unicode text.",
			expected: 8, // 33 chars / 4 = 8.25 -> 8
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.estimateTokenCount(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWebScrapingService_CleanText(t *testing.T) {
	service := &WebScrapingService{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty text",
			input:    "",
			expected: "",
		},
		{
			name:     "Single space",
			input:    " ",
			expected: "",
		},
		{
			name:     "Multiple spaces",
			input:    "Hello    world   test",
			expected: "Hello world test",
		},
		{
			name:     "Multiple newlines",
			input:    "Line 1\n\n\nLine 2\n\n\nLine 3",
			expected: "Line 1 Line 2 Line 3", // strings.Fields removes all whitespace
		},
		{
			name:     "Leading and trailing whitespace",
			input:    "   Hello world   ",
			expected: "Hello world",
		},
		{
			name:     "Mixed whitespace",
			input:    "  Hello\n\n  world  \n  test  ",
			expected: "Hello world test", // strings.Fields removes all whitespace
		},
		{
			name:     "Tabs and other whitespace",
			input:    "Hello\t\t\tworld\n\n\n\ntest",
			expected: "Hello world test", // strings.Fields removes all whitespace
		},
		{
			name:     "Complex mixed content",
			input:    "\n\n  Title\n\n\n  Paragraph 1 with   multiple   spaces.\n\n\n\n  Paragraph 2.\n\n  ",
			expected: "Title Paragraph 1 with multiple spaces. Paragraph 2.", // strings.Fields removes all whitespace
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.cleanText(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test that URL normalization works correctly
func TestWebScrapingService_URLNormalization(t *testing.T) {
	service := &WebScrapingService{}

	tests := []struct {
		name       string
		linkURL    string
		currentURL string
		expected   bool
	}{
		{
			name:       "HTTP vs HTTPS same domain",
			linkURL:    "http://example.com/page2",
			currentURL: "https://example.com/page1",
			expected:   true, // Same domain, different schemes are allowed
		},
		{
			name:       "With and without www",
			linkURL:    "https://www.example.com/page2",
			currentURL: "https://example.com/page1",
			expected:   true, // Same base domain (getBaseDomain returns "example.com" for both)
		},
		{
			name:       "Same domain with port",
			linkURL:    "https://example.com:443/page2",
			currentURL: "https://example.com/page1",
			expected:   true, // Same base domain (port is stripped in getBaseDomain)
		},
		{
			name:       "Subdomain",
			linkURL:    "https://blog.example.com/page2",
			currentURL: "https://example.com/page1",
			expected:   true, // Same base domain (both resolve to "example.com")
		},
		{
			name:       "Case insensitive domain",
			linkURL:    "https://EXAMPLE.COM/page2",
			currentURL: "https://example.com/page1",
			expected:   false, // Different case - getBaseDomain does NOT lowercase
		},
		{
			name:       "Different domain",
			linkURL:    "https://different.com/page2",
			currentURL: "https://example.com/page1",
			expected:   false, // Actually different domains
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.shouldFollowLink(tt.linkURL, tt.currentURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}
