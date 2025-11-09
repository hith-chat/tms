package service

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"

	"github.com/bareuptime/tms/internal/db"
	"github.com/bareuptime/tms/internal/logger"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
)

// PublicTenantID is the fixed tenant ID for all public AI widget builder projects
var PublicTenantID = uuid.MustParse("79634696-84ca-11f0-b71d-063b1faa412d")

// PublicProjectID is the fixed project ID for all public AI widget builder projects
var PublicProjectID = uuid.MustParse("ec33f64a-3560-4bbe-80aa-f97c5409c9f0")

// PublicAIBuilderService handles public AI widget builder requests
type PublicAIBuilderService struct {
	projectRepo        repo.ProjectRepository
	aiBuilderService   *AIBuilderService
	webScrapingService *WebScrapingService
}

// parallelPageJob represents a URL to be processed in the parallel pipeline
type parallelPageJob struct {
	URL   string
	Title string
	Depth int
}

// parallelScrapedPage represents scraped content ready for embedding
type parallelScrapedPage struct {
	Job     parallelPageJob
	Content string
	Error   error
}

// parallelEmbeddedPage represents embedded content ready for storage
type parallelEmbeddedPage struct {
	Job       parallelPageJob
	Content   string
	Embedding pgvector.Vector
	Error     error
}

// parallelProcessingStats tracks parallel processing metrics
type parallelProcessingStats struct {
	TotalURLs       int
	ScrapedPages    int
	FailedScrapes   int
	EmbeddedPages   int
	FailedEmbeddings int
	StoredPages     int
	FailedStores    int
}

// NewPublicAIBuilderService creates a new public AI builder service
func NewPublicAIBuilderService(
	projectRepo repo.ProjectRepository,
	aiBuilderService *AIBuilderService,
	webScrapingService *WebScrapingService,
) *PublicAIBuilderService {
	return &PublicAIBuilderService{
		projectRepo:        projectRepo,
		aiBuilderService:   aiBuilderService,
		webScrapingService: webScrapingService,
	}
}

// processURLsInParallel processes a list of URLs through the parallel pipeline:
// URLs → Scrape (parallel) → Embed (parallel) → Store (parallel)
func (s *PublicAIBuilderService) processURLsInParallel(
	ctx context.Context,
	tenantID, projectID uuid.UUID,
	urls []string,
	targetURL string,
	depth int,
	events chan<- AIBuilderEvent,
) error {
	if len(urls) == 0 {
		return fmt.Errorf("no URLs to process")
	}

	// Create a scraping job for tracking
	job, err := s.webScrapingService.CreateScrapingJob(ctx, tenantID, projectID, &models.CreateScrapingJobRequest{
		URL:      targetURL,
		MaxDepth: depth,
	})
	if err != nil {
		return fmt.Errorf("failed to create scraping job: %w", err)
	}

	// Create job list
	jobs := make([]parallelPageJob, 0, len(urls))
	for _, url := range urls {
		jobs = append(jobs, parallelPageJob{
			URL:   url,
			Title: url, // Will be updated after scraping
			Depth: 0,
		})
	}

	stats := &parallelProcessingStats{
		TotalURLs: len(jobs),
	}

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "parallel_processing_started",
		Stage:   "knowledge_building",
		Message: fmt.Sprintf("Starting parallel processing of %d URLs", len(jobs)),
		Data: map[string]any{
			"total_urls": len(jobs),
			"workers":    map[string]int{
				"scraping":  10,
				"embedding": 5,
				"storage":   3,
			},
		},
	})

	// Create pipeline channels
	scrapedChan := make(chan parallelScrapedPage, 100)
	embeddedChan := make(chan parallelEmbeddedPage, 100)

	// Stage 1: Parallel Scraping Workers (10 workers)
	scrapingDone := make(chan struct{})
	go s.runScrapingWorkers(ctx, jobs, scrapedChan, stats, events, scrapingDone, 10)

	// Stage 2: Parallel Embedding Workers (5 workers)
	embeddingDone := make(chan struct{})
	go s.runEmbeddingWorkers(ctx, scrapedChan, embeddedChan, stats, events, embeddingDone, 5)

	// Stage 3: Parallel Storage Workers (3 workers)
	storageDone := make(chan struct{})
	go s.runStorageWorkers(ctx, tenantID, projectID, job.ID, embeddedChan, stats, events, storageDone, 3)

	// Wait for all stages to complete
	<-scrapingDone
	<-embeddingDone
	<-storageDone

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "parallel_processing_completed",
		Stage:   "knowledge_building",
		Message: fmt.Sprintf("Parallel processing completed. Stored %d/%d pages", stats.StoredPages, stats.TotalURLs),
		Data: map[string]any{
			"total_urls":        stats.TotalURLs,
			"scraped_pages":     stats.ScrapedPages,
			"failed_scrapes":    stats.FailedScrapes,
			"embedded_pages":    stats.EmbeddedPages,
			"failed_embeddings": stats.FailedEmbeddings,
			"stored_pages":      stats.StoredPages,
			"failed_stores":     stats.FailedStores,
		},
	})

	if stats.StoredPages == 0 {
		return fmt.Errorf("failed to process any pages successfully")
	}

	return nil
}

