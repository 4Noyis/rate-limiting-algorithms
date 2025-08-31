package ratelimiters

import (
	"fmt"
	"sync"
	"time"
)

type FixedWindowCounter struct {
	limit      int           // maximum requests allowed in the window
	windowSize time.Duration // time window duration
	count      int           // current count of requests in the window
	startTime  time.Time     // start time of the current window
	mutex      sync.Mutex    // mutex for thread safety
}

func NewFixedWindowCounter(limit int, windowSize time.Duration) *FixedWindowCounter {
	return &FixedWindowCounter{
		limit:      limit,
		windowSize: windowSize,
		count:      0,
		startTime:  time.Now(),
	}
}

func (fwc *FixedWindowCounter) AllowRequest() bool {
	fwc.mutex.Lock()
	defer fwc.mutex.Unlock()

	currentTime := time.Now()

	// Check if we need to reset the window
	if currentTime.Sub(fwc.startTime) >= fwc.windowSize {
		// Reset the window
		fwc.startTime = currentTime
		fwc.count = 0
	}

	// Check if request can be allowed
	if fwc.count < fwc.limit {
		fwc.count++
		return true
	}
	return false
}

// GetCurrentCount returns the current count of requests in the window
func (fwc *FixedWindowCounter) GetCurrentCount() int {
	fwc.mutex.Lock()
	defer fwc.mutex.Unlock()

	currentTime := time.Now()

	// Check if window has expired
	if currentTime.Sub(fwc.startTime) >= fwc.windowSize {
		return 0 // Window has expired, count would be reset
	}

	return fwc.count

}

// GetTimeUntilReset returns the time remaining until the window resets
func (fwc *FixedWindowCounter) GetTimeUntilReset() time.Duration {
	fwc.mutex.Lock()
	defer fwc.mutex.Unlock()

	currentTime := time.Now()
	elapsed := currentTime.Sub(fwc.startTime)

	if elapsed >= fwc.windowSize {
		return 0 // Window should be reset
	}

	return fwc.windowSize - elapsed

}

// GetWindowInfo returns detailed information about the current window state
func (fwc *FixedWindowCounter) GetWindowInfo() (count, limit int, timeUntilReset time.Duration, utilizationPercent float64) {
	fwc.mutex.Lock()
	defer fwc.mutex.Unlock()

	currentTime := time.Now()
	elapsed := currentTime.Sub(fwc.startTime)

	// Check if window has expired
	if elapsed >= fwc.windowSize {
		return 0, fwc.limit, 0, 0.0
	}

	utilization := float64(fwc.count) / float64(fwc.limit) * 100
	timeRemaining := fwc.windowSize - elapsed

	return fwc.count, fwc.limit, timeRemaining, utilization
}

func FixedWindowCounterRateLimiter(requestLimit int, windowSize time.Duration) {
	// Example usage: 5 requests per 10 seconds
	window := NewFixedWindowCounter(requestLimit, windowSize)

	fmt.Println("Fixed Window Counter Rate Limiter Demo")
	fmt.Println("Limit: 5 requests per 10 seconds")
	fmt.Println("Request interval: 2 seconds")
	fmt.Println("=====================================")

	for i := 0; i < 12; i++ {
		count, limit, timeUntilReset, utilization := window.GetWindowInfo()

		if window.AllowRequest() {
			fmt.Printf("Request %2d: allowed  (count: %d/%d, utilization: %5.1f%%, reset in: %6.1fs)\n",
				i+1, count+1, limit, utilization, timeUntilReset.Seconds())
		} else {
			fmt.Printf("Request %2d: denied   (count: %d/%d, utilization: %5.1f%%, reset in: %6.1fs)\n",
				i+1, count, limit, utilization, timeUntilReset.Seconds())
		}

		// Show window reset
		if i == 4 {
			fmt.Println("--- Window will reset after next request ---")
		}

		time.Sleep(2 * time.Second) // Simulate request interval
	}

	fmt.Println("\nFinal window state:")
	count, limit, timeUntilReset, utilization := window.GetWindowInfo()
	fmt.Printf("Count: %d/%d, Utilization: %.1f%%, Time until reset: %.1fs\n",
		count, limit, utilization, timeUntilReset.Seconds())
}
