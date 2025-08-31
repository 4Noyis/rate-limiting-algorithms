package ratelimiters

import (
	"fmt"
	"sync"
	"time"
)

type TokenBucket struct {
	capacity       float64    // maximum number of tokens
	tokens         float64    // current number of tokens
	refillRate     float64    // tokens added per second
	lastRefillTime time.Time  // last time the bucket was refilled
	mutex          sync.Mutex // mutex for thread safety
}

// creates new token bucket with given capacity and refill rate
func NewTokenBucket(capacity, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:       capacity,
		tokens:         capacity,
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
	}
}

// adds tokens to bucket based on elapsed time
func (tb *TokenBucket) refill() {
	currentTime := time.Now()
	elapsedTime := currentTime.Sub(tb.lastRefillTime).Seconds()

	// Calculate new tokens based on elapsed time
	newTokens := elapsedTime * tb.refillRate
	tb.tokens = min(tb.capacity, tb.tokens+newTokens)
	tb.lastRefillTime = currentTime
}

// AllowRequest checks if a request can be allowed and consumes a token if available
func (tb *TokenBucket) AllowRequest() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	tb.refill()

	if tb.tokens >= 1 {
		tb.tokens -= 1
		return true
	}
	return false
}

// GetTokens returns the current number of tokens (for debugging/monitoring)
func (tb *TokenBucket) GetTokens() float64 {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	tb.refill()
	return tb.tokens
}

func TokenBucketRateLimiter(BucketCapacity float64, refillRate float64) {
	// Example usage: 5 tokens max, refill 1 token per second
	bucket := NewTokenBucket(BucketCapacity, refillRate)

	for i := 0; i < 10; i++ {
		if bucket.AllowRequest() {
			fmt.Printf("Request %d: allowed (tokens remaining: %.2f)\n", i+1, bucket.GetTokens())
		} else {
			fmt.Printf("Request %d: denied (tokens remaining: %.2f)\n", i+1, bucket.GetTokens())
		}
		time.Sleep(500 * time.Millisecond) // Simulate request interval
	}
}
