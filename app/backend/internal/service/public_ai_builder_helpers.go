package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/logger"
	"github.com/bareuptime/tms/internal/models"
)

// ScrapedFileInfo holds information about a scraped URL and its file
type ScrapedFileInfo struct {
	URL      string
	FilePath string
	Title    string
	Preview  string // First 300 chars of content
}

// scrapeURLsToFiles scrapes all URLs and saves content to text files
// Returns a list of file info for successfully scraped pages
func (s *PublicAIBuilderService) scrapeURLsToFiles(
	ctx context.Context,
	urls []string,
	workDir string,
	events chan<- AIBuilderEvent,
) ([]ScrapedFileInfo, error) {
	type scrapResult struct {
		URL      string
		Content  string
		Title    string
		FilePath string
		Error    error
	}

	resultChan := make(chan scrapResult, len(urls))
	jobChan := make(chan string, len(urls))

	// Add jobs to channel
	for _, url := range urls {
		jobChan <- url
	}
	close(jobChan)

	// Start 10 parallel workers
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for urlStr := range jobChan {
				// Scrape content
				content, err := s.webScrapingService.ScrapePageContent(ctx, urlStr)
				if err != nil {
					resultChan <- scrapResult{URL: urlStr, Error: err}
					continue
				}

				// Extract title (first line or first 100 chars)
				title := urlStr
				lines := strings.Split(content, "\n")
				if len(lines) > 0 && len(lines[0]) > 0 {
					title = strings.TrimSpace(lines[0])
					if len(title) > 100 {
						title = title[:100]
					}
				}

				// Generate SHA256 hash of URL for filename
				hash := sha256.Sum256([]byte(urlStr))
				filename := hex.EncodeToString(hash[:]) + ".txt"
				filePath := filepath.Join(workDir, filename)

				// Prepare file content with metadata
				fileContent := fmt.Sprintf("URL: %s\nTitle: %s\nScraped: %s\n\n%s",
					urlStr, title, time.Now().Format(time.RFC3339), content)

				// Write to file
				if err := os.WriteFile(filePath, []byte(fileContent), 0644); err != nil {
					resultChan <- scrapResult{URL: urlStr, Error: fmt.Errorf("failed to write file: %w", err)}
					continue
				}

				resultChan <- scrapResult{
					URL:      urlStr,
					Content:  content,
					Title:    title,
					FilePath: filePath,
					Error:    nil,
				}
			}
		}(i)
	}

	// Wait for all workers
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var scrapedFiles []ScrapedFileInfo
	successCount := 0
	failCount := 0

	for result := range resultChan {
		if result.Error != nil {
			failCount++
			s.emit(ctx, events, AIBuilderEvent{
				Type:    "scraping_failed",
				Stage:   "scraping",
				Message: fmt.Sprintf("Failed to scrape %s", result.URL),
				Detail:  result.Error.Error(),
			})
		} else {
			successCount++
			preview := result.Content
			if len(preview) > 300 {
				preview = preview[:300]
			}

			scrapedFiles = append(scrapedFiles, ScrapedFileInfo{
				URL:      result.URL,
				FilePath: result.FilePath,
				Title:    result.Title,
				Preview:  preview,
			})

			if successCount%5 == 0 {
				s.emit(ctx, events, AIBuilderEvent{
					Type:    "scraping_progress",
					Stage:   "scraping",
					Message: fmt.Sprintf("Scraped %d/%d URLs", successCount, len(urls)),
				})
			}
		}
	}

	logger.GetTxLogger(ctx).Info().
		Int("success", successCount).
		Int("failed", failCount).
		Int("total", len(urls)).
		Msg("Scraping to files completed")

	return scrapedFiles, nil
}

