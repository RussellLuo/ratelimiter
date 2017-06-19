package ratelimiter_test

import (
	"sort"
	"sync"
	"time"
)

const delayedError = 6 * time.Millisecond

type Func func(int64) (bool, time.Duration, error)

type arg struct {
	WaitDuration time.Duration
	Concurrency  int
	Amount       int64
}

type rv struct {
	ok      bool
	delayed time.Duration
	err     error
}

type result struct {
	Passed         int
	Dropped        int
	Delayed        int
	DelayDurations []time.Duration
}

func concurrentlyDo(f Func, args []arg) []result {
	times := len(args)
	rvChans := make([]chan rv, times)

	var wg sync.WaitGroup
	for i, a := range args {
		rvChans[i] = make(chan rv, a.Concurrency)
		for j := 0; j < a.Concurrency; j++ {
			wg.Add(1)
			go func(i int, a arg) {
				time.Sleep(a.WaitDuration)
				ok, delayed, err := f(a.Amount)
				rvChans[i] <- rv{ok: ok, delayed: delayed, err: err}
				wg.Done()
			}(i, a)
		}
	}
	wg.Wait()

	for _, c := range rvChans {
		close(c)
	}

	result := make([]result, times)
	for i, c := range rvChans {
		for rv := range c {
			if rv.ok {
				if rv.delayed < delayedError {
					result[i].Passed++
				} else {
					result[i].Delayed++
					result[i].DelayDurations = append(result[i].DelayDurations, rv.delayed)
				}
			} else {
				result[i].Dropped++
			}
		}

		// sorted in ascending order
		d := result[i].DelayDurations
		sort.SliceStable(d, func(i, j int) bool {
			return d[i] < d[j]
		})
	}

	return result
}

func deepEqual(got []result, want []result, careDelayed bool) bool {
	for i, r := range got {
		if !(r.Passed == want[i].Passed && r.Dropped == want[i].Dropped &&
			r.Delayed == want[i].Delayed) {
			return false
		}
		if careDelayed {
			for j, d := range r.DelayDurations {
				wantD := want[i].DelayDurations[j]
				if !(d > wantD-delayedError && d < wantD+delayedError) {
					return false
				}
			}
		}
	}
	return true
}
