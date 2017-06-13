package ratelimiter_test

import (
	"testing"
	"time"

	"github.com/RussellLuo/ratelimiter"
	"github.com/go-redis/redis"
)

func BenchmarkRateLimiter_Take(b *testing.B) {
	rl := ratelimiter.NewRateLimiter(
		&Redis{redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})},
		"ratelimiter:benchmark",
		&ratelimiter.Bucket{
			Interval: 1 * time.Second,
			Quantum:  2,
			Capacity: 10,
		},
	)
	for i := 0; i < b.N; i++ {
		rl.Take(1)
	}
}
