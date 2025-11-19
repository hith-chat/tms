package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/logger"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
)

type WebScrapingService struct {
	knowledgeRepo            *repo.KnowledgeRepository
	embeddingService         *EmbeddingService
	config                   *config.KnowledgeConfig
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

// URLExtractionEvent represents streamed progress updates for URL extraction debugging
type URLExtractionEvent struct {
	Type         string            `json:"type"`
	Message      string            `json:"message,omitempty"`
	URL          string            `json:"url,omitempty"`
	CurrentDepth int               `json:"current_depth,omitempty"`
	MaxDepth     int               `json:"max_depth,omitempty"`
	LinksFound   int               `json:"links_found,omitempty"`
	URLs         []string          `json:"urls,omitempty"`
	FailedURLs   map[string]string `json:"failed_urls,omitempty"`
	Timestamp    time.Time         `json:"timestamp"`
	Metrics      *CrawlMetrics     `json:"metrics,omitempty"`       // Overall performance metrics
	DepthMetrics *DepthMetrics     `json:"depth_metrics,omitempty"` // Per-depth metrics
}

func NewWebScrapingService(knowledgeRepo *repo.KnowledgeRepository, embeddingService *EmbeddingService, cfg *config.KnowledgeConfig) *WebScrapingService {
	// Initialize headless browser extractor with 30 second timeout
	headlessExtractor := NewHeadlessBrowserURLExtractor(30*time.Second, "")

	return &WebScrapingService{
		knowledgeRepo:            knowledgeRepo,
		embeddingService:         embeddingService,
		config:                   cfg,
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

// CreateScrapingJobWithStreamAndBrowser creates a new web scraping job with shared browser context and depth-based strategy
func (s *WebScrapingService) CreateScrapingJobWithStreamAndBrowser(ctx context.Context, tenantID, projectID uuid.UUID, req *models.CreateScrapingJobRequest, sharedBrowser *SharedBrowserContext, maxDepth int, events chan<- ScrapingEvent) (*models.KnowledgeScrapingJob, error) {
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

	// Start scraping with depth-based strategy (Playwright for 0-1, Colly for 2+)
	go s.startScrapingJobWithStreamAndBrowser(ctx, job, sharedBrowser, maxDepth, events)

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

// extractURLsManually extracts URLs using Colly-based comprehensive extraction
func (s *WebScrapingService) extractURLsManually(ctx context.Context, targetURL string, maxDepth int) ([]discoveredLink, error) {
	logger.InfofCtx(ctx, "Starting URL extraction with Colly - target: %s, max_depth: %d", targetURL, maxDepth)
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

	// Parse target URL to extract root domain for filtering
	parsedTargetURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target URL: %w", err)
	}
	rootHost := parsedTargetURL.Host
	rootBaseDomain := getBaseDomain(rootHost)

	txLogger.Debug().
		Str("root_host", rootHost).
		Str("root_base_domain", rootBaseDomain).
		Msg("Extracted root domain information")

	// Try headless browser extraction
	var allLinks []discoveredLink
	visited := make(map[string]bool)
	failedURLs := make(map[string]string) // Track failed URLs and their error messages
	toVisit := []struct {
		url   string
		depth int
	}{{targetURL, 0}}

	for len(toVisit) > 0 && len(allLinks) < 10 {
		current := toVisit[0]
		toVisit = toVisit[1:]

		// Normalize URL to check if we've already visited this page
		normalizedURL := normalizeURLForDeduplication(current.url)
		if current.depth > maxDepth || visited[normalizedURL] {
			continue
		}

		visited[normalizedURL] = true

		// Extract URLs from current page using headless browser
		pageLogger := txLogger.With().
			Str("current_url", current.url).
			Int("depth", current.depth).
			Logger()

		pageLogger.Debug().Msg("Extracting URLs from page")
		extractedURLs, err := s.headlessBrowserExtractor.ExtractURLsFromPage(ctx, current.url)
		if err != nil {
			// Log the error but continue processing other pages
			pageLogger.Warn().
				Err(err).
				Msg("Failed to extract URLs from page - continuing with other pages")
			failedURLs[current.url] = err.Error()
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
				// Check if URL belongs to the same base domain as the root
				parsedURL, parseErr := url.Parse(urlInfo.URL)
				if parseErr != nil {
					continue // Skip invalid URLs
				}

				// Only follow links on the same base domain (allows subdomains)
				if !isSameBaseDomain(parsedURL.Host, rootHost) {
					urlBaseDomain := getBaseDomain(parsedURL.Host)
					pageLogger.Debug().
						Str("url", urlInfo.URL).
						Str("url_host", parsedURL.Host).
						Str("url_base_domain", urlBaseDomain).
						Str("root_base_domain", rootBaseDomain).
						Msg("Skipping URL - different base domain")
					continue
				}

				// Check file extension before processing
				if !s.shouldCrawlURLByExtension(urlInfo.URL) {
					continue
				}

				// Normalize URL for deduplication check
				normalizedDiscoveredURL := normalizeURLForDeduplication(urlInfo.URL)
				if !visited[normalizedDiscoveredURL] && s.shouldFollowLink(urlInfo.URL, current.url) {
					// Visit the normalized URL (without query params)
					toVisit = append(toVisit, struct {
						url   string
						depth int
					}{normalizedDiscoveredURL, current.depth + 1})
				}
			}
		}
	}

	// Log summary of the extraction
	txLogger.Info().
		Int("total_links_found", len(allLinks)).
		Int("total_visited", len(visited)).
		Int("total_failed", len(failedURLs)).
		Msg("Headless browser extraction completed")

	if len(failedURLs) > 0 {
		txLogger.Warn().
			Int("failed_count", len(failedURLs)).
			Interface("failed_urls", failedURLs).
			Msg("Some pages failed to load but extraction continued successfully")
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
	// Don't return error on bad status codes - we'll handle them in OnError
	c.IgnoreRobotsTxt = true

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

				// Check file extension before visiting
				if !s.shouldCrawlURLByExtension(extractedURL) {
					continue
				}

				if s.shouldFollowLink(extractedURL, currentURL) && !visitedURLs[extractedURL] {
					visitedURLs[extractedURL] = true
					e.Request.Visit(extractedURL)
				}
			}
		}
	})

	// Track if we successfully scraped at least something
	var initialError error

	// Error logging - but don't fail the whole crawl for individual page errors
	c.OnError(func(r *colly.Response, err error) {
		// If this is the initial URL, record the error
		if r.Request.URL.String() == targetURL && len(discoveredLinks) == 0 {
			initialError = err
		}
		logger.GetTxLogger(ctx).Warn().
			Str("component", "web_scraper").
			Str("operation", "colly_error").
			Str("url", r.Request.URL.String()).
			Int("status_code", r.StatusCode).
			Err(err).
			Msg("Error visiting URL during comprehensive extraction")
	})

	// Start crawl
	if err := c.Visit(targetURL); err != nil {
		// Log but continue - the OnError callback will capture details
		logger.GetTxLogger(ctx).Warn().
			Str("url", targetURL).
			Err(err).
			Msg("Initial visit returned error, checking if we got any content")
	}

	c.Wait()

	// If we got no links and had an initial error, report it
	if len(discoveredLinks) == 0 && initialError != nil {
		return nil, fmt.Errorf("failed to scrape URL: %w", initialError)
	}
	if len(discoveredLinks) == 0 {
		return nil, fmt.Errorf("no content could be extracted from URL: %s", targetURL)
	}

	return discoveredLinks, nil
}

