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

func ExampleTokenBucket_Take() {
	tb := ratelimiter.NewTokenBucket(
		&Redis{redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})},
		"ratelimiter:tokenbucket:example",
		&ratelimiter.Config{
			Interval: 1 * time.Second / 2,
			Capacity: 5,
		},
	)
	if ok, err := tb.Take(1); ok {
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

func ExampleLeakyBucket_Give() {
	lb := ratelimiter.NewLeakyBucket(
		&Redis{redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})},
		"ratelimiter:leakybucket:example",
		&ratelimiter.Config{
			Interval: 1 * time.Second / 2,
			Capacity: 5,
		},
	)
	if ok, delayed, err := lb.Give(1); ok {
		if delayed == 0 {
			fmt.Println("PASS")
		} else {
			fmt.Println("DELAY")
		}
	} else {
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println("DROP")
	}
	// Output:
	// PASS
}
