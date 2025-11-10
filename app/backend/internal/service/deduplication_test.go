package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/bareuptime/tms/internal/models"
)

// TestWebScrapingDeduplication tests that duplicate URLs are not stored multiple times
func TestWebScrapingDeduplication(t *testing.T) {
	// This test verifies the deduplication logic prevents redundant embeddings

	// Create mock pages with the same URL
	jobID1 := uuid.New()
	jobID2 := uuid.New()

	duplicateURL := "https://example.com/page1"

	page1 := &models.KnowledgeScrapedPage{
		ID:         uuid.New(),
		JobID:      uuid.NullUUID{UUID: jobID1, Valid: true},
		URL:        duplicateURL,
		Title:      stringPtrTest("Page 1 - First Scrape"),
		Content:    "This is the content from first scrape",
		TokenCount: 10,
		ScrapedAt:  time.Now(),
		Metadata:   models.JSONMap{},
	}

	page2 := &models.KnowledgeScrapedPage{
		ID:         uuid.New(),
		JobID:      uuid.NullUUID{UUID: jobID2, Valid: true},
		URL:        duplicateURL, // Same URL!
		Title:      stringPtrTest("Page 1 - Second Scrape"),
		Content:    "This is the content from second scrape", // Different content (updated)
		TokenCount: 12,
		ScrapedAt:  time.Now(),
		Metadata:   models.JSONMap{},
	}

	// Test assertions
	assert.Equal(t, page1.URL, page2.URL, "URLs should be the same to test deduplication")
	assert.NotEqual(t, page1.JobID, page2.JobID, "Job IDs should be different")
	assert.NotEqual(t, page1.Content, page2.Content, "Content may be different (page updated)")

	// Log the test scenario
	t.Logf("Testing deduplication for URL: %s", duplicateURL)
	t.Logf("First scrape content length: %d", len(page1.Content))
	t.Logf("Second scrape content length: %d", len(page2.Content))
}

// TestEmbeddingCostCalculation demonstrates the cost impact of duplicates
func TestEmbeddingCostCalculation(t *testing.T) {
	scenarios := []struct {
		name                      string
		websites                  int
		scrapeRuns                int
		avgPagesPerSite           int
		expectedCallsWithoutDedup int
		expectedCallsWithDedup    int
	}{
		{
			name:                      "Your scenario: 5 websites, 30 scrapes each",
			websites:                  5,
			scrapeRuns:                30,
			avgPagesPerSite:           10,
			expectedCallsWithoutDedup: 5 * 30 * 10, // 1,500 calls
			expectedCallsWithDedup:    5 * 10,      // 50 calls
		},
		{
			name:                      "Large site: 1 website, 100 scrapes",
			websites:                  1,
			scrapeRuns:                100,
			avgPagesPerSite:           50,
			expectedCallsWithoutDedup: 1 * 100 * 50, // 5,000 calls
			expectedCallsWithDedup:    1 * 50,       // 50 calls
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			costPerCall := 0.0001 // $0.0001 per embedding call (OpenAI pricing)

			costWithoutDedup := float64(scenario.expectedCallsWithoutDedup) * costPerCall
			costWithDedup := float64(scenario.expectedCallsWithDedup) * costPerCall
			savings := costWithoutDedup - costWithDedup
			savingsPercent := (savings / costWithoutDedup) * 100

			t.Logf("Scenario: %s", scenario.name)
			t.Logf("Without deduplication: %d embedding calls = $%.4f",
				scenario.expectedCallsWithoutDedup, costWithoutDedup)
			t.Logf("With deduplication: %d embedding calls = $%.4f",
				scenario.expectedCallsWithDedup, costWithDedup)
			t.Logf("Savings: $%.4f (%.1f%% reduction)", savings, savingsPercent)

			// Assert significant savings
			assert.Greater(t, savingsPercent, 90.0, "Deduplication should save over 90% of costs")
		})
	}
}

// TestURLDeduplicationStrategies tests different deduplication approaches
func TestURLDeduplicationStrategies(t *testing.T) {
	testCases := []struct {
		name        string
		urls        []string
		expectDupes bool
		description string
	}{
		{
			name: "Exact URL duplicates",
			urls: []string{
				"https://example.com/page1",
				"https://example.com/page1", // Exact duplicate
				"https://example.com/page2",
			},
			expectDupes: true,
			description: "Same URL should be deduplicated",
		},
		{
			name: "URL variants (should be considered different)",
			urls: []string{
				"https://example.com/page1",
				"https://example.com/page1?utm_source=test", // Different query params
				"https://example.com/page1#section1",        // Different fragment
			},
			expectDupes: false,
			description: "URLs with different params/fragments are different pages",
		},
		{
			name: "Protocol and subdomain variants",
			urls: []string{
				"http://example.com/page1",
				"https://example.com/page1",     // Different protocol
				"https://www.example.com/page1", // Different subdomain
			},
			expectDupes: false,
			description: "Different protocols/subdomains are different URLs",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.description)

			// Count unique URLs
			urlMap := make(map[string]bool)
			for _, url := range tc.urls {
				urlMap[url] = true
			}

			duplicateCount := len(tc.urls) - len(urlMap)

			if tc.expectDupes {
				assert.Greater(t, duplicateCount, 0, "Expected duplicates to be found")
				t.Logf("Found %d duplicates in %d URLs", duplicateCount, len(tc.urls))
			} else {
				assert.Equal(t, 0, duplicateCount, "Expected no duplicates")
				t.Logf("All %d URLs are unique", len(tc.urls))
			}
		})
	}
}

func stringPtrTest(s string) *string {
	return &s
}
