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

    // Choose encoding based on configured model where possible. If the configured model
    // is not recognized by tiktoken.GetEncoding, fall back to cl100k_base which is
    // compatible with many OpenAI models.
    model := "cl100k_base"
    if s.config != nil && s.config.OpenAIEmbeddingModel != "" {
        model = s.config.OpenAIEmbeddingModel
    }

    enc, err := tiktoken.GetEncoding(model)
    if err != nil {
        enc, err = tiktoken.GetEncoding("cl100k_base")
        if err != nil {
            // As a final fallback, approximate
            return len(text) / 4
        }
    }

    tokens := enc.Encode(text, nil, nil)
    return len(tokens)
}
