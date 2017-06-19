package ratelimiter_test

import (
	"testing"
	"time"

	"github.com/RussellLuo/ratelimiter"
	"github.com/go-redis/redis"
)

func BenchmarkGCRA_Transmit(b *testing.B) {
	gcra := ratelimiter.NewGCRA(
		&Redis{redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})},
		"ratelimiter:gcra:benchmark",
		&ratelimiter.Config{
			Interval: 1 * time.Second / 2,
			Capacity: 5,
		},
	)
	for i := 0; i < b.N; i++ {
		gcra.Transmit(1)
	}
}
