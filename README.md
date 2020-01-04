# global-rate-limiter
A fixed window rate limiter

## Installation
```bash
go get github.com/m7mdkamal/global-rate-limiter
```

## Usage
```golang
func main() {
	redisClient := redis.NewClient(&redis.Options{})

	rateLimiter := ratelimiter.NewRateLimiter(redisClient, time.Hour, 1000) // 1000 request per 1 hour

	_ = rateLimiter.Incr() // returns error on limit reached
}
```