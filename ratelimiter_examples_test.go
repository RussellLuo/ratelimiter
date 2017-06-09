package ratelimiter_test

import (
	"fmt"
	"sync"
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

func ExampleTake() {
	limiter := ratelimiter.New(
		&Redis{redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})},
		"ratelimiter:test1",
		&ratelimiter.Bucket{
			Interval: 1 * time.Second,
			Quantum:  5,
			Capacity: 10,
		},
	)
	if ok, _ := limiter.Take(1); ok {
		fmt.Println("PASS")
	} else {
		fmt.Println("DROP")
	}
	// Output:
	// PASS
}

func ExampleTake_concurrency() {
	concurrency := 5
	limiter := ratelimiter.New(
		&Redis{redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})},
		"ratelimiter:test2",
		&ratelimiter.Bucket{
			Interval: 1 * time.Second,
			Quantum:  5,
			Capacity: 10,
		},
	)

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			if ok, _ := limiter.Take(1); ok {
				fmt.Println("PASS")
			} else {
				fmt.Println("DROP")
			}
			wg.Done()
		}()
	}
	wg.Wait()
	// Output:
	// PASS
	// PASS
	// PASS
	// PASS
	// PASS
}
