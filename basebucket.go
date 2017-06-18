package ratelimiter

import (
	"sync"
	"time"
)

// Config is the bucket configuration.
// Both the leaky and token bucket algorithms share the same bucket configuration.
type Config struct {
	// Token bucket:
	//     the interval between each addition of one token
	// Leaky bucket:
	//     the interval between each leak of one unit of water
	Interval time.Duration

	// the capacity of the bucket
	Capacity int64
}

// baseBucket is a basic structure both for TokenBucket and LeakyBucket.
type baseBucket struct {
	mu     sync.RWMutex
	config *Config
}

// Config returns the bucket configuration in a concurrency-safe way.
func (b *baseBucket) Config() Config {
	b.mu.RLock()
	config := *b.config
	b.mu.RUnlock()
	return config
}

// SetConfig updates the bucket configuration in a concurrency-safe way.
func (b *baseBucket) SetConfig(config *Config) {
	b.mu.Lock()
	b.config = config
	b.mu.Unlock()
}
