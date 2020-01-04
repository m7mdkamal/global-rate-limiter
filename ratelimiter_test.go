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
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	start := time.Now()

	rl := NewRateLimiter(c, time.Second*10, 100)
	go func() {
		l2 := NewRateLimiter(c2, time.Second * 10, 100  )
		for i := 0; i < 10000; i++ {
			//start := time.Now()
			go l2.Incr()
			time.Sleep(time.Millisecond * 50)
			//fmt.Println(time.Now().Sub(start))
		}
	}()
	for i := 0; i < 10000; i++ {
		//start := time.Now()
		go rl.Incr()
		time.Sleep(time.Millisecond * 50)
		//fmt.Println(time.Now().Sub(start))
	}
	fmt.Println(time.Now().Sub(start))
	time.Sleep(time.Second * 100)
}
