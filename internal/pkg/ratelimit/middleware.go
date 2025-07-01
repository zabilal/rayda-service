package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// RateLimitConfig holds the configuration for rate limiting
type RateLimitConfig struct {
	// Rate is the number of requests allowed per duration
	Rate int
	// Duration is the time window for the rate limit
	Duration time.Duration
	// KeyPrefix is the prefix for the Redis key
	KeyPrefix string
}

// DefaultConfig returns the default rate limit configuration
func DefaultConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Rate:      100,             // 100 requests
		Duration:  1 * time.Minute, // per minute
		KeyPrefix: "rate_limit:",
	}
}

// WithRate sets the number of requests allowed per duration
func (c *RateLimitConfig) WithRate(rate int) *RateLimitConfig {
	c.Rate = rate
	return c
}

// WithDuration sets the time window for the rate limit
func (c *RateLimitConfig) WithDuration(duration time.Duration) *RateLimitConfig {
	c.Duration = duration
	return c
}

// Middleware creates a new rate limiting middleware
func Middleware(redisClient *redis.Client, config *RateLimitConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultConfig()
	}

	return func(c *gin.Context) {
		// Skip rate limiting for certain paths (e.g., health checks)
		if shouldSkipRateLimit(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Get the client IP or API key
		clientID := getClientIdentifier(c)
		if clientID == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "unable to identify client"})
			return
		}

		// Create a rate limit key
		key := fmt.Sprintf("%s%s:%s", config.KeyPrefix, clientID, time.Now().Truncate(config.Duration).Format(time.RFC3339))

		// Use Redis INCR to increment the counter
		ctx := context.Background()
		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			// If Redis is down, log the error but allow the request to proceed
			c.Next()
			return
		}

		// Set expiration on the key if this is the first request in the window
		if count == 1 {
			redisClient.Expire(ctx, key, config.Duration+time.Second)
		}

		// Set rate limit headers
		c.Writer.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.Rate))
		c.Writer.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(int64(config.Rate)-count, 10))
		c.Writer.Header().Set("X-RateLimit-Reset", time.Now().Add(config.Duration).Format(time.RFC1123))

		// Check if rate limit is exceeded
		if count > int64(config.Rate) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
				"retry_after": config.Duration.String(),
			})
			return
		}

		c.Next()
	}
}

// getClientIdentifier returns a unique identifier for the client
func getClientIdentifier(c *gin.Context) string {
	// Try to get API key from header first
	apiKey := c.GetHeader("X-API-Key")
	if apiKey != "" {
		return "api_key:" + apiKey
	}

	// Fall back to IP address
	ip := c.ClientIP()
	if ip != "" {
		return "ip:" + ip
	}

	// If we can't identify the client, return an empty string
	return ""
}

// shouldSkipRateLimit checks if the request path should be excluded from rate limiting
func shouldSkipRateLimit(path string) bool {
	skipPaths := []string{
		"/health",
		"/metrics",
		"/favicon.ico",
	}

	for _, p := range skipPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}