// startScrapingJobWithStreamAndBrowser discovers links using depth-based strategy with shared browser
func (s *WebScrapingService) startScrapingJobWithStreamAndBrowser(ctx context.Context, job *models.KnowledgeScrapingJob, sharedBrowser *SharedBrowserContext, maxDepth int, events chan<- ScrapingEvent) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	defer close(events)

	var runErr error
	logger.InfofCtx(ctx, "Starting scraping job with browser reuse - jobID: %s, URL: %s, maxDepth: %d", job.ID.String(), job.URL, maxDepth)

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
		Message:   fmt.Sprintf("Starting link discovery for %s (depth-based strategy)", job.URL),
		URL:       job.URL,
		MaxDepth:  job.MaxDepth,
		Timestamp: time.Now(),
	})

	var discoveredLinks []discoveredLink

	// Use Playwright for depth 0-1, Colly for depth 2+
	if maxDepth <= 1 {
		s.sendScrapingEvent(ctx, events, ScrapingEvent{
			Type:      "info",
			JobID:     job.ID,
			Message:   "Using Playwright browser for shallow crawl (depth 0-1)",
			Timestamp: time.Now(),
		})

		// Use Playwright with shared browser for depth 0-1
		playwrightLinks, playwrightErr := s.extractURLsWithHeadlessBrowser(ctx, job.URL, job.MaxDepth)
		if playwrightErr != nil {
			runErr = fmt.Errorf("playwright extraction failed: %w", playwrightErr)
			return
		}

		discoveredLinks = playwrightLinks
	} else {
		s.sendScrapingEvent(ctx, events, ScrapingEvent{
			Type:      "info",
			JobID:     job.ID,
			Message:   "Using lightweight Colly for deep crawl (depth 2+)",
			Timestamp: time.Now(),
		})

		// Use Colly for deeper crawls (more efficient)
		collyLinks, collyErr := s.extractURLsManually(ctx, job.URL, job.MaxDepth)
		if collyErr != nil {
			runErr = fmt.Errorf("colly extraction failed: %w", collyErr)
			return
		}

		discoveredLinks = collyLinks
	}

	// Stream discovered links
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

	totalLinks := len(discoveredLinks)
	if totalLinks == 0 {
		runErr = fmt.Errorf("no links were discovered")
		return
	}

	// Mark job as completed with link count (two-phase workflow deprecated - use ScrapeURLs endpoint instead)
	if err := s.knowledgeRepo.UpdateScrapingJobProgress(job.ID, totalLinks, 0); err != nil {
		runErr = fmt.Errorf("failed to update job progress: %w", err)
		return
	}

	if err := s.knowledgeRepo.UpdateScrapingJobStatus(job.ID, "completed", nil); err != nil {
		runErr = fmt.Errorf("failed to mark job completed: %w", err)
		return
	}

	s.sendScrapingEvent(ctx, events, ScrapingEvent{
		Type:       "completed",
		JobID:      job.ID,
		Message:    fmt.Sprintf("Link discovery completed! Found %d links.", totalLinks),
		LinksFound: totalLinks,
		Timestamp:  time.Now(),
	})
	logger.InfofCtx(ctx, "Scraping job completed - jobID: %s, total_links: %d", job.ID.String(), totalLinks)
	runErr = nil
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

	// Use Colly-based extraction
	s.sendScrapingEvent(ctx, events, ScrapingEvent{
		Type:      "info",
		JobID:     job.ID,
		Message:   "Starting URL discovery with Colly...",
		URL:       job.URL,
		Timestamp: time.Now(),
	})

	discoveredLinks, extractErr := s.extractURLsManually(ctx, job.URL, job.MaxDepth)
	if extractErr != nil {
		runErr = fmt.Errorf("URL extraction failed: %v", extractErr)
		return
	}

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

	totalLinks := len(discoveredLinks)
	if totalLinks == 0 {
		runErr = fmt.Errorf("no links were discovered")
		return
	}

	// Mark job as completed with link count (two-phase workflow deprecated - use ScrapeURLs endpoint instead)
	if err := s.knowledgeRepo.UpdateScrapingJobProgress(job.ID, totalLinks, 0); err != nil {
		runErr = fmt.Errorf("failed to update job progress: %w", err)
		return
	}

	if err := s.knowledgeRepo.UpdateScrapingJobStatus(job.ID, "completed", nil); err != nil {
		runErr = fmt.Errorf("failed to mark job completed: %w", err)
		return
	}

	s.sendScrapingEvent(ctx, events, ScrapingEvent{
		Type:       "completed",
		JobID:      job.ID,
		Message:    fmt.Sprintf("Link discovery completed! Found %d links.", totalLinks),
		LinksFound: totalLinks,
		Timestamp:  time.Now(),
	})
	logger.InfofCtx(ctx, "Scraping job completed - jobID: %s, total_links: %d", job.ID.String(), totalLinks)
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
// DEPRECATED: The two-phase workflow (discover -> select -> index) is no longer supported.
// Use the ScrapeURLs endpoint to directly scrape and index URLs.
func (s *WebScrapingService) GetStagedLinks(ctx context.Context, jobID, tenantID, projectID uuid.UUID) ([]*models.ScrapedLinkPreview, error) {
	return nil, fmt.Errorf("the two-phase scraping workflow is deprecated; use the POST /knowledge/scrape-urls endpoint instead")
}

