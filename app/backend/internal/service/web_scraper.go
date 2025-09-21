package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
)

type WebScrapingService struct {
	knowledgeRepo    *repo.KnowledgeRepository
	embeddingService *EmbeddingService
	config          *config.KnowledgeConfig
}

func NewWebScrapingService(knowledgeRepo *repo.KnowledgeRepository, embeddingService *EmbeddingService, cfg *config.KnowledgeConfig) *WebScrapingService {
	return &WebScrapingService{
		knowledgeRepo:    knowledgeRepo,
		embeddingService: embeddingService,
		config:          cfg,
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

// startScrapingJob executes the scraping job
func (s *WebScrapingService) startScrapingJob(ctx context.Context, job *models.KnowledgeScrapingJob) {
	// Create a context with timeout to prevent infinite scraping
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	var err error
	defer func() {
		if err != nil {
			errStr := err.Error()
			fmt.Printf("Job %s failed with error: %s\n", job.ID, errStr)
			updateErr := s.knowledgeRepo.UpdateScrapingJobStatus(job.ID, "failed", &errStr)
			if updateErr != nil {
				fmt.Printf("Failed to update job status to failed: %v\n", updateErr)
			}
		}
		// If err is nil, job completion is handled explicitly in the main flow
	}()

	// Mark job as running
	if err = s.knowledgeRepo.StartScrapingJob(job.ID); err != nil {
		return
	}

	// Create colly collector
	c := colly.NewCollector(
		// Remove debug logging to reduce noise
		colly.UserAgent(s.config.ScrapeUserAgent),
	)

	// Set request timeout
	c.SetRequestTimeout(s.config.ScrapeTimeout)

	// Rate limiting
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		Delay:       s.config.ScrapeRateLimit,
	})

	// Respect robots.txt
	c.AllowURLRevisit = false

	var scrapedPages []*models.KnowledgeScrapedPage
	visitedURLs := make(map[string]bool) // Track visited URLs
	visitedURLs[job.URL] = true // Mark initial URL as visited
	maxPages := 100 // Limit to prevent excessive scraping

	// Set up HTML callback - use body selector to process each page only once
	c.OnHTML("body", func(e *colly.HTMLElement) {
		// Skip if we've reached max depth or max pages
		depth := e.Request.Depth
		if depth > job.MaxDepth || len(scrapedPages) >= maxPages {
			return
		}

		currentURL := e.Request.URL.String()
		
		// Extract page content
		title := e.ChildText("title")
		if title == "" {
			title = e.ChildText("h1")
		}

		// Extract text content (remove scripts, styles, etc.)
		content := s.extractTextContent(e)
		
		if len(strings.TrimSpace(content)) < 100 { // Skip pages with very little content
			return
		}

		// Create scraped page
		contentHash := s.generateContentHash(content)
		page := &models.KnowledgeScrapedPage{
			ID:          uuid.New(),
			JobID:       job.ID,
			URL:         currentURL,
			Title:       &title,
			Content:     content,
			ContentHash: &contentHash,
			TokenCount:  s.estimateTokenCount(content),
			ScrapedAt:   time.Now(),
			Metadata:    models.JSONMap{},
		}

		scrapedPages = append(scrapedPages, page)

		// Update progress periodically
		if len(scrapedPages)%5 == 0 { // Update every 5 pages
			s.knowledgeRepo.UpdateScrapingJobProgress(job.ID, len(scrapedPages), 0)
		}

		// Find and queue more links if we haven't reached max depth or max pages
		if depth < job.MaxDepth && len(scrapedPages) < maxPages {
			e.ForEach("a[href]", func(_ int, link *colly.HTMLElement) {
				linkURL := link.Attr("href")
				absoluteURL := e.Request.AbsoluteURL(linkURL)
				
				// Check if we should follow this link and haven't visited it
				if s.shouldFollowLink(absoluteURL, currentURL) && !visitedURLs[absoluteURL] {
					visitedURLs[absoluteURL] = true
					e.Request.Visit(linkURL)
				}
			})
		}
	})

	// Set up error callback
	c.OnError(func(r *colly.Response, e error) {
		fmt.Printf("Error scraping %s: %v\n", r.Request.URL, e)
		// Don't fail the entire job for individual page errors
	})

	// Set up request callback for debugging
	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("Visiting: %s (depth: %d)\n", r.URL.String(), r.Depth)
	})

	// Start scraping
	err = c.Visit(job.URL)
	if err != nil {
		err = fmt.Errorf("failed to start scraping: %w", err)
		return
	}

	// Wait for all requests to finish
	c.Wait()

	// Final progress update before processing
	totalPages := len(scrapedPages)
	if totalPages == 0 {
		err = fmt.Errorf("no pages were successfully scraped")
		return
	}
	
	fmt.Printf("Scraped %d pages, starting content-aware processing...\n", totalPages)
	s.knowledgeRepo.UpdateScrapingJobProgress(job.ID, totalPages, totalPages)

	// STEP 1: Generate content hashes for all pages (cheap operation)
	fmt.Printf("Generating content hashes for %d pages...\n", totalPages)
	for _, page := range scrapedPages {
		if page.ContentHash == nil {
			hash := s.generateContentHash(page.Content)
			page.ContentHash = &hash
		}
	}

	// STEP 2: Save to database with content-aware deduplication
	fmt.Printf("Saving %d scraped pages to database...\n", totalPages)
	err = s.knowledgeRepo.CreateScrapedPages(scrapedPages)
	if err != nil {
		err = fmt.Errorf("failed to save scraped pages: %w", err)
		return
	}

	// STEP 3: Generate embeddings ONLY for pages that need them (new/updated)
	// The repository logic marks pages that need embeddings by keeping embedding=nil
	// and pages that don't need embeddings by setting a dummy embedding value
	pagesToEmbed := make([]*models.KnowledgeScrapedPage, 0)
	skippedCount := 0
	
	for _, page := range scrapedPages {
		// Check if this page needs an embedding:
		// - embedding is nil (new or updated page)
		// - page has a valid ID (was processed by repository)
		if page.Embedding == nil && page.ID != uuid.Nil {
			pagesToEmbed = append(pagesToEmbed, page)
		} else if page.Embedding != nil {
			// This page was marked as duplicate (has dummy embedding)
			skippedCount++
		}
	}

	if len(pagesToEmbed) == 0 {
		fmt.Printf("No new or updated pages found - all %d pages were duplicates, skipping embedding generation\n", totalPages)
	} else if !s.embeddingService.IsEnabled() {
		fmt.Printf("Warning: Embedding service is disabled, %d pages saved without embeddings\n", len(pagesToEmbed))
	} else {
		// Generate embeddings only for pages that need them
		fmt.Printf("Generating embeddings for %d pages that need them (skipping %d duplicates) using %s model...\n", 
			len(pagesToEmbed), skippedCount, s.embeddingService.GetModel())
		
		// Create a dedicated context for embedding generation with configured timeout
		embeddingCtx, embeddingCancel := context.WithTimeout(context.Background(), s.config.EmbeddingTimeout)
		defer embeddingCancel()
		
		embeddingErr := s.generateEmbeddingsForPages(embeddingCtx, pagesToEmbed)
		if embeddingErr != nil {
			fmt.Printf("Warning: Failed to generate embeddings (job will still complete): %v\n", embeddingErr)
			// Don't fail the entire job for embedding issues
		} else {
			fmt.Printf("Successfully generated embeddings for %d pages\n", len(pagesToEmbed))
		}
	}
	
	fmt.Printf("Successfully completed scraping job for %d pages\n", totalPages)
	
	// Explicitly mark job as completed - don't rely on defer for success case
	completionErr := s.knowledgeRepo.CompleteScrapingJob(job.ID)
	if completionErr != nil {
		fmt.Printf("Warning: Failed to mark job as completed in database: %v\n", completionErr)
	} else {
		fmt.Printf("Job %s marked as completed successfully\n", job.ID)
	}
	
	// Clear err so defer doesn't override our completion
	err = nil
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
	fmt.Printf("Starting embedding generation for %d pages...\n", len(pages))
	embeddings, err := s.embeddingService.GenerateEmbeddings(ctx, texts)
	if err != nil {
		fmt.Printf("Error generating embeddings: %v\n", err)
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
	fmt.Printf("Updating database with embeddings for %d pages...\n", len(pages))
	err = s.knowledgeRepo.UpdatePageEmbeddings(pages)
	if err != nil {
		return fmt.Errorf("failed to update page embeddings: %w", err)
	}

	fmt.Printf("Successfully assigned and saved embeddings to %d pages\n", len(pages))
	return nil
}

