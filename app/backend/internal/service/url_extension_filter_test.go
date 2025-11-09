package service

import (
	"testing"

	"github.com/bareuptime/tms/internal/config"
)

func TestShouldCrawlURLByExtension(t *testing.T) {
	// Create test service with default allowed extensions
	s := &WebScrapingService{
		config: &config.KnowledgeConfig{
			AllowedFileExtensions: []string{".html", ".htm", ".md", ".markdown", ".txt", ".pdf"},
		},
	}

	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "HTML file should be allowed",
			url:      "https://example.com/page.html",
			expected: true,
		},
		{
			name:     "HTM file should be allowed",
			url:      "https://example.com/index.htm",
			expected: true,
		},
		{
			name:     "Markdown file should be allowed",
			url:      "https://docs.example.com/guide.md",
			expected: true,
		},
		{
			name:     "Markdown file with .markdown extension should be allowed",
			url:      "https://example.com/README.markdown",
			expected: true,
		},
		{
			name:     "PDF file should be allowed",
			url:      "https://example.com/document.pdf",
			expected: true,
		},
		{
			name:     "TXT file should be allowed",
			url:      "https://example.com/notes.txt",
			expected: true,
		},
		{
			name:     "URL without extension should be allowed",
			url:      "https://example.com/about",
			expected: true,
		},
		{
			name:     "URL with path segments and no extension should be allowed",
			url:      "https://example.com/blog/my-post",
			expected: true,
		},
		{
			name:     "Root URL should be allowed",
			url:      "https://example.com/",
			expected: true,
		},
		{
			name:     "Root domain should be allowed",
			url:      "https://example.com",
			expected: true,
		},
		{
			name:     "JavaScript file should be rejected",
			url:      "https://example.com/script.js",
			expected: false,
		},
		{
			name:     "CSS file should be rejected",
			url:      "https://example.com/styles.css",
			expected: false,
		},
		{
			name:     "Image file (JPG) should be rejected",
			url:      "https://example.com/photo.jpg",
			expected: false,
		},
		{
			name:     "Image file (PNG) should be rejected",
			url:      "https://example.com/logo.png",
			expected: false,
		},
		{
			name:     "Image file (GIF) should be rejected",
			url:      "https://example.com/animation.gif",
			expected: false,
		},
		{
			name:     "SVG file should be rejected",
			url:      "https://example.com/icon.svg",
			expected: false,
		},
		{
			name:     "Custom extension (.suman) should be rejected",
			url:      "https://docs.penify.dev/docs/guide.suman",
			expected: false,
		},
		{
			name:     "JSON file should be rejected",
			url:      "https://api.example.com/data.json",
			expected: false,
		},
		{
			name:     "XML file should be rejected",
			url:      "https://example.com/sitemap.xml",
			expected: false,
		},
		{
			name:     "ZIP file should be rejected",
			url:      "https://example.com/archive.zip",
			expected: false,
		},
		{
			name:     "Case-insensitive extension check (HTML uppercase)",
			url:      "https://example.com/page.HTML",
			expected: true,
		},
		{
			name:     "Case-insensitive extension check (JS uppercase)",
			url:      "https://example.com/script.JS",
			expected: false,
		},
		{
			name:     "URL with query parameters and allowed extension",
			url:      "https://example.com/page.html?ref=abc",
			expected: true,
		},
		{
			name:     "URL with fragment and allowed extension",
			url:      "https://example.com/page.html#section",
			expected: true,
		},
		{
			name:     "URL with dots in path but no extension",
			url:      "https://example.com/v1.2/api",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.shouldCrawlURLByExtension(tt.url)
			if result != tt.expected {
				t.Errorf("shouldCrawlURLByExtension(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestShouldCrawlURLByExtension_CustomConfig(t *testing.T) {
	// Test with custom allowed extensions
	s := &WebScrapingService{
		config: &config.KnowledgeConfig{
			AllowedFileExtensions: []string{".json", ".xml"}, // Only allow JSON and XML
		},
	}

	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "JSON file should be allowed with custom config",
			url:      "https://api.example.com/data.json",
			expected: true,
		},
		{
			name:     "XML file should be allowed with custom config",
			url:      "https://example.com/feed.xml",
			expected: true,
		},
		{
			name:     "HTML file should be rejected with custom config",
			url:      "https://example.com/page.html",
			expected: false,
		},
		{
			name:     "No extension should still be allowed",
			url:      "https://example.com/api/endpoint",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.shouldCrawlURLByExtension(tt.url)
			if result != tt.expected {
				t.Errorf("shouldCrawlURLByExtension(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestShouldCrawlURLByExtension_EmptyConfig(t *testing.T) {
	// Test with empty allowed extensions list
	s := &WebScrapingService{
		config: &config.KnowledgeConfig{
			AllowedFileExtensions: []string{},
		},
	}

	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "HTML file should be rejected with empty config",
			url:      "https://example.com/page.html",
			expected: false,
		},
		{
			name:     "No extension should still be allowed",
			url:      "https://example.com/about",
			expected: true,
		},
		{
			name:     "Root should be allowed",
			url:      "https://example.com/",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.shouldCrawlURLByExtension(tt.url)
			if result != tt.expected {
				t.Errorf("shouldCrawlURLByExtension(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestShouldCrawlURLByExtension_EdgeCases(t *testing.T) {
	s := &WebScrapingService{
		config: &config.KnowledgeConfig{
			AllowedFileExtensions: []string{".html", ".md"},
		},
	}

	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "URL with trailing slash after extension",
			url:      "https://example.com/page.html/",
			expected: true, // Treated as directory
		},
		{
			name:     "File in subdirectory with allowed extension",
			url:      "https://example.com/docs/guide.md",
			expected: true,
		},
		{
			name:     "File in subdirectory with disallowed extension",
			url:      "https://example.com/assets/image.jpg",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.shouldCrawlURLByExtension(tt.url)
			if result != tt.expected {
				t.Errorf("shouldCrawlURLByExtension(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}
