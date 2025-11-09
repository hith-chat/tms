package service

import (
	"testing"
)

func TestNormalizeURLForDeduplication(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "www prefix with uppercase and query params",
			input:    "https://WWW.Example.COM/Contact-Us/?ref=A#top",
			expected: "https://example.com/contact-us",
		},
		{
			name:     "www prefix with trailing slash",
			input:    "http://www.example.com/about/",
			expected: "http://example.com/about",
		},
		{
			name:     "already normalized URL",
			input:    "https://example.com/page",
			expected: "https://example.com/page",
		},
		{
			name:     "root path with www",
			input:    "https://www.example.com/",
			expected: "https://example.com/",
		},
		{
			name:     "www with multiple paths",
			input:    "https://WWW.EXAMPLE.COM/Products/Category/",
			expected: "https://example.com/products/category",
		},
		{
			name:     "subdomain other than www (should keep)",
			input:    "https://docs.example.com/API/",
			expected: "https://docs.example.com/api",
		},
		{
			name:     "www1 subdomain (should keep)",
			input:    "https://www1.example.com/page",
			expected: "https://www1.example.com/page",
		},
		{
			name:     "no www, with query and fragment",
			input:    "https://example.com/search?q=test#results",
			expected: "https://example.com/search",
		},
		{
			name:     "uppercase path without www",
			input:    "https://EXAMPLE.COM/ABOUT/US/",
			expected: "https://example.com/about/us",
		},
		{
			name:     "port with www",
			input:    "http://www.example.com:8080/api",
			expected: "http://example.com:8080/api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeURLForDeduplication(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeURLForDeduplication(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeURLForDeduplication_SameNormalization(t *testing.T) {
	// Test that URLs that should be treated as duplicates normalize to the same value
	duplicateSets := [][]string{
		{
			"https://www.example.com/contact",
			"https://example.com/contact",
			"https://WWW.EXAMPLE.COM/CONTACT",
			"https://www.example.com/contact/",
			"https://example.com/Contact?utm_source=google#section",
		},
		{
			"http://www.example.com/",
			"http://example.com/",
			"http://WWW.EXAMPLE.COM/",
		},
		{
			"https://www.example.com/products/item",
			"https://example.com/products/item/",
			"https://WWW.Example.Com/Products/Item/?ref=123",
		},
	}

	for i, set := range duplicateSets {
		normalized := make(map[string]bool)
		for _, url := range set {
			result := normalizeURLForDeduplication(url)
			normalized[result] = true
		}

		if len(normalized) != 1 {
			t.Errorf("Set %d: Expected all URLs to normalize to the same value, got %d different values: %v", i, len(normalized), normalized)
			for _, url := range set {
				t.Logf("  %s -> %s", url, normalizeURLForDeduplication(url))
			}
		}
	}
}

func TestNormalizeURLForDeduplication_DifferentURLs(t *testing.T) {
	// Test that URLs that should be treated as different don't normalize to the same value
	differentURLs := []string{
		"https://example.com/about",
		"https://example.com/contact",
		"https://docs.example.com/about",     // Different subdomain
		"https://www1.example.com/about",    // www1 is kept
		"https://example.com/about/team",    // Different path
	}

	normalized := make(map[string]bool)
	for _, url := range differentURLs {
		result := normalizeURLForDeduplication(url)
		if normalized[result] {
			t.Errorf("URL %s normalized to %s which was already seen - should be unique", url, result)
		}
		normalized[result] = true
	}

	if len(normalized) != len(differentURLs) {
		t.Errorf("Expected %d unique normalized URLs, got %d", len(differentURLs), len(normalized))
	}
}
