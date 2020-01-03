package ratelimiter

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"testing"
	"time"
)

func TestTwosdd(t *testing.T) {
	c := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	start := time.Now()

	rl := NewRateLimiter(c, time.Second*1, 11)
	for i := 0; i < 20; i++ {
		func() {
			err := rl.Incr()
			if err != nil {
				fmt.Println(err)
			}
		}()
	}
	fmt.Println(time.Now().Sub(start))
	time.Sleep(time.Second * 10)
}
