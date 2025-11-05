package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewComprehensiveURLExtractor_InvalidBaseURL(t *testing.T) {
	t.Parallel()

	extractor, err := NewComprehensiveURLExtractor("://bad-url")
	require.Error(t, err)
	require.Nil(t, extractor)
}

func TestComprehensiveURLExtractor_ExtractURLsFromHTML(t *testing.T) {
	t.Parallel()

	extractor, err := NewComprehensiveURLExtractor("https://example.com")
	require.NoError(t, err)

	const html = `<!DOCTYPE html>
<html>
<head>
<link rel="canonical" href="/docs/">
<meta property="og:url" content="/current/overview#top">
<link rel="stylesheet" href="/styles/main.css">
</head>
<body>
<a href="/docs/">Docs</a>
<div style="background-image: url('/assets/pattern');"></div>
<img srcset="/image-1x 1x, https://cdn.example.com/photo 2x">
<img src="/image.png">
<script>var api = "https://sub.example.com/api";</script>
<span data-url="//external.example.org/home/"></span>
<p>Check https://plain.example.net/path and https://example.com/#ignored</p>
</body>
</html>`

	urls := extractor.ExtractURLsFromHTML([]byte(html), "https://example.com/blog/article?lang=en#section")
	require.GreaterOrEqual(t, len(urls), 6)

	urlSet := make(map[string]struct{}, len(urls))
	for _, u := range urls {
		urlSet[u] = struct{}{}
	}

	for _, expected := range []string{
		"https://example.com/docs",
		"https://example.com/assets/pattern",
		"https://cdn.example.com/photo",
		"https://external.example.org/home",
		"https://example.com/current/overview",
	} {
		_, ok := urlSet[expected]
		require.Truef(t, ok, "expected URL %q to be extracted", expected)
	}

	for _, disallowed := range []string{
		"https://example.com/styles/main.css",
		"https://example.com/image.png",
	} {
		_, ok := urlSet[disallowed]
		require.Falsef(t, ok, "did not expect asset URL %q", disallowed)
	}
}

func TestComprehensiveURLExtractor_SkipsVisitedAndInvalid(t *testing.T) {
	t.Parallel()

	extractor, err := NewComprehensiveURLExtractor("https://example.com")
	require.NoError(t, err)

	extractor.visitedURLs["https://example.com/visited"] = true

	urls := extractor.ExtractURLsFromHTML([]byte(`<a href="/visited">Visited</a><a href="mailto:test@example.com">Mail</a>`), "https://example.com/page")

	require.Empty(t, urls)
}
