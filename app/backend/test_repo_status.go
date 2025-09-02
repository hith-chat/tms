//go:build ignore

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/db"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/google/uuid"
)

func main() {
	fmt.Println("=== Testing Job Status Update Repository Methods ===")
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

	// Create repository
	knowledgeRepo := repo.NewKnowledgeRepository(database.DB)

	// Test 1: Create job and verify status transitions
	fmt.Println("🔍 Test 1: Job Status Transitions")
	testJobStatusTransitions(knowledgeRepo)

	fmt.Println()

	// Test 2: Test job completion
	fmt.Println("🔍 Test 2: Job Completion Status")
	testJobCompletion(knowledgeRepo)

	fmt.Println()

	// Test 3: Test job failure handling
	fmt.Println("🔍 Test 3: Job Failure Status")
	testJobFailure(knowledgeRepo)

	fmt.Println()
	fmt.Println("=== All Job Status Tests Complete ===")
}

func testJobStatusTransitions(repo *repo.KnowledgeRepository) {
	// Use existing test tenant and project IDs from migrations
	tenantID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	projectID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")

	job := &models.KnowledgeScrapingJob{
		ID:           uuid.New(),
		TenantID:     tenantID,
		ProjectID:    projectID,
		URL:          "https://test.example.com",
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
	savedJob, err := repo.GetScrapingJob(job.ID)
	if err != nil {
		log.Printf("❌ Failed to get job: %v", err)
		return
	}

	fmt.Printf("📊 Initial state: status=%s, pages=%d, started=%v, completed=%v\n",
		savedJob.Status, savedJob.PagesScraped, savedJob.StartedAt, savedJob.CompletedAt)

	// Test transition to "running"
	fmt.Printf("🚀 Transitioning to 'running' status...\n")
	err = repo.StartScrapingJob(job.ID)
	if err != nil {
		log.Printf("❌ Failed to start job: %v", err)
		return
	}

	// Verify running status
	runningJob, err := repo.GetScrapingJob(job.ID)
	if err != nil {
		log.Printf("❌ Failed to get running job: %v", err)
		return
	}

	fmt.Printf("📊 Running state: status=%s, started=%v\n", runningJob.Status, runningJob.StartedAt)

	if runningJob.Status == "running" {
		fmt.Printf("✅ SUCCESS: Job correctly transitioned to 'running'\n")
	} else {
		fmt.Printf("❌ ISSUE: Expected 'running' but got '%s'\n", runningJob.Status)
	}

	if runningJob.StartedAt != nil {
		fmt.Printf("✅ SUCCESS: StartedAt timestamp set\n")
	} else {
		fmt.Printf("❌ ISSUE: StartedAt timestamp not set\n")
	}

	// Test progress update
	fmt.Printf("📈 Updating progress...\n")
	err = repo.UpdateScrapingJobProgress(job.ID, 5, 10)
	if err != nil {
		log.Printf("❌ Failed to update progress: %v", err)
		return
	}

	// Verify progress update
	progressJob, err := repo.GetScrapingJob(job.ID)
	if err != nil {
		log.Printf("❌ Failed to get job with progress: %v", err)
		return
	}

	fmt.Printf("📊 Progress state: pages_scraped=%d, total_pages=%d\n",
		progressJob.PagesScraped, progressJob.TotalPages)

	if progressJob.PagesScraped == 5 && progressJob.TotalPages == 10 {
		fmt.Printf("✅ SUCCESS: Progress correctly updated\n")
	} else {
		fmt.Printf("❌ ISSUE: Progress not updated correctly\n")
	}
}

func testJobCompletion(repo *repo.KnowledgeRepository) {
	// Use existing test tenant and project IDs from migrations
	tenantID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	projectID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")

	job := &models.KnowledgeScrapingJob{
		ID:           uuid.New(),
		TenantID:     tenantID,
		ProjectID:    projectID,
		URL:          "https://completion-test.example.com",
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

	// Save and start job
	err := repo.CreateScrapingJob(job)
	if err != nil {
		log.Printf("❌ Failed to create job: %v", err)
		return
	}

	err = repo.StartScrapingJob(job.ID)
	if err != nil {
		log.Printf("❌ Failed to start job: %v", err)
		return
	}

	fmt.Printf("✅ Created and started job %s\n", job.ID)

	// Update progress to simulate scraping
	err = repo.UpdateScrapingJobProgress(job.ID, 3, 3)
	if err != nil {
		log.Printf("❌ Failed to update progress: %v", err)
		return
	}

	// Complete the job
	fmt.Printf("🏁 Completing job...\n")
	err = repo.CompleteScrapingJob(job.ID)
	if err != nil {
		log.Printf("❌ Failed to complete job: %v", err)
		return
	}

	// Verify completion
	completedJob, err := repo.GetScrapingJob(job.ID)
	if err != nil {
		log.Printf("❌ Failed to get completed job: %v", err)
		return
	}

	fmt.Printf("📊 Final state: status=%s, pages=%d, completed=%v\n",
		completedJob.Status, completedJob.PagesScraped, completedJob.CompletedAt)

	if completedJob.Status == "completed" {
		fmt.Printf("✅ SUCCESS: Job correctly marked as 'completed'\n")
	} else {
		fmt.Printf("❌ ISSUE: Expected 'completed' but got '%s'\n", completedJob.Status)
	}

	if completedJob.CompletedAt != nil {
		fmt.Printf("✅ SUCCESS: CompletedAt timestamp set\n")
	} else {
		fmt.Printf("❌ ISSUE: CompletedAt timestamp not set\n")
	}

	if completedJob.ErrorMessage == nil {
		fmt.Printf("✅ SUCCESS: No error message for successful completion\n")
	} else {
		fmt.Printf("❌ ISSUE: Unexpected error message: %s\n", *completedJob.ErrorMessage)
	}
}

func testJobFailure(repo *repo.KnowledgeRepository) {
	// Use existing test tenant and project IDs from migrations
	tenantID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	projectID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")

	job := &models.KnowledgeScrapingJob{
		ID:           uuid.New(),
		TenantID:     tenantID,
		ProjectID:    projectID,
		URL:          "https://failure-test.example.com",
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

	// Save and start job
	err := repo.CreateScrapingJob(job)
	if err != nil {
		log.Printf("❌ Failed to create job: %v", err)
		return
	}

	err = repo.StartScrapingJob(job.ID)
	if err != nil {
		log.Printf("❌ Failed to start job: %v", err)
		return
	}

	fmt.Printf("✅ Created and started job %s\n", job.ID)

	// Simulate failure
	errorMessage := "Network timeout while scraping"
	fmt.Printf("💥 Marking job as failed with error: %s\n", errorMessage)
	err = repo.UpdateScrapingJobStatus(job.ID, "failed", &errorMessage)
	if err != nil {
		log.Printf("❌ Failed to mark job as failed: %v", err)
		return
	}

	// Verify failure
	failedJob, err := repo.GetScrapingJob(job.ID)
	if err != nil {
		log.Printf("❌ Failed to get failed job: %v", err)
		return
	}

	fmt.Printf("📊 Failed state: status=%s, error=%v\n",
		failedJob.Status, failedJob.ErrorMessage)

	if failedJob.Status == "failed" {
		fmt.Printf("✅ SUCCESS: Job correctly marked as 'failed'\n")
	} else {
		fmt.Printf("❌ ISSUE: Expected 'failed' but got '%s'\n", failedJob.Status)
	}

	if failedJob.ErrorMessage != nil && *failedJob.ErrorMessage == errorMessage {
		fmt.Printf("✅ SUCCESS: Error message correctly set\n")
	} else {
		fmt.Printf("❌ ISSUE: Error message not set correctly\n")
	}

	// Test error status (alternative failure state)
	job2 := &models.KnowledgeScrapingJob{
		ID:           uuid.New(),
		TenantID:     tenantID,
		ProjectID:    projectID,
		URL:          "https://error-test.example.com",
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

	// Save and test error status
	err = repo.CreateScrapingJob(job2)
	if err != nil {
		log.Printf("❌ Failed to create second job: %v", err)
		return
	}

	err = repo.StartScrapingJob(job2.ID)
	if err != nil {
		log.Printf("❌ Failed to start second job: %v", err)
		return
	}

	errorMessage2 := "Invalid URL format"
	fmt.Printf("🚫 Marking second job as error with: %s\n", errorMessage2)
	err = repo.UpdateScrapingJobStatus(job2.ID, "error", &errorMessage2)
	if err != nil {
		log.Printf("❌ Failed to mark second job as error: %v", err)
		return
	}

	// Verify error status
	errorJob, err := repo.GetScrapingJob(job2.ID)
	if err != nil {
		log.Printf("❌ Failed to get error job: %v", err)
		return
	}

	fmt.Printf("📊 Error state: status=%s, error=%v\n",
		errorJob.Status, errorJob.ErrorMessage)

	if errorJob.Status == "error" {
		fmt.Printf("✅ SUCCESS: Job correctly marked as 'error'\n")
	} else {
		fmt.Printf("❌ ISSUE: Expected 'error' but got '%s'\n", errorJob.Status)
	}
}