// StoreLinkSelection saves the user-selected URLs that should proceed to indexing
// DEPRECATED: The two-phase workflow (discover -> select -> index) is no longer supported.
// Use the ScrapeURLs endpoint to directly scrape and index URLs.
func (s *WebScrapingService) StoreLinkSelection(ctx context.Context, jobID, tenantID, projectID uuid.UUID, urls []string) error {
	return fmt.Errorf("the two-phase scraping workflow is deprecated; use the POST /knowledge/scrape-urls endpoint instead")
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

		// Normalize URL before storing
		normalizedURL := MustNormalizeURL(url)

		page := &models.KnowledgeScrapedPage{
			ID:         uuid.New(),
			JobID:      uuid.NullUUID{UUID: jobID, Valid: true},
			URL:        normalizedURL, // Use normalized URL
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
				Type:        "reused",
				Message:     "Duplicate content detected; reusing existing embedding",
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

// getBaseDomain extracts the base domain from a host
// e.g., "www.penify.dev" -> "penify.dev", "docs.penify.dev" -> "penify.dev"
func getBaseDomain(host string) string {
	// Remove port if present
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	parts := strings.Split(host, ".")
	if len(parts) <= 2 {
		return host
	}

	// For hosts with 3+ parts, take the last 2 parts as the base domain
	// This handles www.example.com, docs.example.com, etc.
	return strings.Join(parts[len(parts)-2:], ".")
}

// isSameBaseDomain checks if two hosts belong to the same base domain
func isSameBaseDomain(host1, host2 string) bool {
	return getBaseDomain(host1) == getBaseDomain(host2)
}

// normalizeURLForDeduplication normalizes URLs for deduplication by:
// - Removing query parameters and fragments
// - Stripping "www." prefix from hostname
// - Removing trailing slashes (except root)
// - Converting to lowercase
// Examples:
//   - "https://WWW.Example.COM/Contact-Us/?ref=A#top" -> "https://example.com/contact-us"
//   - "http://www.example.com/about/" -> "http://example.com/about"
//   - "https://example.com/page" -> "https://example.com/page"
func normalizeURLForDeduplication(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Remove query parameters and fragment
	parsed.RawQuery = ""
	parsed.Fragment = ""

	// Strip "www." prefix from hostname (case-insensitive)
	host := strings.ToLower(parsed.Host)
	if strings.HasPrefix(host, "www.") {
		parsed.Host = strings.TrimPrefix(host, "www.")
	} else {
		parsed.Host = host
	}

	// Normalize trailing slash (except for root path)
	if parsed.Path != "/" && strings.HasSuffix(parsed.Path, "/") {
		parsed.Path = strings.TrimSuffix(parsed.Path, "/")
	}

	// Lowercase the path for consistency
	parsed.Path = strings.ToLower(parsed.Path)

	return parsed.String()
}

// shouldCrawlURLByExtension checks if a URL should be crawled based on its file extension
// Returns true if:
// - URL has no file extension (appears to be a page/directory)
// - URL has an extension in the allowed list (configured via allowed_file_extensions)
// Returns false if:
// - URL has an extension not in the allowed list (.js, .css, .jpg, etc.)
func (s *WebScrapingService) shouldCrawlURLByExtension(urlStr string) bool {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	path := parsed.Path

	// Empty path or root path - allow
	if path == "" || path == "/" {
		return true
	}

	// Extract the last segment of the path
	segments := strings.Split(path, "/")
	lastSegment := segments[len(segments)-1]

	// If no dot in the last segment, it's likely a directory or clean URL - allow
	if !strings.Contains(lastSegment, ".") {
		return true
	}

	// Extract file extension (everything after the last dot)
	parts := strings.Split(lastSegment, ".")
	if len(parts) < 2 {
		return true // No extension
	}

	extension := "." + strings.ToLower(parts[len(parts)-1])

	// Check if extension is in the allowed list
	for _, allowed := range s.config.AllowedFileExtensions {
		if extension == strings.ToLower(allowed) {
			return true
		}
	}

	// Extension not in allowed list - reject
	logger.Debugf("Skipping URL %s with disallowed file extension: %s", urlStr, extension)

	return false
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

	// Only follow links on the same base domain (allows subdomains like docs.penify.dev, www.penify.dev)
	if !isSameBaseDomain(parsedLink.Host, parsedCurrent.Host) {
		return false
	}

	// Skip certain file types and problematic patterns
	path := strings.ToLower(parsedLink.Path)

	skipKeyWords := []string{
		"logout", "signout", "subscribe", "unsubscribe", "cart", "checkout", "min.js", "jquery", "bootstrap", "analytics", "facebook", "twitter", "linkedin", "instagram", "mailto:", "tel:", "javascript:", "cloudflareinsights", "visualwebsiteoptimizer", "hotjar", "googletagmanager", "google-analytics", "blogs",
	}

	for _, skipKeyWord := range skipKeyWords {
		if strings.Contains(currentURL, skipKeyWord) {
			fmt.Println("Skipping URL: "+currentURL, " due to keyword ", skipKeyWord)
			return false
		}
	}

	skipExtensions := []string{
		// Documents
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
		// Archives
		".zip", ".tar", ".gz", ".rar", ".7z",
		// Media files
		".jpg", ".jpeg", ".png", ".gif", ".svg", ".webp", ".ico",
		".mp4", ".mp3", ".avi", ".mov", ".wmv", ".flv", ".mkv", ".wav",
		// Manifests and configs
		".webmanifest", ".manifest", ".json", ".xml", ".rss",
		// Other
		".css", ".js", ".woff", ".woff2", ".ttf", ".eot",
	}
	for _, ext := range skipExtensions {
		if strings.HasSuffix(path, ext) || strings.Contains(path, ext+"?") || strings.Contains(path, ext+"#") || strings.Contains(path, ext+"&") || strings.Contains(path, ext+";") || strings.Contains(path, ext+",") || strings.Contains(path, ext+"/") {
			fmt.Println("Skipping URL: "+linkURL, " due to extension ", ext)
			return false
		}
	}

	// Skip common non-content paths
	skipPaths := []string{"/admin", "/api", "/login", "/register", "/download", "/upload", "site.webmanifest"}
	for _, skipPath := range skipPaths {
		if strings.Contains(path, skipPath) {
			fmt.Println("Skipping URL: "+linkURL, " due to path ", skipPath)
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

// extractPageWithColly extracts URLs, title, and content from a page using Colly (lightweight, fast)
func (s *WebScrapingService) extractPageWithColly(ctx context.Context, targetURL string) ([]ExtractedURLInfo, string, string, error) {
	var pageTitle string
	var pageContent string
	var htmlContent []byte
	var extractErr error

	c := colly.NewCollector(
		colly.UserAgent(s.config.ScrapeUserAgent),
	)
	c.SetRequestTimeout(s.config.ScrapeTimeout)

	// Extract title and content
	c.OnHTML("html", func(e *colly.HTMLElement) {
		// Extract title
		pageTitle = e.ChildText("title")
		if pageTitle == "" {
			pageTitle = e.ChildText("h1")
		}
		pageTitle = strings.TrimSpace(pageTitle)

		// Extract text content for knowledge base
		pageContent = s.extractTextContent(e)

		// Store raw HTML for URL extraction
		htmlContent = []byte(e.Response.Body)
	})

	c.OnError(func(r *colly.Response, err error) {
		extractErr = err
	})

	// Visit the URL
	if err := c.Visit(targetURL); err != nil {
		return nil, "", "", fmt.Errorf("failed to visit URL with Colly: %w", err)
	}

	c.Wait()

	if extractErr != nil {
		return nil, "", "", fmt.Errorf("colly extraction error: %w", extractErr)
	}

	// Extract URLs using ComprehensiveURLExtractor
	extractor, err := NewComprehensiveURLExtractor(targetURL)
	if err != nil {
		return nil, pageTitle, pageContent, fmt.Errorf("failed to create URL extractor: %w", err)
	}

	// Get URLs from HTML
	urlStrings := extractor.ExtractURLsFromHTML(htmlContent, targetURL)

	// Convert to ExtractedURLInfo format
	extractedURLs := make([]ExtractedURLInfo, 0, len(urlStrings))
	for _, urlStr := range urlStrings {
		extractedURLs = append(extractedURLs, ExtractedURLInfo{
			URL:    urlStr,
			Source: "colly",
		})
	}

	return extractedURLs, pageTitle, pageContent, nil
}

// extractPageWithPlaywright extracts URLs, title, and content using Playwright with optional shared browser
func (s *WebScrapingService) extractPageWithPlaywright(ctx context.Context, targetURL string, sharedBrowser *SharedBrowserContext) ([]ExtractedURLInfo, string, string, error) {
	// Extract URLs
	extractedURLs, err := s.headlessBrowserExtractor.ExtractURLsFromPage(ctx, targetURL)
	if err != nil {
		return nil, "", "", fmt.Errorf("playwright URL extraction failed: %w", err)
	}

	// Extract title and content
	title, _ := s.headlessBrowserExtractor.GetPageTitle(ctx, targetURL)
	content, _ := s.headlessBrowserExtractor.GetPageContent(ctx, targetURL)

	return extractedURLs, title, content, nil
}

// urlWorkItem represents a URL to be processed with its depth
type urlWorkItem struct {
	url   string
	depth int
}

// extractionResult represents the result of extracting a single page
type extractionResult struct {
	url           string
	depth         int
	title         string
	content       string
	tokenCount    int
	extractedURLs []ExtractedURLInfo
	err           error
}

// DepthMetrics represents performance metrics for a single depth level
type DepthMetrics struct {
	Depth             int           `json:"depth"`
	URLsProcessed     int           `json:"urls_processed"`
	URLsFailed        int           `json:"urls_failed"`
	URLsDiscovered    int           `json:"urls_discovered"`
	WorkerCount       int           `json:"worker_count"`
	Method            string        `json:"method"` // "playwright" or "colly"
	StartTime         time.Time     `json:"start_time"`
	EndTime           time.Time     `json:"end_time"`
	Duration          time.Duration `json:"duration"`
	AvgProcessingTime time.Duration `json:"avg_processing_time"`
	TotalTokens       int           `json:"total_tokens"`
}

// CrawlMetrics represents overall performance metrics for the entire crawl
type CrawlMetrics struct {
	TotalURLsProcessed  int            `json:"total_urls_processed"`
	TotalURLsFailed     int            `json:"total_urls_failed"`
	TotalURLsDiscovered int            `json:"total_urls_discovered"`
	TotalTokens         int            `json:"total_tokens"`
	StartTime           time.Time      `json:"start_time"`
	EndTime             time.Time      `json:"end_time"`
	TotalDuration       time.Duration  `json:"total_duration"`
	DepthMetrics        []DepthMetrics `json:"depth_metrics"`
	PlaywrightURLs      int            `json:"playwright_urls"`
	CollyURLs           int            `json:"colly_urls"`
	PlaywrightTime      time.Duration  `json:"playwright_time"`
	CollyTime           time.Duration  `json:"colly_time"`
	AvgURLsPerSecond    float64        `json:"avg_urls_per_second"`
}

// ExtractURLsWithStream extracts URLs from a website and streams progress for debugging
// Uses a hybrid approach with parallel processing:
// - Playwright (2-5 workers) for depth 0-1
// - Colly (10-20 workers) for depth >= 2
func (s *WebScrapingService) ExtractURLsWithStream(ctx context.Context, targetURL string, maxDepth int, events chan<- URLExtractionEvent, sharedBrowser *SharedBrowserContext, buildID string) error {
	defer close(events)

	// Validate URL
	if err := s.validateURL(targetURL); err != nil {
		s.sendURLExtractionEvent(ctx, events, URLExtractionEvent{
			Type:      "error",
			Message:   fmt.Sprintf("Invalid URL: %v", err),
			Timestamp: time.Now(),
		})
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Parse target URL to extract root domain
	parsedTargetURL, err := url.Parse(targetURL)
	if err != nil {
		s.sendURLExtractionEvent(ctx, events, URLExtractionEvent{
			Type:      "error",
			Message:   fmt.Sprintf("Failed to parse URL: %v", err),
			Timestamp: time.Now(),
		})
		return fmt.Errorf("failed to parse target URL: %w", err)
	}
	rootHost := parsedTargetURL.Host
	rootBaseDomain := getBaseDomain(rootHost)

	// URL limit counter (in-memory)
	const maxURLLimit = 30
	var urlCounter int64

	s.sendURLExtractionEvent(ctx, events, URLExtractionEvent{
		Type:      "started",
		Message:   fmt.Sprintf("Starting parallel URL extraction for %s (depth: %d, base domain: %s, limit: %d URLs)", targetURL, maxDepth, rootBaseDomain, maxURLLimit),
		URL:       targetURL,
		MaxDepth:  maxDepth,
		Timestamp: time.Now(),
	})

	// Shared state (protected by mutexes)
	var mu sync.Mutex
	var allURLs []string
	var allLinks []discoveredLink
	visited := make(map[string]bool)
	failedURLs := make(map[string]string)

	// Initialize performance metrics
	crawlStartTime := time.Now()
	var crawlMetrics CrawlMetrics
	crawlMetrics.StartTime = crawlStartTime
	crawlMetrics.DepthMetrics = []DepthMetrics{}

	// Initialize with root URL
	currentLevel := []urlWorkItem{{url: targetURL, depth: 0}}

	// Process URLs level by level (BFS) with parallel workers per level
	for len(currentLevel) > 0 && currentLevel[0].depth <= maxDepth {
		currentDepth := currentLevel[0].depth

		// Determine worker count from config based on depth
		// Playwright (depth 0-1): Use configured playwright worker count
		// Colly (depth >= 2): Use configured colly worker count
		workerCount := s.config.PlaywrightWorkerCount
		if currentDepth >= 2 {
			workerCount = s.config.CollyWorkerCount
		}

		// Fallback to defaults if config not set
		if workerCount <= 0 {
			workerCount = 3
			if currentDepth >= 2 {
				workerCount = 15
			}
		}

		// Initialize depth metrics
		depthStartTime := time.Now()
		method := "playwright"
		if currentDepth >= 2 {
			method = "colly"
		}

		s.sendURLExtractionEvent(ctx, events, URLExtractionEvent{
			Type:         "info",
			Message:      fmt.Sprintf("Processing depth %d with %d parallel workers (%d URLs queued)", currentDepth, workerCount, len(currentLevel)),
			CurrentDepth: currentDepth,
			MaxDepth:     maxDepth,
			Timestamp:    time.Now(),
		})

		// Create work channel and results channel
		workChan := make(chan urlWorkItem, len(currentLevel))
		resultsChan := make(chan extractionResult, len(currentLevel))

		// Add work items to channel
		for _, item := range currentLevel {
			// Check file extension before processing
			if !s.shouldCrawlURLByExtension(item.url) {
				continue
			}

			normalizedURL := normalizeURLForDeduplication(item.url)

			mu.Lock()
			alreadyVisited := visited[normalizedURL]
			if !alreadyVisited {
				visited[normalizedURL] = true
			}
			mu.Unlock()

			if !alreadyVisited {
				workChan <- item
			}
		}
		close(workChan)

		// Start worker pool
		var wg sync.WaitGroup
		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for item := range workChan {
					s.sendURLExtractionEvent(ctx, events, URLExtractionEvent{
						Type:         "visiting",
						Message:      fmt.Sprintf("Worker %d visiting: %s", workerID, item.url),
						URL:          item.url,
						CurrentDepth: item.depth,
						MaxDepth:     maxDepth,
						Timestamp:    time.Now(),
					})

					// Extract based on depth
					var extractedURLs []ExtractedURLInfo
					var title, content string
					var err error

					if item.depth <= 1 {
						extractedURLs, title, content, err = s.extractPageWithPlaywright(ctx, item.url, sharedBrowser)
					} else {
						extractedURLs, title, content, err = s.extractPageWithColly(ctx, item.url)
					}

					tokenCount := s.estimateTokenCount(content)

					resultsChan <- extractionResult{
						url:           item.url,
						depth:         item.depth,
						title:         title,
						content:       content,
						tokenCount:    tokenCount,
						extractedURLs: extractedURLs,
						err:           err,
					}

					logger.GetTxLogger(ctx).Info().
						Str("url", item.url).
						Int("depth", item.depth).
						Int("worker", workerID).
						Int("extracted_urls", len(extractedURLs)).
						Str("method", map[bool]string{true: "playwright", false: "colly"}[item.depth <= 1]).
						Msg("worker completed extraction")
				}
			}(i)
		}

		// Close results channel when all workers are done
		go func() {
			wg.Wait()
			close(resultsChan)
		}()

		// Collect results and prepare next level
		var nextLevel []urlWorkItem
		depthURLsProcessed := 0
		depthURLsFailed := 0
		depthURLsDiscovered := 0
		depthTotalTokens := 0

		for result := range resultsChan {
			depthURLsProcessed++
			if result.err != nil {
				depthURLsFailed++
				mu.Lock()
				failedURLs[result.url] = result.err.Error()
				mu.Unlock()

				s.sendURLExtractionEvent(ctx, events, URLExtractionEvent{
					Type:         "warning",
					Message:      fmt.Sprintf("Failed to extract from %s: %v", result.url, result.err),
					URL:          result.url,
					CurrentDepth: result.depth,
					Timestamp:    time.Now(),
				})
				continue
			}

			depthTotalTokens += result.tokenCount

			// Add to results (thread-safe)
			mu.Lock()
			allLinks = append(allLinks, discoveredLink{
				URL:        result.url,
				Title:      result.title,
				Depth:      result.depth,
				TokenCount: result.tokenCount,
			})
			allURLs = append(allURLs, result.url)
			currentLinksCount := len(allURLs)
			mu.Unlock()

			s.sendURLExtractionEvent(ctx, events, URLExtractionEvent{
				Type:         "url_found",
				Message:      fmt.Sprintf("Found: %s (depth: %d)", result.title, result.depth),
				URL:          result.url,
				CurrentDepth: result.depth,
				LinksFound:   currentLinksCount,
				Timestamp:    time.Now(),
			})

			// Process discovered URLs for next level
			if result.depth < maxDepth {
				newURLsCount := 0
				skippedSubdomains := 0

				for _, urlInfo := range result.extractedURLs {
					parsedURL, parseErr := url.Parse(urlInfo.URL)
					if parseErr != nil {
						continue
					}

					if !isSameBaseDomain(parsedURL.Host, rootHost) {
						skippedSubdomains++
						continue
					}

					// Check file extension before processing
					if !s.shouldCrawlURLByExtension(urlInfo.URL) {
						continue
					}

					normalizedDiscoveredURL := normalizeURLForDeduplication(urlInfo.URL)

					mu.Lock()
					alreadyVisited := visited[normalizedDiscoveredURL]
					mu.Unlock()

					if !alreadyVisited && s.shouldFollowLink(urlInfo.URL, result.url) {
						// Check URL counter limit
						shouldAdd := true
						mu.Lock()
						urlCounter++
						if urlCounter > int64(maxURLLimit) {
							shouldAdd = false
						}
						mu.Unlock()

						if !shouldAdd {
							logger.GetTxLogger(ctx).Debug().Msg("Reached URL limit, stopping further URL additions")
						} else {
							nextLevel = append(nextLevel, urlWorkItem{
								url:   normalizedDiscoveredURL,
								depth: result.depth + 1,
							})
							newURLsCount++
							depthURLsDiscovered++
						}
					}
				}

				if newURLsCount > 0 {
					s.sendURLExtractionEvent(ctx, events, URLExtractionEvent{
						Type:         "info",
						Message:      fmt.Sprintf("Discovered %d new URLs from %s (filtered %d external)", newURLsCount, result.url, skippedSubdomains),
						CurrentDepth: result.depth,
						Timestamp:    time.Now(),
					})
				}
			}
		}

		// Move to next level
		currentLevel = nextLevel

		// Calculate depth metrics
		depthEndTime := time.Now()
		depthDuration := depthEndTime.Sub(depthStartTime)
		avgProcessingTime := time.Duration(0)
		if depthURLsProcessed > 0 {
			avgProcessingTime = depthDuration / time.Duration(depthURLsProcessed)
		}

		depthMetric := DepthMetrics{
			Depth:             currentDepth,
			URLsProcessed:     depthURLsProcessed,
			URLsFailed:        depthURLsFailed,
			URLsDiscovered:    depthURLsDiscovered,
			WorkerCount:       workerCount,
			Method:            method,
			StartTime:         depthStartTime,
			EndTime:           depthEndTime,
			Duration:          depthDuration,
			AvgProcessingTime: avgProcessingTime,
			TotalTokens:       depthTotalTokens,
		}

		// Update crawl metrics
		crawlMetrics.DepthMetrics = append(crawlMetrics.DepthMetrics, depthMetric)
		crawlMetrics.TotalTokens += depthTotalTokens
		if method == "playwright" {
			crawlMetrics.PlaywrightURLs += depthURLsProcessed
			crawlMetrics.PlaywrightTime += depthDuration
		} else {
			crawlMetrics.CollyURLs += depthURLsProcessed
			crawlMetrics.CollyTime += depthDuration
		}

		mu.Lock()
		totalFound := len(allURLs)
		mu.Unlock()

		// Send depth completion event with metrics (if enabled)
		event := URLExtractionEvent{
			Type:         "info",
			Message:      fmt.Sprintf("Completed depth %d in %v. Processed: %d, Failed: %d, Discovered: %d. Total URLs: %d", currentDepth, depthDuration.Round(time.Millisecond), depthURLsProcessed, depthURLsFailed, depthURLsDiscovered, totalFound),
			CurrentDepth: currentDepth,
			LinksFound:   totalFound,
			Timestamp:    time.Now(),
		}

		if s.config.EnablePerformanceMetrics {
			event.DepthMetrics = &depthMetric
		}

		s.sendURLExtractionEvent(ctx, events, event)
	}

	// Calculate final crawl metrics
	crawlMetrics.EndTime = time.Now()
	crawlMetrics.TotalDuration = crawlMetrics.EndTime.Sub(crawlMetrics.StartTime)
	crawlMetrics.TotalURLsProcessed = len(allURLs)
	crawlMetrics.TotalURLsFailed = len(failedURLs)
	crawlMetrics.TotalURLsDiscovered = len(allURLs)

	// Calculate average URLs per second
	if crawlMetrics.TotalDuration.Seconds() > 0 {
		crawlMetrics.AvgURLsPerSecond = float64(crawlMetrics.TotalURLsProcessed) / crawlMetrics.TotalDuration.Seconds()
	}

	// Send final summary with metrics
	completionEvent := URLExtractionEvent{
		Type:       "completed",
		Message:    fmt.Sprintf("URL extraction completed in %v. Found %d URLs from base domain %s (Playwright: %d URLs in %v, Colly: %d URLs in %v, Avg: %.2f URLs/sec)", crawlMetrics.TotalDuration.Round(time.Millisecond), len(allURLs), rootBaseDomain, crawlMetrics.PlaywrightURLs, crawlMetrics.PlaywrightTime.Round(time.Millisecond), crawlMetrics.CollyURLs, crawlMetrics.CollyTime.Round(time.Millisecond), crawlMetrics.AvgURLsPerSecond),
		LinksFound: len(allURLs),
		URLs:       allURLs,
		FailedURLs: failedURLs,
		Timestamp:  time.Now(),
	}

	if s.config.EnablePerformanceMetrics {
		completionEvent.Metrics = &crawlMetrics
	}

	s.sendURLExtractionEvent(ctx, events, completionEvent)

	return nil
}

// sendURLExtractionEvent safely sends an event to the channel
func (s *WebScrapingService) sendURLExtractionEvent(ctx context.Context, ch chan<- URLExtractionEvent, event URLExtractionEvent) {
	select {
	case <-ctx.Done():
		return
	case ch <- event:
		return
	}
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

// ScrapeWebsiteThemeWithBrowser extracts theme information using an optional shared browser context
func (s *WebScrapingService) ScrapeWebsiteThemeWithBrowser(ctx context.Context, targetURL string, sharedBrowser *SharedBrowserContext) (*WebsiteThemeData, *SharedBrowserContext, error) {
	// Validate URL
	if err := s.validateURL(targetURL); err != nil {
		return nil, sharedBrowser, fmt.Errorf("invalid URL: %w", err)
	}

	var themeMap map[string]interface{}
	var err error

	// Use shared browser if provided, otherwise create new
	if sharedBrowser != nil {
		themeMap, err = s.headlessBrowserExtractor.ExtractThemeDataWithBrowser(ctx, targetURL, sharedBrowser)
	} else {
		// Create new browser instance for this request
		sharedBrowser, err = s.headlessBrowserExtractor.CreateSharedBrowserContext(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create browser context: %w", err)
		}
		themeMap, err = s.headlessBrowserExtractor.ExtractThemeDataWithBrowser(ctx, targetURL, sharedBrowser)
	}

	if err != nil {
		// Fallback to Colly if Playwright fails
		logger.GetTxLogger(ctx).Warn().
			Err(err).
			Str("url", targetURL).
			Msg("Playwright theme extraction failed, falling back to Colly")
		themeData, collyErr := s.scrapeThemeWithColly(ctx, targetURL)
		return themeData, sharedBrowser, collyErr
	}

	// Convert map to WebsiteThemeData
	themeData := &WebsiteThemeData{
		Colors:         []string{},
		BackgroundHues: []string{},
		FontFamilies:   []string{},
	}

	if pageTitle, ok := themeMap["pageTitle"].(string); ok {
		themeData.PageTitle = pageTitle
	}

	if metaDesc, ok := themeMap["metaDescription"].(string); ok {
		themeData.MetaDesc = metaDesc
	}

	if brandName, ok := themeMap["brandName"].(string); ok {
		themeData.BrandName = brandName
	}

	if colors, ok := themeMap["colors"].([]interface{}); ok {
		for _, c := range colors {
			if colorStr, ok := c.(string); ok {
				themeData.Colors = append(themeData.Colors, colorStr)
			}
		}
	}

	if bgColors, ok := themeMap["backgroundColors"].([]interface{}); ok {
		for _, c := range bgColors {
			if colorStr, ok := c.(string); ok {
				themeData.BackgroundHues = append(themeData.BackgroundHues, colorStr)
			}
		}
	}

	if fonts, ok := themeMap["fontFamilies"].([]interface{}); ok {
		for _, f := range fonts {
			if fontStr, ok := f.(string); ok {
				themeData.FontFamilies = append(themeData.FontFamilies, fontStr)
			}
		}
	}

	// Store CSS variables as JSON string
	if cssVars, ok := themeMap["cssVariables"].(map[string]interface{}); ok {
		cssVarsJSON, _ := json.Marshal(cssVars)
		themeData.CSS = string(cssVarsJSON)
	}

	return themeData, sharedBrowser, nil
}

// ScrapeWebsiteTheme extracts theme information from a website for AI analysis
func (s *WebScrapingService) ScrapeWebsiteTheme(ctx context.Context, targetURL string) (*WebsiteThemeData, error) {
	// Validate URL
	if err := s.validateURL(targetURL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Use Playwright for comprehensive theme extraction
	themeMap, err := s.headlessBrowserExtractor.ExtractThemeData(ctx, targetURL)
	if err != nil {
		// Fallback to Colly if Playwright fails
		logger.GetTxLogger(ctx).Warn().
			Err(err).
			Str("url", targetURL).
			Msg("Playwright theme extraction failed, falling back to Colly")
		return s.scrapeThemeWithColly(ctx, targetURL)
	}

	// Convert map to WebsiteThemeData
	themeData := &WebsiteThemeData{
		Colors:         []string{},
		BackgroundHues: []string{},
		FontFamilies:   []string{},
	}

	if pageTitle, ok := themeMap["pageTitle"].(string); ok {
		themeData.PageTitle = pageTitle
	}

	if metaDesc, ok := themeMap["metaDescription"].(string); ok {
		themeData.MetaDesc = metaDesc
	}

	if brandName, ok := themeMap["brandName"].(string); ok {
		themeData.BrandName = brandName
	}

	if colors, ok := themeMap["colors"].([]interface{}); ok {
		for _, c := range colors {
			if colorStr, ok := c.(string); ok {
				themeData.Colors = append(themeData.Colors, colorStr)
			}
		}
	}

	if bgColors, ok := themeMap["backgroundColors"].([]interface{}); ok {
		for _, c := range bgColors {
			if colorStr, ok := c.(string); ok {
				themeData.BackgroundHues = append(themeData.BackgroundHues, colorStr)
			}
		}
	}

	if fonts, ok := themeMap["fontFamilies"].([]interface{}); ok {
		for _, f := range fonts {
			if fontStr, ok := f.(string); ok {
				themeData.FontFamilies = append(themeData.FontFamilies, fontStr)
			}
		}
	}

	// Store CSS variables as JSON string
	if cssVars, ok := themeMap["cssVariables"].(map[string]interface{}); ok {
		cssVarsJSON, _ := json.Marshal(cssVars)
		themeData.CSS = string(cssVarsJSON)
	}

	return themeData, nil
}

// CaptureScreenshotWithBrowser captures a screenshot using an existing shared browser context
func (s *WebScrapingService) CaptureScreenshotWithBrowser(ctx context.Context, targetURL string, sharedBrowser *SharedBrowserContext) ([]byte, error) {
	return s.headlessBrowserExtractor.CaptureScreenshotWithBrowser(ctx, targetURL, sharedBrowser)
}

// scrapeThemeWithColly is a fallback method using Colly (original implementation)
func (s *WebScrapingService) scrapeThemeWithColly(ctx context.Context, targetURL string) (*WebsiteThemeData, error) {
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

// ScrapePageContent scrapes and extracts text content from a single URL
func (s *WebScrapingService) ScrapePageContent(ctx context.Context, targetURL string) (string, error) {
	// Use headless browser extractor for reliability
	content, err := s.headlessBrowserExtractor.GetPageContent(ctx, targetURL)
	if err != nil {
		return "", fmt.Errorf("failed to scrape page content: %w", err)
	}

	return content, nil
}

// StorePageInVectorDBWithTenantID stores a scraped page with tenant_id for cross-project deduplication
// Returns the created page ID
// Note: Sets job_id to nil for widget-created pages (tracked via widget_knowledge_pages instead)
// Normalizes URLs by removing query parameters to ensure uniqueness across the database
func (s *WebScrapingService) StorePageInVectorDBWithTenantID(ctx context.Context, tenantID, projectID uuid.UUID, url, title, content string, embedding pgvector.Vector, jobID uuid.UUID) (uuid.UUID, error) {
	// Normalize URL (remove query params, fragment, lowercase, trim trailing slash)
	normalizedURL, err := NormalizeURL(url)
	if err != nil {
		// If URL normalization fails, log but use the original URL trimmed
		logger.GetTxLogger(ctx).Warn().
			Str("url", url).
			Err(err).
			Msg("Failed to normalize URL, using original")
		normalizedURL = strings.TrimSpace(url)
	}

	// Calculate content hash
	hash := sha256.Sum256([]byte(content))
	contentHash := fmt.Sprintf("%x", hash)

	// Calculate token count
	tokenCount := len(strings.Fields(content))

	// Use provided title, fallback to URL if empty
	if title == "" {
		title = normalizedURL
	}
	// Truncate title if too long
	if len(title) > 200 {
		title = title[:200]
	}

	// Check if this URL already exists in the database (by normalized URL)
	existingPages, err := s.knowledgeRepo.GetExistingPagesByURL(ctx, normalizedURL)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to check for existing page: %w", err)
	}

	// If page exists with same content, return existing page ID
	if len(existingPages) > 0 {
		existingPage := existingPages[0]

		// Check if content has changed
		if existingPage.ContentHash != nil && *existingPage.ContentHash == contentHash {
			// Content is the same, reuse existing page
			logger.GetTxLogger(ctx).Info().
				Str("url", normalizedURL).
				Str("existing_page_id", existingPage.ID.String()).
				Msg("Reusing existing page with same content")
			return existingPage.ID, nil
		}

		// Content has changed, update the existing page
		logger.GetTxLogger(ctx).Info().
			Str("url", normalizedURL).
			Str("existing_page_id", existingPage.ID.String()).
			Msg("Updating existing page with new content")

		err = s.knowledgeRepo.UpdatePageContentAndEmbedding(ctx, existingPage.ID, title, content, contentHash, embedding, tokenCount)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to update existing page: %w", err)
		}
		return existingPage.ID, nil
	}

	// Create a new knowledge scraped page record
	// job_id is set to nil for widget pages - they're tracked via widget_knowledge_pages
	page := &models.KnowledgeScrapedPage{
		ID:          uuid.New(),
		JobID:       uuid.NullUUID{Valid: false}, // nil for widget pages
		URL:         normalizedURL,               // Use normalized URL
		Title:       &title,
		Content:     content,
		ContentHash: &contentHash,
		TokenCount:  tokenCount,
		ScrapedAt:   time.Now(),
		Embedding:   &embedding,
	}

	// Store in database with tenant_id
	if err := s.knowledgeRepo.CreateScrapedPageWithTenantID(ctx, page, tenantID); err != nil {
		return uuid.Nil, fmt.Errorf("failed to store page in vector DB: %w", err)
	}

	logger.GetTxLogger(ctx).Info().
		Str("url", normalizedURL).
		Str("page_id", page.ID.String()).
		Msg("Created new knowledge page")

	return page.ID, nil
}

// UpdatePageInVectorDB updates an existing page's content and embedding when content has changed
func (s *WebScrapingService) UpdatePageInVectorDB(ctx context.Context, pageID uuid.UUID, title, content string, embedding pgvector.Vector) error {
	// Calculate content hash
	hash := sha256.Sum256([]byte(content))
	contentHash := fmt.Sprintf("%x", hash)

	// Calculate token count
	tokenCount := len(strings.Fields(content))

	// Update the existing page
	if err := s.knowledgeRepo.UpdatePageContentAndEmbedding(ctx, pageID, title, content, contentHash, embedding, tokenCount); err != nil {
		return fmt.Errorf("failed to update page in vector DB: %w", err)
	}

	return nil
}

// ScrapeURLsResult represents the result of the simplified URL scraping
type ScrapeURLsResult struct {
	JobID        uuid.UUID `json:"job_id"`
	TotalURLs    int       `json:"total_urls"`
	PagesAdded   int       `json:"pages_added"`
	PagesSkipped int       `json:"pages_skipped"`
	PagesFailed  int       `json:"pages_failed"`
}

// ScrapeURLs performs simplified URL scraping - scrapes exact URLs provided without crawling
func (s *WebScrapingService) ScrapeURLs(ctx context.Context, tenantID, projectID uuid.UUID, urls []string, forceRefresh bool) (*ScrapeURLsResult, error) {
	// Create a scraping job to track progress
	job := &models.KnowledgeScrapingJob{
		ID:         uuid.New(),
		TenantID:   tenantID,
		ProjectID:  projectID,
		URL:        urls[0], // Use first URL as primary
		MaxDepth:   0,       // No crawling
		Status:     "running",
		TotalPages: len(urls),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	now := time.Now()
	job.StartedAt = &now

	if err := s.knowledgeRepo.CreateScrapingJob(job); err != nil {
		return nil, fmt.Errorf("failed to create scraping job: %w", err)
	}

	result := &ScrapeURLsResult{
		JobID:     job.ID,
		TotalURLs: len(urls),
	}

	var pageIDs []uuid.UUID
	oneWeekAgo := time.Now().Add(-7 * 24 * time.Hour)

	for _, urlStr := range urls {
		// Normalize URL
		normalizedURL := normalizeURLForDeduplication(urlStr)

		// Check if page exists and was scraped recently
		existingPages, err := s.knowledgeRepo.GetExistingPagesByURL(ctx, normalizedURL)
		if err != nil {
			logger.GetTxLogger(ctx).Error().Err(err).Str("url", urlStr).Msg("Failed to check existing page")
			result.PagesFailed++
			continue
		}

		// If page exists and was scraped within 1 week, skip unless force_refresh
		if len(existingPages) > 0 && !forceRefresh {
			existingPage := existingPages[0]
			if existingPage.ScrapedAt.After(oneWeekAgo) {
				logger.GetTxLogger(ctx).Info().
					Str("url", urlStr).
					Time("scraped_at", existingPage.ScrapedAt).
					Msg("Skipping URL - scraped within last week")
				pageIDs = append(pageIDs, existingPage.ID)
				result.PagesSkipped++
				continue
			}
		}

		// Scrape the page content using Colly
		content, title, err := s.scrapePageWithColly(ctx, urlStr)
		if err != nil {
			logger.GetTxLogger(ctx).Error().Err(err).Str("url", urlStr).Msg("Failed to scrape page")
			result.PagesFailed++
			continue
		}

		// Calculate content hash
		hash := sha256.Sum256([]byte(content))
		contentHash := fmt.Sprintf("%x", hash)

		// Check if content changed for existing page
		if len(existingPages) > 0 {
			existingPage := existingPages[0]
			if existingPage.ContentHash != nil && *existingPage.ContentHash == contentHash {
				// Content unchanged, just update scraped_at timestamp
				logger.GetTxLogger(ctx).Info().
					Str("url", urlStr).
					Msg("Content unchanged, reusing existing embedding")
				pageIDs = append(pageIDs, existingPage.ID)
				result.PagesSkipped++
				continue
			}
		}

		// Generate embedding for new/changed content
		embeddings, err := s.embeddingService.GenerateEmbeddings(ctx, []string{content})
		if err != nil {
			logger.GetTxLogger(ctx).Error().Err(err).Str("url", urlStr).Msg("Failed to generate embedding")
			result.PagesFailed++
			continue
		}

		embedding := embeddings[0]
		tokenCount := len(strings.Fields(content))

		// Store or update the page
		if len(existingPages) > 0 {
			// Update existing page
			existingPage := existingPages[0]
			if err := s.knowledgeRepo.UpdatePageContentAndEmbedding(ctx, existingPage.ID, title, content, contentHash, embedding, tokenCount); err != nil {
				logger.GetTxLogger(ctx).Error().Err(err).Str("url", urlStr).Msg("Failed to update page")
				result.PagesFailed++
				continue
			}
			pageIDs = append(pageIDs, existingPage.ID)
		} else {
			// Create new page
			page := &models.KnowledgeScrapedPage{
				ID:          uuid.New(),
				JobID:       uuid.NullUUID{UUID: job.ID, Valid: true},
				URL:         normalizedURL,
				Title:       &title,
				Content:     content,
				ContentHash: &contentHash,
				TokenCount:  tokenCount,
				ScrapedAt:   time.Now(),
				Embedding:   &embedding,
			}

			if err := s.knowledgeRepo.CreateScrapedPageWithTenantID(ctx, page, tenantID); err != nil {
				logger.GetTxLogger(ctx).Error().Err(err).Str("url", urlStr).Msg("Failed to create page")
				result.PagesFailed++
				continue
			}
			pageIDs = append(pageIDs, page.ID)
		}

		result.PagesAdded++
	}

	// Create project_knowledge_pages mappings
	if len(pageIDs) > 0 {
		if err := s.knowledgeRepo.CreateProjectKnowledgePageMappings(ctx, tenantID, projectID, pageIDs); err != nil {
			logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to create project-page mappings")
			// Don't fail the whole job for this
		}
	}

	// Update job progress and status
	pagesScraped := result.PagesAdded + result.PagesSkipped
	if err := s.knowledgeRepo.UpdateScrapingJobProgress(job.ID, pagesScraped, result.TotalURLs); err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to update job progress")
	}

	if err := s.knowledgeRepo.UpdateScrapingJobStatus(job.ID, "completed", nil); err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to update job status")
	}

	return result, nil
}

// scrapePageWithColly scrapes a single page and returns content and title
func (s *WebScrapingService) scrapePageWithColly(ctx context.Context, urlStr string) (string, string, error) {
	var content, title string
	var metaContent []string
	var scrapeErr error

	c := colly.NewCollector(
		colly.UserAgent(s.config.ScrapeUserAgent),
	)

	c.SetRequestTimeout(30 * time.Second)

	c.OnHTML("html", func(e *colly.HTMLElement) {
		// Get title
		title = e.ChildText("title")
		if title == "" {
			title = e.ChildText("h1")
		}

		// Extract meta tags for better content extraction
		e.ForEach("meta", func(_ int, el *colly.HTMLElement) {
			name := el.Attr("name")
			property := el.Attr("property")
			content := el.Attr("content")

			if content == "" {
				return
			}

			// Standard meta tags
			switch name {
			case "description", "keywords", "author":
				metaContent = append(metaContent, content)
			}

			// Open Graph tags
			switch property {
			case "og:title", "og:description", "og:site_name":
				metaContent = append(metaContent, content)
			}

			// Twitter tags
			switch name {
			case "twitter:title", "twitter:description":
				metaContent = append(metaContent, content)
			}
		})

		// Extract text content from body
		bodyContent := s.extractTextContent(e)

		// Combine meta content with body content
		var allContent []string
		if len(metaContent) > 0 {
			allContent = append(allContent, strings.Join(metaContent, " "))
		}
		if bodyContent != "" {
			allContent = append(allContent, bodyContent)
		}

		content = strings.Join(allContent, "\n\n")
	})

	c.OnError(func(r *colly.Response, err error) {
		scrapeErr = fmt.Errorf("scraping failed: %w", err)
	})

	if err := c.Visit(urlStr); err != nil {
		return "", "", fmt.Errorf("failed to visit URL: %w", err)
	}

	if scrapeErr != nil {
		return "", "", scrapeErr
	}

	if content == "" {
		return "", "", fmt.Errorf("no content extracted from page")
	}

	return content, title, nil
}
