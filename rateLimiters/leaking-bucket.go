package ratelimiters

import (
	"fmt"
	"sync"
	"time"
)

type LeakyBucket struct {
	capacity float64    // maximum bucket size
	water    float64    // current amount of "water" in the bucket
	leakRate float64    // rate at which water leaks per second
	lastTime time.Time  // last time the bucket was updated
	mutex    sync.Mutex // mutex for thread safety
}

func NewLeakyBucket(capacity, leakRate float64) *LeakyBucket {
	return &LeakyBucket{
		capacity: capacity,
		water:    0,
		leakRate: leakRate,
		lastTime: time.Now(),
	}
}

// leak removes water from the bucket based on elapsed time
func (lb *LeakyBucket) leak() {
	currentTime := time.Now()
	elapsedTime := currentTime.Sub(lb.lastTime).Seconds()

	// Remove water based on elapsed time and leak rate
	leaked := elapsedTime * lb.leakRate
	lb.water = max(0, lb.water-leaked)
	lb.lastTime = currentTime
}

// AllowRequest checks if a request can be allowed and adds water if there's capacity
func (lb *LeakyBucket) AllowRequest() bool {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	lb.leak()

	if lb.water < lb.capacity {
		lb.water += 1 // Add 1 unit of "water" for each request
		return true
	}
	return false
}

// GetWaterLevel returns the current water level (for debugging/monitoring)
func (lb *LeakyBucket) GetWaterLevel() float64 {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	lb.leak()
	return lb.water
}

func (lb *LeakyBucket) GetCapacityUsed() float64 {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	lb.leak()
	return (lb.water / lb.capacity) * 100
}

func LeakingBucketRateLimiter(bucketCapacity float64, leakRate float64) {
	// Example usage: max 5 requests in bucket, leak 1 per second
	bucket := NewLeakyBucket(bucketCapacity, leakRate)

	fmt.Println("Leaky Bucket Rate Limiter Demo")
	fmt.Println("Bucket capacity: 5, Leak rate: 1 per second")
	fmt.Println("Request interval: 500ms")
	fmt.Println("================================")

	for i := 0; i < 10; i++ {
		if bucket.AllowRequest() {
			fmt.Printf("Request %d: allowed (water level: %.2f/%.0f, usage: %.1f%%)\n",
				i+1, bucket.GetWaterLevel(), bucket.capacity, bucket.GetCapacityUsed())
		} else {
			fmt.Printf("Request %d: denied (water level: %.2f/%.0f, usage: %.1f%%)\n",
				i+1, bucket.GetWaterLevel(), bucket.capacity, bucket.GetCapacityUsed())
		}
		time.Sleep(500 * time.Millisecond) // Simulate request interval
	}

	fmt.Println("\nWaiting 3 seconds to see bucket drain...")
	time.Sleep(3 * time.Second)
	fmt.Printf("Final water level: %.2f/%.0f (%.1f%% full)\n",
		bucket.GetWaterLevel(), bucket.capacity, bucket.GetCapacityUsed())
}
