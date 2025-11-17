package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
)

type KnowledgeService struct {
	knowledgeRepo    *repo.KnowledgeRepository
	embeddingService *EmbeddingService
}

func NewKnowledgeService(knowledgeRepo *repo.KnowledgeRepository, embeddingService *EmbeddingService) *KnowledgeService {
	return &KnowledgeService{
		knowledgeRepo:    knowledgeRepo,
		embeddingService: embeddingService,
	}
}

// SearchKnowledgeBase searches for relevant content in the knowledge base
func (s *KnowledgeService) SearchKnowledgeBase(ctx context.Context, tenantID, projectID uuid.UUID, req *models.KnowledgeSearchRequest) (*models.KnowledgeSearchResponse, error) {
	startTime := time.Now()

	// Get knowledge settings for the project, or create default ones if they don't exist
	settings, err := s.knowledgeRepo.GetSettings(projectID)
	if err != nil {
		// If settings don't exist, create default ones
		if err == sql.ErrNoRows {
			defaultSettings := &models.KnowledgeSettings{
				ID:                  uuid.New(),
				TenantID:            tenantID,
				ProjectID:           projectID,
				Enabled:             true, // Enable by default
				EmbeddingModel:      "text-embedding-ada-002",
				ChunkSize:           1000,
				ChunkOverlap:        200,
				MaxContextChunks:    5,
				SimilarityThreshold: 0.7,
			}

			err = s.knowledgeRepo.CreateSettings(defaultSettings)
			if err != nil {
				return nil, fmt.Errorf("failed to create default knowledge settings: %w", err)
			}
			settings = defaultSettings
		} else {
			return nil, fmt.Errorf("failed to get knowledge settings: %w", err)
		}
	}

	if !settings.Enabled {
		return &models.KnowledgeSearchResponse{
			Results:     []models.KnowledgeSearchResult{},
			TotalCount:  0,
			Query:       req.Query,
			ProcessedIn: time.Since(startTime).String(),
		}, nil
	}

	// Generate embedding for the search query
	fmt.Printf("Generating embedding for search query: %s\n", req.Query)
	queryEmbedding, err := s.embeddingService.QueryEmbedding(ctx, req.Query)
	if err != nil {
		fmt.Printf("ERROR: Failed to generate query embedding: %v\n", err)
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}
	fmt.Printf("Successfully generated embedding for query\n")

	// Use settings for search parameters
	threshold := settings.SimilarityThreshold
	if req.SimilarityScore > 0 {
		threshold = req.SimilarityScore
	}

	maxResults := req.MaxResults
	if maxResults == 0 {
		maxResults = settings.MaxContextChunks
	}

	// Search knowledge base
	fmt.Printf("Searching knowledge base with %d max results, threshold %.2f\n", maxResults, threshold)
	results, err := s.knowledgeRepo.SearchKnowledgeBase(
		tenantID,
		projectID,
		queryEmbedding,
		maxResults,
		threshold,
		req.IncludeDocuments,
		req.IncludePages,
	)
	if err != nil {
		fmt.Printf("ERROR: Failed to search knowledge base: %v\n", err)
		return nil, fmt.Errorf("failed to search knowledge base: %w", err)
	}
	fmt.Printf("Search completed, found %d results\n", len(results))

	response := &models.KnowledgeSearchResponse{
		Results:     make([]models.KnowledgeSearchResult, len(results)),
		TotalCount:  len(results),
		Query:       req.Query,
		ProcessedIn: time.Since(startTime).String(),
	}

	// Convert to response format
	for i, result := range results {
		response.Results[i] = *result
	}

	return response, nil
}

// GetRelevantContext gets relevant context for AI chat based on the message
func (s *KnowledgeService) GetRelevantContext(ctx context.Context, tenantID, projectID uuid.UUID, message string) ([]models.KnowledgeSearchResult, error) {
	// Get knowledge settings
	settings, err := s.knowledgeRepo.GetSettings(projectID)
	if err != nil {
		// If no settings found, knowledge management is not enabled
		return []models.KnowledgeSearchResult{}, nil
	}

	if !settings.Enabled {
		return []models.KnowledgeSearchResult{}, nil
	}

	// Generate embedding for the message
	messageEmbedding, err := s.embeddingService.QueryEmbedding(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to generate message embedding: %w", err)
	}

	// Search for relevant context
	results, err := s.knowledgeRepo.SearchKnowledgeBase(
		tenantID,
		projectID,
		messageEmbedding,
		settings.MaxContextChunks,
		settings.SimilarityThreshold,
		true, // Include documents
		true, // Include pages
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search for relevant context: %w", err)
	}

	// Convert to results format
	contextResults := make([]models.KnowledgeSearchResult, len(results))
	for i, result := range results {
		contextResults[i] = *result
	}

	return contextResults, nil
}

