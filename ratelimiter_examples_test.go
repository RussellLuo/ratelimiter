package ratelimiter_test

import (
	"fmt"
	"strings"
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


func (r *Redis) EvalSha(sha1 string, keys []string, args ...interface{}) (interface{}, error, bool) {
	result, err := r.client.EvalSha(sha1, keys, args...).Result()
	noScript := err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT ")
	return result, err, noScript
}

func Example() {
	rl := ratelimiter.NewRateLimiter(
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
	if ok, err := rl.Take(1); ok {
		fmt.Println("PASS")
	} else {
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println("DROP")
	}
	// Output:
	// PASS
}
