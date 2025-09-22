package service

import (
	"strings"

	tiktoken "github.com/pkoukk/tiktoken-go"
)

// estimateTokenCount uses tiktoken to estimate tokens for a given model.
// This is now the default implementation and requires the tiktoken-go module.
func (s *WebScrapingService) estimateTokenCount(text string) int {
	if text == "" {
		return 0
	}

	// Normalize whitespace to make counts more consistent
	text = strings.TrimSpace(text)
	// If no embedding model is configured, use the original simple heuristic
	// to avoid pulling tiktoken into test harnesses or default runs.
	if s == nil || s.config == nil || s.config.OpenAIEmbeddingModel == "" {
		return len(text) / 4
	}

	// Otherwise, attempt to use tiktoken for a model-aware estimate.
	model := s.config.OpenAIEmbeddingModel
	enc, err := tiktoken.GetEncoding(model)
	if err != nil {
		// Try a generic encoding as fallback
		enc, err = tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			// As a final fallback, approximate
			return len(text) / 4
		}
	}

	tokens := enc.Encode(text, nil, nil)
	return len(tokens)
}
