package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/logger"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
)

type WebScrapingService struct {
	knowledgeRepo            *repo.KnowledgeRepository
	embeddingService         *EmbeddingService
	config                   *config.KnowledgeConfig
	redisClient              redis.UniversalClient
	headlessBrowserExtractor *HeadlessBrowserURLExtractor
}

const (
	stagingDirName     = "tms-webscrape"
	maxSelectableLinks = 10
)

const MaxSelectableLinks = maxSelectableLinks

// discoveredLink represents a link discovered during scraping
type discoveredLink struct {
	URL        string `json:"url"`
	Title      string `json:"title,omitempty"`
	Depth      int    `json:"depth"`
	TokenCount int    `json:"token_count"`
}

// URLExtractionRequest represents the request payload for yourgpt.ai URL extraction API
type URLExtractionRequest struct {
	URL string `json:"url"`
}

// URLExtractionResponse represents the response from yourgpt.ai URL extraction API
type URLExtractionResponse struct {
	URLs []string `json:"urls"`
}

type linkDiscoveryResult struct {
	JobID       uuid.UUID        `json:"job_id"`
	RootURL     string           `json:"root_url"`
	MaxDepth    int              `json:"max_depth"`
	GeneratedAt time.Time        `json:"generated_at"`
	Links       []discoveredLink `json:"links"`
}

