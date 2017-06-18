package ratelimiter_test

import (
	"testing"
	"time"

	"github.com/RussellLuo/ratelimiter"
	"github.com/go-redis/redis"
)

func BenchmarkTokenBucket_Take(b *testing.B) {
	tb := ratelimiter.NewTokenBucket(
		&Redis{redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})},
		"ratelimiter:tokenbucket:benchmark",
		&ratelimiter.Config{
			Interval: 1 * time.Second / 2,
			Capacity: 5,
		},
	)
	for i := 0; i < b.N; i++ {
		tb.Take(1)
	}
}
