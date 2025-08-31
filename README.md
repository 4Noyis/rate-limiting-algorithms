# Rate Limiting Algorithms

Rate Limiting Algorithms are mechanisms designed to control the rate at which requests are processed or served by a system. They help prevent system overload, ensure fair resource allocation, and protect against abuse or denial-of-service attacks.


## Token Bucket Algorithm
The token bucket algorithm maintains a bucket that holds tokens. Each request consumes one token, and tokens are refilled at a constant rate. This algorithm allows for burst traffic while maintaining an average rate limit.

### How it Works
- Bucket Capacity: Maximum number of tokens the bucket can hold
- Refill Rate: Rate at which new tokens are addet to the bucket
- Request Proccessing: Each request consumes one tokens; if no tokens are available, the request is denied

#### Key Features

- ✅ Burst Handling: Allows temporary bursts up to bucket capacity
- ✅ Flexible: Good balance between strictness and allowing traffic spikes
- ✅ Memory Efficient: Only stores current token count and last refill time

#### Use Cases

- API rate limiting with burst allowance
- Network traffic shaping
- Resource allocation systems
---

## Leaky Bucket Algorithm
The leaky bucket algorithm processes incoming requests at a predetermined constant rate, similar to water leaking from a bucket with a small hole. Requests are queued and processed in a first-in, first-out (FIFO) order.

### How It Works
Incoming requests are added to a bucket (queue) and processed at a fixed rate. If the bucket is full when a new request arrives, the request is dropped.

### Algorithm Flow 
1. Request Arrives -> Check current water level
2. If bucket not full -> Add 1 unit of water, allow request
3. If bucket full -> Deny request (overflow protection)
4. Water continuously leaks at the the specified rate

Incoming request = water poured in from the top
leak rate = size of the hole (how fast water drains)
capacity = bucket size (prevents overflow)

#### Key Features
- ✅ Smooth Output: Ensures steady processing rate
- ✅ Queue Management: Handles request queuing automatically
- ⚠️ Burst Limitation: Not ideal for handling large traffic spikes

#### Use Cases
- Traffic smoothing
- Queue management systems
- Preventing system overload
---

## Fixed Window Counter Algorithm
The fixed window counter algorithm tracks incoming requests within predetermined timeframes. It divides the timeline into discrete, non-overlapping windows and counts requests within each window according to a set limit.

### How It Works
Time is divided into fixed windows (e.g., every 10 seconds). Within each window, requests are counted up to a specified limit. When a window expires, the counter resets to zero.

### Window Lifecycle
1. Window Start: t=0s -> Counter starts at 0
2. During Window: t=0s to t=10s -> Count requests up to limit 
3. Window End: t=10s -> Reset counter to 0, start new window
4. Repeat New window begins immediately 

Example: ```NewFixedWindowCounter(5, 10*time.Second)```
```
Time: 0s  2s  4s  6s  8s  10s 12s 14s 16s 18s 20s
Req:  1   2   3   4   5   6   1   2   3   4   5
      ↑   ↑   ↑   ↑   ✗   ↑   ↑   ↑   ↑   ↑   ✗
      └─────Window 1──────┘   └─────Window 2──────┘
```

#### Key Features
- ✅ Simple: Easy to understand and implement
- ✅ Memory Efficient: Only stores count and window start time
- ⚠️ Boundary Issues: Can allow 2x limit across window boundaries
- ⚠️ Unfair: Early requests in window get priority

#### Use Cases
- Simple API quotas
- Resource limiting with predictable reset times
- Systems where simplicity is preferred over precision
---

## Sliding Window Log Algorithm
Unlike fixed windows, the sliding window continuously moves with each request, maintaining an exact log of when each request occurred within the time window. This provides the most accurate rate limiting.

### How It Works
The algorithm maintains a log of all request timestamps within the current window. As time progresses, old requests automatically fall out of the window, and new requests can be accepted.

### Continuous Sliding
```
Timeline: 0s  2s  4s  6s  8s  10s 12s 14s 16s 18s
Requests: 1   2   3   4   5   6   7   8   9   10

At t=10s: Window covers [0s-10s] → requests 1,2,3,4,5 ✓
At t=12s: Window covers [2s-12s] → requests 2,3,4,5,6 ✓  
At t=14s: Window covers [4s-14s] → requests 3,4,5,6,7 ✓
```

#### Request Processing
1. Cleanup: Remove timestamps older than window size
2. Check: Count remaining requests < limit?
3. Record: If allowed, add current timestamp
4. Slide: Window automatically slides with time

#### Key Features
- ✅ Precise: Exact enforcement of rate limits
- ✅ Fair: No boundary effects like fixed windows
- ✅ Smooth: Gradual capacity recovery as old requests expire
- ⚠️ Memory Intensive: Stores every request timestamp
- ⚠️ Performance: O(n) cleanup operation per request

#### Use Cases
- Critical APIs requiring precise rate limiting
- Systems where fairness is essential
- Applications with audit requirements

---
## Comparison of Algorithms
```
| Algorithm    | Memory | Accuracy | Burst Handling | Complexity | Best For                 |
|───────────────────────────────────────────────────────────────────────────────────────────|
| Token Bucket | O(1)   | Good     | Excellent      | O(1)       | APIs with burst allowance|
| Leaky Bucket | O(1)   | Good     | Good           | O(1)       | Traffic smoothing        |
| Fixed Window | O(1)   | Good     | Poor           | O(1)       | Simple quotos            |
| Sliding Log  | O(n)   | Excellent| Excellent      | O(n)       | Precise rate limiting    |
└──────────────┘────────┘──────────┘────────────────┘────────────┘──────────────────────────┘ 
```

## Choosing the Right Algorithm

Use Token Bucket when:
- You need to allow burst traffic
- API clients expect some flexibility
- Memory efficiency is important

Use Leaky Bucket when:
- Smooth, consistent output rate is required
- You want to prevent traffic spikes from reaching downstream systems
- Queue management is beneficial

Use Fixed Window when:
- Simplicity is more important than precision
- You need predictable reset times
- Memory usage must be minimal

Use Sliding Window Log when:
- Precise rate limiting is critical
- Fairness across all time periods is required
- You can afford higher memory usage
- Audit trails are needed

## Implementation Considerations
### Thread Safety
All implementations should include proper synchronization mechanisms (mutexes, atomic operations) when used in concurrent environments.

#### Memory Management

- Token/Leaky Bucket: Constant memory usage
- Fixed Window: Constant memory usage
- Sliding Window: Memory grows with request rate - consider cleanup strategies

#### Performance

- Token/Leaky/Fixed: O(1) operations
- Sliding Window: O(n) operations - optimize cleanup frequency