// IndexingEvent represents streamed progress updates for the indexing phase
type IndexingEvent struct {
	Type        string    `json:"type"`
	Message     string    `json:"message,omitempty"`
	URL         string    `json:"url,omitempty"`
	Total       int       `json:"total,omitempty"`
	Completed   int       `json:"completed,omitempty"`
	Pending     int       `json:"pending,omitempty"`
	TokenCount  int       `json:"token_count,omitempty"`
	TotalTokens int       `json:"total_tokens,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// ScrapingEvent represents streamed progress updates for the scraping phase
type ScrapingEvent struct {
	Type         string    `json:"type"`
	JobID        uuid.UUID `json:"job_id"`
	Message      string    `json:"message,omitempty"`
	URL          string    `json:"url,omitempty"`
	CurrentDepth int       `json:"current_depth,omitempty"`
	MaxDepth     int       `json:"max_depth,omitempty"`
	LinksFound   int       `json:"links_found,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

func NewWebScrapingService(knowledgeRepo *repo.KnowledgeRepository, embeddingService *EmbeddingService, cfg *config.KnowledgeConfig, redisClient redis.UniversalClient) *WebScrapingService {
	// Initialize headless browser extractor with 30 second timeout
	headlessExtractor := NewHeadlessBrowserURLExtractor(30*time.Second, "")

	return &WebScrapingService{
		knowledgeRepo:            knowledgeRepo,
		embeddingService:         embeddingService,
		config:                   cfg,
		redisClient:              redisClient,
		headlessBrowserExtractor: headlessExtractor,
	}
}

// CreateScrapingJob creates a new web scraping job
func (s *WebScrapingService) CreateScrapingJob(ctx context.Context, tenantID, projectID uuid.UUID, req *models.CreateScrapingJobRequest) (*models.KnowledgeScrapingJob, error) {
	// Validate URL
	if err := s.validateURL(req.URL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Create scraping job
	job := &models.KnowledgeScrapingJob{
		ID:        uuid.New(),
		TenantID:  tenantID,
		ProjectID: projectID,
		URL:       req.URL,
		MaxDepth:  req.MaxDepth,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database
	if err := s.knowledgeRepo.CreateScrapingJob(job); err != nil {
		return nil, fmt.Errorf("failed to create scraping job: %w", err)
	}

	// Start scraping asynchronously
	go s.startScrapingJob(ctx, job)

	return job, nil
}

// CreateScrapingJobWithStream creates a new web scraping job and streams progress
func (s *WebScrapingService) CreateScrapingJobWithStream(ctx context.Context, tenantID, projectID uuid.UUID, req *models.CreateScrapingJobRequest, events chan<- ScrapingEvent) (*models.KnowledgeScrapingJob, error) {
	// Validate URL
	if err := s.validateURL(req.URL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Create scraping job
	job := &models.KnowledgeScrapingJob{
		ID:        uuid.New(),
		TenantID:  tenantID,
		ProjectID: projectID,
		URL:       req.URL,
		MaxDepth:  req.MaxDepth,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database
	if err := s.knowledgeRepo.CreateScrapingJob(job); err != nil {
		return nil, fmt.Errorf("failed to create scraping job: %w", err)
	}

	// Start scraping with streaming
	go s.startScrapingJobWithStream(ctx, job, events)

	return job, nil
}

// validateURL validates the URL for scraping
func (s *WebScrapingService) validateURL(rawURL string) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("only HTTP and HTTPS URLs are supported")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}

	// Block localhost and internal IPs for security
	if strings.Contains(parsedURL.Host, "localhost") ||
		strings.Contains(parsedURL.Host, "127.0.0.1") ||
		strings.Contains(parsedURL.Host, "0.0.0.0") ||
		strings.Contains(parsedURL.Host, "::1") ||
		strings.HasPrefix(parsedURL.Host, "10.") ||
		strings.HasPrefix(parsedURL.Host, "192.168.") ||
		strings.HasPrefix(parsedURL.Host, "172.") {
		return fmt.Errorf("internal and localhost URLs are not allowed")
	}

	return nil
}

// extractURLsFromAPI attempts to extract URLs using yourgpt.ai API
func (s *WebScrapingService) extractURLsFromAPI(ctx context.Context, targetURL string) ([]string, error) {
	requestPayload := URLExtractionRequest{URL: targetURL}
	jsonData, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://yourgpt.ai/api/extractUrls", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers as provided in the curl command
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	req.Header.Set("Origin", "https://yourgpt.ai")
	req.Header.Set("Referer", "https://yourgpt.ai/tools/url-extractor")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response URLExtractionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.URLs, nil
}

// extractURLsManually extracts URLs using multiple fallback strategies
// Priority order: 1. Playwright (headless browser), 2. YourGPT API, 3. Comprehensive (Colly)
func (s *WebScrapingService) extractURLsManually(ctx context.Context, targetURL string, maxDepth int) ([]discoveredLink, error) {
	logger.InfofCtx(ctx, "Starting URL extraction - target: %s, max_depth: %d", targetURL, maxDepth)

	// Step 1: Try Playwright headless browser extraction first
	logger.DebugCtx(ctx, "Step 1: Attempting Playwright headless browser extraction")
	headlessLinks, err := s.extractURLsWithHeadlessBrowser(ctx, targetURL, maxDepth)

	logger.InfofCtx(ctx, "Playwright headless browser found %d links", len(headlessLinks))

	if err == nil && len(headlessLinks) > 0 {
		logger.InfofCtx(ctx, "Playwright extraction successful - found %d links", len(headlessLinks))
		return headlessLinks, nil
	}

	// Log Playwright failure reason
	if err != nil {
		if strings.Contains(err.Error(), "please install the driver") {
			logger.WarnCtx(ctx, "Playwright not installed, trying next method")
		} else {
			logger.WarnfCtx(ctx, "Playwright extraction failed: %v - trying next method", err)
		}
	}

	// Step 2: Try YourGPT API extraction as second fallback
	logger.InfoCtx(ctx, "Step 2: Attempting YourGPT API extraction")
	apiURLs, apiErr := s.extractURLsFromAPI(ctx, targetURL)

	if apiErr == nil && len(apiURLs) > 0 {
		logger.InfofCtx(ctx, "YourGPT API extraction successful - found %d URLs", len(apiURLs))

		// Convert API URLs to discoveredLink format
		var apiLinks []discoveredLink
		for _, urlStr := range apiURLs {
			apiLinks = append(apiLinks, discoveredLink{
				URL:        urlStr,
				Title:      "",
				Depth:      0,
				TokenCount: 0, // Will be estimated later if needed
			})
		}

		return apiLinks, nil
	}

	if apiErr != nil {
		logger.WarnfCtx(ctx, "YourGPT API extraction failed: %v - trying final fallback", apiErr)
	}

	// Step 3: Final fallback to comprehensive extraction (Colly)
	logger.InfoCtx(ctx, "Step 3: Falling back to comprehensive extraction (Colly)")
	return s.extractURLsComprehensively(ctx, targetURL, maxDepth)
}

// extractURLsWithHeadlessBrowser extracts URLs using a comprehensive HTTP-based approach
func (s *WebScrapingService) extractURLsWithHeadlessBrowser(ctx context.Context, targetURL string, maxDepth int) ([]discoveredLink, error) {
	txLogger := logger.GetTxLogger(ctx).With().
		Str("component", "web_scraper").
		Str("operation", "extract_urls_headless").
		Str("target_url", targetURL).
		Int("max_depth", maxDepth).
		Logger()

	txLogger.Info().Msg("Starting headless browser extraction")

	// Try headless browser extraction
	var allLinks []discoveredLink
	visited := make(map[string]bool)
	toVisit := []struct {
		url   string
		depth int
	}{{targetURL, 0}}

	for len(toVisit) > 0 && len(allLinks) < 10 {
		current := toVisit[0]
		toVisit = toVisit[1:]

		if current.depth > maxDepth || visited[current.url] {
			continue
		}

		visited[current.url] = true

		// Extract URLs from current page using headless browser
		pageLogger := txLogger.With().
			Str("current_url", current.url).
			Int("depth", current.depth).
			Logger()

		pageLogger.Debug().Msg("Extracting URLs from page")
		extractedURLs, err := s.headlessBrowserExtractor.ExtractURLsFromPage(ctx, current.url)
		if err != nil {
			pageLogger.Error().
				Err(err).
				Msg("Failed to extract URLs from page")
			continue
		}

		// Get page title and content for token estimation
		title, _ := s.headlessBrowserExtractor.GetPageTitle(ctx, current.url)
		content, _ := s.headlessBrowserExtractor.GetPageContent(ctx, current.url)
		tokenCount := s.estimateTokenCount(content)

		// Add current page to results
		allLinks = append(allLinks, discoveredLink{
			URL:        current.url,
			Title:      title,
			Depth:      current.depth,
			TokenCount: tokenCount,
		})

		// Add discovered URLs for next depth level
		if current.depth < maxDepth {
			for _, urlInfo := range extractedURLs {
				if !visited[urlInfo.URL] && s.shouldFollowLink(urlInfo.URL, current.url) {
					toVisit = append(toVisit, struct {
						url   string
						depth int
					}{urlInfo.URL, current.depth + 1})
				}
			}
		}
	}

	return allLinks, nil
}

func (s *WebScrapingService) extractURLsComprehensively(ctx context.Context, targetURL string, maxDepth int) ([]discoveredLink, error) {
	// Initialize the comprehensive URL extractor
	extractor, err := NewComprehensiveURLExtractor(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL extractor: %w", err)
	}

	c := colly.NewCollector(
		colly.UserAgent(s.config.ScrapeUserAgent),
	)

	c.SetRequestTimeout(s.config.ScrapeTimeout)
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2, // Conservative parallelism for comprehensive extraction
		Delay:       s.config.ScrapeRateLimit,
	})
	c.AllowURLRevisit = false

	var discoveredLinks []discoveredLink
	visitedURLs := map[string]bool{targetURL: true}
	maxLinks := 500 // Increased limit for comprehensive extraction

	// Process each page comprehensively
	c.OnHTML("html", func(e *colly.HTMLElement) {
		depth := e.Request.Depth
		if depth > maxDepth || len(discoveredLinks) >= maxLinks {
			return
		}

		currentURL := e.Request.URL.String()

		// Extract title
		title := e.ChildText("title")
		if title == "" {
			title = e.ChildText("h1")
		}
		title = strings.TrimSpace(title)

		// Extract text content and estimate tokens
		pageContent := s.extractTextContent(e)
		tokenCount := s.estimateTokenCount(pageContent)

		// Add current page to discovered links
		link := discoveredLink{
			URL:        currentURL,
			Title:      title,
			Depth:      depth,
			TokenCount: tokenCount,
		}
		discoveredLinks = append(discoveredLinks, link)

		// Use comprehensive URL extraction if we haven't reached max depth
		if depth < maxDepth && len(discoveredLinks) < maxLinks {
			// Get the HTML content as bytes
			htmlContent := []byte(e.Response.Body)

			// Extract URLs comprehensively
			extractedURLs := extractor.ExtractURLsFromHTML(htmlContent, currentURL)

			// Visit extracted URLs
			for _, extractedURL := range extractedURLs {
				if len(discoveredLinks) >= maxLinks {
					break
				}

				if s.shouldFollowLink(extractedURL, currentURL) && !visitedURLs[extractedURL] {
					visitedURLs[extractedURL] = true
					e.Request.Visit(extractedURL)
				}
			}
		}
	})

	// Error logging
	c.OnError(func(r *colly.Response, err error) {
		logger.GetTxLogger(ctx).Error().
			Str("component", "web_scraper").
			Str("operation", "colly_error").
			Str("url", r.Request.URL.String()).
			Err(err).
			Msg("Error visiting URL during comprehensive extraction")
	})

	// Start crawl
	if err := c.Visit(targetURL); err != nil {
		return nil, fmt.Errorf("failed to start comprehensive crawl: %w", err)
	}

	c.Wait()
	return discoveredLinks, nil
}

// startScrapingJobWithStream discovers links with real-time progress streaming
func (s *WebScrapingService) startScrapingJobWithStream(ctx context.Context, job *models.KnowledgeScrapingJob, events chan<- ScrapingEvent) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute) // Shorter timeout for link discovery
	defer cancel()
	defer close(events)

	var runErr error
	logger.InfofCtx(ctx, "Starting scraping job stream - jobID: %s, URL: %s", job.ID.String(), job.URL)

	defer func() {
		if runErr != nil {
			errStr := runErr.Error()
			logger.ErrorfCtx(ctx, runErr, "Scraping job failed: %v", runErr)

			s.sendScrapingEvent(ctx, events, ScrapingEvent{
				Type:      "error",
				JobID:     job.ID,
				Message:   errStr,
				Timestamp: time.Now(),
			})

			if updateErr := s.knowledgeRepo.UpdateScrapingJobStatus(job.ID, "failed", &errStr); updateErr != nil {
				logger.ErrorfCtx(ctx, updateErr, "Failed to update job status to failed: %v", updateErr)
			}
		}
	}()

	if runErr = s.knowledgeRepo.StartScrapingJob(job.ID); runErr != nil {
		return
	}

	s.sendScrapingEvent(ctx, events, ScrapingEvent{
		Type:      "started",
		JobID:     job.ID,
		Message:   fmt.Sprintf("Starting link discovery for %s", job.URL),
		URL:       job.URL,
		MaxDepth:  job.MaxDepth,
		Timestamp: time.Now(),
	})

	var discoveredLinks []discoveredLink

	// Try API-based URL extraction first
	s.sendScrapingEvent(ctx, events, ScrapingEvent{
		Type:      "info",
		JobID:     job.ID,
		Message:   "Attempting API-based URL extraction...",
		URL:       job.URL,
		Timestamp: time.Now(),
	})

	apiURLs, apiErr := s.extractURLsFromAPI(ctx, job.URL)
	if apiErr == nil && len(apiURLs) > 0 {
		s.sendScrapingEvent(ctx, events, ScrapingEvent{
			Type:       "info",
			JobID:      job.ID,
			Message:    fmt.Sprintf("API extracted %d URLs successfully", len(apiURLs)),
			URL:        job.URL,
			LinksFound: len(apiURLs),
			Timestamp:  time.Now(),
		})

		// Convert API URLs to discoveredLink format
		for i, url := range apiURLs {
			if i >= 100 { // Limit to avoid overwhelming
				break
			}
			if err := s.validateURL(url); err != nil {
				continue // Skip invalid URLs
			}

			link := discoveredLink{
				URL:        url,
				Title:      "", // Will be fetched during indexing
				Depth:      0,  // API URLs are at depth 0
				TokenCount: 0,  // Set to 0 as requested for API-extracted URLs
			}
			discoveredLinks = append(discoveredLinks, link)

			s.sendScrapingEvent(ctx, events, ScrapingEvent{
				Type:       "link_found",
				JobID:      job.ID,
				Message:    fmt.Sprintf("Found URL from API: %s", url),
				URL:        url,
				LinksFound: len(discoveredLinks),
				Timestamp:  time.Now(),
			})
		}
	} else {
		// Fallback to manual extraction
		s.sendScrapingEvent(ctx, events, ScrapingEvent{
			Type:      "warning",
			JobID:     job.ID,
			Message:   fmt.Sprintf("API extraction failed (%v), falling back to manual discovery...", apiErr),
			URL:       job.URL,
			Timestamp: time.Now(),
		})

		manualLinks, manualErr := s.extractURLsManually(ctx, job.URL, job.MaxDepth)
		if manualErr != nil {
			runErr = fmt.Errorf("both API and manual URL extraction failed: API error: %v, Manual error: %v", apiErr, manualErr)
			return
		}

		discoveredLinks = manualLinks
		for _, link := range discoveredLinks {
			s.sendScrapingEvent(ctx, events, ScrapingEvent{
				Type:         "link_found",
				JobID:        job.ID,
				Message:      fmt.Sprintf("Found link: %s", link.Title),
				URL:          link.URL,
				CurrentDepth: link.Depth,
				MaxDepth:     job.MaxDepth,
				LinksFound:   len(discoveredLinks),
				Timestamp:    time.Now(),
			})
		}
	}

	totalLinks := len(discoveredLinks)
	if totalLinks == 0 {
		runErr = fmt.Errorf("no links were discovered")
		return
	}

	// Store discovered links in Redis with 2-hour expiry
	redisKey, persistErr := s.storeDiscoveredLinksInRedis(ctx, job, discoveredLinks)
	if persistErr != nil {
		runErr = fmt.Errorf("failed to store discovered links: %w", persistErr)
		return
	}

	if err := s.knowledgeRepo.MarkJobAwaitingSelection(job.ID, totalLinks, redisKey); err != nil {
		runErr = fmt.Errorf("failed to mark job awaiting selection: %w", err)
		return
	}

	s.sendScrapingEvent(ctx, events, ScrapingEvent{
		Type:       "completed",
		JobID:      job.ID,
		Message:    fmt.Sprintf("Link discovery completed! Found %d links. Ready for review.", totalLinks),
		LinksFound: totalLinks,
		Timestamp:  time.Now(),
	})
	logger.InfofCtx(ctx, "Scraping job completed - jobID: %s, total_links: %d, redis_key: %s", job.ID.String(), totalLinks, redisKey)
	runErr = nil
}

