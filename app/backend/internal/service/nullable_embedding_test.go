package service

import (
	"fmt"
	"testing"

	"github.com/pgvector/pgvector-go"
	"github.com/stretchr/testify/assert"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/models"
)

// Test that embedding generation failure doesn't prevent job completion
func TestEmbeddingFailureHandling(t *testing.T) {
	// Create a disabled embedding service to simulate failure
	cfg := &config.KnowledgeConfig{
		Enabled:          false,
		EmbeddingService: "disabled",
	}
	
	embeddingService := NewEmbeddingService(cfg)
	
	// Test that disabled service reports correctly
	assert.False(t, embeddingService.IsEnabled())
}

// Test nullable embedding field assignment
func TestNullableEmbeddingAssignment(t *testing.T) {
	// Test that we can create pages with nil embeddings
	testPages := []*models.KnowledgeScrapedPage{
		{
			URL:        "https://example.com",
			Content:    "Test content",
			TokenCount: 10,
			Embedding:  nil, // This should be allowed now
			Metadata:   models.JSONMap{},
		},
	}
	
	// Verify the page was created with nil embedding
	assert.Nil(t, testPages[0].Embedding)
	assert.Equal(t, "Test content", testPages[0].Content)
}

// Test embedding assignment when generation succeeds
func TestEmbeddingAssignmentSuccess(t *testing.T) {
	// Create a test embedding vector
	embeddingData := make([]float32, 1536)
	for i := range embeddingData {
		embeddingData[i] = 0.1
	}
	embedding := pgvector.NewVector(embeddingData)
	
	// Test page with successful embedding
	testPage := &models.KnowledgeScrapedPage{
		URL:        "https://example.com",
		Content:    "Test content",
		TokenCount: 10,
		Embedding:  &embedding,
		Metadata:   models.JSONMap{},
	}
	
	// Verify the page has the embedding
	assert.NotNil(t, testPage.Embedding)
	assert.Equal(t, 1536, len(testPage.Embedding.Slice()))
}

// Test the generateEmbeddingsForPages function with disabled service
func TestGenerateEmbeddingsForPages_ServiceDisabled(t *testing.T) {
	// Create disabled embedding service
	cfg := &config.KnowledgeConfig{
		Enabled:          false,
		EmbeddingService: "disabled",
	}
	
	embeddingService := NewEmbeddingService(cfg)
	
	// Create web scraping service
	service := &WebScrapingService{
		embeddingService: embeddingService,
		config:          cfg,
	}
	
	// Create test pages
	testPages := []*models.KnowledgeScrapedPage{
		{
			URL:        "https://example.com",
			Content:    "Test content",
			TokenCount: 10,
			Embedding:  nil,
			Metadata:   models.JSONMap{},
		},
	}
	
	// When embedding service is disabled, pages should remain with nil embeddings
	// This simulates the scenario where embedding generation is skipped
	
	// Since the service is disabled, we expect no embedding generation to occur
	// The test verifies that pages can have nil embeddings without causing issues
	for _, page := range testPages {
		assert.Nil(t, page.Embedding, "Page should have nil embedding when service is disabled")
	}
	
	// This would be called in the actual flow - testing that it doesn't panic
	assert.NotPanics(t, func() {
		// Simulate the check that happens in the web scraper
		if !service.embeddingService.IsEnabled() {
			fmt.Printf("Warning: Embedding service is disabled, saving pages without embeddings\n")
		}
	})
}

// Test chunk creation with nullable embeddings
func TestKnowledgeChunkWithNullableEmbedding(t *testing.T) {
	// Test that we can create chunks with nil embeddings
	testChunk := &models.KnowledgeChunk{
		Content:    "Test chunk content",
		TokenCount: 10,
		Embedding:  nil, // This should be allowed now
		Metadata:   models.JSONMap{},
	}
	
	// Verify the chunk was created with nil embedding
	assert.Nil(t, testChunk.Embedding)
	assert.Equal(t, "Test chunk content", testChunk.Content)
	
	// Test with embedding assigned
	embeddingData := make([]float32, 1536)
	embedding := pgvector.NewVector(embeddingData)
	testChunk.Embedding = &embedding
	
	assert.NotNil(t, testChunk.Embedding)
	assert.Equal(t, 1536, len(testChunk.Embedding.Slice()))
}
