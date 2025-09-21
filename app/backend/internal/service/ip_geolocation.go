// Package service provides IP geolocation functionality with Redis caching.
// This file integrates the IP geolocation service into the service layer.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/bareuptime/tms/internal/redis"
)

// IPDetails represents the response structure from IP geolocation APIs
// This structure contains comprehensive location and network information for an IP address
type IPDetails struct {
	IP            string  `json:"ip"`                   // The IP address being queried
	Version       string  `json:"version"`              // IP version (IPv4 or IPv6)
	City          string  `json:"city"`                 // City name
	Region        string  `json:"region"`               // State/Province/Region name
	RegionCode    string  `json:"region_code"`          // State/Province/Region code
	Country       string  `json:"country"`              // Country name
	CountryName   string  `json:"country_name"`         // Full country name
	CountryCode   string  `json:"country_code"`         // ISO country code (2-letter)
	CountryCode3  string  `json:"country_code_iso3"`    // ISO country code (3-letter)
	CountryTLD    string  `json:"country_tld"`          // Country top-level domain
	ContinentCode string  `json:"continent_code"`       // Continent code
	InEU          bool    `json:"in_eu"`                // Whether the country is in EU
	Postal        string  `json:"postal"`               // Postal/ZIP code
	Latitude      float64 `json:"latitude"`             // Geographic latitude
	Longitude     float64 `json:"longitude"`            // Geographic longitude
	Timezone      string  `json:"timezone"`             // IANA timezone identifier
	UTCOffset     string  `json:"utc_offset"`           // UTC offset (e.g., "+0200")
	CountryCode2  string  `json:"country_calling_code"` // International calling code
	Currency      string  `json:"currency"`             // Currency code (ISO 4217)
	CurrencyName  string  `json:"currency_name"`        // Full currency name
	Languages     string  `json:"languages"`            // Spoken languages (comma-separated)
	ASN           string  `json:"asn"`                  // Autonomous System Number
	Organization  string  `json:"org"`                  // ISP/Organization name
}

// IPGeolocationService provides IP geolocation functionality with Redis caching
// It follows the distributed cache pattern similar to Python's @distributed_cache decorator
type IPGeolocationService struct {
	redisService *redis.Service // Redis service for caching
	httpClient   *http.Client   // HTTP client for API calls
	cachePrefix  string         // Cache key prefix
	cacheTTL     time.Duration  // Cache time-to-live duration
}

// NewIPGeolocationService creates a new instance of IPGeolocationService
//
// Parameters:
//   - redisService: Redis service instance for caching operations
//   - cachePrefix: Prefix for cache keys (defaults to "ip_track" if empty)
//   - cacheTTL: Cache duration (defaults to 15 days if zero)
//
// Returns:
//   - *IPGeolocationService: Configured service instance ready for use
//
// Example usage:
//
//	service := NewIPGeolocationService(redisService, "ip_geo", 24*time.Hour)
//	details, err := service.GetIPDetails(context.Background(), "8.8.8.8")
func NewIPGeolocationService(redisService *redis.Service, cachePrefix string, cacheTTL time.Duration) *IPGeolocationService {
	// Set default values if not provided
	if cachePrefix == "" {
		cachePrefix = "ip_track"
	}
	if cacheTTL == 0 {
		cacheTTL = 15 * 24 * time.Hour // 15 days default
	}

	return &IPGeolocationService{
		redisService: redisService,
		httpClient: &http.Client{
			Timeout: 10 * time.Second, // 10 second timeout for API calls
		},
		cachePrefix: cachePrefix,
		cacheTTL:    cacheTTL,
	}
}

