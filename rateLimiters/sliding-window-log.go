package ratelimiters

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

// SlidingWindowLog represents a sliding window log rate limiter
type SlidingWindowLog struct {
	limit      int           // maximum requests allowed in the window
	windowSize time.Duration // time window duration
	requests   *list.List    // stores timestamps of requests (acts as deque)
	mutex      sync.Mutex    // mutex for thread safety
}

// NewSlidingWindowLog creates a new sliding window log with the specified limit and window size
func NewSlidingWindowLog(limit int, windowSize time.Duration) *SlidingWindowLog {
	return &SlidingWindowLog{
		limit:      limit,
		windowSize: windowSize,
		requests:   list.New(),
	}
}

// cleanupOldRequests removes requests outside the current window
func (swl *SlidingWindowLog) cleanupOldRequests(currentTime time.Time) {
	cutoffTime := currentTime.Add(-swl.windowSize)

	// Remove requests older than the window
	for swl.requests.Len() > 0 {
		front := swl.requests.Front()
		if front == nil {
			break
		}

		requestTime := front.Value.(time.Time)
		if requestTime.After(cutoffTime) {
			break // All remaining requests are within the window
		}

		swl.requests.Remove(front)
	}
}

// AllowRequest checks if a request can be allowed within the sliding window
func (swl *SlidingWindowLog) AllowRequest() bool {
	swl.mutex.Lock()
	defer swl.mutex.Unlock()

	currentTime := time.Now()

	// Clean up old requests outside the window
	swl.cleanupOldRequests(currentTime)

	// Check if request can be allowed
	if swl.requests.Len() < swl.limit {
		swl.requests.PushBack(currentTime)
		return true
	}
	return false
}

// GetCurrentCount returns the current count of requests in the sliding window
func (swl *SlidingWindowLog) GetCurrentCount() int {
	swl.mutex.Lock()
	defer swl.mutex.Unlock()

	currentTime := time.Now()
	swl.cleanupOldRequests(currentTime)

	return swl.requests.Len()
}

// GetOldestRequestAge returns how long ago the oldest request in the window occurred
func (swl *SlidingWindowLog) GetOldestRequestAge() time.Duration {
	swl.mutex.Lock()
	defer swl.mutex.Unlock()

	currentTime := time.Now()
	swl.cleanupOldRequests(currentTime)

	if swl.requests.Len() == 0 {
		return 0
	}

	oldestRequest := swl.requests.Front().Value.(time.Time)
	return currentTime.Sub(oldestRequest)
}

// GetTimeUntilSlotAvailable returns when the next request slot will become available
func (swl *SlidingWindowLog) GetTimeUntilSlotAvailable() time.Duration {
	swl.mutex.Lock()
	defer swl.mutex.Unlock()

	currentTime := time.Now()
	swl.cleanupOldRequests(currentTime)

	// If we're not at the limit, a slot is available now
	if swl.requests.Len() < swl.limit {
		return 0
	}

	// Find when the oldest request will expire
	oldestRequest := swl.requests.Front().Value.(time.Time)
	expirationTime := oldestRequest.Add(swl.windowSize)

	if expirationTime.After(currentTime) {
		return expirationTime.Sub(currentTime)
	}

	return 0
}

// GetWindowInfo returns detailed information about the current window state
func (swl *SlidingWindowLog) GetWindowInfo() (count, limit int, utilizationPercent float64, oldestAge time.Duration) {
	swl.mutex.Lock()
	defer swl.mutex.Unlock()

	currentTime := time.Now()
	swl.cleanupOldRequests(currentTime)

	count = swl.requests.Len()
	utilization := float64(count) / float64(swl.limit) * 100

	var oldest time.Duration
	if count > 0 {
		oldestRequest := swl.requests.Front().Value.(time.Time)
		oldest = currentTime.Sub(oldestRequest)
	}

	return count, swl.limit, utilization, oldest
}

// GetRequestTimestamps returns all request timestamps in the current window (for debugging)
func (swl *SlidingWindowLog) GetRequestTimestamps() []time.Time {
	swl.mutex.Lock()
	defer swl.mutex.Unlock()

	currentTime := time.Now()
	swl.cleanupOldRequests(currentTime)

	timestamps := make([]time.Time, 0, swl.requests.Len())
	for e := swl.requests.Front(); e != nil; e = e.Next() {
		timestamps = append(timestamps, e.Value.(time.Time))
	}

	return timestamps
}

func SlidingWindowLogRateLimiter(requestLimit int, windowSize time.Duration) {
	// Example usage: 5 requests per 10 seconds
	slidingWindow := NewSlidingWindowLog(requestLimit, windowSize)

	fmt.Println("Sliding Window Log Rate Limiter Demo")
	fmt.Println("Limit: 5 requests per 10 seconds")
	fmt.Println("Request interval: 2 seconds")
	fmt.Println("====================================")

	for i := 0; i < 12; i++ {
		count, limit, utilization, oldestAge := slidingWindow.GetWindowInfo()
		timeUntilSlot := slidingWindow.GetTimeUntilSlotAvailable()

		if slidingWindow.AllowRequest() {
			fmt.Printf("Request %2d: allowed  (count: %d/%d, util: %5.1f%%, oldest: %4.1fs ago)\n",
				i+1, count+1, limit, utilization, oldestAge.Seconds())
		} else {
			fmt.Printf("Request %2d: denied   (count: %d/%d, util: %5.1f%%, next slot in: %4.1fs)\n",
				i+1, count, limit, utilization, timeUntilSlot.Seconds())
		}

		// Show the sliding nature after some requests
		if i == 4 {
			fmt.Println("--- Notice how the window slides continuously ---")
		}

		time.Sleep(2 * time.Second) // Simulate request interval
	}

	fmt.Println("\nRequest timeline in current window:")
	timestamps := slidingWindow.GetRequestTimestamps()
	now := time.Now()

	for i, ts := range timestamps {
		age := now.Sub(ts)
		fmt.Printf("Request %d: %.1fs ago\n", i+1, age.Seconds())
	}
}
