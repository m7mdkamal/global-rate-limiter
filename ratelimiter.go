package ratelimiter

import (
	"github.com/go-redis/redis/v7"
	"time"
)

type ratelimiter struct {
	redisClient *redis.Client
	threshold   int
	interval    time.Duration
}

func NewRateLimiter(c *redis.Client, interval time.Duration, threshold int) *ratelimiter {
	r := ratelimiter{
		redisClient: c,
		threshold:   threshold,
		interval:    interval,
	}
	initRedis(r.redisClient, r)
	go r.resetLoop()
	return &r
}

func (r *ratelimiter) resetLoop() {
	for {
		select {
		case <-time.Tick(r.interval):
			r.reset()
		}
	}
}

func (r *ratelimiter) Incr() error {
	return increase(r.redisClient)
}

func (r *ratelimiter) reset() {
	reset(r.redisClient)
}
