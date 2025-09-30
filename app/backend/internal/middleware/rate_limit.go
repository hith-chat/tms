package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/bareuptime/tms/internal/rate"
	"github.com/gin-gonic/gin"
)

// RateLimitMiddleware creates a rate limiting middleware based on IP address
func RateLimitMiddleware(rateLimiter *rate.RateLimiter, limitType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract client IP address
		clientIP := getClientIP(c)
		if clientIP == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to determine client IP"})
			c.Abort()
			return
		}

		// Create rate limit key based on IP and limit type
		key := fmt.Sprintf("%s:%s", limitType, clientIP)

		// Get the rate limit configuration for this limit type
		config, exists := rate.DefaultLimits[limitType]
		if !exists {
			// Use default public API limit if specific type not found
			config = rate.DefaultLimits["public_api"]
		}

		// Check rate limit
		result, err := rateLimiter.CheckRateLimit(c.Request.Context(), key, config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limit check failed"})
			c.Abort()
			return
		}

		// Add rate limit headers to response
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.Requests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", result.RequestsLeft))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", result.ResetTime.Unix()))

		// If rate limit exceeded, return 429 Too Many Requests
		if !result.Allowed {
			c.Header("Retry-After", fmt.Sprintf("%.0f", result.RetryAfter.Seconds()))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": result.RetryAfter.Seconds(),
				"reset_time":  result.ResetTime.Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getClientIP extracts the real client IP address from various headers
func getClientIP(c *gin.Context) string {
	// Try to get IP from X-Forwarded-For header (most common with load balancers)
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xForwardedFor, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if isValidIP(ip) && !isPrivateIP(ip) {
				return ip
			}
		}
	}

	// Try X-Real-IP header (used by some proxies)
	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP != "" && isValidIP(xRealIP) && !isPrivateIP(xRealIP) {
		return xRealIP
	}

	// Try X-Forwarded-For again, but accept private IPs if no public IP found
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if isValidIP(ip) {
				return ip
			}
		}
	}

	// Try X-Real-IP again, accept private IP if no public IP found
	if xRealIP != "" && isValidIP(xRealIP) {
		return xRealIP
	}

	// Fall back to remote address
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	
	if isValidIP(ip) {
		return ip
	}

	return ""
}

// isValidIP checks if the given string is a valid IP address
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// isPrivateIP checks if the given IP is a private IP address
func isPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// Check for private IPv4 ranges
	private := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8", // localhost
	}

	for _, cidr := range private {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(parsedIP) {
			return true
		}
	}

	// Check for IPv6 private ranges
	if parsedIP.To4() == nil {
		// IPv6 addresses
		if strings.HasPrefix(ip, "::1") || // localhost
			strings.HasPrefix(ip, "fc00:") || // unique local address
			strings.HasPrefix(ip, "fd00:") || // unique local address
			strings.HasPrefix(ip, "fe80:") { // link-local
			return true
		}
	}

	return false
}

// PublicAPIRateLimit creates rate limiting middleware for public API endpoints
func PublicAPIRateLimit(rateLimiter *rate.RateLimiter) gin.HandlerFunc {
	return RateLimitMiddleware(rateLimiter, "public_api")
}

// AuthRateLimit creates rate limiting middleware for authentication endpoints
func AuthRateLimit(rateLimiter *rate.RateLimiter) gin.HandlerFunc {
	return RateLimitMiddleware(rateLimiter, "auth")
}

// PaymentRateLimit creates rate limiting middleware for payment endpoints
func PaymentRateLimit(rateLimiter *rate.RateLimiter) gin.HandlerFunc {
	return RateLimitMiddleware(rateLimiter, "payment")
}

// StrictRateLimit creates rate limiting middleware for endpoints that need stricter limits
func StrictRateLimit(rateLimiter *rate.RateLimiter) gin.HandlerFunc {
	return RateLimitMiddleware(rateLimiter, "api_strict")
}

// CustomRateLimit creates rate limiting middleware with custom configuration
func CustomRateLimit(rateLimiter *rate.RateLimiter, limitType string, requests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract client IP address
		clientIP := getClientIP(c)
		if clientIP == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to determine client IP"})
			c.Abort()
			return
		}

		// Create rate limit key based on IP and limit type
		key := fmt.Sprintf("%s:%s", limitType, clientIP)

		// Use custom configuration
		config := rate.LimitConfig{
			Requests: requests,
			Window:   window,
		}

		// Check rate limit
		result, err := rateLimiter.CheckRateLimit(c.Request.Context(), key, config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limit check failed"})
			c.Abort()
			return
		}

		// Add rate limit headers to response
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.Requests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", result.RequestsLeft))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", result.ResetTime.Unix()))

		// If rate limit exceeded, return 429 Too Many Requests
		if !result.Allowed {
			c.Header("Retry-After", fmt.Sprintf("%.0f", result.RetryAfter.Seconds()))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": result.RetryAfter.Seconds(),
				"reset_time":  result.ResetTime.Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// PublicWidgetBuilderRateLimit creates rate limiting middleware for public AI widget builder endpoint
// Limits to 2 requests per 6 hours per IP address
func PublicWidgetBuilderRateLimit(rateLimiter *rate.RateLimiter) gin.HandlerFunc {
	return RateLimitMiddleware(rateLimiter, "public_widget_builder")
}
