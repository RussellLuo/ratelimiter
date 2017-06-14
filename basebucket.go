package ratelimiter

import (
	"sync"
	"time"
)

// Config is the bucket configuration.
// Both the leaky and token bucket algorithms share the same bucket configuration.
type Config struct {
	// Token bucket:
	//     the fixed time duration between each addition
	// Leaky bucket:
	//     the fixed time duration between each leak
	Interval time.Duration

	// Token bucket:
	//     the number of tokens added in the interval
	// Leaky bucket:
	//     the amount of water leaks in the interval
	Quantum int64

	// the capacity of the bucket
	Capacity int64
}

// baseBucket is a basic structure both for TokenBucket and LeakyBucket.
type baseBucket struct {
	mu     sync.RWMutex
	config *Config
}

// Config returns the bucket configuration in a concurrency-safe way.
func (lb *baseBucket) Config() Config {
	lb.mu.RLock()
	config := *lb.config
	lb.mu.RUnlock()
	return config
}

// SetConfig updates the bucket configuration in a concurrency-safe way.
func (lb *baseBucket) SetConfig(config *Config) {
	lb.mu.Lock()
	lb.config = config
	lb.mu.Unlock()
}
