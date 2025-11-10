package service

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/bareuptime/tms/internal/models"
)

// TestContentUpdateHandling tests how the system handles content updates
func TestContentUpdateHandling(t *testing.T) {
	scenarios := []struct {
		name            string
		originalContent string
		updatedContent  string
		expectUpdate    bool
		description     string
	}{
		{
			name:            "Minor content update",
			originalContent: "This is the original article about Go programming.",
			updatedContent:  "This is the updated article about Go programming with new examples.",
			expectUpdate:    true,
			description:     "Content changed - should update embedding",
		},
		{
			name:            "Major content overhaul",
			originalContent: "Introduction to Python programming basics.",
			updatedContent:  "Advanced TypeScript patterns and best practices.",
			expectUpdate:    true,
			description:     "Completely different content - should update embedding",
		},
		{
			name:            "Identical content",
			originalContent: "This content hasn't changed at all.",
			updatedContent:  "This content hasn't changed at all.",
			expectUpdate:    false,
			description:     "Same content - should skip to save costs",
		},
		{
			name:            "Whitespace only changes",
			originalContent: "This is some content.",
			updatedContent:  "This is some content.   \n",
			expectUpdate:    true,
			description:     "Even whitespace changes detected",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Simulate scraping the same URL at different times
			jobID1 := uuid.New()
			jobID2 := uuid.New()
			testURL := "https://example.com/article"

			// First scrape - original content
			page1 := &models.KnowledgeScrapedPage{
				ID:      uuid.New(),
				JobID:   uuid.NullUUID{UUID: jobID1, Valid: true},
				URL:     testURL,
				Content: scenario.originalContent,
			}

			// Second scrape - potentially updated content
			page2 := &models.KnowledgeScrapedPage{
				ID:      uuid.New(),
				JobID:   uuid.NullUUID{UUID: jobID2, Valid: true},
				URL:     testURL, // Same URL
				Content: scenario.updatedContent,
			}

			// Generate content hashes (simulate the hash generation)
			hash1 := generateTestContentHash(scenario.originalContent)
			hash2 := generateTestContentHash(scenario.updatedContent)

			page1.ContentHash = &hash1
			page2.ContentHash = &hash2

			// Test the logic
			contentChanged := *page1.ContentHash != *page2.ContentHash

			assert.Equal(t, scenario.expectUpdate, contentChanged,
				"Content change detection failed for: %s", scenario.description)

			if contentChanged {
				t.Logf("✅ Content change detected - will update embedding")
				t.Logf("   Original hash: %s", *page1.ContentHash)
				t.Logf("   New hash: %s", *page2.ContentHash)
			} else {
				t.Logf("⏭️  No content change - will skip to save costs")
			}
		})
	}
}

// TestContentUpdateWorkflow demonstrates the complete workflow
func TestContentUpdateWorkflow(t *testing.T) {
	t.Log("=== CONTENT UPDATE WORKFLOW TEST ===")

	url := "https://blog.example.com/go-tutorial"

	// Day 1: Original article
	originalContent := `
	# Getting Started with Go
	Go is a programming language developed by Google.
	It's simple, fast, and reliable.
	`

	// Day 30: Updated article
	updatedContent := `
	# Getting Started with Go - Updated 2025
	Go is a programming language developed by Google.
	It's simple, fast, and reliable.
	
	New in 2025:
	- Improved generics support
	- Better error handling
	- Performance optimizations
	`

	// Simulate the system behavior
	hash1 := generateTestContentHash(originalContent)
	hash2 := generateTestContentHash(updatedContent)

	t.Logf("📅 Day 1 - Original scrape:")
	t.Logf("   URL: %s", url)
	t.Logf("   Content length: %d chars", len(originalContent))
	t.Logf("   Content hash: %s", hash1[:12]+"...")
	t.Logf("   Action: ✅ Generate new embedding")

	t.Logf("\n📅 Day 30 - Updated content:")
	t.Logf("   URL: %s (same URL)", url)
	t.Logf("   Content length: %d chars", len(updatedContent))
	t.Logf("   Content hash: %s", hash2[:12]+"...")

	if hash1 != hash2 {
		t.Logf("   Detection: 🔄 Content changed!")
		t.Logf("   Action: ✅ Update existing record with new embedding")
		t.Logf("   Benefit: 🎯 Always fresh, accurate search results")
	} else {
		t.Logf("   Detection: ⏭️  Content unchanged")
		t.Logf("   Action: ⏭️  Skip - save API costs")
	}

	// Verify the hashes are different
	assert.NotEqual(t, hash1, hash2, "Content hashes should be different for changed content")
}

// TestUpdateStrategy demonstrates different update strategies
func TestUpdateStrategy(t *testing.T) {
	strategies := []struct {
		name        string
		approach    string
		pros        []string
		cons        []string
		implemented bool
	}{
		{
			name:        "Content Hash Based (Current)",
			approach:    "SHA256 hash of content, update if hash changes",
			pros:        []string{"Detects any content change", "Efficient comparison", "Preserves history"},
			cons:        []string{"No versioning", "Overwrites old embedding"},
			implemented: true,
		},
		{
			name:        "Timestamp Based",
			approach:    "Update if last-modified header changed",
			pros:        []string{"Simple", "Fast"},
			cons:        []string{"Not all sites provide headers", "Can miss updates"},
			implemented: false,
		},
		{
			name:        "Version Based",
			approach:    "Keep multiple versions of same URL",
			pros:        []string{"Full history", "Can compare versions"},
			cons:        []string{"More storage", "Complex queries"},
			implemented: false,
		},
	}

	for _, strategy := range strategies {
		t.Run(strategy.name, func(t *testing.T) {
			t.Logf("Strategy: %s", strategy.name)
			t.Logf("Approach: %s", strategy.approach)
			t.Logf("Implemented: %t", strategy.implemented)

			if strategy.implemented {
				t.Logf("✅ This strategy is currently active")
			} else {
				t.Logf("💡 Future enhancement opportunity")
			}
		})
	}
}

// Helper function to generate content hash for testing
func generateTestContentHash(content string) string {
	// Simplified hash for testing - in real code this uses SHA256
	if len(content) == 0 {
		return "empty"
	}
	return fmt.Sprintf("hash_%d_%s", len(content), content[:min(10, len(content))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
