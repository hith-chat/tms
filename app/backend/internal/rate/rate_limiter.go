package rate

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/bareuptime/tms/internal/redis"
)

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	redisService *redis.Service
}

// LimitConfig represents rate limiting configuration
type LimitConfig struct {
	Requests int           // Number of requests allowed
	Window   time.Duration // Time window for the limit
}

// RateLimitResult represents the result of a rate limit check
type RateLimitResult struct {
	Allowed       bool          // Whether the request is allowed
	RequestsLeft  int           // Number of requests left in current window
	ResetTime     time.Time     // When the rate limit window resets
	RetryAfter    time.Duration // How long to wait before retrying (if blocked)
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(redisService *redis.Service) *RateLimiter {
	return &RateLimiter{
		redisService: redisService,
	}
}

// CheckRateLimit checks if a request should be allowed based on rate limiting
func (rl *RateLimiter) CheckRateLimit(ctx context.Context, key string, config LimitConfig) (*RateLimitResult, error) {
	redisKey := fmt.Sprintf("rate_limit:%s", key)
	windowStart := time.Now().Truncate(config.Window)
	windowEnd := windowStart.Add(config.Window)
	
	// Use Redis pipeline for atomic operations
	client := rl.redisService.GetClient()
	pipe := client.Pipeline()
	
	// Increment count
	incrCmd := pipe.Incr(ctx, redisKey)
	
	// Set expiry if this is the first request in the window
	pipe.ExpireAt(ctx, redisKey, windowEnd)
	
	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute rate limit check: %w", err)
	}
	
	// Get the incremented value
	currentCount := int(incrCmd.Val())
	
	// Check if limit exceeded
	allowed := currentCount <= config.Requests
	requestsLeft := config.Requests - currentCount
	if requestsLeft < 0 {
		requestsLeft = 0
	}
	
	// Calculate retry after time
	var retryAfter time.Duration
	if !allowed {
		retryAfter = windowEnd.Sub(time.Now())
		if retryAfter < 0 {
			retryAfter = 0
		}
	}
	
	return &RateLimitResult{
		Allowed:      allowed,
		RequestsLeft: requestsLeft,
		ResetTime:    windowEnd,
		RetryAfter:   retryAfter,
	}, nil
}

// GetRateLimitStatus gets the current rate limit status without incrementing
func (rl *RateLimiter) GetRateLimitStatus(ctx context.Context, key string, config LimitConfig) (*RateLimitResult, error) {
	redisKey := fmt.Sprintf("rate_limit:%s", key)
	windowStart := time.Now().Truncate(config.Window)
	windowEnd := windowStart.Add(config.Window)
	
	// Get current count
	client := rl.redisService.GetClient()
	currentCountStr, err := client.Get(ctx, redisKey).Result()
	if err != nil {
		// If key doesn't exist, treat as 0
		if err.Error() == "redis: nil" {
			return &RateLimitResult{
				Allowed:      true,
				RequestsLeft: config.Requests,
				ResetTime:    windowEnd,
				RetryAfter:   0,
			}, nil
		}
		return nil, fmt.Errorf("failed to get rate limit status: %w", err)
	}
	
	currentCount, err := strconv.Atoi(currentCountStr)
	if err != nil {
		return nil, fmt.Errorf("invalid count value in Redis: %w", err)
	}
	
	// Check if limit would be exceeded
	allowed := currentCount < config.Requests
	requestsLeft := config.Requests - currentCount
	if requestsLeft < 0 {
		requestsLeft = 0
	}
	
	// Calculate retry after time
	var retryAfter time.Duration
	if !allowed {
		retryAfter = windowEnd.Sub(time.Now())
		if retryAfter < 0 {
			retryAfter = 0
		}
	}
	
	return &RateLimitResult{
		Allowed:      allowed,
		RequestsLeft: requestsLeft,
		ResetTime:    windowEnd,
		RetryAfter:   retryAfter,
	}, nil
}

// ResetRateLimit resets the rate limit for a specific key
func (rl *RateLimiter) ResetRateLimit(ctx context.Context, key string) error {
	redisKey := fmt.Sprintf("rate_limit:%s", key)
	client := rl.redisService.GetClient()
	err := client.Del(ctx, redisKey).Err()
	if err != nil {
		return fmt.Errorf("failed to reset rate limit: %w", err)
	}
	return nil
}

// DefaultLimits defines common rate limiting configurations
var DefaultLimits = map[string]LimitConfig{
	"public_api":      {Requests: 100, Window: time.Minute},     // 100 requests per minute for public APIs
	"auth":            {Requests: 10, Window: time.Minute},      // 10 auth attempts per minute
	"password_reset":  {Requests: 3, Window: time.Hour},        // 3 password resets per hour
	"registration":    {Requests: 5, Window: time.Hour},        // 5 registrations per hour per IP
	"payment":         {Requests: 20, Window: time.Minute},     // 20 payment requests per minute
	"file_upload":     {Requests: 50, Window: time.Minute},     // 50 file uploads per minute
	"api_strict":      {Requests: 30, Window: time.Minute},     // 30 requests per minute for strict endpoints
}
