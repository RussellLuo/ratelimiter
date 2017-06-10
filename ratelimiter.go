package ratelimiter

import (
	"sync"
	"time"
)

const lua = `
local key = KEYS[1]
local interval = tonumber(ARGV[1])
local quantum = tonumber(ARGV[2])
local capacity = tonumber(ARGV[3])
local now = tonumber(ARGV[4])
local count = tonumber(ARGV[5])

local bucket = {tc=quantum, ts=now}
local value = redis.call("get", key)
if value then
  bucket = cjson.decode(value)
end

local cycles = math.floor((now - bucket.ts) / interval)
if cycles > 0 then
  bucket.tc = math.min(bucket.tc + cycles * quantum, capacity)
  bucket.ts = bucket.ts + cycles * interval
end

if bucket.tc >= count then
  bucket.tc = bucket.tc - count
  if redis.call("set", key, cjson.encode(bucket)) then
    return 1
  end
end

return 0
`

type Redis interface {
	Eval(script string, keys []string, args ...interface{}) (interface{}, error)
}

type Bucket struct {
	Interval time.Duration // the fixed time duration between each addition
	Quantum  int64         // the number of tokens will be added in the interval
	Capacity int64         // the depth of the bucket
}

type RateLimiter struct {
	redis Redis
	key   string

	mu     sync.RWMutex
	bucket *Bucket
}

// New returns a new rate limiter special for key in redis
// with the specified bucket configuration.
func New(redis Redis, key string, bucket *Bucket) *RateLimiter {
	return &RateLimiter{
		redis:  redis,
		key:    key,
		bucket: bucket,
	}
}

// SetBucket updates the bucket configuration.
func (rl *RateLimiter) SetBucket(bucket *Bucket) {
	rl.mu.Lock()
	rl.bucket = bucket
	rl.mu.Unlock()
}

// Take takes count tokens from the bucket stored at rl.key in Redis.
func (rl *RateLimiter) Take(count int64) (bool, error) {
	rl.mu.RLock()
	bucket := *rl.bucket
	rl.mu.RUnlock()

	now := time.Now().Unix()
	status, err := rl.redis.Eval(
		lua,
		[]string{rl.key},
		int64(bucket.Interval/time.Second),
		bucket.Quantum,
		bucket.Capacity,
		now,
		count,
	)
	if err != nil {
		return false, err
	} else {
		return status == int64(1), nil
	}
}
