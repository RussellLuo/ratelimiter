package ratelimiter_test

import (
	"testing"
	"time"

	"github.com/RussellLuo/ratelimiter"
	"github.com/go-redis/redis"
)

func TestGCRA_Transmit(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	key := "ratelimiter:gcra:test"

	gcra := ratelimiter.NewGCRA(
		&Redis{client},
		key,
		&ratelimiter.Config{
			Interval: 1 * time.Second / 2,
			Capacity: 5,
		},
	)

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
					Passed:  1,
					Dropped: 0,
					Delayed: 1,
					DelayDurations: []time.Duration{
						time.Duration(500 * time.Millisecond),
					},
				},
				{
					Passed:  1,
					Dropped: 0,
					Delayed: 1,
					DelayDurations: []time.Duration{
						time.Duration(500 * time.Millisecond),
					},
				},
				{
					Passed:  1,
					Dropped: 0,
					Delayed: 1,
					DelayDurations: []time.Duration{
						time.Duration(500 * time.Millisecond),
					},
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
					Passed:  1,
					Dropped: 0,
					Delayed: 3,
					DelayDurations: []time.Duration{
						time.Duration(500 * time.Millisecond),
						time.Duration(1 * time.Second),
						time.Duration(1500 * time.Millisecond),
					},
				},
				{
					Passed:  0,
					Dropped: 1,
					Delayed: 3,
					DelayDurations: []time.Duration{
						time.Duration(1 * time.Second),
						time.Duration(1500 * time.Millisecond),
						time.Duration(2 * time.Second),
					},
				},
				{
					Passed:  0,
					Dropped: 2,
					Delayed: 2,
					DelayDurations: []time.Duration{
						time.Duration(1500 * time.Millisecond),
						time.Duration(2 * time.Second),
					},
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
					Delayed: 0,
				},
				{
					Passed:  0,
					Dropped: 1,
					Delayed: 1,
					DelayDurations: []time.Duration{
						time.Duration(2 * time.Second),
					},
				},
				{
					Passed:  0,
					Dropped: 2,
					Delayed: 3,
					DelayDurations: []time.Duration{
						time.Duration(1 * time.Second),
						time.Duration(1500 * time.Millisecond),
						time.Duration(2 * time.Second),
					},
				},
			},
		},
	}
	for _, c := range cases {
		client.Del(key)
		got := concurrentlyDo(gcra.Transmit, c.in)
		if !deepEqual(got, c.want, true) {
			t.Errorf("Got (%#v) != Want (%#v)", got, c.want)
		}
	}
}