// runScrapingWorkers spawns parallel workers to scrape content from URLs
func (s *PublicAIBuilderService) runScrapingWorkers(
	ctx context.Context,
	jobs []parallelPageJob,
	output chan<- parallelScrapedPage,
	stats *parallelProcessingStats,
	events chan<- AIBuilderEvent,
	done chan<- struct{},
	workerCount int,
) {
	defer close(output)
	defer close(done)

	jobChan := make(chan parallelPageJob, len(jobs))
	for _, job := range jobs {
		jobChan <- job
	}
	close(jobChan)

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for job := range jobChan {
				// Scrape content
				content, err := s.webScrapingService.ScrapePageContent(ctx, job.URL)

				result := parallelScrapedPage{
					Job:     job,
					Content: content,
					Error:   err,
				}

				if err != nil {
					stats.FailedScrapes++
					s.emit(ctx, events, AIBuilderEvent{
						Type:    "scraping_failed",
						Stage:   "scraping",
						Message: fmt.Sprintf("Failed to scrape %s", job.URL),
						Detail:  err.Error(),
					})
				} else {
					stats.ScrapedPages++
					if stats.ScrapedPages%10 == 0 {
						s.emit(ctx, events, AIBuilderEvent{
							Type:    "scraping_progress",
							Stage:   "scraping",
							Message: fmt.Sprintf("Scraped %d/%d pages", stats.ScrapedPages, stats.TotalURLs),
						})
					}
				}

				output <- result
			}
		}(i)
	}

	wg.Wait()
}

// runEmbeddingWorkers spawns parallel workers to create embeddings
func (s *PublicAIBuilderService) runEmbeddingWorkers(
	ctx context.Context,
	input <-chan parallelScrapedPage,
	output chan<- parallelEmbeddedPage,
	stats *parallelProcessingStats,
	events chan<- AIBuilderEvent,
	done chan<- struct{},
	workerCount int,
) {
	defer close(output)
	defer close(done)

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for scraped := range input {
				// Skip if scraping failed
				if scraped.Error != nil {
					continue
				}

				// Generate embedding
				embedding, err := s.webScrapingService.embeddingService.GenerateEmbedding(ctx, scraped.Content)

				result := parallelEmbeddedPage{
					Job:       scraped.Job,
					Content:   scraped.Content,
					Embedding: embedding,
					Error:     err,
				}

				if err != nil {
					stats.FailedEmbeddings++
					s.emit(ctx, events, AIBuilderEvent{
						Type:    "embedding_failed",
						Stage:   "embedding",
						Message: fmt.Sprintf("Failed to create embedding for %s", scraped.Job.URL),
						Detail:  err.Error(),
					})
				} else {
					stats.EmbeddedPages++
					if stats.EmbeddedPages%10 == 0 {
						s.emit(ctx, events, AIBuilderEvent{
							Type:    "embedding_progress",
							Stage:   "embedding",
							Message: fmt.Sprintf("Created embeddings for %d/%d pages", stats.EmbeddedPages, stats.ScrapedPages),
						})
					}
				}

				output <- result
			}
		}(i)
	}

	wg.Wait()
}

