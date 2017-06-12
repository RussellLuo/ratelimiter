package ratelimiter

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
)

type Redis interface {
	Eval(script string, keys []string, args ...interface{}) (interface{}, error)
	EvalSha(sha1 string, keys []string, args ...interface{}) (interface{}, error, bool)
}

type Script struct {
	redis Redis
	src   string
	hash  string
}

func NewScript(redis Redis, src string) *Script {
	h := sha1.New()
	io.WriteString(h, src)
	return &Script{
		redis: redis,
		src:   src,
		hash:  hex.EncodeToString(h.Sum(nil)),
	}
}

func (s *Script) Run(keys []string, args ...interface{}) (interface{}, error) {
	result, err, noScript := s.redis.EvalSha(s.hash, keys, args...)
	if noScript {
		result, err = s.redis.Eval(s.src, keys, args...)
	}
	return result, err
}
