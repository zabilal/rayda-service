package config

// CacheConfig holds configuration for the cache
type CacheConfig struct {
    // TTL is the default time-to-live for cache entries in seconds
    TTL    int    `mapstructure:"CACHE_TTL"`
    // Prefix is the prefix used for all cache keys
    Prefix string `mapstructure:"CACHE_KEY_PREFIX"`
    // Enabled enables or disables the cache
    Enabled bool   `mapstructure:"CACHE_ENABLED"`
}

// DefaultCacheConfig returns the default cache configuration
func DefaultCacheConfig() CacheConfig {
    return CacheConfig{
        TTL:     300, // 5 minutes
        Prefix:  "rayda",
        Enabled: true,
    }
}
