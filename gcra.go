package ratelimiter

import (
	"time"
)

// the Lua script that implements the generic cell rate algorithm.
const luaGCRA = `
local key = KEYS[1]
local interval = tonumber(ARGV[1])
local tolerance = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local amount = tonumber(ARGV[4])

local tat = redis.call("get", key)
if tat then
  tat = tonumber(tat)
else
  tat = now
end

local new_tat = math.max(now, tat) + amount

if now >= new_tat - tolerance - interval then
  local ttl = math.ceil((new_tat - now)/ 1000000)
  if redis.call("setex", key, ttl, new_tat) then
    return math.max(tat - now, 0)
  end
end

return -1
`

// GCRA implements the generic cell rate algorithm.
// See https://en.wikipedia.org/wiki/Generic_cell_rate_algorithm.
type GCRA struct {
	baseBucket

	script *Script
	key    string
}

// NewGCRA returns a new GCRA rate limiter special for key in redis
// with the specified parameters.
func NewGCRA(redis Redis, key string, config *Config) *GCRA {
	return &GCRA{
		baseBucket: baseBucket{config: config},
		script:     NewScript(redis, luaGCRA),
		key:        key,
	}
}

// Transmit transmits a message to the bucket.
// Think of count 1 represents a message containing only one cell, and
// count greater than 1 represents a message containing multiple cells.
func (g *GCRA) Transmit(amount int64) (bool, time.Duration, error) {
	config := g.Config()
	if amount > config.Capacity {
		return false, -1, nil
	}

	// the interval between each arrival of one cell
	emissionInterval := config.Interval
	// how much earlier a cell can arrive than it would
	delayVariationTolerance := time.Duration(config.Capacity-1) * config.Interval

	now := time.Now().UnixNano()
	result, err := g.script.Run(
		[]string{g.key},
		int64(emissionInterval/time.Microsecond),
		int64(delayVariationTolerance/time.Microsecond),
		int64(time.Duration(now)/time.Microsecond),
		amount*int64(emissionInterval/time.Microsecond),
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
