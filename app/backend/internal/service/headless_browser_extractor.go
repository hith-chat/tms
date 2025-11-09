package service

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"

	"github.com/bareuptime/tms/internal/logger"
)

// SharedBrowserContext holds a reusable Playwright browser instance
type SharedBrowserContext struct {
	PW      *playwright.Playwright
	Browser playwright.Browser
	Context playwright.BrowserContext
}

// Close cleans up the browser context
func (s *SharedBrowserContext) Close() error {
	if s.Context != nil {
		if err := s.Context.Close(); err != nil {
			return err
		}
	}
	if s.Browser != nil {
		if err := s.Browser.Close(); err != nil {
			return err
		}
	}
	if s.PW != nil {
		if err := s.PW.Stop(); err != nil {
			return err
		}
	}
	return nil
}

// HeadlessBrowserURLExtractor uses a headless browser for comprehensive URL extraction
type HeadlessBrowserURLExtractor struct {
	timeout   time.Duration
	userAgent string
}

// NewHeadlessBrowserURLExtractor creates a new headless browser URL extractor
func NewHeadlessBrowserURLExtractor(timeout time.Duration, userAgent string) *HeadlessBrowserURLExtractor {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	if userAgent == "" {
		userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	}

	return &HeadlessBrowserURLExtractor{
		timeout:   timeout,
		userAgent: userAgent,
	}
}

// CreateSharedBrowserContext creates a reusable Playwright browser instance
func (e *HeadlessBrowserURLExtractor) CreateSharedBrowserContext(ctx context.Context) (*SharedBrowserContext, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to start playwright: %w", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
		Args: []string{
			"--disable-gpu",
			"--disable-extensions",
			"--disable-plugins",
			"--disable-background-networking",
			"--disable-sync",
			"--disable-translate",
			"--disable-default-apps",
			"--no-first-run",
			"--no-default-browser-check",
			"--disable-blink-features=AutomationControlled",
		},
	})
	if err != nil {
		pw.Stop()
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	browserContext, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: &e.userAgent,
	})
	if err != nil {
		browser.Close()
		pw.Stop()
		return nil, fmt.Errorf("failed to create browser context: %w", err)
	}

	return &SharedBrowserContext{
		PW:      pw,
		Browser: browser,
		Context: browserContext,
	}, nil
}

// ExtractedURLInfo contains information about an extracted URL
type ExtractedURLInfo struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Source      string `json:"source"` // where it was found: "link", "image", "script", "meta", etc.
	Text        string `json:"text"`   // link text or alt text
}