// startScrapingJob executes the scraping job (legacy - discovers links like the streaming version)
func (s *WebScrapingService) startScrapingJob(ctx context.Context, job *models.KnowledgeScrapingJob) {
	// Simply call the link discovery without streaming for legacy support
	events := make(chan ScrapingEvent, 100) // Buffered channel to avoid blocking
	go func() {
		// Drain events to prevent blocking
		for range events {
		}
	}()

	s.startScrapingJobWithStream(ctx, job, events)
}

func (s *WebScrapingService) storeDiscoveredLinksInRedis(ctx context.Context, job *models.KnowledgeScrapingJob, links []discoveredLink) (string, error) {
	result := linkDiscoveryResult{
		JobID:       job.ID,
		RootURL:     job.URL,
		MaxDepth:    job.MaxDepth,
		GeneratedAt: time.Now(),
		Links:       links,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to serialize discovered links: %w", err)
	}

	redisKey := fmt.Sprintf("scraping:links:%s", job.ID)
	// Store in Redis with 2 hours expiration (short-lived)
	err = s.redisClient.Set(ctx, redisKey, data, 2*time.Hour).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store discovered links in Redis: %w", err)
	}

	return redisKey, nil
}

func (s *WebScrapingService) loadDiscoveredLinksFromRedis(ctx context.Context, redisKey string) (*linkDiscoveryResult, error) {
	data, err := s.redisClient.Get(ctx, redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("discovered links not found or expired")
		}
		return nil, fmt.Errorf("failed to read discovered links from Redis: %w", err)
	}

	var result linkDiscoveryResult
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, fmt.Errorf("failed to parse discovered links: %w", err)
	}

	return &result, nil
}

