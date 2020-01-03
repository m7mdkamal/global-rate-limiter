package ratelimiter

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"math/rand"
	"strconv"
	"time"
)

const (
	namespacePrefix = "global-rate-limiter"
	thresholdKey = namespacePrefix + ":threshold"
	intervalKey = namespacePrefix + ":interval"
	resetTimeKey = namespacePrefix + ":reset-key"
	lockedKey = namespacePrefix + ":locked"
	counterKey = namespacePrefix + ":counter"
	heldKey = namespacePrefix + ":held"
)

var AcquireLockErr = errors.New("failed to acquire lock")

func initRedis(c *redis.Client, rl ratelimiter) {
	id, err := acquireLockWithTimeout(c.Conn(), time.Second, time.Second)
	if err != nil {
		if errors.Is(err, AcquireLockErr) {

		} else {

		}
	}

	c.Watch(func(tx *redis.Tx) error {
		tx.Pipelined(func(pip redis.Pipeliner) error {
			pip.MSet(
				thresholdKey, rl.threshold,
				intervalKey, rl.interval.Seconds(),
				resetTimeKey, time.Now().Add(rl.interval).Unix(),
				//namespacePrefix+":held", 0,
			)
			return nil
		})
		return nil
	})
	_ = releaseLock(c, id)
}

func acquireLockWithTimeout(conn *redis.Conn, acquireTimout, lockTimout time.Duration) (string, error) {
	identifier := fmt.Sprint(rand.Int())
	end := time.Now().Add(acquireTimout)
	for time.Now().Before(end) {
		if conn.SetNX(lockedKey, identifier, lockTimout).Val() {
			return identifier, nil
		}
		time.Sleep(time.Millisecond * 10)
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

func increase(c *redis.Client) error{
	id, _ := acquireLockWithTimeout(c.Conn(), time.Second, time.Second)
	err := c.Watch(func(tx *redis.Tx) error {
		if tx.Get(heldKey).Val() == "1" {
			return errors.New("LIMIT REACHED")
		}
		value := int(tx.Incr(counterKey).Val())
		threshold, err := strconv.Atoi(tx.Get(thresholdKey).Val())
		if err != nil {
			panic(err)
		}
		if value >= threshold {
			tx.Set(heldKey, "1", time.Second * 1000)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return releaseLock(c, id)
}

func reset(c *redis.Client) {
	id, _ := acquireLockWithTimeout(c.Conn(), time.Second, time.Second)
	c.Watch(func(tx *redis.Tx) error {
		tx.Pipelined(func(p redis.Pipeliner) error {
			p.Del(heldKey, counterKey)
			return nil
		})
		return nil
	})
	_ = releaseLock(c, id)
}
