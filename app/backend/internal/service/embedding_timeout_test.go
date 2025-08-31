package service

import (
	"context"
	"testing"
	"time"

	"github.com/bareuptime/tms/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestEmbeddingTimeoutConfiguration(t *testing.T) {
	tests := []struct {
		name             string
		embeddingTimeout time.Duration
		expectedTimeout  time.Duration
	}{
		{
			name:             "default timeout",
			embeddingTimeout: 120 * time.Second,
			expectedTimeout:  120 * time.Second,
		},
		{
			name:             "custom timeout",
			embeddingTimeout: 60 * time.Second,
			expectedTimeout:  60 * time.Second,
		},
		{
			name:             "long timeout for large batches",
			embeddingTimeout: 300 * time.Second,
			expectedTimeout:  300 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.KnowledgeConfig{
				Enabled:          true,
				EmbeddingService: "openai",
				OpenAIAPIKey:     "test-key",
				EmbeddingTimeout: tt.embeddingTimeout,
			}

			service := NewEmbeddingService(cfg)
			
			// Verify the HTTP client timeout is set correctly
			assert.Equal(t, tt.expectedTimeout, service.httpClient.Timeout)
		})
	}
}

func TestEmbeddingServiceTimeoutRespect(t *testing.T) {
	cfg := &config.KnowledgeConfig{
		Enabled:          true,
		EmbeddingService: "openai",
		OpenAIAPIKey:     "test-key",
		EmbeddingTimeout: 100 * time.Millisecond, // Very short timeout for testing
	}

	service := NewEmbeddingService(cfg)

	// Create a context that will be canceled quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// This should fail quickly due to context timeout
	start := time.Now()
	_, err := service.GenerateEmbedding(ctx, "test text")
	elapsed := time.Since(start)

	// Should fail within the context timeout (50ms) + some buffer
	assert.Error(t, err)
	assert.Less(t, elapsed, 200*time.Millisecond, "Request should timeout quickly")
	assert.Contains(t, err.Error(), "context", "Error should mention context cancellation")
}

func TestEmbeddingServiceConfiguration(t *testing.T) {
	cfg := &config.KnowledgeConfig{
		Enabled:              true,
		EmbeddingService:     "openai",
		OpenAIEmbeddingModel: "text-embedding-ada-002",
		OpenAIAPIKey:         "test-key",
		EmbeddingTimeout:     120 * time.Second,
	}

	service := NewEmbeddingService(cfg)

	assert.True(t, service.IsEnabled())
	assert.Equal(t, "text-embedding-ada-002", service.GetModel())
	assert.Equal(t, 1536, service.GetDimension())
}