func (s *WebScrapingService) sendIndexingEvent(ctx context.Context, ch chan<- IndexingEvent, event IndexingEvent) {
	select {
	case <-ctx.Done():
		return
	case ch <- event:
		return
	}
}

func (s *WebScrapingService) sendScrapingEvent(ctx context.Context, ch chan<- ScrapingEvent, event ScrapingEvent) {
	select {
	case <-ctx.Done():
		return
	case ch <- event:
		return
	}
}

// GetStagedLinks returns the discovered links awaiting user confirmation
func (s *WebScrapingService) GetStagedLinks(ctx context.Context, jobID, tenantID, projectID uuid.UUID) ([]*models.ScrapedLinkPreview, error) {
	job, err := s.knowledgeRepo.GetScrapingJob(jobID, tenantID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to load scraping job: %w", err)
	}

	if job.Status != "awaiting_selection" {
		return nil, fmt.Errorf("job is not ready for selection (current status: %s)", job.Status)
	}

	if job.StagingFilePath == nil || *job.StagingFilePath == "" {
		return nil, fmt.Errorf("discovered links not available for this job yet")
	}

	discovered, err := s.loadDiscoveredLinksFromRedis(ctx, *job.StagingFilePath)
	if err != nil {
		return nil, err
	}

	selectedSet := make(map[string]struct{}, len(job.SelectedLinks))
	for _, url := range job.SelectedLinks {
		selectedSet[url] = struct{}{}
	}

	previews := make([]*models.ScrapedLinkPreview, 0, len(discovered.Links))
	for _, link := range discovered.Links {
		preview := &models.ScrapedLinkPreview{
			URL:            link.URL,
			Title:          link.Title,
			Depth:          link.Depth,
			TokenCount:     link.TokenCount,
			ContentPreview: "Content will be fetched during indexing",
		}
		if _, ok := selectedSet[link.URL]; ok {
			preview.Selected = true
		}
		previews = append(previews, preview)
	}

	sort.Slice(previews, func(i, j int) bool {
		if previews[i].Depth == previews[j].Depth {
			return previews[i].URL < previews[j].URL
		}
		return previews[i].Depth < previews[j].Depth
	})

	return previews, nil
}

