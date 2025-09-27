package service

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ComprehensiveURLExtractor handles comprehensive URL extraction from web pages
type ComprehensiveURLExtractor struct {
	baseURL     *url.URL
	visitedURLs map[string]bool
	urlPatterns []*regexp.Regexp
}

// NewComprehensiveURLExtractor creates a new comprehensive URL extractor
func NewComprehensiveURLExtractor(baseURL string) (*ComprehensiveURLExtractor, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// Compile URL patterns for text extraction
	urlPatterns := []*regexp.Regexp{
		// HTTP/HTTPS URLs
		regexp.MustCompile(`https?://[^\s<>"']+[^\s<>"'.,;!?]`),
		// www. URLs without protocol
		regexp.MustCompile(`www\.[^\s<>"']+[^\s<>"'.,;!?]`),
		// Domain-like patterns (basic heuristic)
		regexp.MustCompile(`[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.([a-zA-Z]{2,}|[a-zA-Z]{2,}\.[a-zA-Z]{2,})`),
	}

	return &ComprehensiveURLExtractor{
		baseURL:     parsedURL,
		visitedURLs: make(map[string]bool),
		urlPatterns: urlPatterns,
	}, nil
}

// ExtractURLsFromHTML extracts URLs from HTML content using multiple methods
func (e *ComprehensiveURLExtractor) ExtractURLsFromHTML(htmlContent []byte, currentURL string) []string {
	var urls []string
	uniqueURLs := make(map[string]bool)

	// Parse HTML document
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlContent))
	if err != nil {
		return urls
	}

	// Method 1: Extract from standard HTML elements
	e.extractFromHTMLElements(doc, currentURL, uniqueURLs)

	// Method 2: Extract from inline styles and CSS
	e.extractFromStyles(doc, currentURL, uniqueURLs)

	// Method 3: Extract from JavaScript and data attributes
	e.extractFromScriptsAndData(doc, currentURL, uniqueURLs)

	// Method 4: Extract from text content using regex
	e.extractFromTextContent(doc, currentURL, uniqueURLs)

	// Method 5: Extract from meta tags and headers
	e.extractFromMetaTags(doc, currentURL, uniqueURLs)

	// Convert map to slice
	for urlStr := range uniqueURLs {
		if e.isValidURL(urlStr) && !e.visitedURLs[urlStr] {
			urls = append(urls, urlStr)
		}
	}

	return urls
}

// extractFromHTMLElements extracts URLs from standard HTML elements
func (e *ComprehensiveURLExtractor) extractFromHTMLElements(doc *goquery.Document, currentURL string, uniqueURLs map[string]bool) {
	// Links
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			if absoluteURL := e.resolveURL(href, currentURL); absoluteURL != "" {
				uniqueURLs[absoluteURL] = true
			}
		}
	})

	// Images
	doc.Find("img[src]").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			if absoluteURL := e.resolveURL(src, currentURL); absoluteURL != "" {
				uniqueURLs[absoluteURL] = true
			}
		}
	})

	// Image srcset attribute
	doc.Find("img[srcset]").Each(func(i int, s *goquery.Selection) {
		if srcset, exists := s.Attr("srcset"); exists {
			e.extractFromSrcset(srcset, currentURL, uniqueURLs)
		}
	})

	// Picture sources
	doc.Find("source[src], source[srcset]").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			if absoluteURL := e.resolveURL(src, currentURL); absoluteURL != "" {
				uniqueURLs[absoluteURL] = true
			}
		}
		if srcset, exists := s.Attr("srcset"); exists {
			e.extractFromSrcset(srcset, currentURL, uniqueURLs)
		}
	})

	// Stylesheets
	doc.Find("link[rel='stylesheet'][href]").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			if absoluteURL := e.resolveURL(href, currentURL); absoluteURL != "" {
				uniqueURLs[absoluteURL] = true
			}
		}
	})

	// Scripts
	doc.Find("script[src]").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			if absoluteURL := e.resolveURL(src, currentURL); absoluteURL != "" {
				uniqueURLs[absoluteURL] = true
			}
		}
	})

	// Iframes
	doc.Find("iframe[src]").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			if absoluteURL := e.resolveURL(src, currentURL); absoluteURL != "" {
				uniqueURLs[absoluteURL] = true
			}
		}
	})

	// Audio and Video sources
	doc.Find("audio[src], video[src]").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			if absoluteURL := e.resolveURL(src, currentURL); absoluteURL != "" {
				uniqueURLs[absoluteURL] = true
			}
		}
	})

	// Source elements (for audio/video)
	doc.Find("source[src]").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			if absoluteURL := e.resolveURL(src, currentURL); absoluteURL != "" {
				uniqueURLs[absoluteURL] = true
			}
		}
	})

	// Track elements
	doc.Find("track[src]").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			if absoluteURL := e.resolveURL(src, currentURL); absoluteURL != "" {
				uniqueURLs[absoluteURL] = true
			}
		}
	})

	// Object and embed elements
	doc.Find("object[data]").Each(func(i int, s *goquery.Selection) {
		if data, exists := s.Attr("data"); exists {
			if absoluteURL := e.resolveURL(data, currentURL); absoluteURL != "" {
				uniqueURLs[absoluteURL] = true
			}
		}
	})

	doc.Find("embed[src]").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			if absoluteURL := e.resolveURL(src, currentURL); absoluteURL != "" {
				uniqueURLs[absoluteURL] = true
			}
		}
	})

	// Form actions
	doc.Find("form[action]").Each(func(i int, s *goquery.Selection) {
		if action, exists := s.Attr("action"); exists {
			if absoluteURL := e.resolveURL(action, currentURL); absoluteURL != "" {
				uniqueURLs[absoluteURL] = true
			}
		}
	})
}

