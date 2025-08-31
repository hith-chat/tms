package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pgvector/pgvector-go"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/models"
)

type EmbeddingService struct {
	config     *config.KnowledgeConfig
	httpClient *http.Client
}

func NewEmbeddingService(cfg *config.KnowledgeConfig) *EmbeddingService {
	return &EmbeddingService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.EmbeddingTimeout,
		},
	}
}

// GenerateEmbedding generates an embedding vector for the given text
func (s *EmbeddingService) GenerateEmbedding(ctx context.Context, text string) (pgvector.Vector, error) {
	switch s.config.EmbeddingService {
	case "openai":
		return s.generateOpenAIEmbedding(ctx, text)
	default:
		return pgvector.Vector{}, fmt.Errorf("unsupported embedding service: %s", s.config.EmbeddingService)
	}
}

// GenerateEmbeddings generates embeddings for multiple texts in batch
func (s *EmbeddingService) GenerateEmbeddings(ctx context.Context, texts []string) ([]pgvector.Vector, error) {
	if len(texts) == 0 {
		return []pgvector.Vector{}, nil
	}

	fmt.Printf("Generating embeddings for %d texts using service: %s\n", len(texts), s.config.EmbeddingService)

	switch s.config.EmbeddingService {
	case "openai":
		return s.generateOpenAIEmbeddings(ctx, texts)
	default:
		return nil, fmt.Errorf("unsupported embedding service: %s", s.config.EmbeddingService)
	}
}

// generateOpenAIEmbedding generates embedding using OpenAI API
func (s *EmbeddingService) generateOpenAIEmbedding(ctx context.Context, text string) (pgvector.Vector, error) {
	embeddings, err := s.generateOpenAIEmbeddings(ctx, []string{text})
	if err != nil {
		return pgvector.Vector{}, err
	}

	if len(embeddings) == 0 {
		return pgvector.Vector{}, fmt.Errorf("no embeddings returned")
	}

	return embeddings[0], nil
}

// generateOpenAIEmbeddings generates embeddings using OpenAI API for multiple texts
func (s *EmbeddingService) generateOpenAIEmbeddings(ctx context.Context, texts []string) ([]pgvector.Vector, error) {
	if s.config.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	fmt.Printf("Calling OpenAI API for %d texts using model: %s\n", len(texts), s.config.OpenAIEmbeddingModel)

	// Prepare request
	reqBody := models.EmbeddingRequest{
		Input: texts,
		Model: s.config.OpenAIEmbeddingModel,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.OpenAIAPIKey)

	// Make request
	fmt.Printf("Making OpenAI embeddings API request...\n")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("OpenAI API error - Status: %d, Response: %s\n", resp.StatusCode, string(body))
		return nil, fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var embeddingResp models.EmbeddingResponse
	if err := json.Unmarshal(body, &embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(embeddingResp.Data) != len(texts) {
		return nil, fmt.Errorf("expected %d embeddings, got %d", len(texts), len(embeddingResp.Data))
	}

	// Convert to pgvector format
	embeddings := make([]pgvector.Vector, len(embeddingResp.Data))
	for i, data := range embeddingResp.Data {
		embeddings[i] = models.NewVectorFromFloat32Slice(data.Embedding)
	}

	fmt.Printf("Successfully received %d embeddings from OpenAI API\n", len(embeddings))
	return embeddings, nil
}

// QueryEmbedding creates an embedding for a search query
func (s *EmbeddingService) QueryEmbedding(ctx context.Context, query string) (pgvector.Vector, error) {
	return s.GenerateEmbedding(ctx, query)
}

// IsEnabled returns whether the embedding service is enabled and configured
func (s *EmbeddingService) IsEnabled() bool {
	fmt.Println("Checking if embedding service is enabled...")
	fmt.Println("Embedding service enabled:", s.config.Enabled)
	fmt.Println("OpenAI API key present:", s.config.OpenAIAPIKey != "")
	return s.config.Enabled && s.config.OpenAIAPIKey != ""
}

// GetModel returns the current embedding model being used
func (s *EmbeddingService) GetModel() string {
	return s.config.OpenAIEmbeddingModel
}

// GetDimension returns the embedding dimension for the current model
func (s *EmbeddingService) GetDimension() int {
	// OpenAI text-embedding-ada-002 returns 1536 dimensions
	// OpenAI text-embedding-3-small returns 1536 dimensions
	// OpenAI text-embedding-3-large returns 3072 dimensions
	switch s.config.OpenAIEmbeddingModel {
	case "text-embedding-3-large":
		return 3072
	case "text-embedding-ada-002", "text-embedding-3-small":
		return 1536
	default:
		return 1536 // Default to ada-002 dimensions
	}
}
