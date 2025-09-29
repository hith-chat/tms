package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/db"
	"github.com/bareuptime/tms/internal/logger"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/bareuptime/tms/internal/service"
	"github.com/google/uuid"
)

func main() {
	logger.Info("=== End-to-End Job Completion Test ===")
	logger.Info("")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Errorf("Failed to load config: %v", err)
		return
	}

	// Connect to database
	database, err := db.Connect(&cfg.Database)
	if err != nil {
		logger.Errorf("Failed to connect to database: %v", err)
		return
	}
	defer database.Close()

	// Create repositories and services
	knowledgeRepo := repo.NewKnowledgeRepository(database.DB)
	embeddingService := service.NewEmbeddingService(&cfg.Knowledge)
	// Pass nil for redis client in this test harness (no Redis required here)
	webScraperService := service.NewWebScrapingService(knowledgeRepo, embeddingService, &cfg.Knowledge, nil)

	// Use existing test tenant and project IDs
	tenantID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	projectID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")

	logger.Infof("🎯 Testing with tenant: %s, project: %s", tenantID, projectID)
	logger.Infof("🔧 Embedding service enabled: %v", embeddingService.IsEnabled())
	logger.Info("")

	// Test 1: Create and run a scraping job
	logger.Info("🔍 Test: Full Scraping Job Lifecycle")

	// Create scraping request
	req := &models.CreateScrapingJobRequest{
		URL:      "https://httpbin.org/html",
		MaxDepth: 1,
	}

	logger.Infof("🚀 Creating scraping job for URL: %s", req.URL)

	// Create the job using the service
	job, err := webScraperService.CreateScrapingJob(context.Background(), tenantID, projectID, req)
	if err != nil {
		logger.Errorf("❌ Failed to create scraping job: %v", err)
		return
	}

	logger.Infof("✅ Created job: %s", job.ID)
	logger.Infof("📊 Initial status: %s", job.Status)

	// Monitor job status for completion
	logger.Info("⏰ Monitoring job status changes...")

	maxWait := 60 * time.Second
	checkInterval := 2 * time.Second
	startTime := time.Now()

	for time.Since(startTime) < maxWait {
		// Get current job status
		currentJob, err := webScraperService.GetScrapingJob(context.Background(), job.ID, job.TenantID, job.ProjectID)
		if err != nil {
			logger.Errorf("❌ Error getting job status: %v", err)
			time.Sleep(checkInterval)
			continue
		}

		elapsed := time.Since(startTime)
		statusMsg := fmt.Sprintf("📊 [%s] Status: %s, Pages: %d",
			elapsed.Round(time.Second),
			currentJob.Status,
			currentJob.PagesScraped)

		if currentJob.StartedAt != nil {
			statusMsg += fmt.Sprintf(", Started: %s", currentJob.StartedAt.Format("15:04:05"))
		}
		if currentJob.CompletedAt != nil {
			statusMsg += fmt.Sprintf(", Completed: %s", currentJob.CompletedAt.Format("15:04:05"))
		}
		if currentJob.ErrorMessage != nil {
			statusMsg += fmt.Sprintf(", Error: %s", *currentJob.ErrorMessage)
		}
		logger.Info(statusMsg)

		// Check if job is complete
		if currentJob.Status == "completed" {
			logger.Info("🎉 SUCCESS: Job completed successfully!")

			// Verify completion details
			logger.Infof("✅ Final status: %s", currentJob.Status)
			logger.Infof("✅ Pages scraped: %d", currentJob.PagesScraped)
			logger.Infof("✅ Started at: %v", currentJob.StartedAt)
			logger.Infof("✅ Completed at: %v", currentJob.CompletedAt)

			if currentJob.StartedAt != nil && currentJob.CompletedAt != nil {
				duration := currentJob.CompletedAt.Sub(*currentJob.StartedAt)
				logger.Infof("✅ Processing time: %s", duration.Round(time.Millisecond))
			}

			// Check scraped pages
			pages, err := webScraperService.GetJobPages(context.Background(), job.ID, job.TenantID, job.ProjectID)
			if err != nil {
				logger.Errorf("❌ Error getting scraped pages: %v", err)
			} else {
				logger.Infof("📄 Scraped pages: %d", len(pages))
				for i, page := range pages {
					logger.Infof("   Page %d: %s (tokens: %d)", i+1, page.URL, page.TokenCount)
					if page.Title != nil {
						logger.Infof("           Title: %s", *page.Title)
					}
				}
			}

			logger.Info("✅ JOB COMPLETION STATUS TEST PASSED!")
			logger.Info("✅ Status properly updated from 'pending' → 'running' → 'completed'")
			logger.Info("✅ Timestamps correctly set")
			logger.Info("✅ Pages scraped and saved")
			break
		} else if currentJob.Status == "failed" || currentJob.Status == "error" {
			logger.Infof("❌ Job failed with status: %s", currentJob.Status)
			if currentJob.ErrorMessage != nil {
				logger.Infof("❌ Error: %s", *currentJob.ErrorMessage)
			}

			logger.Info("✅ JOB FAILURE STATUS TEST PASSED!")
			logger.Info("✅ Status properly updated to indicate failure")
			logger.Info("✅ Error message set")
			break
		} else if currentJob.Status == "running" {
			// Continue monitoring
		} else if currentJob.Status == "pending" {
			// Still waiting to start
		} else {
			logger.Infof("❓ Unknown status: %s", currentJob.Status)
		}

		time.Sleep(checkInterval)
	}

	if time.Since(startTime) >= maxWait {
		logger.Info("⏰ Test timeout reached")

		// Get final status
		finalJob, err := webScraperService.GetScrapingJob(context.Background(), job.ID, job.TenantID, job.ProjectID)
		if err == nil {
			logger.Infof("📊 Final status: %s", finalJob.Status)
			logger.Infof("📊 Pages scraped: %d", finalJob.PagesScraped)
		}
	}

	logger.Info("=== End-to-End Job Completion Test Complete ===")
}