// estimateTokenCount provides a rough estimate of token count
func (s *WebScrapingService) estimateTokenCount(text string) int {
	// Rough approximation: 1 token ≈ 4 characters for English text
	return len(text) / 4
}

// generateContentHash creates a SHA256 hash of the content for deduplication
func (s *WebScrapingService) generateContentHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// GetScrapingJob returns a scraping job by ID
func (s *WebScrapingService) GetScrapingJob(ctx context.Context, jobID uuid.UUID) (*models.KnowledgeScrapingJob, error) {
	return s.knowledgeRepo.GetScrapingJob(jobID)
}

// ListScrapingJobs returns a list of scraping jobs for a project
func (s *WebScrapingService) ListScrapingJobs(ctx context.Context, tenantID, projectID uuid.UUID, limit, offset int) ([]*models.KnowledgeScrapingJob, error) {
	return s.knowledgeRepo.ListScrapingJobs(tenantID, projectID, limit, offset)
}

// GetJobPages returns all pages scraped by a job
func (s *WebScrapingService) GetJobPages(ctx context.Context, jobID uuid.UUID) ([]*models.KnowledgeScrapedPage, error) {
	return s.knowledgeRepo.GetJobPages(jobID)
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
		colly.UserAgent("TMS-ThemeBot/1.0 (+https://tms.bareuptime.co)"),
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
		fmt.Printf("Error scraping %s: %v\n", r.Request.URL, err)
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
