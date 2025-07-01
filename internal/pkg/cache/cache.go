// Package cache provides caching functionality for the application.
// It includes implementations for in-memory and Redis-based caching.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// Cache defines the interface for cache operations
type Cache interface {
	// Get retrieves a value from the cache by key
	Get(ctx context.Context, key string, dest interface{}) (bool, error)
	// Set stores a value in the cache with the specified TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	// Delete removes a value from the cache
	Delete(ctx context.Context, key string) error
	// Exists checks if a key exists in the cache
	Exists(ctx context.Context, key string) (bool, error)
}

// RedisCache is a Redis-based implementation of the Cache interface
type RedisCache struct {
	client *redis.Client
	prefix string
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(client *redis.Client, prefix string) *RedisCache {
	return &RedisCache{
		client: client,
		prefix: prefix,
	}
}

// Get retrieves a value from Redis cache
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	fullKey := c.getFullKey(key)
	val, err := c.client.Get(ctx, fullKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, fmt.Errorf("failed to get from cache: %w", err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return false, fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	return true, nil
}

// Set stores a value in Redis cache
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := c.getFullKey(key)

	val, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.client.Set(ctx, fullKey, val, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache value: %w", err)
	}

	return nil
}

// Delete removes a value from Redis cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := c.getFullKey(key)
	return c.client.Del(ctx, fullKey).Err()
}

// Exists checks if a key exists in Redis cache
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := c.getFullKey(key)
	exists, err := c.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}

	return exists > 0, nil
}

// getFullKey returns the full cache key with prefix
func (c *RedisCache) getFullKey(key string) string {
	if c.prefix != "" {
		return fmt.Sprintf("%s:%s", c.prefix, key)
	}
	return key
}

// InMemoryCache is an in-memory implementation of the Cache interface for testing and development
type InMemoryCache struct {
	store  map[string][]byte
	ttls   map[string]time.Time
	prefix string
}

// NewInMemoryCache creates a new in-memory cache instance
func NewInMemoryCache(prefix string) *InMemoryCache {
	return &InMemoryCache{
		store:  make(map[string][]byte),
		ttls:   make(map[string]time.Time),
		prefix: prefix,
	}
}

// Get retrieves a value from in-memory cache
func (c *InMemoryCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	fullKey := c.getFullKey(key)

	// Check if key exists and is not expired
	if expiry, exists := c.ttls[fullKey]; exists && time.Now().Before(expiry) {
		val, exists := c.store[fullKey]
		if !exists {
			return false, nil
		}

		if err := json.Unmarshal(val, dest); err != nil {
			return false, fmt.Errorf("failed to unmarshal cached value: %w", err)
		}

		return true, nil
	}

	// Clean up expired key
	if _, exists := c.ttls[fullKey]; exists {
		delete(c.ttls, fullKey)
		delete(c.store, fullKey)
	}

	return false, nil
}

// Set stores a value in in-memory cache
func (c *InMemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := c.getFullKey(key)

	val, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	c.store[fullKey] = val
	if ttl > 0 {
		c.ttls[fullKey] = time.Now().Add(ttl)
	}

	return nil
}

// Delete removes a value from in-memory cache
func (c *InMemoryCache) Delete(ctx context.Context, key string) error {
	fullKey := c.getFullKey(key)
	delete(c.store, fullKey)
	delete(c.ttls, fullKey)
	return nil
}

// Exists checks if a key exists in in-memory cache
func (c *InMemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := c.getFullKey(key)

	// Check if key exists and is not expired
	if expiry, exists := c.ttls[fullKey]; exists && time.Now().Before(expiry) {
		_, exists := c.store[fullKey]
		return exists, nil
	}

	// Clean up expired key
	if _, exists := c.ttls[fullKey]; exists {
		delete(c.ttls, fullKey)
		delete(c.store, fullKey)
	}

	return false, nil
}

// getFullKey returns the full cache key with prefix
func (c *InMemoryCache) getFullKey(key string) string {
	if c.prefix != "" {
		return fmt.Sprintf("%s:%s", c.prefix, key)
	}
	return key
}
