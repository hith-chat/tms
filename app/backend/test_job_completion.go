package main
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/db"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/bareuptime/tms/internal/service"
)

func main() {
	fmt.Println("=== Testing Web Scraper Job Completion Status ===")
	fmt.Println()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	database, err := db.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Create repositories and services
	knowledgeRepo := repo.NewKnowledgeRepository(database)
	embeddingService := service.NewEmbeddingService(&cfg.Knowledge)
	webScraperService := service.NewWebScrapingService(knowledgeRepo, embeddingService, &cfg.Knowledge)

	ctx := context.Background()
	
	// Test 1: Successful scraping should update job to "completed"
	fmt.Println("🔍 Test 1: Successful Scraping → Job Status 'completed'")
	testSuccessfulScraping(ctx, knowledgeRepo, webScraperService)
	
	fmt.Println()
	
	// Test 2: Invalid URL should update job to "failed"
	fmt.Println("🔍 Test 2: Invalid URL → Job Status 'failed'")
	testInvalidURLScraping(ctx, knowledgeRepo, webScraperService)
	
	fmt.Println()
	
	// Test 3: Network error should update job appropriately
	fmt.Println("🔍 Test 3: Network Error → Job Status handling")
	testNetworkErrorScraping(ctx, knowledgeRepo, webScraperService)

	fmt.Println()
	fmt.Println("=== All Job Completion Tests Complete ===")
}