// extractFromSrcset extracts URLs from srcset attributes
func (e *ComprehensiveURLExtractor) extractFromSrcset(srcset, currentURL string, uniqueURLs map[string]bool) {
	// Srcset format: "url1 1x, url2 2x" or "url1 100w, url2 200w"
	parts := strings.Split(srcset, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		// Extract URL (everything before the last space)
		spaceIndex := strings.LastIndex(part, " ")
		var urlPart string
		if spaceIndex != -1 {
			urlPart = strings.TrimSpace(part[:spaceIndex])
		} else {
			urlPart = part
		}

		if absoluteURL := e.resolveURL(urlPart, currentURL); absoluteURL != "" {
			uniqueURLs[absoluteURL] = true
		}
	}
}

// extractFromStyles extracts URLs from CSS styles
func (e *ComprehensiveURLExtractor) extractFromStyles(doc *goquery.Document, currentURL string, uniqueURLs map[string]bool) {
	// Inline styles
	doc.Find("*[style]").Each(func(i int, s *goquery.Selection) {
		if style, exists := s.Attr("style"); exists {
			e.extractURLsFromCSS(style, currentURL, uniqueURLs)
		}
	})

	// Style tags
	doc.Find("style").Each(func(i int, s *goquery.Selection) {
		css := s.Text()
		e.extractURLsFromCSS(css, currentURL, uniqueURLs)
	})
}

// extractURLsFromCSS extracts URLs from CSS content
func (e *ComprehensiveURLExtractor) extractURLsFromCSS(css, currentURL string, uniqueURLs map[string]bool) {
	// Match url() functions in CSS
	urlRegex := regexp.MustCompile(`url\s*\(\s*['"]?([^'")\s]+)['"]?\s*\)`)
	matches := urlRegex.FindAllStringSubmatch(css, -1)

	for _, match := range matches {
		if len(match) > 1 {
			if absoluteURL := e.resolveURL(match[1], currentURL); absoluteURL != "" {
				uniqueURLs[absoluteURL] = true
			}
		}
	}
}

// extractFromScriptsAndData extracts URLs from JavaScript and data attributes
func (e *ComprehensiveURLExtractor) extractFromScriptsAndData(doc *goquery.Document, currentURL string, uniqueURLs map[string]bool) {
	// Script content
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		scriptContent := s.Text()
		e.extractURLsFromText(scriptContent, currentURL, uniqueURLs)
	})

	// Data attributes
	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		// Check all attributes for data-* and other common URL attributes
		attrs := []string{"data-src", "data-url", "data-href", "data-link", "data-background",
			"data-image", "data-video", "data-audio", "data-file", "data-download"}

		for _, attr := range attrs {
			if value, exists := s.Attr(attr); exists {
				if absoluteURL := e.resolveURL(value, currentURL); absoluteURL != "" {
					uniqueURLs[absoluteURL] = true
				}
			}
		}
	})
}

