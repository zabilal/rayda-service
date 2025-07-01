package cache

import (
	"context"
	"time"
)

// noOpCache is a no-operation cache implementation that doesn't store anything.
// It's useful for testing or when caching needs to be disabled.
type noOpCache struct{}

// NewNoOpCache creates a new no-op cache instance.
func NewNoOpCache() Cache {
	return &noOpCache{}
}

// Get always returns false and no error.
func (n *noOpCache) Get(_ context.Context, _ string, _ interface{}) (bool, error) {
	return false, nil
}

// Set is a no-op and always returns nil.
func (n *noOpCache) Set(_ context.Context, _ string, _ interface{}, _ time.Duration) error {
	return nil
}

// Delete is a no-op and always returns nil.
func (n *noOpCache) Delete(_ context.Context, _ string) error {
	return nil
}

// Exists always returns false and no error.
func (n *noOpCache) Exists(_ context.Context, _ string) (bool, error) {
	return false, nil
}
