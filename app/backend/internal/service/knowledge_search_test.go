package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/models"
)

// MockKnowledgeRepository is a mock implementation for testing
type MockKnowledgeRepository struct {
	mock.Mock
}

func (m *MockKnowledgeRepository) GetSettings(projectID uuid.UUID) (*models.KnowledgeSettings, error) {
	args := m.Called(projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.KnowledgeSettings), args.Error(1)
}

func (m *MockKnowledgeRepository) CreateSettings(settings *models.KnowledgeSettings) error {
	args := m.Called(settings)
	return args.Error(0)
}

func (m *MockKnowledgeRepository) SearchKnowledgeBase(tenantID, projectID uuid.UUID, embedding interface{}, limit int, threshold float64, includeDocuments, includePages bool) ([]*models.KnowledgeSearchResult, error) {
	args := m.Called(tenantID, projectID, embedding, limit, threshold, includeDocuments, includePages)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.KnowledgeSearchResult), args.Error(1)
}

// MockEmbeddingService is a mock implementation for testing
type MockEmbeddingService struct {
	mock.Mock
}

func (m *MockEmbeddingService) QueryEmbedding(ctx context.Context, query string) (interface{}, error) {
	args := m.Called(ctx, query)
	return args.Get(0), args.Error(1)
}

func (m *MockEmbeddingService) IsEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

// TestKnowledgeSearchWithMissingSettings tests the scenario where knowledge settings don't exist
func TestKnowledgeSearchWithMissingSettings(t *testing.T) {
	mockRepo := new(MockKnowledgeRepository)
	mockEmbedding := new(MockEmbeddingService)

	service := &KnowledgeService{
		knowledgeRepo:    mockRepo,
		embeddingService: mockEmbedding,
	}

	tenantID := uuid.New()
	projectID := uuid.New()
	req := &models.KnowledgeSearchRequest{
		Query:            "test search",
		MaxResults:       10,
		SimilarityScore:  0.7,
		IncludeDocuments: true,
		IncludePages:     true,
	}

	// Mock: GetSettings returns sql.ErrNoRows (settings don't exist)
	mockRepo.On("GetSettings", projectID).Return(nil, sql.ErrNoRows)

	// Mock: CreateSettings succeeds
	mockRepo.On("CreateSettings", mock.AnythingOfType("*models.KnowledgeSettings")).Return(nil)

	// Mock: QueryEmbedding succeeds
	mockEmbedding.On("QueryEmbedding", mock.Anything, "test search").Return([]float32{0.1, 0.2, 0.3}, nil)

	// Mock: SearchKnowledgeBase returns empty results
	mockRepo.On("SearchKnowledgeBase", tenantID, projectID, mock.Anything, 10, 0.7, true, true).Return([]*models.KnowledgeSearchResult{}, nil)

	// Execute
	ctx := context.Background()
	response, err := service.SearchKnowledgeBase(ctx, tenantID, projectID, req)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "test search", response.Query)
	assert.Equal(t, 0, response.TotalCount)
	assert.Len(t, response.Results, 0)

	// Verify mocks were called
	mockRepo.AssertExpectations(t)
	mockEmbedding.AssertExpectations(t)
}

// TestKnowledgeSearchEmbeddingFailure tests the scenario where embedding generation fails
func TestKnowledgeSearchEmbeddingFailure(t *testing.T) {
	mockRepo := new(MockKnowledgeRepository)
	mockEmbedding := new(MockEmbeddingService)

	service := &KnowledgeService{
		knowledgeRepo:    mockRepo,
		embeddingService: mockEmbedding,
	}

	tenantID := uuid.New()
	projectID := uuid.New()
	req := &models.KnowledgeSearchRequest{
		Query:            "test search",
		MaxResults:       10,
		SimilarityScore:  0.7,
		IncludeDocuments: true,
		IncludePages:     true,
	}

	// Mock: GetSettings returns existing settings
	settings := &models.KnowledgeSettings{
		ID:                  uuid.New(),
		TenantID:           tenantID,
		ProjectID:          projectID,
		Enabled:            true,
		EmbeddingModel:     "text-embedding-ada-002",
		ChunkSize:          1000,
		ChunkOverlap:       200,
		MaxContextChunks:   5,
		SimilarityThreshold: 0.7,
	}
	mockRepo.On("GetSettings", projectID).Return(settings, nil)

	// Mock: QueryEmbedding fails (this is the main test case)
	mockEmbedding.On("QueryEmbedding", mock.Anything, "test search").Return(nil, assert.AnError)

	// Execute
	ctx := context.Background()
	response, err := service.SearchKnowledgeBase(ctx, tenantID, projectID, req)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to generate query embedding")

	// Verify mocks were called
	mockRepo.AssertExpectations(t)
	mockEmbedding.AssertExpectations(t)
}

// TestKnowledgeSearchDisabledSettings tests the scenario where knowledge is disabled
func TestKnowledgeSearchDisabledSettings(t *testing.T) {
	mockRepo := new(MockKnowledgeRepository)
	mockEmbedding := new(MockEmbeddingService)

	service := &KnowledgeService{
		knowledgeRepo:    mockRepo,
		embeddingService: mockEmbedding,
	}

	tenantID := uuid.New()
	projectID := uuid.New()
	req := &models.KnowledgeSearchRequest{
		Query:            "test search",
		MaxResults:       10,
		SimilarityScore:  0.7,
		IncludeDocuments: true,
		IncludePages:     true,
	}

	// Mock: GetSettings returns disabled settings
	settings := &models.KnowledgeSettings{
		ID:                  uuid.New(),
		TenantID:           tenantID,
		ProjectID:          projectID,
		Enabled:            false, // DISABLED
		EmbeddingModel:     "text-embedding-ada-002",
		ChunkSize:          1000,
		ChunkOverlap:       200,
		MaxContextChunks:   5,
		SimilarityThreshold: 0.7,
	}
	mockRepo.On("GetSettings", projectID).Return(settings, nil)

	// Execute
	ctx := context.Background()
	response, err := service.SearchKnowledgeBase(ctx, tenantID, projectID, req)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "test search", response.Query)
	assert.Equal(t, 0, response.TotalCount)
	assert.Len(t, response.Results, 0)

	// Verify mocks were called (note: embedding service should NOT be called when disabled)
	mockRepo.AssertExpectations(t)
	mockEmbedding.AssertExpectations(t)
}
