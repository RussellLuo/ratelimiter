package ratelimiter_test

import (
	"fmt"
	"time"

	"github.com/RussellLuo/ratelimiter"
	"github.com/go-redis/redis"
)

type Redis struct {
	client *redis.Client
}

func (r *Redis) Eval(script string, keys []string, args ...interface{}) (interface{}, error) {
	return r.client.Eval(script, keys, args...).Result()
}

func Example() {
	rl := ratelimiter.New(
		&Redis{redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})},
		"ratelimiter:example",
		&ratelimiter.Bucket{
			Interval: 1 * time.Second,
			Quantum:  2,
			Capacity: 10,
		},
	)
	if ok, _ := rl.Take(1); ok {
		fmt.Println("PASS")
	} else {
		fmt.Println("DROP")
	}
	// Output:
	// PASS
}
