package ratelimiter

import (
	"time"
)

// the Lua script that implements the Token Bucket Algorithm.
// bucket.tc represents the token count.
// bucket.ts represents the timestamp of the last time the bucket was refilled.
const luaTokenBucket = `
local key = KEYS[1]
local interval = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local amount = tonumber(ARGV[4])

local bucket = {tc=capacity, ts=now}
local value = redis.call("get", key)
if value then
  bucket = cjson.decode(value)
end

local added = math.floor((now - bucket.ts) / interval)
if added > 0 then
  bucket.tc = math.min(bucket.tc + added, capacity)
  bucket.ts = bucket.ts + added * interval
end

if bucket.tc >= amount then
  bucket.tc = bucket.tc - amount
  bucket.ts = string.format("%.f", bucket.ts)
  if redis.call("set", key, cjson.encode(bucket)) then
    return 1
  end
end

return 0
`

// TokenBucket implements the Token Bucket Algorithm.
// See https://en.wikipedia.org/wiki/Token_bucket.
type TokenBucket struct {
	baseBucket

	script *Script
	key    string
}

// NewTokenBucket returns a new token-bucket rate limiter special for key in redis
// with the specified bucket configuration.
func NewTokenBucket(redis Redis, key string, config *Config) *TokenBucket {
	return &TokenBucket{
		baseBucket: baseBucket{config: config},
		script:     NewScript(redis, luaTokenBucket),
		key:        key,
	}
}

// Take takes amount tokens from the bucket.
func (b *TokenBucket) Take(amount int64) (bool, error) {
	config := b.Config()
	if amount > config.Capacity {
		return false, nil
	}

	now := time.Now().UnixNano()
	result, err := b.script.Run(
		[]string{b.key},
		int64(config.Interval/time.Microsecond),
		config.Capacity,
		int64(time.Duration(now)/time.Microsecond),
		amount,
	)
	if err != nil {
		return false, err
	} else {
		return result == int64(1), nil
	}
}
