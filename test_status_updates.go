package tms
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/db"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/bareuptime/tms/internal/service"
)

func main() {
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

	// Create repositories
	knowledgeRepo := repo.NewKnowledgeRepository(database)

	// Create services
	embeddingService := service.NewEmbeddingService(&cfg.Knowledge)
	webScraperService := service.NewWebScrapingService(knowledgeRepo, embeddingService, &cfg.Knowledge)

	// Setup test data
	tenantID := uuid.New()
	projectID := uuid.New()
	testURL := "https://httpbin.org/html" // Simple HTML endpoint for testing

	// Create a scraping job
	job := &models.KnowledgeScrapingJob{
		ID:          uuid.New(),
		TenantID:    tenantID,
		ProjectID:   projectID,
		URL:         testURL,
		MaxDepth:    1,
		Status:      "pending",
		PagesScraped: 0,
		TotalPages:  0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save the job
	err = knowledgeRepo.CreateScrapingJob(job)
	if err != nil {
		log.Fatalf("Failed to create scraping job: %v", err)
	}

	fmt.Printf("Created scraping job: %s\n", job.ID)
	fmt.Printf("Initial status: %s\n", job.Status)

	// Start scraping and monitor status updates
	ctx := context.Background()
	
	fmt.Println("\n=== Starting web scraping ===")
	fmt.Printf("Target URL: %s\n", testURL)
	fmt.Printf("Embedding service enabled: %v\n", embeddingService.IsEnabled())
	
	if !embeddingService.IsEnabled() {
		fmt.Println("WARNING: Embedding service disabled (missing OpenAI API key)")
		fmt.Println("This will test our logging improvements when embedding fails")
	}

	// Start scraping
	err = webScraperService.StartScraping(ctx, job.ID, testURL, 1)
	if err != nil {
		log.Printf("Scraping error: %v", err)
	}

	// Check final status
	updatedJob, err := knowledgeRepo.GetScrapingJob(job.ID)
	if err != nil {
		log.Fatalf("Failed to get updated job: %v", err)
	}

	fmt.Println("\n=== Scraping Results ===")
	fmt.Printf("Final status: %s\n", updatedJob.Status)
	fmt.Printf("Pages scraped: %d\n", updatedJob.PagesScraped)
	fmt.Printf("Completed at: %v\n", updatedJob.CompletedAt)
	
	if updatedJob.ErrorMessage != nil {
		fmt.Printf("Error message: %s\n", *updatedJob.ErrorMessage)
	}

	// Check scraped pages
	scrapedPages, err := knowledgeRepo.GetScrapedPages(job.ID)
	if err != nil {
		log.Printf("Failed to get scraped pages: %v", err)
	} else {
		fmt.Printf("Number of scraped pages: %d\n", len(scrapedPages))
		
		for i, page := range scrapedPages {
			fmt.Printf("Page %d:\n", i+1)
			fmt.Printf("  URL: %s\n", page.URL)
			if page.Title != nil {
				fmt.Printf("  Title: %s\n", *page.Title)
			}
			fmt.Printf("  Content length: %d characters\n", len(page.Content))
			fmt.Printf("  Token count: %d\n", page.TokenCount)
			
			// Check if embedding was generated
			if len(page.Embedding) > 0 {
				fmt.Printf("  Embedding: Generated (%d dimensions)\n", len(page.Embedding))
			} else {
				fmt.Printf("  Embedding: Not generated (expected if no API key)\n")
			}
		}
	}

	fmt.Println("\n=== Test Complete ===")
	fmt.Println("Check the logs above for:")
	fmt.Println("1. Status update messages (pending -> running -> completed)")
	fmt.Println("2. Embedding generation logs (success or failure)")
	fmt.Println("3. Performance improvements (should process only body element)")
}