func testSuccessfulScraping(ctx context.Context, repo *repo.KnowledgeRepository, scraper *service.WebScrapingService) {
	// Create test job
	tenantID := uuid.New()
	projectID := uuid.New()
	testURL := "https://httpbin.org/html" // Simple HTML endpoint that should work

	job := &models.KnowledgeScrapingJob{
		ID:           uuid.New(),
		TenantID:     tenantID,
		ProjectID:    projectID,
		URL:          testURL,
		MaxDepth:     1,
		Status:       "pending",
		PagesScraped: 0,
		TotalPages:   0,
		StartedAt:    nil,
		CompletedAt:  nil,
		ErrorMessage: nil,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save initial job
	err := repo.CreateScrapingJob(job)
	if err != nil {
		log.Printf("❌ Failed to create job: %v", err)
		return
	}

	fmt.Printf("✅ Created job %s with initial status: %s\n", job.ID, job.Status)

	// Verify initial status
	initialJob, err := repo.GetScrapingJob(job.ID)
	if err != nil {
		log.Printf("❌ Failed to get initial job: %v", err)
		return
	}
	fmt.Printf("✅ Initial job status: %s, pages: %d\n", initialJob.Status, initialJob.PagesScraped)

	// Start scraping
	fmt.Printf("🚀 Starting scraping for URL: %s\n", testURL)
	err = scraper.StartScraping(ctx, job.ID, testURL, 1)
	if err != nil {
		log.Printf("❌ Scraping failed: %v", err)
	} else {
		fmt.Printf("✅ Scraping completed without error\n")
	}

	// Check final job status
	finalJob, err := repo.GetScrapingJob(job.ID)
	if err != nil {
		log.Printf("❌ Failed to get final job: %v", err)
		return
	}

	fmt.Printf("📊 Final Job Status Report:\n")
	fmt.Printf("   Status: %s\n", finalJob.Status)
	fmt.Printf("   Pages Scraped: %d\n", finalJob.PagesScraped)
	fmt.Printf("   Started At: %v\n", finalJob.StartedAt)
	fmt.Printf("   Completed At: %v\n", finalJob.CompletedAt)
	if finalJob.ErrorMessage != nil {
		fmt.Printf("   Error: %s\n", *finalJob.ErrorMessage)
	}

	// Verify expectations
	if finalJob.Status == "completed" {
		fmt.Printf("✅ SUCCESS: Job correctly updated to 'completed'\n")
		if finalJob.PagesScraped > 0 {
			fmt.Printf("✅ SUCCESS: Pages scraped count updated (%d)\n", finalJob.PagesScraped)
		}
		if finalJob.CompletedAt != nil {
			fmt.Printf("✅ SUCCESS: CompletedAt timestamp set\n")
		}
	} else {
		fmt.Printf("❌ ISSUE: Expected 'completed' but got '%s'\n", finalJob.Status)
	}

	// Check scraped pages
	pages, err := repo.GetScrapedPages(job.ID)
	if err == nil {
		fmt.Printf("📄 Scraped pages: %d\n", len(pages))
		for i, page := range pages {
			fmt.Printf("   Page %d: %s (tokens: %d)\n", i+1, page.URL, page.TokenCount)
		}
	}
}

func testInvalidURLScraping(ctx context.Context, repo *repo.KnowledgeRepository, scraper *service.WebScrapingService) {
	// Create test job with invalid URL
	tenantID := uuid.New()
	projectID := uuid.New()
	invalidURL := "not-a-valid-url-at-all"

	job := &models.KnowledgeScrapingJob{
		ID:           uuid.New(),
		TenantID:     tenantID,
		ProjectID:    projectID,
		URL:          invalidURL,
		MaxDepth:     1,
		Status:       "pending",
		PagesScraped: 0,
		TotalPages:   0,
		StartedAt:    nil,
		CompletedAt:  nil,
		ErrorMessage: nil,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save initial job
	err := repo.CreateScrapingJob(job)
	if err != nil {
		log.Printf("❌ Failed to create job: %v", err)
		return
	}

	fmt.Printf("✅ Created job %s with invalid URL: %s\n", job.ID, invalidURL)

	// Start scraping (should fail)
	fmt.Printf("🚀 Starting scraping for invalid URL: %s\n", invalidURL)
	err = scraper.StartScraping(ctx, job.ID, invalidURL, 1)
	if err != nil {
		fmt.Printf("✅ Expected error for invalid URL: %v\n", err)
	} else {
		fmt.Printf("❓ No error returned for invalid URL (unexpected)\n")
	}

	// Check final job status
	finalJob, err := repo.GetScrapingJob(job.ID)
	if err != nil {
		log.Printf("❌ Failed to get final job: %v", err)
		return
	}

	fmt.Printf("📊 Final Job Status Report:\n")
	fmt.Printf("   Status: %s\n", finalJob.Status)
	fmt.Printf("   Pages Scraped: %d\n", finalJob.PagesScraped)
	if finalJob.ErrorMessage != nil {
		fmt.Printf("   Error: %s\n", *finalJob.ErrorMessage)
	}

	// Verify expectations for failure case
	if finalJob.Status == "failed" {
		fmt.Printf("✅ SUCCESS: Job correctly updated to 'failed' for invalid URL\n")
		if finalJob.ErrorMessage != nil {
			fmt.Printf("✅ SUCCESS: Error message set for failed job\n")
		}
	} else {
		fmt.Printf("❌ ISSUE: Expected 'failed' but got '%s' for invalid URL\n", finalJob.Status)
	}
}

func testNetworkErrorScraping(ctx context.Context, repo *repo.KnowledgeRepository, scraper *service.WebScrapingService) {
	// Create test job with URL that should cause network error
	tenantID := uuid.New()
	projectID := uuid.New()
	errorURL := "https://this-domain-should-not-exist-12345.com"

	job := &models.KnowledgeScrapingJob{
		ID:           uuid.New(),
		TenantID:     tenantID,
		ProjectID:    projectID,
		URL:          errorURL,
		MaxDepth:     1,
		Status:       "pending",
		PagesScraped: 0,
		TotalPages:   0,
		StartedAt:    nil,
		CompletedAt:  nil,
		ErrorMessage: nil,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save initial job
	err := repo.CreateScrapingJob(job)
	if err != nil {
		log.Printf("❌ Failed to create job: %v", err)
		return
	}

	fmt.Printf("✅ Created job %s with unreachable URL: %s\n", job.ID, errorURL)

	// Start scraping (should handle network error gracefully)
	fmt.Printf("🚀 Starting scraping for unreachable URL: %s\n", errorURL)
	err = scraper.StartScraping(ctx, job.ID, errorURL, 1)
	if err != nil {
		fmt.Printf("ℹ️  Error returned for unreachable URL: %v\n", err)
	} else {
		fmt.Printf("ℹ️  No error returned (service may handle network errors gracefully)\n")
	}

	// Check final job status
	finalJob, err := repo.GetScrapingJob(job.ID)
	if err != nil {
		log.Printf("❌ Failed to get final job: %v", err)
		return
	}

	fmt.Printf("📊 Final Job Status Report:\n")
	fmt.Printf("   Status: %s\n", finalJob.Status)
	fmt.Printf("   Pages Scraped: %d\n", finalJob.PagesScraped)
	if finalJob.ErrorMessage != nil {
		fmt.Printf("   Error: %s\n", *finalJob.ErrorMessage)
	}

	// For network errors, the service might handle it as completed with 0 pages or failed
	if finalJob.Status == "failed" || (finalJob.Status == "completed" && finalJob.PagesScraped == 0) {
		fmt.Printf("✅ SUCCESS: Job handled network error appropriately (status: %s)\n", finalJob.Status)
	} else {
		fmt.Printf("❓ UNEXPECTED: Job status '%s' with %d pages for network error\n", finalJob.Status, finalJob.PagesScraped)
	}
}