// runStorageWorkers spawns parallel workers to store pages in vector database
func (s *PublicAIBuilderService) runStorageWorkers(
	ctx context.Context,
	tenantID, projectID, jobID uuid.UUID,
	input <-chan parallelEmbeddedPage,
	stats *parallelProcessingStats,
	events chan<- AIBuilderEvent,
	done chan<- struct{},
	workerCount int,
) {
	defer close(done)

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for embedded := range input {
				// Skip if embedding failed
				if embedded.Error != nil {
					continue
				}

				// Store in vector database
				err := s.webScrapingService.StorePageInVectorDB(ctx, tenantID, projectID, embedded.Job.URL, embedded.Content, embedded.Embedding, jobID)

				if err != nil {
					stats.FailedStores++
					s.emit(ctx, events, AIBuilderEvent{
						Type:    "storage_failed",
						Stage:   "storage",
						Message: fmt.Sprintf("Failed to store %s in vector DB", embedded.Job.URL),
						Detail:  err.Error(),
					})
				} else {
					stats.StoredPages++
					if stats.StoredPages%10 == 0 {
						s.emit(ctx, events, AIBuilderEvent{
							Type:    "storage_progress",
							Stage:   "storage",
							Message: fmt.Sprintf("Stored %d/%d pages in vector DB", stats.StoredPages, stats.EmbeddedPages),
						})
					}
				}
			}
		}(i)
	}

	wg.Wait()
}

// BuildPublicWidget creates a public AI widget for the given URL
// It uses the fixed public tenant and project IDs
// Returns streaming events via the events channel
func (s *PublicAIBuilderService) BuildPublicWidget(
	ctx context.Context,
	targetURL string,
	depth int,
	events chan<- AIBuilderEvent,
) (uuid.UUID, error) {
	if depth <= 0 {
		depth = 3 // Default depth
	}

	// Parse and validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "initialization",
			Message: "Invalid URL provided",
			Detail:  err.Error(),
		})
		return uuid.Nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "initialization",
			Message: "Invalid URL provided",
			Detail:  "URL must include both scheme (http/https) and host",
		})
		return uuid.Nil, fmt.Errorf("invalid URL: missing scheme or host")
	}

	if parsedURL.Scheme != "https" {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "initialization",
			Message: "Invalid URL provided",
			Detail:  "Only HTTPS URLs are supported",
		})
		return uuid.Nil, fmt.Errorf("invalid URL: only HTTPS URLs are supported")
	}

	domain := parsedURL.Host
	// Remove www. prefix for consistency
	domain = strings.TrimPrefix(domain, "www.")

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "builder_started",
		Stage:   "initialization",
		Message: fmt.Sprintf("Starting public AI widget builder for %s", domain),
		Data: map[string]any{
			"domain":     domain,
			"url":        targetURL,
			"tenant_id":  PublicTenantID.String(),
			"project_id": PublicProjectID.String(),
		},
	})

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "project_ready",
		Stage:   "initialization",
		Message: fmt.Sprintf("Using public project with ID: %s", PublicProjectID),
		Data: map[string]any{
			"project_id": PublicProjectID.String(),
			"tenant_id":  PublicTenantID.String(),
			"build_id":   uuid.New().String(),
		},
	})

	// Step 1: Build widget theme (uses existing AIBuilderService)
	s.emit(ctx, events, AIBuilderEvent{
		Type:    "theme_generation_started",
		Stage:   "theme",
		Message: "Generating chat widget theme from website",
	})

	widget, sharedBrowser, err := s.aiBuilderService.buildWidget(ctx, PublicTenantID, PublicProjectID, targetURL, events)
	if err != nil {
		logger.GetTxLogger(ctx).Error().
			Str("project_id", PublicProjectID.String()).
			Err(err).
			Msg("Widget theme generation failed")
		return PublicProjectID, fmt.Errorf("widget generation failed: %w", err)
	}

	// Ensure browser is cleaned up when done
	defer func() {
		if sharedBrowser != nil {
			sharedBrowser.Close()
		}
	}()

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "theme_generation_completed",
		Stage:   "theme",
		Message: fmt.Sprintf("Widget theme created: %s", widget.Name),
		Data: map[string]any{
			"widget_id": widget.ID.String(),
		},
	})

	// Step 2: Extract all URLs using parallel extraction
	s.emit(ctx, events, AIBuilderEvent{
		Type:    "url_extraction_started",
		Stage:   "url_extraction",
		Message: fmt.Sprintf("Extracting URLs from %s (depth: %d)", targetURL, depth),
	})

	urlEvents := make(chan URLExtractionEvent, 100)
	var extractedURLs []string

	// Run URL extraction in background and collect results
	go func() {
		if err := s.webScrapingService.ExtractURLsWithStream(ctx, targetURL, depth, urlEvents, sharedBrowser); err != nil {
			logger.GetTxLogger(ctx).Error().Err(err).Msg("URL extraction failed")
		}
	}()

	// Collect extracted URLs
	for event := range urlEvents {
		// Forward URL extraction events to main event stream
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "url_extraction_progress",
			Stage:   "url_extraction",
			Message: event.Message,
			Data: map[string]any{
				"event_type":    event.Type,
				"links_found":   event.LinksFound,
				"current_depth": event.CurrentDepth,
			},
		})

		// Collect final URL list
		if event.Type == "completed" && len(event.URLs) > 0 {
			extractedURLs = event.URLs
		}
	}

	if len(extractedURLs) == 0 {
		err := fmt.Errorf("no URLs were extracted from the website")
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "url_extraction",
			Message: err.Error(),
		})
		return PublicProjectID, err
	}

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "url_extraction_completed",
		Stage:   "url_extraction",
		Message: fmt.Sprintf("Extracted %d URLs", len(extractedURLs)),
		Data: map[string]any{
			"total_urls": len(extractedURLs),
		},
	})

	// Steps 3, 4, 5: Process URLs in parallel (scrape → embed → store)
	err = s.processURLsInParallel(ctx, PublicTenantID, PublicProjectID, extractedURLs, targetURL, depth, events)
	if err != nil {
		logger.GetTxLogger(ctx).Error().
			Str("project_id", PublicProjectID.String()).
			Err(err).
			Msg("Parallel processing failed")
		return PublicProjectID, fmt.Errorf("parallel processing failed: %w", err)
	}

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "builder_completed",
		Stage:   "completion",
		Message: "AI widget successfully built and deployed",
		Data: map[string]any{
			"project_id": PublicProjectID.String(),
			"widget_id":  widget.ID.String(),
			"total_urls": len(extractedURLs),
		},
	})

	return PublicProjectID, nil
}

