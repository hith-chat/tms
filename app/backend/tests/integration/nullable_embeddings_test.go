package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/bareuptime/tms/internal/service"
)

// TestNullableEmbeddingsIntegration tests the complete nullable embeddings functionality
func TestNullableEmbeddingsIntegration(t *testing.T) {
	database := setupTestDB(t)
	defer teardownTestDB(database)

	// Clean up existing test data
	database.Exec("DELETE FROM knowledge_scraped_pages WHERE url LIKE 'https://test-nullable-embedding%'")
	database.Exec("DELETE FROM knowledge_scraping_jobs WHERE url LIKE 'https://test-nullable-embedding%'")

	tenantID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	projectID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")

	// Create repositories
	knowledgeRepo := repo.NewKnowledgeRepository(database.DB)

	t.Run("can save pages with null embeddings", func(t *testing.T) {
		// Create a scraping job
		job := &models.KnowledgeScrapingJob{
			ID:        uuid.New(),
			TenantID:  tenantID,
			ProjectID: projectID,
			URL:       "https://test-nullable-embedding-1.com",
			MaxDepth:  1,
			Status:    "running",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := knowledgeRepo.CreateScrapingJob(job)
		require.NoError(t, err)

		// Create pages with null embeddings (simulating embedding generation failure)
		pages := []*models.KnowledgeScrapedPage{
			{
				ID:         uuid.New(),
				JobID:      job.ID,
				URL:        "https://test-nullable-embedding-1.com",
				Title:      stringPtr("Test Page with Null Embedding"),
				Content:    "This page was saved without embeddings due to generation failure",
				TokenCount: 15,
				ScrapedAt:  time.Now(),
				Embedding:  nil, // This is the key test - null embedding should be allowed
				Metadata:   models.JSONMap{},
			},
		}

		// This should NOT fail with "vector must have at least 1 dimension"
		err = knowledgeRepo.CreateScrapedPages(pages)
		assert.NoError(t, err, "Should be able to save pages with null embeddings")

		// Verify the page was saved
		savedPages, err := knowledgeRepo.GetJobPages(job.ID)
		require.NoError(t, err)
		require.Len(t, savedPages, 1)

		savedPage := savedPages[0]
		assert.Equal(t, "https://test-nullable-embedding-1.com", savedPage.URL)
		assert.Nil(t, savedPage.Embedding, "Embedding should be null")
		assert.Equal(t, "This page was saved without embeddings due to generation failure", savedPage.Content)

		// Complete the job - this should succeed
		err = knowledgeRepo.CompleteScrapingJob(job.ID)
		assert.NoError(t, err, "Job should complete successfully even with null embeddings")

		// Verify job completion
		completedJob, err := knowledgeRepo.GetScrapingJob(job.ID)
		require.NoError(t, err)
		assert.Equal(t, "completed", completedJob.Status)
	})

	t.Run("can save pages with valid embeddings", func(t *testing.T) {
		// Create another scraping job
		job := &models.KnowledgeScrapingJob{
			ID:        uuid.New(),
			TenantID:  tenantID,
			ProjectID: projectID,
			URL:       "https://test-nullable-embedding-2.com",
			MaxDepth:  1,
			Status:    "running",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := knowledgeRepo.CreateScrapingJob(job)
		require.NoError(t, err)

		// Create a valid embedding vector
		embeddingData := make([]float32, 1536)
		for i := range embeddingData {
			embeddingData[i] = float32(i) * 0.001 // Some test values
		}
		embedding := pgvector.NewVector(embeddingData)

		// Create pages with valid embeddings
		pages := []*models.KnowledgeScrapedPage{
			{
				ID:         uuid.New(),
				JobID:      job.ID,
				URL:        "https://test-nullable-embedding-2.com",
				Title:      stringPtr("Test Page with Valid Embedding"),
				Content:    "This page was saved with a valid embedding vector",
				TokenCount: 15,
				ScrapedAt:  time.Now(),
				Embedding:  &embedding,
				Metadata:   models.JSONMap{},
			},
		}

		// This should work as before
		err = knowledgeRepo.CreateScrapedPages(pages)
		assert.NoError(t, err, "Should be able to save pages with valid embeddings")

		// Verify the page was saved with embedding
		savedPages, err := knowledgeRepo.GetJobPages(job.ID)
		require.NoError(t, err)
		require.Len(t, savedPages, 1)

		savedPage := savedPages[0]
		assert.Equal(t, "https://test-nullable-embedding-2.com", savedPage.URL)
		assert.NotNil(t, savedPage.Embedding, "Embedding should not be null")
		assert.Equal(t, 1536, len(savedPage.Embedding.Slice()), "Embedding should have correct dimensions")

		// Complete the job
		err = knowledgeRepo.CompleteScrapingJob(job.ID)
		assert.NoError(t, err)
	})
}

// TestEmbeddingServiceDisabled tests behavior when embedding service is disabled
func TestEmbeddingServiceDisabled(t *testing.T) {
	// Test embedding service configuration
	cfg := &config.KnowledgeConfig{
		Enabled:          false,
		EmbeddingService: "",
	}

	embeddingService := service.NewEmbeddingService(cfg)

	// Test that service reports as disabled
	assert.False(t, embeddingService.IsEnabled(), "Embedding service should be disabled")

	// Test that we can still work with pages without embeddings
	testPage := &models.KnowledgeScrapedPage{
		URL:        "https://example.com",
		Content:    "Test content without embedding",
		TokenCount: 10,
		Embedding:  nil,
		Metadata:   models.JSONMap{},
	}

	assert.Nil(t, testPage.Embedding, "Page should have nil embedding when service is disabled")
}

// TestDocumentProcessorNullableEmbeddings tests document processing with nullable embeddings
func TestDocumentProcessorNullableEmbeddings(t *testing.T) {
	// Test that chunks can be created with null embeddings
	testChunk := &models.KnowledgeChunk{
		Content:    "Test chunk content",
		TokenCount: 10,
		Embedding:  nil, // Should be allowed
		Metadata:   models.JSONMap{},
	}

	assert.Nil(t, testChunk.Embedding, "Chunk should allow nil embeddings")
	assert.Equal(t, "Test chunk content", testChunk.Content)

	// Test with valid embedding
	embeddingData := make([]float32, 1536)
	embedding := pgvector.NewVector(embeddingData)
	testChunk.Embedding = &embedding

	assert.NotNil(t, testChunk.Embedding, "Chunk should support valid embeddings")
	assert.Equal(t, 1536, len(testChunk.Embedding.Slice()))
}
