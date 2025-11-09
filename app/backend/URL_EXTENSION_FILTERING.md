# URL Extension Filtering Configuration

## Overview

The web scraping system now includes configurable URL extension filtering to prevent crawling unwanted file types (e.g., `.js`, `.css`, `.jpg`, `.suman`). This feature helps optimize crawling performance and reduces unnecessary downloads.

## How It Works

The system checks each discovered URL's file extension before adding it to the crawl queue:

- **URLs with no extension** (e.g., `https://example.com/about`) → ✅ **Allowed**
- **URLs with allowed extensions** (e.g., `https://example.com/page.html`) → ✅ **Allowed**
- **URLs with disallowed extensions** (e.g., `https://example.com/script.js`) → ❌ **Rejected**

## Default Configuration

By default, the following file extensions are allowed:

```yaml
knowledge:
  allowed_file_extensions:
    - .html
    - .htm
    - .md
    - .markdown
    - .txt
    - .pdf
```

## Configuration Methods

### Method 1: YAML Configuration File

Edit your `config.yaml` file:

```yaml
knowledge:
  enabled: true
  # ... other config ...

  # Allowed file extensions for web scraping
  allowed_file_extensions:
    - .html
    - .htm
    - .md
    - .markdown
    - .txt
    - .pdf
    # Add custom extensions as needed:
    # - .rst
    # - .adoc
```

### Method 2: Environment Variable

Set the `KNOWLEDGE_ALLOWED_FILE_EXTENSIONS` environment variable:

```bash
# Single extension
export KNOWLEDGE_ALLOWED_FILE_EXTENSIONS=".html"

# Multiple extensions (comma-separated)
export KNOWLEDGE_ALLOWED_FILE_EXTENSIONS=".html,.htm,.md,.txt,.pdf"
```

### Method 3: Docker Compose

Add to your `docker-compose.yml`:

```yaml
services:
  backend:
    environment:
      - KNOWLEDGE_ALLOWED_FILE_EXTENSIONS=.html,.htm,.md,.markdown,.txt,.pdf
```

## Use Cases

### 1. Documentation Sites Only

For sites that only serve Markdown documentation:

```yaml
knowledge:
  allowed_file_extensions:
    - .md
    - .markdown
```

### 2. Web Pages Only

For traditional HTML websites:

```yaml
knowledge:
  allowed_file_extensions:
    - .html
    - .htm
```

### 3. API Documentation

For sites serving JSON/XML documentation:

```yaml
knowledge:
  allowed_file_extensions:
    - .json
    - .xml
    - .html
```

### 4. Allow Everything with Extensions

To disable filtering and allow all file types:

```yaml
knowledge:
  allowed_file_extensions: []  # Empty list = only URLs without extensions
```

**Note:** An empty list will reject ALL files with extensions, but still allows clean URLs (e.g., `/docs/guide`).

## Examples

### Example 1: Preventing Image/Asset Crawling

**Problem:** Your crawler is downloading images, stylesheets, and JavaScript files.

**Solution:**

```yaml
knowledge:
  allowed_file_extensions:
    - .html
    - .htm
    - .md
    # Excluded: .jpg, .png, .css, .js, .svg, etc.
```

**Result:**
- ✅ `https://example.com/docs/guide.html`
- ✅ `https://example.com/blog/post.md`
- ✅ `https://example.com/about` (no extension)
- ❌ `https://example.com/logo.png`
- ❌ `https://example.com/styles.css`
- ❌ `https://example.com/app.js`

### Example 2: Custom File Extensions

**Problem:** A site uses custom extensions like `.suman`, `.custom`.

**Solution:**

```yaml
knowledge:
  allowed_file_extensions:
    - .html
    - .md
    # .suman is NOT in the list
```

**Result:**
- ✅ `https://docs.example.com/guide.html`
- ❌ `https://docs.penify.dev/docs/guide.suman` (rejected)

### Example 3: Supporting Specific Documentation Formats

**Scenario:** Crawling technical documentation that uses reStructuredText and AsciiDoc.

```yaml
knowledge:
  allowed_file_extensions:
    - .html
    - .rst       # reStructuredText
    - .adoc      # AsciiDoc
    - .asciidoc  # AsciiDoc variant
    - .md
```

## Behavior Details

### Case-Insensitive Matching

Extensions are matched case-insensitively:

- `.HTML`, `.html`, `.HtMl` → All treated as `.html`

### Query Parameters and Fragments

Query parameters and fragments are ignored when checking extensions:

- `https://example.com/page.html?ref=123#section` → Extension is `.html`

### Trailing Slashes

URLs with trailing slashes after an extension are treated as directories:

- `https://example.com/page.html/` → No extension (treated as directory)
- `https://example.com/page.html` → Extension is `.html`

### Subdirectories

The filter only checks the last path segment:

- `https://example.com/v1.2.3/docs/guide.html` → Extension is `.html`
- `https://example.com/api/v1.0/users` → No extension

## Performance Impact

### Before Extension Filtering

```
Crawling https://example.com
├── /docs/guide.html      ✅ Useful content
├── /docs/logo.png        ❌ Wasted bandwidth
├── /docs/styles.css      ❌ Wasted bandwidth
├── /docs/app.js          ❌ Wasted bandwidth
├── /docs/favicon.ico     ❌ Wasted bandwidth
└── /docs/guide.suman     ❌ Unsupported format
Total: 6 URLs crawled (only 1 useful)
```

### After Extension Filtering

```
Crawling https://example.com
├── /docs/guide.html      ✅ Useful content
└── /docs/index.html      ✅ Useful content
Total: 2 URLs crawled (both useful)
⚡ 66% reduction in unnecessary requests
```

## Integration Points

The extension filter is applied at 4 key points in the crawling pipeline:

1. **Initial URL discovery** (`ExtractSelectableLinks`)
2. **Playwright URL extraction** (`ExtractURLsWithStream` - depth 0-1)
3. **Colly URL extraction** (`extractPageWithColly` - depth >= 2)
4. **Parallel processing** (`ExtractURLsWithStream` - worker pool)

## Debugging

To see which URLs are being rejected, check the logs:

```bash
# Enable debug logging
export LOG_LEVEL=debug

# Look for extension filtering messages
docker-compose logs -f backend | grep "Skipping URL"
```

Example log output:

```
DEBUG Skipping URL https://example.com/logo.png with disallowed file extension: .png
DEBUG Skipping URL https://example.com/script.js with disallowed file extension: .js
DEBUG Skipping URL https://docs.penify.dev/guide.suman with disallowed file extension: .suman
```

## Testing

Run the test suite to verify extension filtering behavior:

```bash
cd app/backend/internal/service
go test -run TestShouldCrawlURLByExtension -v
```

**Test Coverage:**
- ✅ 25 test cases for default configuration
- ✅ 4 test cases for custom configurations
- ✅ 3 test cases for empty configuration
- ✅ 3 edge cases (trailing slashes, subdirectories)

## FAQ

### Q: Will this affect existing crawls?

**A:** No. The filter only applies to new URL discoveries. Already-crawled pages remain in the database.

### Q: What if I want to crawl images for OCR?

**A:** Add image extensions to the allowed list:

```yaml
knowledge:
  allowed_file_extensions:
    - .html
    - .md
    - .jpg
    - .jpeg
    - .png
```

### Q: Can I use wildcards like `.*` or `*.html`?

**A:** No. Only exact extension matches are supported. Use `.html` not `*.html`.

### Q: What happens if `allowed_file_extensions` is not set?

**A:** The system uses the default list: `[.html, .htm, .md, .markdown, .txt, .pdf]`

### Q: How do I allow ALL file types?

**A:** This feature is designed to restrict, not allow everything. To crawl all file types:

1. Set a very permissive list in config
2. Or modify the code to disable filtering

**Not recommended** - this will crawl images, scripts, stylesheets, binaries, etc.

## Implementation Details

### Code Location

- **Filter function:** `internal/service/web_scraper.go:1380` (`shouldCrawlURLByExtension`)
- **Configuration:** `internal/config/config.go:189` (`AllowedFileExtensions`)
- **Tests:** `internal/service/url_extension_filter_test.go`

### Algorithm

```go
func shouldCrawlURLByExtension(urlStr string) bool {
    1. Parse URL
    2. If root/empty path → Allow
    3. Extract last path segment
    4. If no dot in segment → Allow (clean URL)
    5. Extract extension after last dot
    6. Check if extension in allowed list
       - Yes → Allow
       - No → Reject (log and skip)
}
```

### Memory Impact

- **Minimal:** Only stores the allowed extensions list (typically 3-10 strings)
- **Per-request overhead:** ~10-20 string comparisons maximum

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2024-11-09 | Initial implementation with configurable extension filtering |

## Related Documentation

- [Web Scraping Configuration Guide](./SCRAPING_CONFIG.md)
- [URL Normalization Guide](./URL_NORMALIZATION.md)
- [Parallel AI Widget Builder](./PARALLEL_AI_WIDGET_BUILDER.md)

## Support

For issues or questions about URL extension filtering:

1. Check the debug logs for rejected URLs
2. Verify your configuration is loaded correctly
3. Run the test suite to ensure expected behavior
4. Open an issue with the rejected URL patterns you're experiencing