// StoreLinkSelection saves the user-selected URLs that should proceed to indexing
func (s *WebScrapingService) StoreLinkSelection(ctx context.Context, jobID, tenantID, projectID uuid.UUID, urls []string) error {
	job, err := s.knowledgeRepo.GetScrapingJob(jobID, tenantID, projectID)
	if err != nil {
		return fmt.Errorf("failed to load scraping job: %w", err)
	}

	if job.Status != "awaiting_selection" {
		return fmt.Errorf("job is not ready for selection (current status: %s)", job.Status)
	}

	if job.StagingFilePath == nil || *job.StagingFilePath == "" {
		return fmt.Errorf("discovered links not available for this job")
	}

	if len(urls) == 0 {
		return fmt.Errorf("at least one URL must be selected for indexing")
	}

	// Load discovered links from Redis to validate selection
	discovered, err := s.loadDiscoveredLinksFromRedis(ctx, *job.StagingFilePath)
	if err != nil {
		return err
	}

	validURLs := make(map[string]struct{}, len(discovered.Links))
	for _, link := range discovered.Links {
		validURLs[link.URL] = struct{}{}
	}

	deduped := make([]string, 0, len(urls))
	seen := make(map[string]struct{})
	for _, raw := range urls {
		candidate := strings.TrimSpace(raw)
		if candidate == "" {
			continue
		}
		if _, ok := validURLs[candidate]; !ok {
			return fmt.Errorf("selected URL is not part of discovered links: %s", candidate)
		}
		if err := s.validateURL(candidate); err != nil {
			return fmt.Errorf("invalid URL selected (%s): %w", candidate, err)
		}
		if _, exists := seen[candidate]; exists {
			continue
		}
		seen[candidate] = struct{}{}
		deduped = append(deduped, candidate)
	}

	if len(deduped) == 0 {
		return fmt.Errorf("no valid URLs provided for selection")
	}

	if len(deduped) > maxSelectableLinks {
		return fmt.Errorf("you can select up to %d links for indexing", maxSelectableLinks)
	}

	if err := s.knowledgeRepo.SaveSelectedLinks(jobID, deduped); err != nil {
		return fmt.Errorf("failed to store selected links: %w", err)
	}

	return nil
}