// GetKnowledgeStats returns statistics about the knowledge base
func (s *KnowledgeService) GetKnowledgeStats(ctx context.Context, tenantID, projectID uuid.UUID) (*models.KnowledgeStats, error) {
	return s.knowledgeRepo.GetStats(tenantID, projectID)
}

// GetKnowledgeSettings returns the knowledge settings for a project
func (s *KnowledgeService) GetKnowledgeSettings(ctx context.Context, projectID uuid.UUID) (*models.KnowledgeSettings, error) {
	return s.knowledgeRepo.GetSettings(projectID)
}

// UpdateKnowledgeSettings updates the knowledge settings for a project
func (s *KnowledgeService) UpdateKnowledgeSettings(ctx context.Context, projectID uuid.UUID, req *models.UpdateKnowledgeSettingsRequest) (*models.KnowledgeSettings, error) {
	// Update settings
	if err := s.knowledgeRepo.UpdateSettings(projectID, req); err != nil {
		return nil, fmt.Errorf("failed to update knowledge settings: %w", err)
	}

	// Return updated settings
	return s.knowledgeRepo.GetSettings(projectID)
}

// FormatContextForAI formats knowledge search results for AI context injection
func (s *KnowledgeService) FormatContextForAI(results []models.KnowledgeSearchResult) string {
	if len(results) == 0 {
		return ""
	}

	var contextBuilder strings.Builder
	contextBuilder.WriteString("Relevant knowledge base information:\n\n")

	for i, result := range results {
		contextBuilder.WriteString(fmt.Sprintf("Source %d (%s):\n", i+1, result.Source))
		if result.Title != nil {
			contextBuilder.WriteString(fmt.Sprintf("Title: %s\n", *result.Title))
		}
		contextBuilder.WriteString(fmt.Sprintf("Content: %s\n", result.Content))
		contextBuilder.WriteString(fmt.Sprintf("Relevance Score: %.2f\n\n", result.Score))
	}

	contextBuilder.WriteString("Please use this information to provide accurate and helpful responses. Always cite the sources when possible.")

	return contextBuilder.String()
}

// ReplaceFAQItems overwrites FAQ items for the provided tenant/project
func (s *KnowledgeService) ReplaceFAQItems(ctx context.Context, tenantID, projectID uuid.UUID, items []*models.KnowledgeFAQItem) error {
	for _, item := range items {
		if item.ID == uuid.Nil {
			item.ID = uuid.New()
		}
		item.TenantID = tenantID
		item.ProjectID = projectID
		if item.Metadata == nil {
			item.Metadata = models.JSONMap{}
		}
	}

	return s.knowledgeRepo.ReplaceFAQItems(ctx, tenantID, projectID, items)
}

// ListFAQItems returns FAQ entries for a project
func (s *KnowledgeService) ListFAQItems(ctx context.Context, tenantID, projectID uuid.UUID) ([]*models.KnowledgeFAQItem, error) {
	items, err := s.knowledgeRepo.ListFAQItems(ctx, tenantID, projectID)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// GetWidgetKnowledgePagesByProject retrieves all knowledge pages associated with widgets in a project
func (s *KnowledgeService) GetWidgetKnowledgePagesByProject(ctx context.Context, projectID uuid.UUID, widgetID *uuid.UUID) ([]*models.WidgetKnowledgePageWithDetails, error) {
	pages, err := s.knowledgeRepo.GetWidgetKnowledgePagesByProject(ctx, projectID, widgetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get widget knowledge pages: %w", err)
	}
	return pages, nil
}

// DeleteWidgetKnowledgePageMapping removes the association between a widget and a knowledge page
func (s *KnowledgeService) DeleteWidgetKnowledgePageMapping(ctx context.Context, mappingID uuid.UUID) error {
	err := s.knowledgeRepo.DeleteWidgetKnowledgePageMapping(ctx, mappingID)
	if err != nil {
		return fmt.Errorf("failed to delete widget knowledge page mapping: %w", err)
	}
	return nil
}
