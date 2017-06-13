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
local quantum = tonumber(ARGV[2])
local capacity = tonumber(ARGV[3])
local now = tonumber(ARGV[4])
local count = tonumber(ARGV[5])

local bucket = {wl=capacity - quantum, ts=now}
local value = redis.call("get", key)
if value then
  bucket = cjson.decode(value)
end

local cycles = math.floor((now - bucket.ts) / interval)
if cycles > 0 then
  bucket.wl = math.max(bucket.wl - cycles * quantum, 0)
  bucket.ts = bucket.ts + cycles * interval
end

if bucket.wl + count <= capacity then
  bucket.wl = bucket.wl + count
  if redis.call("set", key, cjson.encode(bucket)) then
    return 1
  end
end

return 0
`

// LeakyBucket implements the Leaky Bucket Algorithm as a meter.
// See https://en.wikipedia.org/wiki/Leaky_bucket#The_Leaky_Bucket_Algorithm_as_a_Meter.
type LeakyBucket struct {
	BaseBucket

	script *Script
	key    string
}

// New returns a new leaky-bucket rate limiter special for key in redis
// with the specified bucket configuration.
func NewLeakyBucket(redis Redis, key string, config *Config) *LeakyBucket {
	return &LeakyBucket{
		BaseBucket: BaseBucket{config: config},
		script:     NewScript(redis, luaLeakyBucket),
		key:        key,
	}
}

// Give gives count amount of water into the bucket stored at lb.key in Redis.
func (lb *LeakyBucket) Give(count int64) (bool, error) {
	config := lb.Config()

	now := time.Now().Unix()
	result, err := lb.script.Run(
		[]string{lb.key},
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
