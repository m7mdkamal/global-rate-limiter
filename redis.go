package ratelimiter

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"log"
	"math/rand"
	"time"
)

const (
	namespacePrefix = "global-rate-limiter"
	thresholdKey    = namespacePrefix + ":threshold"
	intervalKey     = namespacePrefix + ":interval"
	resetTimeKey    = namespacePrefix + ":reset-time"
	lockedKey       = namespacePrefix + ":locked"
	counterKey      = namespacePrefix + ":counter"
	heldKey         = namespacePrefix + ":held"
)

var AcquireLockErr = errors.New("failed to acquire lock")

func acquireLockWithTimeout(conn *redis.Conn, acquireTimout, lockTimout time.Duration) (string, error) {
	defer conn.Close()
	rand.Seed(time.Now().UnixNano())
	identifier := fmt.Sprint(rand.Int())
	end := time.Now().Add(acquireTimout)
	for time.Now().Before(end) {
		cmd := conn.SetNX(lockedKey, identifier, lockTimout)

		if cmd.Val() {
			return identifier, nil
		} else {
			fmt.Println(cmd.Err())
		}

		time.Sleep(time.Millisecond)
	}
	return "", fmt.Errorf("acqurelockError denied: %w", AcquireLockErr)
}

func releaseLock(c *redis.Client, identifier string) error {
	err := c.Watch(func(tx *redis.Tx) error {
		if tx.Get(lockedKey).Val() != identifier {
			return errors.New("lock's identifier doesn't match")
		}
		tx.Del(lockedKey)
		return nil
	}, lockedKey)
	return err

}

func (r *ratelimiter) increase() error {
	c := r.redisClient
	id, err := acquireLockWithTimeout(c.Conn(), time.Second, time.Millisecond*10)
	if err != nil {
		log.Fatal("failed to acquire lock")
	}
	err = c.Watch(func(tx *redis.Tx) error {
		if tx.Exists(heldKey).Val() == 1 {
			return errors.New("LIMIT REACHED")
		}
		value := int(tx.Incr(counterKey).Val())
		if value >= r.threshold {
			tx.Pipelined(func(p redis.Pipeliner) error {
				tx.MSet(heldKey, "1")
				return nil
			})
		}
		return nil
	}, lockedKey)
	if err != nil {
		releaseLock(c, id)
		fmt.Println("watch value changed ", err)
		return err
	}
	return releaseLock(c, id)
}

func (r *ratelimiter) reset() {
	c := r.redisClient
	id, err := acquireLockWithTimeout(c.Conn(), time.Second*5, time.Millisecond*600)
	if err != nil {
		log.Fatalln("failed to acquire lock on reset ", err)
	}
	c.Watch(func(tx *redis.Tx) error {
		if tx.Exists(resetTimeKey).Val() == 1 {
			return nil
		}
		r.nextResetTime = time.Now().Add(r.interval)
		fmt.Println(r.nextResetTime)
		tx.Pipelined(func(p redis.Pipeliner) error {
			p.Del(heldKey, counterKey)
			p.MSet(
				thresholdKey, r.threshold,
				intervalKey, r.interval.Seconds(),
				resetTimeKey, r.nextResetTime.Unix(),
			)
			p.Expire(resetTimeKey, r.interval)
			return nil
		})
		return nil
	})
	_ = releaseLock(c, id)
}
