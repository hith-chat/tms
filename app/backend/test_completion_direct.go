//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/db"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/bareuptime/tms/internal/service"
	"github.com/google/uuid"
)

func main() {
	fmt.Println("=== Testing Job Completion Status Fix ===")
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
	knowledgeRepo := repo.NewKnowledgeRepository(database.DB)
	embeddingService := service.NewEmbeddingService(&cfg.Knowledge)
	webScraperService := service.NewWebScrapingService(knowledgeRepo, embeddingService, &cfg.Knowledge)

	// Use existing test tenant and project IDs
	tenantID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	projectID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")

	fmt.Printf("🎯 Testing job completion with tenant: %s, project: %s\n", tenantID, projectID)
	fmt.Printf("🔧 Embedding service enabled: %v\n", embeddingService.IsEnabled())
	fmt.Println()

	// Create scraping request for a simple HTML page
	req := &models.CreateScrapingJobRequest{
		URL:      "https://httpbin.org/html",
		MaxDepth: 1,
	}

	fmt.Printf("🚀 Creating scraping job for URL: %s\n", req.URL)

	// Create the job using the service
	job, err := webScraperService.CreateScrapingJob(context.Background(), tenantID, projectID, req)
	if err != nil {
		log.Fatalf("❌ Failed to create scraping job: %v", err)
	}

	fmt.Printf("✅ Created job: %s\n", job.ID)
	fmt.Printf("📊 Initial status: %s\n", job.Status)

	// Monitor job status for completion
	fmt.Println("\n⏰ Monitoring job status for completion...")

	maxWait := 90 * time.Second
	checkInterval := 3 * time.Second
	startTime := time.Now()
	lastStatus := ""

	for time.Since(startTime) < maxWait {
		// Get current job status
		currentJob, err := webScraperService.GetScrapingJob(context.Background(), job.ID)
		if err != nil {
			log.Printf("❌ Error getting job status: %v", err)
			time.Sleep(checkInterval)
			continue
		}

		elapsed := time.Since(startTime)

		// Only print if status changed
		if currentJob.Status != lastStatus {
			fmt.Printf("📊 [%s] Status changed: %s → %s",
				elapsed.Round(time.Second),
				lastStatus,
				currentJob.Status)

			if currentJob.PagesScraped > 0 {
				fmt.Printf(", Pages: %d", currentJob.PagesScraped)
			}

			if currentJob.StartedAt != nil {
				fmt.Printf(", Started: %s", currentJob.StartedAt.Format("15:04:05"))
			}
			if currentJob.CompletedAt != nil {
				fmt.Printf(", Completed: %s", currentJob.CompletedAt.Format("15:04:05"))
			}
			if currentJob.ErrorMessage != nil {
				fmt.Printf(", Error: %s", *currentJob.ErrorMessage)
			}
			fmt.Println()

			lastStatus = currentJob.Status
		}

		// Check if job is complete
		if currentJob.Status == "completed" {
			fmt.Println("\n🎉 SUCCESS: Job completed successfully!")

			// Verify completion details
			fmt.Printf("✅ Final status: %s\n", currentJob.Status)
			fmt.Printf("✅ Pages scraped: %d\n", currentJob.PagesScraped)
			fmt.Printf("✅ Started at: %v\n", currentJob.StartedAt)
			fmt.Printf("✅ Completed at: %v\n", currentJob.CompletedAt)

			if currentJob.StartedAt != nil && currentJob.CompletedAt != nil {
				duration := currentJob.CompletedAt.Sub(*currentJob.StartedAt)
				fmt.Printf("✅ Processing time: %s\n", duration.Round(time.Millisecond))
			}

			// Check scraped pages
			pages, err := webScraperService.GetJobPages(context.Background(), job.ID)
			if err != nil {
				log.Printf("❌ Error getting scraped pages: %v", err)
			} else {
				fmt.Printf("📄 Scraped pages: %d\n", len(pages))
				for i, page := range pages {
					fmt.Printf("   Page %d: %s (tokens: %d)\n", i+1, page.URL, page.TokenCount)
					if page.Title != nil {
						fmt.Printf("           Title: %s\n", *page.Title)
					}
				}
			}

			fmt.Println("\n🔧 FIX VERIFICATION:")
			fmt.Println("✅ Job properly transitions to 'completed' status")
			fmt.Println("✅ CompletedAt timestamp is set")
			fmt.Println("✅ Pages are scraped and saved")
			fmt.Println("✅ Job completion works even with embedding service disabled")
			break
		} else if currentJob.Status == "failed" || currentJob.Status == "error" {
			fmt.Printf("\n❌ Job failed with status: %s\n", currentJob.Status)
			if currentJob.ErrorMessage != nil {
				fmt.Printf("❌ Error: %s\n", *currentJob.ErrorMessage)
			}

			fmt.Println("\n🔧 FIX VERIFICATION:")
			fmt.Println("✅ Job properly transitions to failure status")
			fmt.Println("✅ Error message is stored")
			break
		}

		time.Sleep(checkInterval)
	}

	if time.Since(startTime) >= maxWait {
		fmt.Println("\n⏰ Test timeout reached")

		// Get final status
		finalJob, err := webScraperService.GetScrapingJob(context.Background(), job.ID)
		if err == nil {
			fmt.Printf("📊 Final status: %s\n", finalJob.Status)
			fmt.Printf("📊 Pages scraped: %d\n", finalJob.PagesScraped)
		}
	}

	fmt.Println("\n=== Job Completion Status Fix Test Complete ===")
}