// GetIPDetails fetches IP geolocation details with automatic Redis caching
//
// This method implements distributed caching similar to Python's @distributed_cache decorator:
// 1. First checks Redis cache for existing data
// 2. If cache hit: returns cached data immediately
// 3. If cache miss: calls external API, caches result, then returns data
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - clientIP: IP address to look up (IPv4 or IPv6)
//
// Returns:
//   - *IPDetails: Comprehensive IP geolocation information
//   - error: Any error encountered during lookup or caching
//
// Cache behavior:
//   - Cache key format: "{prefix}:{ip_address}"
//   - TTL: Configurable (default 15 days)
//   - Automatic JSON serialization/deserialization
//   - Graceful fallback on cache failures
//
// Example usage:
//
//	details, err := service.GetIPDetails(ctx, "192.168.1.1")
//	if err != nil {
//	    log.Printf("Failed to get IP details: %v", err)
//	    return err
//	}
//	fmt.Printf("Location: %s, %s", details.City, details.Country)
func (s *IPGeolocationService) GetIPDetails(ctx context.Context, clientIP string) (*IPDetails, error) {
	// Check if IP is private or localhost - skip caching for these
	isPrivate := isPrivateOrLocalIP(clientIP)

	// Generate cache key with prefix
	cacheKey := fmt.Sprintf("%s:%s", s.cachePrefix, clientIP)

	// For private/localhost IPs, return default values instead of API call
	if isPrivate {
		log.Printf("Returning default values for private/local IP: %s", clientIP)
		return s.getDefaultIPDetails(clientIP), nil
	}

	// Step 1: Try to get data from Redis cache
	cachedData, err := s.redisService.GetClient().Get(ctx, cacheKey).Result()
	if err == nil && cachedData != "" {
		// Cache hit - deserialize and return cached data
		var details IPDetails
		if err := json.Unmarshal([]byte(cachedData), &details); err == nil {
			log.Printf("Cache hit for IP: %s", clientIP)
			return &details, nil
		}
		// If deserialization fails, continue to API call
		log.Printf("Failed to deserialize cached data for IP %s: %v", clientIP, err)
	}

	// Step 2: Cache miss - fetch from external API
	log.Printf("Cache miss for IP: %s, fetching from API", clientIP)
	details, err := s.fetchFromAPI(ctx, clientIP)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IP details from API: %w", err)
	}

	// Step 3: Cache the result for future requests
	if err := s.cacheIPDetails(ctx, cacheKey, details); err != nil {
		// Log caching error but don't fail the request
		log.Printf("Failed to cache IP details for %s: %v", clientIP, err)
	}

	return details, nil
}

// fetchFromAPI makes HTTP request to ipapi.co service to get IP geolocation data
//
// Parameters:
//   - ctx: Context for request cancellation and timeout
//   - clientIP: IP address to query
//
// Returns:
//   - *IPDetails: Parsed IP geolocation data
//   - error: Any error during HTTP request or JSON parsing
//
// API Details:
//   - Provider: ipapi.co (free tier available)
//   - Endpoint: https://ipapi.co/{ip}/json
//   - Timeout: 10 seconds
//   - Rate limits: Applies based on ipapi.co terms
func (s *IPGeolocationService) fetchFromAPI(ctx context.Context, clientIP string) (*IPDetails, error) {
	// Construct API URL
	url := fmt.Sprintf("https://ipapi.co/%s/json", clientIP)

	// Create HTTP request with context for timeout/cancellation
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set user agent for better API compatibility
	req.Header.Set("User-Agent", "BareUptime-Service/1.0")

	// Execute HTTP request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
	}

	// Parse JSON response
	var details IPDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &details, nil
}

// cacheIPDetails stores IP details in Redis with configured TTL
//
// Parameters:
//   - ctx: Context for Redis operation
//   - cacheKey: Redis key for storing the data
//   - details: IP details to cache
//
// Returns:
//   - error: Any error during JSON serialization or Redis storage
//
// Cache format: JSON-serialized IPDetails struct
func (s *IPGeolocationService) cacheIPDetails(ctx context.Context, cacheKey string, details *IPDetails) error {
	// Serialize IP details to JSON
	jsonData, err := json.Marshal(details)
	if err != nil {
		return fmt.Errorf("failed to serialize IP details: %w", err)
	}

	// Store in Redis with TTL
	if err := s.redisService.GetClient().Set(ctx, cacheKey, string(jsonData), s.cacheTTL).Err(); err != nil {
		return fmt.Errorf("failed to store in Redis: %w", err)
	}

	log.Printf("Cached IP details for key: %s (TTL: %v)", cacheKey, s.cacheTTL)
	return nil
}

// GetCurrencyCode is a convenience method to extract currency code from IP details
//
// Parameters:
//   - ctx: Context for the operation
//   - clientIP: IP address to lookup
//
// Returns:
//   - string: ISO 4217 currency code (e.g., "USD", "EUR")
//   - error: Any error during IP lookup
//
// This method provides the same functionality as your Python function's currency extraction
func (s *IPGeolocationService) GetCurrencyCode(ctx context.Context, clientIP string) (string, error) {
	details, err := s.GetIPDetails(ctx, clientIP)
	if err != nil {
		return "", err
	}
	return details.Currency, nil
}