// extractFromTextContent extracts URLs from plain text using regex
func (e *ComprehensiveURLExtractor) extractFromTextContent(doc *goquery.Document, currentURL string, uniqueURLs map[string]bool) {
	// Get all text content
	textContent := doc.Text()
	e.extractURLsFromText(textContent, currentURL, uniqueURLs)
}

// extractURLsFromText extracts URLs from text using regex patterns
func (e *ComprehensiveURLExtractor) extractURLsFromText(text, currentURL string, uniqueURLs map[string]bool) {
	for _, pattern := range e.urlPatterns {
		matches := pattern.FindAllString(text, -1)
		for _, match := range matches {
			// Clean up the match
			match = strings.TrimSpace(match)
			if absoluteURL := e.resolveURL(match, currentURL); absoluteURL != "" {
				uniqueURLs[absoluteURL] = true
			}
		}
	}
}

// extractFromMetaTags extracts URLs from meta tags
func (e *ComprehensiveURLExtractor) extractFromMetaTags(doc *goquery.Document, currentURL string, uniqueURLs map[string]bool) {
	// Common meta tags that contain URLs
	metaSelectors := []string{
		"meta[property='og:url']",
		"meta[property='og:image']",
		"meta[property='og:video']",
		"meta[property='og:audio']",
		"meta[name='twitter:image']",
		"meta[name='twitter:player']",
		"meta[property='article:author']",
		"meta[name='author']",
		"link[rel='canonical']",
		"link[rel='alternate']",
		"link[rel='next']",
		"link[rel='prev']",
		"link[rel='icon']",
		"link[rel='shortcut icon']",
		"link[rel='apple-touch-icon']",
	}

	for _, selector := range metaSelectors {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			var urlValue string
			if content, exists := s.Attr("content"); exists {
				urlValue = content
			} else if href, exists := s.Attr("href"); exists {
				urlValue = href
			}

			if urlValue != "" {
				if absoluteURL := e.resolveURL(urlValue, currentURL); absoluteURL != "" {
					uniqueURLs[absoluteURL] = true
				}
			}
		})
	}
}

// resolveURL resolves a potentially relative URL to an absolute URL
func (e *ComprehensiveURLExtractor) resolveURL(rawURL, currentURL string) string {
	if rawURL == "" {
		return ""
	}

	// Parse the current URL
	currentParsed, err := url.Parse(currentURL)
	if err != nil {
		return ""
	}

	// Parse the raw URL
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	// Resolve against current URL
	resolved := currentParsed.ResolveReference(parsed)

	// Normalize the URL
	return e.normalizeURL(resolved.String())
}

// normalizeURL normalizes a URL for consistency
func (e *ComprehensiveURLExtractor) normalizeURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Remove fragment
	parsed.Fragment = ""

	// Remove trailing slash from path (except for root)
	if parsed.Path != "/" {
		parsed.Path = strings.TrimSuffix(parsed.Path, "/")
	}

	return parsed.String()
}

// isValidURL checks if a URL is valid and should be included
func (e *ComprehensiveURLExtractor) isValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// Must have a scheme and host
	if parsed.Scheme == "" || parsed.Host == "" {
		return false
	}

	// Only HTTP/HTTPS
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}

	// Skip common non-content URLs
	skipPatterns := []string{
		"javascript:",
		"mailto:",
		"tel:",
		"#",
		".css",
		".js",
		".ico",
		".png",
		".jpg",
		".jpeg",
		".gif",
		".svg",
		".webp",
		".pdf",
		".zip",
		".rar",
		".exe",
		".dmg",
	}

	lowercaseURL := strings.ToLower(rawURL)
	for _, pattern := range skipPatterns {
		if strings.Contains(lowercaseURL, pattern) {
			return false
		}
	}

	return true
}
