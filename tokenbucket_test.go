package ratelimiter_test

import (
	"testing"
	"time"

	"github.com/RussellLuo/ratelimiter"
	"github.com/go-redis/redis"
)

func TestTokenBucket_Take(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	key := "ratelimiter:tokenbucket:test"

	bucket := ratelimiter.NewTokenBucket(
		&Redis{client},
		key,
		&ratelimiter.Config{
			Interval: 1 * time.Second / 2,
			Capacity: 5,
		},
	)

	f := func(amount int64) (bool, time.Duration, error) {
		ok, err := bucket.Take(amount)
		return ok, 0, err
	}

	cases := []struct {
		in   []arg
		want []result
	}{
		{
			in: []arg{
				{
					WaitDuration: 0 * time.Second,
					Concurrency:  2,
					Amount:       1,
				},
				{
					WaitDuration: 1 * time.Second,
					Concurrency:  2,
					Amount:       1,
				},
				{
					WaitDuration: 2 * time.Second,
					Concurrency:  2,
					Amount:       1,
				},
			},
			want: []result{
				{
					Passed:  2,
					Dropped: 0,
				},
				{
					Passed:  2,
					Dropped: 0,
				},
				{
					Passed:  2,
					Dropped: 0,
				},
			},
		},
		{
			in: []arg{
				{
					WaitDuration: 0 * time.Second,
					Concurrency:  4,
					Amount:       1,
				},
				{
					WaitDuration: 1 * time.Second,
					Concurrency:  4,
					Amount:       1,
				},
				{
					WaitDuration: 2 * time.Second,
					Concurrency:  4,
					Amount:       1,
				},
			},
			want: []result{
				{
					Passed:  4,
					Dropped: 0,
				},
				{
					Passed:  3,
					Dropped: 1,
				},
				{
					Passed:  2,
					Dropped: 2,
				},
			},
		},
		{
			in: []arg{
				{
					WaitDuration: 0 * time.Second,
					Concurrency:  1,
					Amount:       5,
				},
				{
					WaitDuration: 500 * time.Millisecond,
					Concurrency:  2,
					Amount:       1,
				},
				{
					WaitDuration: 2 * time.Second,
					Concurrency:  5,
					Amount:       1,
				},
			},
			want: []result{
				{
					Passed:  1,
					Dropped: 0,
				},
				{
					Passed:  1,
					Dropped: 1,
				},
				{
					Passed:  3,
					Dropped: 2,
				},
			},
		},
	}
	for _, c := range cases {
		client.Del(key)
		got := concurrentlyDo(f, c.in)
		if !deepEqual(got, c.want, false) {
			t.Errorf("Got (%#v) != Want (%#v)", got, c.want)
		}
	}
}
