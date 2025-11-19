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
	"github.com/pgvector/pgvector-go"

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

	prompt += fmt.Sprintf(`

IMPORTANT: Respond with ONLY a valid JSON array containing the URLs of the most important pages.
- If there are 8 or more pages available, select exactly 8 URLs
- If there are fewer than 8 pages available, return all of them (e.g., if only 2 pages, return those 2)
- Do NOT include any explanations, notes, or other text
- Return ONLY the JSON array

Example format:
["https://example.com/page1", "https://example.com/page2"]

Your JSON response:`)


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

	// Try to extract JSON array if there's extra text
	// Look for the first '[' and last ']' to extract just the JSON array
	startIdx := strings.Index(response, "[")
	endIdx := strings.LastIndex(response, "]")

	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		return nil, fmt.Errorf("failed to parse AI ranking response: no valid JSON array found in response: %s", response)
	}

	jsonStr := response[startIdx : endIdx+1]

	var rankedURLs []string
	if err := json.Unmarshal([]byte(jsonStr), &rankedURLs); err != nil {
		return nil, fmt.Errorf("failed to parse AI ranking response: %w. Extracted JSON: %s", err, jsonStr)
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

// PageContentInfo holds URL content and calculated hash
type PageContentInfo struct {
	URL         string
	Content     string
	Title       string
	ContentHash string
}

// embedAndStoreTop8 reads content from files, generates embeddings with deduplication, and stores in pgvector
// Implements content-hash based deduplication to avoid regenerating embeddings for unchanged content
func (s *PublicAIBuilderService) embedAndStoreTop8(
	ctx context.Context,
	tenantID, projectID, widgetID, jobID uuid.UUID,
	top8URLs []string,
	workDir string,
	events chan<- AIBuilderEvent,
) error {
	// STEP 1: Read content from files for top 8 URLs and calculate content hashes
	var pageContents []PageContentInfo
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

		// Parse metadata and content (first 3 lines are metadata: URL, Title, Scraped)
		lines := strings.Split(string(data), "\n")
		title := urlStr
		contentStart := 0
		for i, line := range lines {
			if strings.HasPrefix(line, "Title:") {
				title = strings.TrimPrefix(line, "Title:")
				title = strings.TrimSpace(title)
			}
			if strings.TrimSpace(line) == "" && i > 0 {
				contentStart = i + 1
				break
			}
		}

		content := strings.Join(lines[contentStart:], "\n")

		// Calculate content hash for deduplication
		contentHash := sha256.Sum256([]byte(content))
		contentHashStr := hex.EncodeToString(contentHash[:])

		pageContents = append(pageContents, PageContentInfo{
			URL:         urlStr,
			Content:     content,
			Title:       title,
			ContentHash: contentHashStr,
		})
	}

	if len(pageContents) == 0 {
		return fmt.Errorf("no content available for top 8 URLs")
	}

	// STEP 2: Query existing pages by tenant_id to check for duplicates
	s.emit(ctx, events, AIBuilderEvent{
		Type:    "deduplication_check",
		Stage:   "embedding_storage",
		Message: "Checking for existing content to avoid duplicate embeddings...",
	})

	urls := make([]string, len(pageContents))
	for i, pc := range pageContents {
		urls[i] = pc.URL
	}

	existingPages, err := s.webScrapingService.knowledgeRepo.GetExistingPagesByTenantAndURLs(ctx, tenantID, urls)
	if err != nil {
		return fmt.Errorf("failed to query existing pages: %w", err)
	}

	// STEP 3: Categorize pages into new, changed, and unchanged
	type pageWithStatus struct {
		PageContentInfo
		Status     string // "new", "changed", "unchanged"
		ExistingID *uuid.UUID
	}

	var pagesWithStatus []pageWithStatus
	var needsEmbedding []PageContentInfo // Only new and changed pages
	var reusedCount int

	for _, pc := range pageContents {
		existing, exists := existingPages[pc.URL]

		if !exists {
			// New URL - needs embedding
			pagesWithStatus = append(pagesWithStatus, pageWithStatus{
				PageContentInfo: pc,
				Status:          "new",
				ExistingID:      nil,
			})
			needsEmbedding = append(needsEmbedding, pc)
		} else if existing.ContentHash == nil || *existing.ContentHash != pc.ContentHash {
			// Content changed - needs new embedding
			pagesWithStatus = append(pagesWithStatus, pageWithStatus{
				PageContentInfo: pc,
				Status:          "changed",
				ExistingID:      &existing.ID,
			})
			needsEmbedding = append(needsEmbedding, pc)
		} else {
			// Content unchanged - reuse existing embedding
			pagesWithStatus = append(pagesWithStatus, pageWithStatus{
				PageContentInfo: pc,
				Status:          "unchanged",
				ExistingID:      &existing.ID,
			})
			reusedCount++
		}
	}

	logger.GetTxLogger(ctx).Info().
		Int("new", len(needsEmbedding)-reusedCount).
		Int("changed", 0). // Will be calculated properly in loop
		Int("unchanged", reusedCount).
		Int("total", len(pageContents)).
		Msg("Content-hash deduplication analysis completed")

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "deduplication_complete",
		Stage:   "embedding_storage",
		Message: fmt.Sprintf("Deduplication: %d new/changed, %d reused (%.1f%% savings)",
			len(needsEmbedding), reusedCount,
			float64(reusedCount)/float64(len(pageContents))*100),
	})

	// STEP 4: Generate embeddings only for new and changed pages
	var newEmbeddings []pgvector.Vector
	if len(needsEmbedding) > 0 {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "embedding_in_progress",
			Stage:   "embedding_storage",
			Message: fmt.Sprintf("Generating embeddings for %d URLs (saved %d API calls)...",
				len(needsEmbedding), reusedCount),
		})

		texts := make([]string, len(needsEmbedding))
		for i, pc := range needsEmbedding {
			texts[i] = pc.Content
		}

		newEmbeddings, err = s.webScrapingService.embeddingService.GenerateEmbeddings(ctx, texts)
		if err != nil {
			return fmt.Errorf("failed to generate embeddings: %w", err)
		}

		s.emit(ctx, events, AIBuilderEvent{
			Type:    "embedding_completed",
			Stage:   "embedding_storage",
			Message: fmt.Sprintf("Generated %d embeddings (%.1f%% cost reduction)",
				len(newEmbeddings), float64(reusedCount)/float64(len(pageContents))*100),
		})
	}

	// STEP 5: Create/update pages in database and collect page IDs
	s.emit(ctx, events, AIBuilderEvent{
		Type:    "storage_in_progress",
		Stage:   "embedding_storage",
		Message: "Storing pages in knowledge base...",
	})

	var pageIDs []uuid.UUID
	embeddingIndex := 0

	for i, ps := range pagesWithStatus {
		var pageID uuid.UUID

		if ps.Status == "unchanged" {
			// Reuse existing page - no database changes needed
			pageID = *ps.ExistingID
		} else if ps.Status == "new" {
			// New page - create with embedding
			if embeddingIndex >= len(newEmbeddings) {
				logger.GetTxLogger(ctx).Error().Msg("Embedding index out of range")
				continue
			}
			embedding := newEmbeddings[embeddingIndex]
			embeddingIndex++

			// Store the new page and get the returned page ID
			createdPageID, err := s.webScrapingService.StorePageInVectorDBWithTenantID(
				ctx, tenantID, projectID, ps.URL, ps.Title, ps.Content, embedding, jobID,
			)
			if err != nil {
				logger.GetTxLogger(ctx).Error().
					Str("url", ps.URL).
					Err(err).
					Msg("Failed to store new page in vector DB")
				continue
			}
			pageID = createdPageID
		} else if ps.Status == "changed" {
			// Changed page - update existing record with new embedding
			if embeddingIndex >= len(newEmbeddings) {
				logger.GetTxLogger(ctx).Error().Msg("Embedding index out of range")
				continue
			}
			embedding := newEmbeddings[embeddingIndex]
			embeddingIndex++

			// Update the existing page (preserves original job_id and page ID)
			err := s.webScrapingService.UpdatePageInVectorDB(
				ctx, *ps.ExistingID, ps.Title, ps.Content, embedding,
			)
			if err != nil {
				logger.GetTxLogger(ctx).Error().
					Str("url", ps.URL).
					Err(err).
					Msg("Failed to update changed page in vector DB")
				continue
			}
			pageID = *ps.ExistingID
		}

		pageIDs = append(pageIDs, pageID)

		s.emit(ctx, events, AIBuilderEvent{
			Type:    "storage_progress",
			Stage:   "embedding_storage",
			Message: fmt.Sprintf("Processed %d/%d pages (%s: %s)",
				i+1, len(pagesWithStatus), ps.Status, ps.URL),
		})
	}

	// STEP 6: Create project_knowledge_pages mappings
	s.emit(ctx, events, AIBuilderEvent{
		Type:    "mapping_in_progress",
		Stage:   "embedding_storage",
		Message: "Creating project-to-page associations...",
	})

	err = s.webScrapingService.knowledgeRepo.CreateProjectKnowledgePageMappings(ctx, tenantID, projectID, pageIDs)
	if err != nil {
		return fmt.Errorf("failed to create project-page mappings: %w", err)
	}

	logger.GetTxLogger(ctx).Info().
		Int("total_pages", len(pageContents)).
		Int("new_embeddings", len(newEmbeddings)).
		Int("reused_embeddings", reusedCount).
		Float64("savings_percent", float64(reusedCount)/float64(len(pageContents))*100).
		Msg("Successfully completed embedding storage with deduplication")

	return nil
}
