package ratelimiter

import (
	"time"
)

// the Lua script that implements the Leaky Bucket Algorithm as a meter.
// bucket.wl represents the water level,
// bucket.ts represents the timestamp of the last time the bucket was refilled.
const luaLeakyBucket = `
local key = KEYS[1]
local interval = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local amount = tonumber(ARGV[4])

local bucket = {wl=0, ts=now}
local value = redis.call("get", key)
if value then
  bucket = cjson.decode(value)
end

local leaks = math.floor((now - bucket.ts) / interval)
if leaks > 0 then
  bucket.wl = math.max(bucket.wl - leaks, 0)
  bucket.ts = bucket.ts + leaks * interval
end

if bucket.wl + amount <= capacity then
  local delayed = bucket.wl * interval
  bucket.wl = bucket.wl + amount
  bucket.ts = string.format("%.f", bucket.ts)
  if redis.call("set", key, cjson.encode(bucket)) then
    return delayed
  end
end

return -1
`

// LeakyBucket implements the Leaky Bucket Algorithm as a meter.
// See https://en.wikipedia.org/wiki/Leaky_bucket#The_Leaky_Bucket_Algorithm_as_a_Meter.
type LeakyBucket struct {
	baseBucket

	script *Script
	key    string
}

// NewLeakyBucket returns a new leaky-bucket rate limiter special for key in redis
// with the specified bucket configuration.
func NewLeakyBucket(redis Redis, key string, config *Config) *LeakyBucket {
	return &LeakyBucket{
		baseBucket: baseBucket{config: config},
		script:     NewScript(redis, luaLeakyBucket),
		key:        key,
	}
}

// Give gives amount units of water into the bucket.
func (b *LeakyBucket) Give(amount int64) (bool, time.Duration, error) {
	config := b.Config()
	if amount > config.Capacity {
		return false, -1, nil
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
		return false, -1, err
	} else {
		switch delayed := result.(int64); delayed {
		case -1:
			return false, -1, nil
		default:
			return true, time.Duration(delayed) * time.Microsecond, nil
		}
	}
}