// StreamIndexing processes selected links and streams progress updates via the provided channel
func (s *WebScrapingService) StreamIndexing(ctx context.Context, jobID, tenantID, projectID uuid.UUID, events chan<- IndexingEvent) error {
	job, err := s.knowledgeRepo.GetScrapingJob(jobID, tenantID, projectID)
	if err != nil {
		return fmt.Errorf("failed to load scraping job: %w", err)
	}

	if job.Status != "awaiting_selection" && job.Status != "indexing" {
		return fmt.Errorf("job is not ready for indexing (status: %s)", job.Status)
	}

	if len(job.SelectedLinks) == 0 {
		return fmt.Errorf("no links have been selected for indexing")
	}

	if err := s.knowledgeRepo.StartIndexingJob(jobID, job.SelectedLinks); err != nil {
		return fmt.Errorf("failed to mark job as indexing: %w", err)
	}

	s.sendIndexingEvent(ctx, events, IndexingEvent{
		Type:      "started",
		Total:     len(job.SelectedLinks),
		Timestamp: time.Now(),
	})

	pagesForIndex := make([]*models.KnowledgeScrapedPage, 0, len(job.SelectedLinks))
	totalTokens := 0
	now := time.Now()

	// Create collector to fetch content fresh from URLs
	c := colly.NewCollector(
		colly.UserAgent(s.config.ScrapeUserAgent),
	)
	c.SetRequestTimeout(s.config.ScrapeTimeout)

	processed := 0
	for _, url := range job.SelectedLinks {
		s.sendIndexingEvent(ctx, events, IndexingEvent{
			Type:      "processing",
			Message:   fmt.Sprintf("Fetching content from: %s", url),
			URL:       url,
			Completed: processed,
			Pending:   len(job.SelectedLinks) - processed,
			Timestamp: time.Now(),
		})

		var title, content string

		c.OnHTML("html", func(e *colly.HTMLElement) {
			title = e.ChildText("title")
			if title == "" {
				title = e.ChildText("h1")
			}
			content = s.extractTextContent(e)
		})

		c.OnError(func(r *colly.Response, e error) {
			s.sendIndexingEvent(ctx, events, IndexingEvent{
				Type:      "warning",
				Message:   fmt.Sprintf("Error fetching %s: %v", url, e),
				URL:       url,
				Timestamp: time.Now(),
			})
		})

		if err := c.Visit(url); err != nil {
			s.sendIndexingEvent(ctx, events, IndexingEvent{
				Type:      "warning",
				Message:   fmt.Sprintf("Failed to fetch %s: %v", url, err),
				URL:       url,
				Timestamp: time.Now(),
			})
			continue
		}

		c.Wait()

		if len(strings.TrimSpace(content)) < 100 {
			s.sendIndexingEvent(ctx, events, IndexingEvent{
				Type:      "warning",
				Message:   "Content too short, skipping",
				URL:       url,
				Timestamp: time.Now(),
			})
			continue
		}

		tokenCount := s.estimateTokenCount(content)
		totalTokens += tokenCount

		page := &models.KnowledgeScrapedPage{
			ID:         uuid.New(),
			JobID:      jobID,
			URL:        url,
			Content:    content,
			TokenCount: tokenCount,
			ScrapedAt:  now,
			Metadata: models.JSONMap{
				"source": "web_scraper",
				"depth":  0, // Will be set by indexing order
			},
		}

		if title != "" {
			page.Title = &title
		}

		hash := s.generateContentHash(content)
		page.ContentHash = &hash

		pagesForIndex = append(pagesForIndex, page)
		processed++

		s.sendIndexingEvent(ctx, events, IndexingEvent{
			Type:        "page_processed",
			Message:     fmt.Sprintf("Processed: %s", title),
			URL:         url,
			Completed:   processed,
			Pending:     len(job.SelectedLinks) - processed,
			TokenCount:  tokenCount,
			TotalTokens: totalTokens,
			Timestamp:   time.Now(),
		})

		// Update progress
		if err := s.knowledgeRepo.UpdateScrapingJobProgress(jobID, processed, len(job.SelectedLinks)); err != nil {
			logger.GetTxLogger(ctx).Warn().
				Str("component", "web_scraper").
				Str("job_id", jobID.String()).
				Err(err).
				Msg("Failed to update scraping progress")
		}
	}

	if len(pagesForIndex) == 0 {
		return fmt.Errorf("no valid selected pages found for indexing")
	}

	s.sendIndexingEvent(ctx, events, IndexingEvent{
		Type:        "started",
		Total:       len(pagesForIndex),
		TotalTokens: totalTokens,
		Timestamp:   time.Now(),
	})

	if err := s.knowledgeRepo.UpdateScrapingJobProgress(jobID, 0, len(pagesForIndex)); err != nil {
		logger.GetTxLogger(ctx).Warn().
			Str("component", "web_scraper").
			Str("job_id", jobID.String()).
			Err(err).
			Msg("Failed to initialise indexing progress")
	}

	if err := s.knowledgeRepo.CreateScrapedPages(pagesForIndex); err != nil {
		s.sendIndexingEvent(ctx, events, IndexingEvent{
			Type:      "error",
			Message:   fmt.Sprintf("Failed to store selected pages: %v", err),
			Timestamp: time.Now(),
		})
		return fmt.Errorf("failed to store selected pages: %w", err)
	}

	pagesToEmbed := make([]*models.KnowledgeScrapedPage, 0, len(pagesForIndex))
	completed := 0
	for _, page := range pagesForIndex {
		if page.Embedding == nil {
			pagesToEmbed = append(pagesToEmbed, page)
		} else {
			completed++
			s.sendIndexingEvent(ctx, events, IndexingEvent{
				Type:        "skipped",
				Message:     "Duplicate content detected; skipping embedding",
				URL:         page.URL,
				Completed:   completed,
				Pending:     len(pagesForIndex) - completed,
				TokenCount:  page.TokenCount,
				TotalTokens: totalTokens,
				Timestamp:   time.Now(),
			})
		}
	}

	if err := s.knowledgeRepo.UpdateScrapingJobProgress(jobID, completed, len(pagesForIndex)); err != nil {
		logger.GetTxLogger(ctx).Warn().
			Str("component", "web_scraper").
			Str("job_id", jobID.String()).
			Int("completed", completed).
			Int("total_pages", len(pagesForIndex)).
			Err(err).
			Msg("Failed to update indexing progress")
	}

	if len(pagesToEmbed) > 0 {
		if !s.embeddingService.IsEnabled() {
			s.sendIndexingEvent(ctx, events, IndexingEvent{
				Type:      "warning",
				Message:   "Embedding service disabled; pages saved without embeddings",
				Timestamp: time.Now(),
			})
		} else {
			s.sendIndexingEvent(ctx, events, IndexingEvent{
				Type:        "embedding_started",
				Total:       len(pagesToEmbed),
				Completed:   completed,
				Pending:     len(pagesForIndex) - completed,
				TotalTokens: totalTokens,
				Timestamp:   time.Now(),
			})

			embeddingCtx, cancel := context.WithTimeout(context.Background(), s.config.EmbeddingTimeout)
			defer cancel()

			if err := s.generateEmbeddingsForPages(embeddingCtx, pagesToEmbed); err != nil {
				s.sendIndexingEvent(ctx, events, IndexingEvent{
					Type:      "error",
					Message:   fmt.Sprintf("Embedding generation failed: %v", err),
					Timestamp: time.Now(),
				})
				return fmt.Errorf("failed to generate embeddings: %w", err)
			}

			completed += len(pagesToEmbed)
			s.sendIndexingEvent(ctx, events, IndexingEvent{
				Type:        "embedding_completed",
				Completed:   completed,
				Pending:     len(pagesForIndex) - completed,
				Total:       len(pagesForIndex),
				TotalTokens: totalTokens,
				Timestamp:   time.Now(),
			})
		}
	}

	if err := s.knowledgeRepo.UpdateScrapingJobProgress(jobID, len(pagesForIndex), len(pagesForIndex)); err != nil {
		logger.GetTxLogger(ctx).Warn().
			Str("component", "web_scraper").
			Str("job_id", jobID.String()).
			Int("total_pages", len(pagesForIndex)).
			Err(err).
			Msg("Failed to finalise indexing progress")
	}

	if err := s.knowledgeRepo.CompleteIndexingJob(jobID); err != nil {
		s.sendIndexingEvent(ctx, events, IndexingEvent{
			Type:      "error",
			Message:   fmt.Sprintf("Failed to finalise indexing job: %v", err),
			Timestamp: time.Now(),
		})
		return fmt.Errorf("failed to mark indexing job complete: %w", err)
	}

	s.sendIndexingEvent(ctx, events, IndexingEvent{
		Type:        "completed",
		Total:       len(pagesForIndex),
		Completed:   len(pagesForIndex),
		Pending:     0,
		TotalTokens: totalTokens,
		Timestamp:   time.Now(),
	})

	return nil
}

