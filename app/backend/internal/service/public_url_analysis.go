package service

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// PublicURLAnalysisService handles public URL analysis operations
type PublicURLAnalysisService struct {
	webScrapingService *WebScrapingService
}

// NewPublicURLAnalysisService creates a new public URL analysis service
func NewPublicURLAnalysisService(webScrapingService *WebScrapingService) *PublicURLAnalysisService {
	return &PublicURLAnalysisService{
		webScrapingService: webScrapingService,
	}
}

// URLAnalysisRequest represents the request for URL analysis
type URLAnalysisRequest struct {
	URL      string `json:"url" validate:"required,url"`
	MaxDepth int    `json:"max_depth,omitempty"`
}

// URLAnalysisEvent represents a streaming event during URL analysis
type URLAnalysisEvent struct {
	Type       string    `json:"type"`                  // "progress", "url_found", "completed", "error"
	Message    string    `json:"message,omitempty"`     // Human readable message
	URL        string    `json:"url,omitempty"`         // Current URL being processed
	Title      string    `json:"title,omitempty"`       // Page title
	Depth      int       `json:"depth,omitempty"`       // Current depth level
	TokenCount int       `json:"token_count,omitempty"` // Token count for this URL
	Total      int       `json:"total,omitempty"`       // Total URLs found so far
	Timestamp  time.Time `json:"timestamp"`             // Event timestamp
}

// URLAnalysisResult represents the final result of URL analysis
type URLAnalysisResult struct {
	RootURL     string        `json:"root_url"`
	MaxDepth    int           `json:"max_depth"`
	TotalURLs   int           `json:"total_urls"`
	TotalTokens int           `json:"total_tokens"`
	URLs        []AnalyzedURL `json:"urls"`
	GeneratedAt time.Time     `json:"generated_at"`
}

// AnalyzedURL represents a single analyzed URL
type AnalyzedURL struct {
	URL        string `json:"url"`
	Title      string `json:"title,omitempty"`
	Depth      int    `json:"depth"`
	TokenCount int    `json:"token_count"`
}

// AnalyzeURLWithStream analyzes a URL and streams progress
func (s *PublicURLAnalysisService) AnalyzeURLWithStream(ctx context.Context, req URLAnalysisRequest, events chan<- URLAnalysisEvent) (*URLAnalysisResult, error) {
	// Validate URL
	if err := s.validateURL(req.URL); err != nil {
		s.sendEvent(ctx, events, URLAnalysisEvent{
			Type:      "error",
			Message:   fmt.Sprintf("Invalid URL: %v", err),
			Timestamp: time.Now(),
		})
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Set default max depth
	maxDepth := req.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 1
	}

	// Send initial event
	s.sendEvent(ctx, events, URLAnalysisEvent{
		Type:      "progress",
		Message:   fmt.Sprintf("Starting URL analysis for %s (max depth: %d)", req.URL, maxDepth),
		URL:       req.URL,
		Timestamp: time.Now(),
	})

	// Extract URLs manually using existing web scraping service
	discoveredLinks, err := s.webScrapingService.extractURLsManually(ctx, req.URL, maxDepth)
	if err != nil {
		s.sendEvent(ctx, events, URLAnalysisEvent{
			Type:      "error",
			Message:   fmt.Sprintf("Failed to analyze URL: %v", err),
			Timestamp: time.Now(),
		})
		return nil, fmt.Errorf("failed to analyze URL: %w", err)
	}

	// Process and stream results
	var analyzedURLs []AnalyzedURL
	totalTokens := 0

	for i, link := range discoveredLinks {
		analyzedURL := AnalyzedURL{
			URL:        link.URL,
			Title:      link.Title,
			Depth:      link.Depth,
			TokenCount: link.TokenCount,
		}
		analyzedURLs = append(analyzedURLs, analyzedURL)
		totalTokens += link.TokenCount

		// Send progress event for each URL found
		s.sendEvent(ctx, events, URLAnalysisEvent{
			Type:       "url_found",
			Message:    fmt.Sprintf("Found URL at depth %d: %s", link.Depth, link.Title),
			URL:        link.URL,
			Title:      link.Title,
			Depth:      link.Depth,
			TokenCount: link.TokenCount,
			Total:      i + 1,
			Timestamp:  time.Now(),
		})
	}

	// Create final result
	result := &URLAnalysisResult{
		RootURL:     req.URL,
		MaxDepth:    maxDepth,
		TotalURLs:   len(analyzedURLs),
		TotalTokens: totalTokens,
		URLs:        analyzedURLs,
		GeneratedAt: time.Now(),
	}

	// Send completion event
	s.sendEvent(ctx, events, URLAnalysisEvent{
		Type:      "completed",
		Message:   fmt.Sprintf("Analysis completed. Found %d URLs with %d total tokens", result.TotalURLs, result.TotalTokens),
		Total:     result.TotalURLs,
		Timestamp: time.Now(),
	})

	return result, nil
}

// validateURL validates the provided URL
func (s *PublicURLAnalysisService) validateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}

	return nil
}

// sendEvent sends an event to the channel safely
func (s *PublicURLAnalysisService) sendEvent(ctx context.Context, events chan<- URLAnalysisEvent, event URLAnalysisEvent) {
	select {
	case events <- event:
		// Event sent successfully
	case <-ctx.Done():
		// Context cancelled, stop sending
		return
	default:
		// Channel full or closed, skip this event
		return
	}
}
