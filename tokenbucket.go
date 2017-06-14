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

// TokenBucket implements the Token Bucket Algorithm.
// See https://en.wikipedia.org/wiki/Token_bucket.
type TokenBucket struct {
	baseBucket

	script *Script
	key    string
}

// New returns a new token-bucket rate limiter special for key in redis
// with the specified bucket configuration.
func NewTokenBucket(redis Redis, key string, config *Config) *TokenBucket {
	return &TokenBucket{
		baseBucket: baseBucket{config: config},
		script:     NewScript(redis, luaTokenBucket),
		key:        key,
	}
}

// Take takes count tokens from the bucket stored at tb.key in Redis.
func (tb *TokenBucket) Take(count int64) (bool, error) {
	config := tb.Config()

	now := time.Now().Unix()
	result, err := tb.script.Run(
		[]string{tb.key},
		int64(config.Interval/time.Second),
		config.Quantum,
		config.Capacity,
		now,
		count,
	)
	if err != nil {
		return false, err
	} else {
		return result == int64(1), nil
	}
}