// extractTextContent extracts clean text content from HTML
func (s *WebScrapingService) extractTextContent(e *colly.HTMLElement) string {
	// Remove script and style elements
	e.ForEach("script, style, nav, header, footer, aside", func(_ int, el *colly.HTMLElement) {
		el.DOM.Remove()
	})

	// Extract main content areas first
	mainContent := e.ChildText("main, [role=main], .content, .main-content, article")
	if mainContent != "" {
		return s.cleanText(mainContent)
	}

	// Fallback to body content
	bodyContent := e.ChildText("body")
	return s.cleanText(bodyContent)
}

// cleanText cleans and normalizes text content
func (s *WebScrapingService) cleanText(text string) string {
	// Replace multiple whitespaces with single space
	text = strings.Join(strings.Fields(text), " ")

	// Remove excessive newlines
	text = strings.ReplaceAll(text, "\n\n\n", "\n\n")

	return strings.TrimSpace(text)
}

// shouldFollowLink determines if a link should be followed
func (s *WebScrapingService) shouldFollowLink(linkURL, currentURL string) bool {
	parsedLink, err := url.Parse(linkURL)
	if err != nil {
		return false
	}

	parsedCurrent, err := url.Parse(currentURL)
	if err != nil {
		return false
	}

	// Only follow links on the same domain
	if parsedLink.Host != parsedCurrent.Host {
		return false
	}

	// Skip certain file types
	path := strings.ToLower(parsedLink.Path)
	skipExtensions := []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".zip", ".tar", ".gz", ".jpg", ".jpeg", ".png", ".gif", ".svg", ".mp4", ".mp3", ".avi", ".mov"}
	for _, ext := range skipExtensions {
		if strings.HasSuffix(path, ext) {
			return false
		}
	}

	// Skip common non-content paths
	skipPaths := []string{"/admin", "/api", "/login", "/register", "/download", "/upload"}
	for _, skipPath := range skipPaths {
		if strings.Contains(path, skipPath) {
			return false
		}
	}

	return true
}