// rankURLsWithAI uses AI to rank scraped URLs and select top 8 most relevant
func (s *PublicAIBuilderService) rankURLsWithAI(
	ctx context.Context,
	scrapedFiles []ScrapedFileInfo,
	workDir string,
	widget *models.ChatWidget,
	events chan<- AIBuilderEvent,
) ([]string, error) {
	// Build the ranking prompt
	prompt := fmt.Sprintf(`You are analyzing a website to build a knowledge base for a chat widget.

Chat Widget Theme:
- Brand/Agent Name: %s
- Welcome Message: %s
- Widget Purpose: Help users with questions about the website

I have scraped %d pages from the website. Please analyze these pages and select the TOP 8 most relevant and valuable pages for the knowledge base.

Select pages that:
1. Contain core information about the product/service
2. Answer common user questions
3. Provide documentation, guides, or tutorials
4. Are most representative of the website's purpose

Pages to analyze:
`, widget.AgentName, widget.WelcomeMessage, len(scrapedFiles))

	// Add each URL with title and preview
	for i, file := range scrapedFiles {
		prompt += fmt.Sprintf("\n%d. URL: %s\n   Title: %s\n   Preview: %s\n",
			i+1, file.URL, file.Title, file.Preview)
	}

	prompt += `\n\nRespond with ONLY a JSON array of exactly 8 URLs (the most important ones). Example format:
["https://example.com/page1", "https://example.com/page2", ...]

Response:`

	// Call AI service to rank
	s.emit(ctx, events, AIBuilderEvent{
		Type:    "ai_ranking_in_progress",
		Stage:   "ai_ranking",
		Message: "AI is analyzing and ranking URLs...",
	})

	response, err := s.aiBuilderService.aiService.CallAIForRanking(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI ranking failed: %w", err)
	}

	// Clean and parse JSON response
	response = stripMarkdownCodeBlocks(response)
	response = strings.TrimSpace(response)

	var rankedURLs []string
	if err := json.Unmarshal([]byte(response), &rankedURLs); err != nil {
		return nil, fmt.Errorf("failed to parse AI ranking response: %w. Response: %s", err, response)
	}

	// Validate we got exactly 8 URLs (or less if fewer available)
	maxURLs := 8
	if len(scrapedFiles) < maxURLs {
		maxURLs = len(scrapedFiles)
	}

	if len(rankedURLs) > maxURLs {
		rankedURLs = rankedURLs[:maxURLs]
	}

	if len(rankedURLs) == 0 {
		return nil, fmt.Errorf("AI returned no URLs")
	}

	logger.GetTxLogger(ctx).Info().
		Int("total_urls", len(scrapedFiles)).
		Int("selected", len(rankedURLs)).
		Strs("top_urls", rankedURLs).
		Msg("AI ranking completed")

	return rankedURLs, nil
}

// embedAndStoreTop8 reads content from files, generates embeddings, and stores in pgvector
func (s *PublicAIBuilderService) embedAndStoreTop8(
	ctx context.Context,
	tenantID, projectID, widgetID, jobID uuid.UUID,
	top8URLs []string,
	workDir string,
	events chan<- AIBuilderEvent,
) error {
	// Read content from files for top 8 URLs
	type urlContent struct {
		URL     string
		Content string
	}

	var contents []urlContent
	for _, urlStr := range top8URLs {
		// Generate filename from URL hash
		hash := sha256.Sum256([]byte(urlStr))
		filename := hex.EncodeToString(hash[:]) + ".txt"
		filePath := filepath.Join(workDir, filename)

		// Read file
		data, err := os.ReadFile(filePath)
		if err != nil {
			logger.GetTxLogger(ctx).Warn().
				Str("url", urlStr).
				Str("file", filePath).
				Err(err).
				Msg("Failed to read scraped file for top URL")
			continue
		}

		// Parse content (skip metadata lines)
		lines := strings.Split(string(data), "\n")
		contentStart := 0
		for i, line := range lines {
			if strings.TrimSpace(line) == "" && i > 0 {
				contentStart = i + 1
				break
			}
		}

		content := strings.Join(lines[contentStart:], "\n")
		contents = append(contents, urlContent{
			URL:     urlStr,
			Content: content,
		})
	}

	if len(contents) == 0 {
		return fmt.Errorf("no content available for top 8 URLs")
	}

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "embedding_in_progress",
		Stage:   "embedding_storage",
		Message: fmt.Sprintf("Generating embeddings for %d URLs...", len(contents)),
	})

	// Generate embeddings in batch
	texts := make([]string, len(contents))
	for i, c := range contents {
		texts[i] = c.Content
	}

	embeddings, err := s.webScrapingService.embeddingService.GenerateEmbeddings(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "embedding_completed",
		Stage:   "embedding_storage",
		Message: fmt.Sprintf("Generated %d embeddings", len(embeddings)),
	})

	// Store in pgvector
	s.emit(ctx, events, AIBuilderEvent{
		Type:    "storage_in_progress",
		Stage:   "embedding_storage",
		Message: "Storing embeddings in knowledge base...",
	})

	for i, content := range contents {
		if i >= len(embeddings) {
			break
		}

		err := s.webScrapingService.StorePageInVectorDB(
			ctx, tenantID, projectID, content.URL, content.Content, embeddings[i], jobID,
		)

		if err != nil {
			logger.GetTxLogger(ctx).Error().
				Str("url", content.URL).
				Err(err).
				Msg("Failed to store page in vector DB")
			continue
		}

		s.emit(ctx, events, AIBuilderEvent{
			Type:    "storage_progress",
			Stage:   "embedding_storage",
			Message: fmt.Sprintf("Stored %d/%d pages", i+1, len(contents)),
		})
	}

	logger.GetTxLogger(ctx).Info().
		Int("stored_count", len(contents)).
		Msg("Successfully stored all embeddings in pgvector")

	return nil
}
