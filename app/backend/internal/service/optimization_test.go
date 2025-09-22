package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestOptimizationLogic tests the core optimization logic without external dependencies
func TestOptimizationLogic(t *testing.T) {
	t.Run("ContentHashGeneration", func(t *testing.T) {
		service := &WebScrapingService{}

		content1 := "Hello, World!"
		content2 := "Hello, World!"
		content3 := "Different content"

		hash1 := service.generateContentHash(content1)
		hash2 := service.generateContentHash(content2)
		hash3 := service.generateContentHash(content3)

		// Same content should produce same hash
		assert.Equal(t, hash1, hash2, "Same content should produce same hash")

		// Different content should produce different hash
		assert.NotEqual(t, hash1, hash3, "Different content should produce different hash")

		// Hash should be consistent
		assert.NotEmpty(t, hash1, "Hash should not be empty")
		assert.Len(t, hash1, 64, "SHA256 hash should be 64 characters")
	})

	t.Run("TokenCountEstimation", func(t *testing.T) {
		service := &WebScrapingService{}

		text := "This is a test text with approximately twenty characters"
		tokenCount := service.estimateTokenCount(text)

		// Should estimate roughly 1 token per 4 characters (allow off-by-one)
		expectedTokens := len(strings.TrimSpace(text)) / 4
		if expectedTokens < 0 {
			expectedTokens = 0
		}
		// Allow a small rounding difference (±1)
		if tokenCount != expectedTokens && tokenCount != expectedTokens+1 && tokenCount != expectedTokens-1 {
			t.Fatalf("Token count estimation mismatch: expected approx %d, got %d", expectedTokens, tokenCount)
		}
	})
}
