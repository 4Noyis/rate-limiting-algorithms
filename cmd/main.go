package main

import (
	"fmt"

	ratelimiters "github.com/4Noyis/rate-limiter-algorithms/rateLimiters"
)

func main() {
	fmt.Println("Go project rate-limiter-algorithms")

	ratelimiters.TokenBucketRateLimiter(5, 1)

	//	ratelimiters.LeakingBucketRateLimiter(5, 0.25)

	//	ratelimiters.FixedWindowCounterRateLimiter(5, 10*time.Second)

	//	ratelimiters.SlidingWindowLogRateLimiter(5, 10*time.Second)

}
