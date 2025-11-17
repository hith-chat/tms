package service

import (
	"net/url"
	"strings"
)

// NormalizeURL normalizes a URL by:
// 1. Removing all query parameters
// 2. Removing the fragment (hash)
// 3. Converting to lowercase
// 4. Trimming trailing slashes (except for root paths)
// This ensures URL uniqueness across the database
func NormalizeURL(rawURL string) (string, error) {
	// Parse the URL
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Remove query parameters and fragment
	parsed.RawQuery = ""
	parsed.Fragment = ""

	// Convert scheme and host to lowercase
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)

	// Get the normalized URL string
	normalized := parsed.String()

	// Trim trailing slash (except for root path like https://example.com/)
	if len(parsed.Path) > 1 && strings.HasSuffix(normalized, "/") {
		normalized = strings.TrimSuffix(normalized, "/")
	}

	return normalized, nil
}

// MustNormalizeURL normalizes a URL and panics if it fails
// Use this only when you're certain the URL is valid
func MustNormalizeURL(rawURL string) string {
	normalized, err := NormalizeURL(rawURL)
	if err != nil {
		// If normalization fails, return the original URL trimmed
		return strings.TrimSpace(rawURL)
	}
	return normalized
}