// ExtractURLsFromPage extracts URLs from a web page using Playwright
func (e *HeadlessBrowserURLExtractor) ExtractURLsFromPage(ctx context.Context, targetURL string) ([]ExtractedURLInfo, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	log.Printf("DEBUG: Starting Playwright browser for URL: %s", targetURL)

	// Start Playwright
	pw, err := playwright.Run()
	if err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to start Playwright")
		return nil, fmt.Errorf("failed to start playwright: %w", err)
	}
	defer pw.Stop()

	// Launch browser
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
		Args: []string{
			"--disable-gpu",
			"--disable-extensions",
			"--disable-plugins",
			"--disable-background-networking",
			"--disable-sync",
			"--disable-translate",
			"--disable-default-apps",
			"--no-first-run",
			"--no-default-browser-check",
			"--disable-blink-features=AutomationControlled",
		},
	})
	if err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to launch Chromium browser")
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}
	defer browser.Close()

	// Create context and page
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: &e.userAgent,
	})
	if err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to create browser context")
		return nil, fmt.Errorf("failed to create browser context: %w", err)
	}
	defer context.Close()

	page, err := context.NewPage()
	if err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to create page")
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	// Navigate to the page
	log.Printf("DEBUG: Navigating to %s", targetURL)
	response, err := page.Goto(targetURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(float64(e.timeout.Milliseconds())),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to navigate to page: %w", err)
	}

	if response.Status() != 200 {
		return nil, fmt.Errorf("bad response status: %d", response.Status())
	}

	// Wait for the page to be fully loaded
	err = page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	if err != nil {
		log.Printf("DEBUG: Warning - failed to wait for network idle: %v", err)
	}

	// Get page title for debugging
	title, err := page.Title()
	if err != nil {
		logger.GetTxLogger(ctx).Warn().Err(err).Msg("Failed to get page title")
	} else {
		logger.GetTxLogger(ctx).Info().Str("title", title).Msg("Page loaded successfully")
	}

	// Execute JavaScript to extract URLs
	result, err := page.Evaluate(`
		(function() {
			console.log('DEBUG: Starting URL extraction for:', window.location.href);
			
			const results = [];
			
			// Extract from all types of elements
			const selectors = [
				{ selector: 'a[href]', attr: 'href', source: 'link' },
				{ selector: 'link[href]', attr: 'href', source: 'link_tag' },
				{ selector: 'iframe[src]', attr: 'src', source: 'image' },
				{ selector: 'img[src]', attr: 'src', source: 'image' },
				{ selector: 'script[src]', attr: 'src', source: 'script' },
				{ selector: 'source[src]', attr: 'src', source: 'media_source' },
				{ selector: 'track[src]', attr: 'src', source: 'track' },
				{ selector: 'embed[src]', attr: 'src', source: 'embed' },
				{ selector: 'object[data]', attr: 'data', source: 'object' },
			];
			
			selectors.forEach(({ selector, attr, source }) => {
				const elements = document.querySelectorAll(selector);
				console.log('DEBUG: Found', elements.length, selector, 'elements');
				
				elements.forEach((el, index) => {
					const url = el.getAttribute(attr);
					const text = (el.textContent || el.innerText || '').trim();
					const title = el.title || el.alt || '';
					
					if (url) {
						try {
							// Resolve relative URLs
							const fullURL = new URL(url, window.location.href).href;
							
							// Only include HTTP/HTTPS URLs
							if (fullURL.startsWith('http://') || fullURL.startsWith('https://')) {
								// Remove fragments
								const cleanUrl = fullURL.split('#')[0];
								
								// Check if we already have this URL
								const exists = results.some(r => r.url === cleanUrl);
								if (!exists) {
									results.push({
										url: cleanUrl,
										source: source,
										text: text.slice(0, 200), // Limit text length
										description: title.slice(0, 200)
									});
									
									if (index < 3) {
										console.log('DEBUG: Added', source, ':', cleanUrl);
									}
								}
							}
						} catch (e) {
							// Invalid URL, skip
						}
					}
				});
			});
			
			// Also extract from meta tags
			const metaSelectors = [
				'meta[property="og:url"]',
				'meta[property="og:image"]', 
				'meta[name="twitter:image"]',
				'link[rel="canonical"]',
				'link[rel="alternate"]'
			];
			
			metaSelectors.forEach(selector => {
				const elements = document.querySelectorAll(selector);
				elements.forEach(el => {
					const url = el.content || el.href;
					if (url) {
						try {
							const fullURL = new URL(url, window.location.href).href;
							if (fullURL.startsWith('http://') || fullURL.startsWith('https://')) {
								const cleanUrl = fullURL.split('#')[0];
								const exists = results.some(r => r.url === cleanUrl);
								if (!exists) {
									results.push({
										url: cleanUrl,
										source: 'meta',
										text: '',
										description: el.getAttribute('property') || el.getAttribute('name') || el.getAttribute('rel') || ''
									});
								}
							}
						} catch (e) {
							// Invalid URL, skip
						}
					}
				});
			});
			
			console.log('DEBUG: Found total', results.length, 'unique URLs');
			return results;
		})()
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to execute JavaScript: %w", err)
	}

	// Convert result to our struct
	var extractedURLs []ExtractedURLInfo

	// The result should be an array of objects
	if resultArray, ok := result.([]interface{}); ok {
		for _, item := range resultArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				urlInfo := ExtractedURLInfo{}

				if url, ok := itemMap["url"].(string); ok {
					urlInfo.URL = url
				}
				if source, ok := itemMap["source"].(string); ok {
					urlInfo.Source = source
				}
				if text, ok := itemMap["text"].(string); ok {
					urlInfo.Text = text
				}
				if desc, ok := itemMap["description"].(string); ok {
					urlInfo.Description = desc
				}

				extractedURLs = append(extractedURLs, urlInfo)
			}
		}
	}

	logger.GetTxLogger(ctx).Debug().Int("count", len(extractedURLs)).Msg("Extracted URLs before cleaning")

	// Post-process and clean the URLs
	var cleanedURLs []ExtractedURLInfo
	seen := make(map[string]bool)

	for _, urlInfo := range extractedURLs {
		// Clean and validate URL
		if cleanedURL := e.cleanAndValidateURL(urlInfo.URL); cleanedURL != "" {
			if !seen[cleanedURL] {
				seen[cleanedURL] = true
				urlInfo.URL = cleanedURL
				cleanedURLs = append(cleanedURLs, urlInfo)
			}
		}
	}

	logger.GetTxLogger(ctx).Debug().Int("count", len(cleanedURLs)).Msg("Cleaned URLs after processing")

	// Log first few cleaned URLs for debugging
	fmt.Println("cleaned urls ->", cleanedURLs)

	log.Printf("Returning %d cleaned URLs", len(cleanedURLs))
	return cleanedURLs, nil
}

// cleanAndValidateURL cleans and validates a URL
func (e *HeadlessBrowserURLExtractor) cleanAndValidateURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	// Parse URL
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	// Must have scheme and host
	if parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}

	// Only HTTP/HTTPS
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}

	// Remove fragment
	parsed.Fragment = ""

	// Skip common non-content files
	path := strings.ToLower(parsed.Path)
	skipExtensions := []string{
		".css", ".js", ".ico", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp",
		".woff", ".woff2", ".ttf", ".eot", ".mp3", ".mp4", ".avi", ".mov", ".wmv",
		".pdf", ".zip", ".rar", ".tar", ".gz", ".exe", ".dmg", ".deb", ".rpm",
	}

	for _, ext := range skipExtensions {
		if strings.HasSuffix(path, ext) {
			return ""
		}
	}

	// Skip common tracking and ad URLs
	skipPatterns := []string{
		"google-analytics.com",
		"googletagmanager.com",
		"facebook.com/tr",
		"doubleclick.net",
		"adsystem.com",
		"googlesyndication.com",
		"amazon-adsystem.com",
		"site.webmanifest",
	}

	host := strings.ToLower(parsed.Host)
	for _, pattern := range skipPatterns {
		if strings.Contains(host, pattern) {
			return ""
		}
	}

	return parsed.String()
}

// GetPageTitle extracts the page title using Playwright
func (e *HeadlessBrowserURLExtractor) GetPageTitle(ctx context.Context, targetURL string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Start Playwright
	pw, err := playwright.Run()
	if err != nil {
		return "", fmt.Errorf("failed to start playwright: %w", err)
	}
	defer pw.Stop()

	// Launch browser
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("failed to launch browser: %w", err)
	}
	defer browser.Close()

	// Create context and page
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: &e.userAgent,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create browser context: %w", err)
	}
	defer context.Close()

	page, err := context.NewPage()
	if err != nil {
		return "", fmt.Errorf("failed to create page: %w", err)
	}

	// Navigate to the page
	_, err = page.Goto(targetURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(float64(e.timeout.Milliseconds())),
	})
	if err != nil {
		return "", fmt.Errorf("failed to navigate to page: %w", err)
	}

	// Get title
	title, err := page.Title()
	if err != nil {
		return "", fmt.Errorf("failed to get page title: %w", err)
	}

	return strings.TrimSpace(title), nil
}

// GetPageContent extracts text content from the page using Playwright
func (e *HeadlessBrowserURLExtractor) GetPageContent(ctx context.Context, targetURL string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Start Playwright
	pw, err := playwright.Run()
	if err != nil {
		return "", fmt.Errorf("failed to start playwright: %w", err)
	}
	defer pw.Stop()

	// Launch browser
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("failed to launch browser: %w", err)
	}
	defer browser.Close()

	// Create context and page
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: &e.userAgent,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create browser context: %w", err)
	}
	defer context.Close()

	page, err := context.NewPage()
	if err != nil {
		return "", fmt.Errorf("failed to create page: %w", err)
	}

	// Navigate to the page
	_, err = page.Goto(targetURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(float64(e.timeout.Milliseconds())),
	})
	if err != nil {
		return "", fmt.Errorf("failed to navigate to page: %w", err)
	}

	// Wait for dynamic content
	page.WaitForTimeout(2000) // 2 seconds

	// Extract text content
	content, err := page.Evaluate(`document.body.textContent || document.body.innerText || ''`)
	if err != nil {
		return "", fmt.Errorf("failed to extract page content: %w", err)
	}

	contentStr, ok := content.(string)
	if !ok {
		return "", fmt.Errorf("failed to convert content to string")
	}

	return strings.TrimSpace(contentStr), nil
}

// ExtractThemeDataWithBrowser extracts theme information using an existing browser context
func (e *HeadlessBrowserURLExtractor) ExtractThemeDataWithBrowser(ctx context.Context, targetURL string, sharedCtx *SharedBrowserContext) (map[string]interface{}, error) {
	if sharedCtx == nil {
		// Fall back to creating a new instance
		return e.ExtractThemeData(ctx, targetURL)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	page, err := sharedCtx.Context.NewPage()
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()

	// Navigate to the page
	_, err = page.Goto(targetURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(float64(e.timeout.Milliseconds())),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to navigate to page: %w", err)
	}

	// Wait for dynamic content
	page.WaitForTimeout(2000)

	return e.extractThemeFromPage(page)
}

// ExtractThemeData extracts comprehensive theme information from a website using Playwright
func (e *HeadlessBrowserURLExtractor) ExtractThemeData(ctx context.Context, targetURL string) (map[string]interface{}, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Start Playwright
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to start playwright: %w", err)
	}
	defer pw.Stop()

	// Launch browser
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
		Args: []string{
			"--disable-gpu",
			"--disable-extensions",
			"--disable-plugins",
			"--disable-background-networking",
			"--disable-sync",
			"--disable-translate",
			"--disable-default-apps",
			"--no-first-run",
			"--no-default-browser-check",
			"--disable-blink-features=AutomationControlled",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}
	defer browser.Close()

	// Create context and page
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: &e.userAgent,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create browser context: %w", err)
	}
	defer context.Close()

	page, err := context.NewPage()
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	// Navigate to the page
	_, err = page.Goto(targetURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(float64(e.timeout.Milliseconds())),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to navigate to page: %w", err)
	}

	// Wait for dynamic content
	page.WaitForTimeout(2000) // 2 seconds for JS to render

	return e.extractThemeFromPage(page)
}

// extractThemeFromPage extracts theme data from an already loaded page
func (e *HeadlessBrowserURLExtractor) extractThemeFromPage(page playwright.Page) (map[string]interface{}, error) {
	// Execute comprehensive theme extraction JavaScript
	result, err := page.Evaluate(`
		(function() {
			const themeData = {
				colors: [],
				backgroundColors: [],
				fontFamilies: [],
				pageTitle: document.title || '',
				metaDescription: '',
				brandName: '',
				cssVariables: {},
				dominantColors: [],
				screenshot: null
			};

			// Extract meta description
			const metaDesc = document.querySelector('meta[name="description"]');
			if (metaDesc) {
				themeData.metaDescription = metaDesc.getAttribute('content') || '';
			}

			// Extract brand name from multiple sources
			const h1 = document.querySelector('h1');
			const ogSiteName = document.querySelector('meta[property="og:site_name"]');
			const logo = document.querySelector('.brand, .logo, .site-title, [class*="brand"], [class*="logo"]');

			themeData.brandName = (h1 && h1.textContent.trim()) ||
			                      (ogSiteName && ogSiteName.getAttribute('content')) ||
			                      (logo && logo.textContent.trim()) || '';

			// Extract CSS variables from :root
			const rootStyles = getComputedStyle(document.documentElement);
			const cssVars = {};
			for (let i = 0; i < rootStyles.length; i++) {
				const propName = rootStyles[i];
				if (propName.startsWith('--')) {
					cssVars[propName] = rootStyles.getPropertyValue(propName).trim();
				}
			}
			themeData.cssVariables = cssVars;

			// Function to get computed styles for important elements
			const importantSelectors = [
				'body',
				'header',
				'nav',
				'main',
				'footer',
				'.hero',
				'.button',
				'button',
				'a',
				'h1, h2, h3',
				'[class*="primary"]',
				'[class*="accent"]',
				'[class*="brand"]'
			];

			const colorSet = new Set();
			const bgColorSet = new Set();
			const fontSet = new Set();

			importantSelectors.forEach(selector => {
				const elements = document.querySelectorAll(selector);
				elements.forEach(el => {
					if (elements.length > 50 && Array.from(elements).indexOf(el) > 10) {
						return; // Limit to first 10 elements for large collections
					}

					const styles = getComputedStyle(el);

					// Extract colors
					const color = styles.color;
					if (color && color !== 'rgba(0, 0, 0, 0)') {
						colorSet.add(color);
					}

					// Extract background colors
					const bgColor = styles.backgroundColor;
					if (bgColor && bgColor !== 'rgba(0, 0, 0, 0)') {
						bgColorSet.add(bgColor);
					}

					// Extract font families
					const fontFamily = styles.fontFamily;
					if (fontFamily) {
						fontSet.add(fontFamily);
					}
				});
			});

			themeData.colors = Array.from(colorSet);
			themeData.backgroundColors = Array.from(bgColorSet);
			themeData.fontFamilies = Array.from(fontSet);

			return themeData;
		})()
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to extract theme data: %w", err)
	}

	// Convert result to map
	themeMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to parse theme data")
	}

	return themeMap, nil
}

// TakeFullPageScreenshot captures a full-page screenshot
func (e *HeadlessBrowserURLExtractor) TakeFullPageScreenshot(ctx context.Context, targetURL string) ([]byte, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Start Playwright
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to start playwright: %w", err)
	}
	defer pw.Stop()

	// Launch browser
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
		Args: []string{
			"--disable-gpu",
			"--disable-extensions",
			"--disable-plugins",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}
	defer browser.Close()

	// Create context and page
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: &e.userAgent,
		// Set a standard viewport for consistency
		Viewport: &playwright.Size{
			Width:  1920,
			Height: 1080,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create browser context: %w", err)
	}
	defer context.Close()

	page, err := context.NewPage()
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	// Navigate to the page
	_, err = page.Goto(targetURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(float64(e.timeout.Milliseconds())),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to navigate to page: %w", err)
	}

	// Wait for page to fully render
	page.WaitForTimeout(2000)

	// Take full page screenshot
	screenshot, err := page.Screenshot(playwright.PageScreenshotOptions{
		FullPage: playwright.Bool(true),
		Type:     playwright.ScreenshotTypePng,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to take screenshot: %w", err)
	}

	return screenshot, nil
}