// createPublicProject creates a new public project for the given domain
func (s *PublicAIBuilderService) createPublicProject(
	ctx context.Context,
	domain string,
	targetURL string,
	events chan<- AIBuilderEvent,
) (*db.Project, error) {
	s.emit(ctx, events, AIBuilderEvent{
		Type:    "project_creation_started",
		Stage:   "initialization",
		Message: "Creating public project",
	})

	// Generate project key from domain
	// e.g., "example.com" -> "EXAMPLE-COM-PUBLIC"
	projectKey := s.generateProjectKey(domain)

	// Generate project name
	projectName := fmt.Sprintf("%s Public Widget", domain)

	// Calculate expiration time (6 hours from now)
	expiresAt := time.Now().Add(6 * time.Hour)

	project := &db.Project{
		ID:        uuid.New(),
		TenantID:  PublicTenantID,
		Key:       projectKey,
		Name:      projectName,
		Status:    "active",
		IsPublic:  true,
		ExpiresAt: &expiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.projectRepo.Create(ctx, project)
	if err != nil {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "initialization",
			Message: "Failed to create public project",
			Detail:  err.Error(),
		})
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	logger.GetTxLogger(ctx).Info().
		Str("project_id", project.ID.String()).
		Str("domain", domain).
		Time("expires_at", expiresAt).
		Msg("Created public project")

	return project, nil
}

// generateProjectKey generates a project key from a domain
// e.g., "example.com" -> "EXAMPLE-COM-PUBLIC"
// e.g., "docs.anthropic.com" -> "DOCS-ANTHROPIC-COM-PUBLIC"
func (s *PublicAIBuilderService) generateProjectKey(domain string) string {
	// Remove any port numbers
	domain = strings.Split(domain, ":")[0]

	// Replace dots with dashes
	key := strings.ReplaceAll(domain, ".", "-")

	// Remove any non-alphanumeric characters except dashes
	reg := regexp.MustCompile("[^a-zA-Z0-9-]+")
	key = reg.ReplaceAllString(key, "")

	// Convert to uppercase and add PUBLIC suffix
	key = strings.ToUpper(key) + "-PUBLIC"

	// Ensure key doesn't exceed 50 characters (database limit)
	if len(key) > 50 {
		key = key[:47] + "PUB" // Truncate and ensure PUBLIC suffix
	}

	return key
}

// emit sends an event with a timestamp
func (s *PublicAIBuilderService) emit(ctx context.Context, events chan<- AIBuilderEvent, event AIBuilderEvent) {
	event.Timestamp = time.Now()
	select {
	case <-ctx.Done():
	case events <- event:
	}
}

// ExtractURLsDebug extracts all URLs from a website for debugging purposes
func (s *PublicAIBuilderService) ExtractURLsDebug(ctx context.Context, targetURL string, depth int, events chan<- URLExtractionEvent) error {
	return s.webScrapingService.ExtractURLsWithStream(ctx, targetURL, depth, events, nil)
}
