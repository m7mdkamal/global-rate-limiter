package ratelimiter

import (
	"github.com/go-redis/redis/v7"
	"time"
)

type ratelimiter struct {
	redisClient   *redis.Client
	threshold     int
	interval      time.Duration
	nextResetTime time.Time

	ResetChan     chan<- struct{}
}

func NewRateLimiter(c *redis.Client, interval time.Duration, threshold int) *ratelimiter {
	r := ratelimiter{
		redisClient: c,
		threshold:   threshold,
		interval:    interval,
	}
	go r.resetLoop()
	return &r
}

func (r *ratelimiter) resetLoop() {
	r.reset()
	for {
		select {
		case <-time.Tick(r.interval):
			r.reset()
		}
	}
}

func (r *ratelimiter) Incr() error {
	return r.increase()
}