// InvalidateCache removes cached data for a specific IP address
//
// Parameters:
//   - ctx: Context for Redis operation
//   - clientIP: IP address to remove from cache
//
// Returns:
//   - error: Any error during cache invalidation
//
// Use this method to force refresh of IP data or for cache management
func (s *IPGeolocationService) InvalidateCache(ctx context.Context, clientIP string) error {
	cacheKey := fmt.Sprintf("%s:%s", s.cachePrefix, clientIP)
	return s.redisService.GetClient().Del(ctx, cacheKey).Err()
}

// GetCacheStats returns basic cache statistics for monitoring
//
// Parameters:
//   - ctx: Context for Redis operations
//   - clientIP: IP address to check cache status
//
// Returns:
//   - bool: Whether the IP is cached
//   - time.Duration: TTL remaining (0 if not cached)
//   - error: Any error during cache check
func (s *IPGeolocationService) GetCacheStats(ctx context.Context, clientIP string) (bool, time.Duration, error) {
	cacheKey := fmt.Sprintf("%s:%s", s.cachePrefix, clientIP)

	exists, err := s.redisService.GetClient().Exists(ctx, cacheKey).Result()
	if err != nil {
		return false, 0, err
	}

	if exists == 0 {
		return false, 0, nil
	}

	ttl, err := s.redisService.GetClient().TTL(ctx, cacheKey).Result()
	return true, ttl, err
}

// isPrivateOrLocalIP checks if the given IP address is private, localhost, or loopback
//
// Parameters:
//   - ipStr: IP address string to check
//
// Returns:
//   - bool: true if IP is private/localhost, false otherwise
//
// Private IP ranges checked:
//   - 127.0.0.0/8 (localhost/loopback)
//   - 10.0.0.0/8 (private)
//   - 172.16.0.0/12 (private)
//   - 192.168.0.0/16 (private)
//   - ::1 (IPv6 localhost)
//   - fc00::/7 (IPv6 private)
func isPrivateOrLocalIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Check for localhost/loopback
	if ip.IsLoopback() {
		return true
	}

	// Check for private IP ranges
	if ip.IsPrivate() {
		return true
	}

	// Additional check for localhost IPv4 (127.0.0.1)
	if ip.Equal(net.IPv4(127, 0, 0, 1)) {
		return true
	}

	// Additional check for IPv6 localhost (::1)
	if ip.Equal(net.IPv6loopback) {
		return true
	}

	return false
}

// getDefaultIPDetails returns default IP details for private/localhost IPs
//
// Parameters:
//   - clientIP: The private/localhost IP address
//
// Returns:
//   - *IPDetails: Default IP details with minimal but sensible values
//
// This method provides default values for private IPs since external geolocation
// APIs cannot provide meaningful location data for private network addresses.
func (s *IPGeolocationService) getDefaultIPDetails(clientIP string) *IPDetails {
	// Determine IP version
	version := "IPv4"
	if ip := net.ParseIP(clientIP); ip != nil && ip.To4() == nil {
		version = "IPv6"
	}

	// Determine if it's localhost
	isLocalhost := clientIP == "127.0.0.1" || clientIP == "::1" || clientIP == "localhost"

	// Set default location based on IP type
	city := "Private Network"
	country := "Private Network"
	countryCode := "XX"

	if isLocalhost {
		city = "Local Machine"
		country = "Local Machine"
		countryCode = "LO"
	}

	return &IPDetails{
		IP:            clientIP,
		Version:       version,
		City:          city,
		Region:        "Private",
		RegionCode:    "XX",
		Country:       country,
		CountryName:   country,
		CountryCode:   countryCode,
		CountryCode3:  "XXX",
		CountryTLD:    ".local",
		ContinentCode: "XX",
		InEU:          false,
		Postal:        "00000",
		Latitude:      0.0,
		Longitude:     0.0,
		Timezone:      "UTC",
		UTCOffset:     "+0000",
		CountryCode2:  "+0",
		Currency:      "USD", // Default currency
		CurrencyName:  "US Dollar",
		Languages:     "en",
		ASN:           "AS0",
		Organization:  "Private Network",
	}
}