// generateEmbeddingsForPages generates embeddings for scraped pages
func (s *WebScrapingService) generateEmbeddingsForPages(ctx context.Context, pages []*models.KnowledgeScrapedPage) error {
	if len(pages) == 0 {
		return nil
	}

	// Extract texts for batch embedding generation
	texts := make([]string, len(pages))
	for i, page := range pages {
		texts[i] = page.Content
	}

	// Generate embeddings in batch
	txLogger := logger.GetTxLogger(ctx).With().
		Str("component", "web_scraper").
		Str("operation", "generate_embeddings").
		Int("page_count", len(pages)).
		Logger()

	txLogger.Info().Msg("Starting embedding generation")
	embeddings, err := s.embeddingService.GenerateEmbeddings(ctx, texts)
	if err != nil {
		txLogger.Error().
			Err(err).
			Msg("Error generating embeddings")
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	if len(embeddings) != len(pages) {
		return fmt.Errorf("embedding count mismatch: expected %d, got %d", len(pages), len(embeddings))
	}

	// Assign embeddings to pages
	for i, embedding := range embeddings {
		if i < len(pages) {
			pages[i].Embedding = &embedding
		}
	}

	// Update the database with the new embeddings
	txLogger.Info().Msg("Updating database with embeddings")
	err = s.knowledgeRepo.UpdatePageEmbeddings(pages)
	if err != nil {
		txLogger.Error().
			Err(err).
			Msg("Failed to update page embeddings")
		return fmt.Errorf("failed to update page embeddings: %w", err)
	}

	txLogger.Info().
		Int("updated_pages", len(pages)).
		Msg("Successfully assigned and saved embeddings to pages")
	return nil
}

// generateContentHash creates a SHA256 hash of the content for deduplication
func (s *WebScrapingService) generateContentHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// GetScrapingJob returns a scraping job by ID
func (s *WebScrapingService) GetScrapingJob(ctx context.Context, jobID, tenantID, projectID uuid.UUID) (*models.KnowledgeScrapingJob, error) {
	return s.knowledgeRepo.GetScrapingJob(jobID, tenantID, projectID)
}

// ListScrapingJobs returns a list of scraping jobs for a project
func (s *WebScrapingService) ListScrapingJobs(ctx context.Context, tenantID, projectID uuid.UUID, limit, offset int) ([]*models.KnowledgeScrapingJob, error) {
	return s.knowledgeRepo.ListScrapingJobs(tenantID, projectID, limit, offset)
}

// GetJobPages returns all pages scraped by a job
func (s *WebScrapingService) GetJobPages(ctx context.Context, jobID, tenantID, projectID uuid.UUID) ([]*models.KnowledgeScrapedPage, error) {
	return s.knowledgeRepo.GetJobPages(jobID, tenantID, projectID)
}

// WebsiteThemeData represents extracted website theme information
type WebsiteThemeData struct {
	Colors         []string `json:"colors"`
	BackgroundHues []string `json:"background_hues"`
	FontFamilies   []string `json:"font_families"`
	BrandName      string   `json:"brand_name"`
	PageTitle      string   `json:"page_title"`
	MetaDesc       string   `json:"meta_description"`
	CSS            string   `json:"css_content"`
}

// Implement ThemeData interface methods
func (w *WebsiteThemeData) GetColors() []string {
	return w.Colors
}

func (w *WebsiteThemeData) GetBackgroundHues() []string {
	return w.BackgroundHues
}

func (w *WebsiteThemeData) GetFontFamilies() []string {
	return w.FontFamilies
}

func (w *WebsiteThemeData) GetBrandName() string {
	return w.BrandName
}

func (w *WebsiteThemeData) GetPageTitle() string {
	return w.PageTitle
}

func (w *WebsiteThemeData) GetMetaDesc() string {
	return w.MetaDesc
}

// ScrapeWebsiteTheme extracts theme information from a website for AI analysis
func (s *WebScrapingService) ScrapeWebsiteTheme(ctx context.Context, targetURL string) (*WebsiteThemeData, error) {
	// Validate URL
	if err := s.validateURL(targetURL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	themeData := &WebsiteThemeData{
		Colors:         []string{},
		BackgroundHues: []string{},
		FontFamilies:   []string{},
	}

	// Create colly collector
	c := colly.NewCollector(
		colly.UserAgent("Hith-ThemeBot/1.0 (+https://api.hith.chat)"),
	)

	// Set timeout
	c.SetRequestTimeout(30 * time.Second)

	// Extract theme information from the page
	c.OnHTML("html", func(e *colly.HTMLElement) {
		// Extract page title
		themeData.PageTitle = e.ChildText("title")

		// Extract meta description
		themeData.MetaDesc = e.ChildAttr("meta[name='description']", "content")

		// Try to extract brand name from various sources
		brandName := e.ChildText("h1")
		if brandName == "" {
			brandName = e.ChildAttr("meta[property='og:site_name']", "content")
		}
		if brandName == "" {
			brandName = e.ChildText(".brand, .logo, .site-title, [class*='brand'], [class*='logo']")
		}
		themeData.BrandName = strings.TrimSpace(brandName)

		// Extract inline CSS and style information
		cssContent := ""
		e.ForEach("style", func(_ int, el *colly.HTMLElement) {
			cssContent += el.Text + "\n"
		})

		// Extract link to external stylesheets (first few)
		e.ForEach("link[rel='stylesheet']", func(i int, el *colly.HTMLElement) {
			if i < 3 { // Limit to first 3 stylesheets
				href := el.Attr("href")
				if href != "" {
					// Convert relative URLs to absolute
					if !strings.HasPrefix(href, "http") {
						if strings.HasPrefix(href, "//") {
							href = "https:" + href
						} else if strings.HasPrefix(href, "/") {
							parsedURL, _ := url.Parse(targetURL)
							href = parsedURL.Scheme + "://" + parsedURL.Host + href
						}
					}
					cssContent += "/* External CSS: " + href + " */\n"
				}
			}
		})

		themeData.CSS = cssContent

		// Extract colors from inline styles
		colorRegex := regexp.MustCompile(`(?i)color\s*:\s*([^;]+)`)
		bgColorRegex := regexp.MustCompile(`(?i)background(?:-color)?\s*:\s*([^;]+)`)

		// Extract from style attributes
		e.ForEach("*[style]", func(_ int, el *colly.HTMLElement) {
			style := el.Attr("style")

			// Extract colors
			matches := colorRegex.FindAllStringSubmatch(style, -1)
			for _, match := range matches {
				color := strings.TrimSpace(match[1])
				if s.isValidColor(color) {
					themeData.Colors = append(themeData.Colors, color)
				}
			}

			// Extract background colors
			matches = bgColorRegex.FindAllStringSubmatch(style, -1)
			for _, match := range matches {
				color := strings.TrimSpace(match[1])
				if s.isValidColor(color) {
					themeData.BackgroundHues = append(themeData.BackgroundHues, color)
				}
			}
		})

		// Extract font families from CSS and computed styles
		fontRegex := regexp.MustCompile(`(?i)font-family\s*:\s*([^;]+)`)
		matches := fontRegex.FindAllStringSubmatch(cssContent, -1)
		for _, match := range matches {
			font := strings.TrimSpace(match[1])
			if font != "" {
				themeData.FontFamilies = append(themeData.FontFamilies, font)
			}
		}
	})

	// Error handling
	c.OnError(func(r *colly.Response, err error) {
		logger.GetTxLogger(ctx).Error().
			Str("component", "web_scraper").
			Str("operation", "scrape_theme_data").
			Str("url", r.Request.URL.String()).
			Err(err).
			Msg("Error scraping theme data from URL")
	})

	// Visit the URL
	err := c.Visit(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape website: %w", err)
	}

	// Clean up and deduplicate extracted data
	themeData.Colors = s.deduplicateStrings(themeData.Colors)
	themeData.BackgroundHues = s.deduplicateStrings(themeData.BackgroundHues)
	themeData.FontFamilies = s.deduplicateStrings(themeData.FontFamilies)

	return themeData, nil
}

// isValidColor checks if a color value is valid
func (s *WebScrapingService) isValidColor(color string) bool {
	color = strings.TrimSpace(color)
	if color == "" {
		return false
	}

	// Check for hex colors
	if strings.HasPrefix(color, "#") && len(color) >= 4 {
		return true
	}

	// Check for rgb/rgba
	if strings.HasPrefix(color, "rgb") {
		return true
	}

	// Check for common color names
	commonColors := []string{"red", "blue", "green", "yellow", "orange", "purple", "pink",
		"brown", "black", "white", "gray", "grey", "navy", "teal", "cyan", "magenta"}
	for _, cc := range commonColors {
		if strings.EqualFold(color, cc) {
			return true
		}
	}

	return false
}

// deduplicateStrings removes duplicate strings from a slice
func (s *WebScrapingService) deduplicateStrings(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] && item != "" {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}
