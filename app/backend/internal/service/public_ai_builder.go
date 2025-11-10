package service

import (
	"context"
	"fmt"
	"net/url"
	"os"
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
var PublicTenantID = uuid.MustParse("7fb17866-be1f-11f0-bc68-bac7f8788f8c")

// PublicProjectID is the fixed project ID for all public AI widget builder projects
var PublicProjectID = uuid.MustParse("7fb1ea91-be1f-11f0-bc68-bac7f8788f8c")

// PublicAIBuilderService handles public AI widget builder requests
type PublicAIBuilderService struct {
	projectRepo        repo.ProjectRepository
	chatWidgetRepo     *repo.ChatWidgetRepo
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
	TotalURLs        int
	ScrapedPages     int
	FailedScrapes    int
	EmbeddedPages    int
	FailedEmbeddings int
	StoredPages      int
	FailedStores     int
}

// NewPublicAIBuilderService creates a new public AI builder service
func NewPublicAIBuilderService(
	projectRepo repo.ProjectRepository,
	chatWidgetRepo *repo.ChatWidgetRepo,
	aiBuilderService *AIBuilderService,
	webScrapingService *WebScrapingService,
) *PublicAIBuilderService {
	return &PublicAIBuilderService{
		projectRepo:        projectRepo,
		chatWidgetRepo:     chatWidgetRepo,
		aiBuilderService:   aiBuilderService,
		webScrapingService: webScrapingService,
	}
}

// processURLsInParallel processes a list of URLs through the new pipeline:
// Stage 1: URLs → Scrape to text files (parallel)
// Stage 2: AI ranking to select top 8 URLs
// Stage 3: Read top 8 → Embed (batch) → Store in pgvector (parallel)
func (s *PublicAIBuilderService) processURLsInParallel(
	ctx context.Context,
	tenantID, projectID uuid.UUID,
	widgetID, buildID uuid.UUID,
	urls []string,
	widget *models.ChatWidget,
	events chan<- AIBuilderEvent,
) error {
	if len(urls) == 0 {
		return fmt.Errorf("no URLs to process")
	}

	// Create a scraping job for tracking
	job, err := s.webScrapingService.CreateScrapingJob(ctx, tenantID, projectID, &models.CreateScrapingJobRequest{
		URL:      urls[0], // Use first URL as the target
		MaxDepth: 2,
	})
	if err != nil {
		return fmt.Errorf("failed to create scraping job: %w", err)
	}

	// Create working directory for this build
	workDir := fmt.Sprintf("/tmp/widgets/%s", buildID.String())
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("failed to create working directory: %w", err)
	}

	logger.GetTxLogger(ctx).Info().
		Str("build_id", buildID.String()).
		Str("work_dir", workDir).
		Int("total_urls", len(urls)).
		Msg("Starting 3-stage parallel pipeline")

	// STAGE 1: Scrape all URLs and save to text files
	s.emit(ctx, events, AIBuilderEvent{
		Type:    "stage_1_started",
		Stage:   "scraping",
		Message: fmt.Sprintf("Stage 1: Scraping %d URLs to text files", len(urls)),
		Data: map[string]any{
			"total_urls": len(urls),
			"workers":    10,
		},
	})

	scrapedFiles, err := s.scrapeURLsToFiles(ctx, urls, workDir, events)
	if err != nil {
		return fmt.Errorf("stage 1 failed: %w", err)
	}

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "stage_1_completed",
		Stage:   "scraping",
		Message: fmt.Sprintf("Stage 1: Successfully scraped %d/%d URLs", len(scrapedFiles), len(urls)),
		Data: map[string]any{
			"scraped": len(scrapedFiles),
			"total":   len(urls),
			"failed":  len(urls) - len(scrapedFiles),
		},
	})

	if len(scrapedFiles) == 0 {
		return fmt.Errorf("no URLs were successfully scraped")
	}

	// STAGE 2: Use AI to rank and select top 8 URLs
	s.emit(ctx, events, AIBuilderEvent{
		Type:    "stage_2_started",
		Stage:   "ai_ranking",
		Message: fmt.Sprintf("Stage 2: Using AI to select top 8 from %d URLs", len(scrapedFiles)),
	})

	top8URLs, err := s.rankURLsWithAI(ctx, scrapedFiles, workDir, widget, events)
	if err != nil {
		return fmt.Errorf("stage 2 failed: %w", err)
	}

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "stage_2_completed",
		Stage:   "ai_ranking",
		Message: fmt.Sprintf("Stage 2: Selected top 8 most relevant URLs"),
		Data: map[string]any{
			"selected_urls": top8URLs,
		},
	})

	// STAGE 3: Embed and store only top 8 URLs
	s.emit(ctx, events, AIBuilderEvent{
		Type:    "stage_3_started",
		Stage:   "embedding_storage",
		Message: fmt.Sprintf("Stage 3: Generating embeddings and storing top 8 URLs"),
	})

	err = s.embedAndStoreTop8(ctx, tenantID, projectID, widgetID, job.ID, top8URLs, workDir, events)
	if err != nil {
		return fmt.Errorf("stage 3 failed: %w", err)
	}

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "stage_3_completed",
		Stage:   "embedding_storage",
		Message: "Stage 3: Successfully stored all 8 URLs in knowledge base",
		Data: map[string]any{
			"stored_count": len(top8URLs),
		},
	})

	logger.GetTxLogger(ctx).Info().
		Str("build_id", buildID.String()).
		Int("total_processed", len(top8URLs)).
		Msg("3-stage parallel pipeline completed successfully")

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

	// Generate unique build ID for this build session
	buildID := uuid.New()

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

	// Check for cached widget created within the last week (if repo is available)
	if s.chatWidgetRepo != nil {
		widgets, err := s.chatWidgetRepo.ListChatWidgets(ctx, PublicTenantID, PublicProjectID)
		if err == nil && len(widgets) > 0 {
			// Check if any widget was created within the last week
			oneWeekAgo := time.Now().Add(-7 * 24 * time.Hour)
			for _, widget := range widgets {
				if widget.CreatedAt.After(oneWeekAgo) {
					// Found a recent widget, reuse it
					logger.GetTxLogger(ctx).Info().
						Str("widget_id", widget.ID.String()).
						Str("domain", domain).
						Time("widget_created_at", widget.CreatedAt).
						Msg("Reusing cached widget from last week")

					s.emit(ctx, events, AIBuilderEvent{
						Type:    "cache_hit",
						Stage:   "initialization",
						Message: fmt.Sprintf("Found existing widget for %s, reusing it", domain),
						Data: map[string]any{
							"widget_id":         widget.ID.String(),
							"widget_name":       widget.Name,
							"widget_created_at": widget.CreatedAt,
						},
					})

					s.emit(ctx, events, AIBuilderEvent{
						Type:    "completed",
						Stage:   "completion",
						Message: "AI widget retrieved from cache",
						Data: map[string]any{
							"project_id": PublicProjectID.String(),
							"widget_id":  widget.ID.String(),
							"cached":     true,
						},
					})

					return PublicProjectID, nil
				}
			}
		}
	}

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "builder_started",
		Stage:   "initialization",
		Message: fmt.Sprintf("Starting public AI widget builder for %s (build ID: %s)", domain, buildID.String()),
		Data: map[string]any{
			"domain":     domain,
			"url":        targetURL,
			"tenant_id":  PublicTenantID.String(),
			"project_id": PublicProjectID.String(),
			"build_id":   buildID.String(),
		},
	})

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "project_ready",
		Stage:   "initialization",
		Message: fmt.Sprintf("Using public project with ID: %s", PublicProjectID),
		Data: map[string]any{
			"project_id": PublicProjectID.String(),
			"tenant_id":  PublicTenantID.String(),
			"build_id":   buildID.String(),
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
		if err := s.webScrapingService.ExtractURLsWithStream(ctx, targetURL, depth, urlEvents, sharedBrowser, buildID.String()); err != nil {
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

	// Steps 3-7: Process URLs in parallel (scrape to files → AI rank top 8 → embed → store)
	err = s.processURLsInParallel(ctx, PublicTenantID, PublicProjectID, widget.ID, buildID, extractedURLs, widget, events)
	if err != nil {
		logger.GetTxLogger(ctx).Error().
			Str("project_id", PublicProjectID.String()).
			Err(err).
			Msg("Parallel processing failed")
		return PublicProjectID, fmt.Errorf("parallel processing failed: %w", err)
	}

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "completed",
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
	return s.webScrapingService.ExtractURLsWithStream(ctx, targetURL, depth, events, nil, "") // No build ID for debug
}
