package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeURLExtractor struct {
	result []discoveredLink
	err    error
}

func (f *fakeURLExtractor) extractURLsManually(ctx context.Context, targetURL string, maxDepth int) ([]discoveredLink, error) {
	return f.result, f.err
}

func TestPublicURLAnalysisService_AnalyzeURLWithStream_InvalidURL(t *testing.T) {
	t.Parallel()

	svc := &PublicURLAnalysisService{
		webScrapingService: &fakeURLExtractor{},
	}

	events := make(chan URLAnalysisEvent, 1)

	res, err := svc.AnalyzeURLWithStream(context.Background(), URLAnalysisRequest{URL: "not-a-url"}, events)

	require.Error(t, err)
	require.Nil(t, res)

	emitted := collectEvents(events)
	require.Len(t, emitted, 1)
	require.Equal(t, "error", emitted[0].Type)
	require.Contains(t, emitted[0].Message, "Invalid URL")
}

func TestPublicURLAnalysisService_AnalyzeURLWithStream_ExtractionError(t *testing.T) {
	t.Parallel()

	svc := &PublicURLAnalysisService{
		webScrapingService: &fakeURLExtractor{err: errors.New("boom")},
	}

	events := make(chan URLAnalysisEvent, 2)

	res, err := svc.AnalyzeURLWithStream(context.Background(), URLAnalysisRequest{URL: "https://example.com"}, events)

	require.Error(t, err)
	require.Nil(t, res)

	emitted := collectEvents(events)
	var foundError bool
	for _, evt := range emitted {
		if evt.Type == "error" {
			foundError = true
			require.Contains(t, evt.Message, "Failed to analyze URL")
		}
	}
	require.True(t, foundError, "expected error event to be emitted")
}

func TestPublicURLAnalysisService_AnalyzeURLWithStream_Success(t *testing.T) {
	t.Parallel()

	links := []discoveredLink{
		{URL: "https://example.com", Title: "Home", Depth: 0, TokenCount: 100},
		{URL: "https://example.com/about", Title: "About", Depth: 1, TokenCount: 50},
	}

	events := make(chan URLAnalysisEvent, len(links)+2)

	svc := &PublicURLAnalysisService{
		webScrapingService: &fakeURLExtractor{result: links},
	}

	result, err := svc.AnalyzeURLWithStream(context.Background(), URLAnalysisRequest{URL: "https://example.com"}, events)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "https://example.com", result.RootURL)
	require.Equal(t, 1, result.MaxDepth)
	require.Equal(t, len(links), result.TotalURLs)
	require.Equal(t, 150, result.TotalTokens)
	require.Len(t, result.URLs, 2)
	require.WithinDuration(t, time.Now(), result.GeneratedAt, time.Second)

	emitted := collectEvents(events)
	var completed bool
	var progress bool
	var urlsFound int
	for _, evt := range emitted {
		switch evt.Type {
		case "progress":
			progress = true
			require.Equal(t, "https://example.com", evt.URL)
		case "url_found":
			urlsFound++
			require.NotEmpty(t, evt.URL)
		case "completed":
			completed = true
			require.Equal(t, result.TotalURLs, evt.Total)
		}
	}

	require.True(t, progress, "expected progress event")
	require.Equal(t, len(links), urlsFound)
	require.True(t, completed, "expected completed event")
}

func collectEvents(ch chan URLAnalysisEvent) []URLAnalysisEvent {
	var events []URLAnalysisEvent
	for len(ch) > 0 {
		events = append(events, <-ch)
	}
	return events
}
