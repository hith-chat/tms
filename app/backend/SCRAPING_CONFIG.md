# Web Scraping Configuration Guide

## Overview

The web scraping service now supports configurable worker counts and performance metrics tracking for optimal parallel processing.

## Configuration Options

### Worker Counts

Configure the number of parallel workers for different extraction methods:

```yaml
knowledge:
  # Playwright workers for depth 0-1 (JavaScript-enabled crawling)
  # Lower count due to browser memory overhead
  # Default: 3
  # Recommended range: 2-5
  playwright_worker_count: 3

  # Colly workers for depth >= 2 (lightweight HTTP crawling)
  # Higher count for faster static content extraction
  # Default: 15
  # Recommended range: 10-30
  colly_worker_count: 15

  # Enable detailed performance metrics tracking
  # Default: true
  enable_performance_metrics: true
```

### Environment Variables

You can also configure via environment variables:

```bash
# Worker counts
export KNOWLEDGE_PLAYWRIGHT_WORKER_COUNT=5
export KNOWLEDGE_COLLY_WORKER_COUNT=20

# Performance metrics
export KNOWLEDGE_ENABLE_PERFORMANCE_METRICS=true
```

## Performance Metrics

When `enable_performance_metrics` is `true`, the scraping service tracks and reports:

### Per-Depth Metrics
- `depth`: The depth level (0, 1, 2, ...)
- `urls_processed`: Number of URLs successfully processed
- `urls_failed`: Number of URLs that failed
- `urls_discovered`: Number of new URLs discovered for next level
- `worker_count`: Number of parallel workers used
- `method`: Extraction method ("playwright" or "colly")
- `duration`: Total time taken for this depth
- `avg_processing_time`: Average time per URL
- `total_tokens`: Total content tokens extracted

### Overall Crawl Metrics
- `total_urls_processed`: Total URLs extracted across all depths
- `total_urls_failed`: Total failed URLs
- `total_tokens`: Total content tokens
- `total_duration`: Complete crawl time
- `playwright_urls`: URLs processed with Playwright
- `colly_urls`: URLs processed with Colly
- `playwright_time`: Time spent in Playwright
- `colly_time`: Time spent in Colly
- `avg_urls_per_second`: Overall throughput

## Example Event with Metrics

```json
{
  "type": "completed",
  "message": "URL extraction completed in 45.2s. Found 156 URLs (Playwright: 21 URLs in 18.3s, Colly: 135 URLs in 26.9s, Avg: 3.45 URLs/sec)",
  "links_found": 156,
  "timestamp": "2024-01-15T10:30:00Z",
  "metrics": {
    "total_urls_processed": 156,
    "total_urls_failed": 4,
    "total_tokens": 245678,
    "total_duration": 45200000000,
    "playwright_urls": 21,
    "colly_urls": 135,
    "playwright_time": 18300000000,
    "colly_time": 26900000000,
    "avg_urls_per_second": 3.45,
    "depth_metrics": [
      {
        "depth": 0,
        "urls_processed": 1,
        "urls_failed": 0,
        "urls_discovered": 20,
        "worker_count": 3,
        "method": "playwright",
        "duration": 3200000000,
        "avg_processing_time": 3200000000,
        "total_tokens": 1234
      },
      {
        "depth": 1,
        "urls_processed": 20,
        "urls_failed": 1,
        "urls_discovered": 135,
        "worker_count": 3,
        "method": "playwright",
        "duration": 15100000000,
        "avg_processing_time": 755000000,
        "total_tokens": 45678
      },
      {
        "depth": 2,
        "urls_processed": 135,
        "urls_failed": 3,
        "urls_discovered": 0,
        "worker_count": 15,
        "method": "colly",
        "duration": 26900000000,
        "avg_processing_time": 199259259,
        "total_tokens": 198766
      }
    ]
  }
}
```

## Performance Tuning

### For Faster Crawls (More Resources)

```yaml
knowledge:
  playwright_worker_count: 5   # Increase if you have sufficient memory
  colly_worker_count: 30       # Increase for faster HTTP requests
```

**Pros:**
- Significantly faster crawling (2-3x speedup)
- Better for time-sensitive operations

**Cons:**
- Higher memory usage (~500MB per Playwright worker)
- Higher CPU usage
- May hit rate limits on target servers

### For Resource-Constrained Environments

```yaml
knowledge:
  playwright_worker_count: 2   # Reduce memory footprint
  colly_worker_count: 10       # Reduce CPU usage
```

**Pros:**
- Lower memory footprint
- Lower CPU usage
- More respectful to target servers

**Cons:**
- Slower crawling (2-3x slower)

## Monitoring Performance

### View Real-Time Metrics

The scraping service streams events with metrics during crawling. Monitor these events to:

1. **Track Progress**: See how many URLs processed per depth
2. **Identify Bottlenecks**: Check avg_processing_time per depth
3. **Optimize Workers**: Adjust worker counts based on throughput
4. **Estimate Completion**: Use avg_urls_per_second to predict finish time

### Example Monitoring

```bash
# Watch depth completion events
curl -N http://localhost:8080/api/scraping/extract-urls?url=https://example.com&depth=3 | \
  jq 'select(.type == "info" and .depth_metrics != null) | .depth_metrics'
```

## Best Practices

1. **Start Conservative**: Use default values (3/15) first
2. **Monitor Memory**: Increase Playwright workers only if you have 2GB+ available RAM
3. **Test Incrementally**: Increase Colly workers by 5 at a time
4. **Enable Metrics**: Always enable metrics to track performance
5. **Watch for Rate Limits**: If you see many failures, reduce worker counts
6. **Target Server Capacity**: Respect target server by not over-parallelizing

## Troubleshooting

### High Memory Usage
- **Reduce** `playwright_worker_count` to 2
- Keep `colly_worker_count` at default (15)

### Many Failed URLs
- **Reduce** both worker counts (target server may be rate limiting)
- Add delays between requests (configure `scrape_rate_limit`)

### Slow Crawling
- **Increase** `colly_worker_count` to 20-30
- Slightly increase `playwright_worker_count` if memory allows

### Inconsistent Performance
- Enable `enable_performance_metrics` to identify bottlenecks
- Check `depth_metrics` to see which depths are slow
