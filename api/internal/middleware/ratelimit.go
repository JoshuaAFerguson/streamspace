// Package middleware provides HTTP middleware for the StreamSpace API.
// This file implements rate limiting to prevent brute force and DoS attacks.
//
// SECURITY ENHANCEMENT (2025-11-14):
// Added in-memory rate limiting for MFA verification to prevent brute force attacks.
//
// Rate limiting is critical for security:
// - MFA codes are only 6 digits (1 million combinations)
// - Without rate limiting, codes can be brute forced in minutes
// - With 5 attempts/minute limit, brute force takes ~160 days
//
// Current Implementation: In-Memory (Development/Single-Server)
// - Fast: No network round-trips
// - Simple: No external dependencies
// - Limitations: Not distributed (doesn't work across multiple API servers)
//
// Production Recommendation: Redis-Backed Rate Limiting
// - Distributed: Works across multiple API servers
// - Persistent: Survives API server restarts
// - Scalable: Handles millions of concurrent rate limit entries
//
// Usage:
//   // In handler
//   limiter := middleware.GetRateLimiter()
//   if !limiter.CheckLimit("user:123:mfa", 5, 1*time.Minute) {
//     return errors.New("rate limit exceeded")
//   }
package middleware

import (
	"sync"
	"time"
)

// RateLimiter implements a simple in-memory sliding window rate limiter.
//
// Thread Safety: Uses sync.RWMutex for concurrent access protection.
//
// Algorithm: Sliding Window
// - Records timestamp of each attempt
// - Filters out attempts outside the time window
// - Counts remaining attempts
// - Allows if count < maxAttempts
//
// Memory Management:
// - Automatic cleanup runs every 5 minutes
// - Removes entries older than 10 minutes
// - Prevents memory leaks from abandoned rate limits
//
// For production use with multiple API servers, replace with Redis-backed
// implementation for distributed rate limiting.
type RateLimiter struct {
	attempts map[string][]time.Time
	mu       sync.RWMutex
}

var (
	globalRateLimiter = &RateLimiter{
		attempts: make(map[string][]time.Time),
	}
	cleanupOnce sync.Once
)

// GetRateLimiter returns the singleton rate limiter instance
func GetRateLimiter() *RateLimiter {
	// Start cleanup goroutine once
	cleanupOnce.Do(func() {
		go globalRateLimiter.cleanup()
	})
	return globalRateLimiter
}

// CheckLimit checks if the rate limit has been exceeded using sliding window algorithm.
//
// This method is the core of the rate limiting system. It implements a sliding window
// counter that accurately tracks requests over time, preventing both burst attacks and
// sustained high-rate attacks.
//
// # Algorithm: Sliding Window Counter
//
// Traditional fixed window problems:
//   - User makes 99 requests at 00:59
//   - Window resets at 01:00
//   - User makes 99 more requests at 01:01
//   - Result: 198 requests in 2 seconds (should be 100/minute max)
//
// Sliding window solution:
//   - Track timestamp of each individual request
//   - Filter requests to only those within the time window from now
//   - Count filtered requests against limit
//   - More accurate but requires storing all timestamps
//
// # Parameters
//
// **key** (string):
//   - Unique identifier for the resource being rate limited
//   - Format: "{resource_type}:{resource_id}:{action}"
//   - Examples:
//     - "user:123:login" (login attempts for user 123)
//     - "user:456:mfa" (MFA verification for user 456)
//     - "ip:192.168.1.1:api" (API requests from IP)
//     - "session:sess-789:create" (session creation attempts)
//
// **maxAttempts** (int):
//   - Maximum number of requests allowed within the window
//   - Examples:
//     - 5 for MFA verification (5 wrong codes/minute)
//     - 10 for login attempts (10 failed logins/minute)
//     - 100 for API requests (100 requests/minute)
//     - 1000 for read operations (1000 reads/minute)
//
// **window** (time.Duration):
//   - Time window for counting requests
//   - Examples:
//     - 1*time.Minute for short-term protection
//     - 5*time.Minute for medium-term protection
//     - 1*time.Hour for long-term protection
//
// # Return Value
//
// Returns true if request is allowed, false if rate limit exceeded:
//   - true: Attempt recorded, request proceeds
//   - false: Limit exceeded, request rejected (attempt NOT recorded)
//
// # Thread Safety
//
// This method is thread-safe:
//   - Uses write lock (rl.mu.Lock()) for exclusive access
//   - Safe for concurrent calls from multiple goroutines
//   - Lock held for entire operation (atomic check-and-increment)
//
// # Performance Characteristics
//
// Time complexity:
//   - O(n) where n is number of attempts in window
//   - Typical n = 5-100 (very fast)
//   - Worst case: n = maxAttempts (still fast)
//
// Memory usage:
//   - ~24 bytes per attempt (time.Time is 24 bytes)
//   - Example: 100 attempts = 2.4 KB
//   - Automatic cleanup prevents unbounded growth
//
// Latency:
//   - Average: <1ms (in-memory operation)
//   - Worst case: <5ms (with many attempts to filter)
//
// # Security Considerations
//
// **Brute Force Protection**:
//   - Example: 6-digit MFA code (1,000,000 combinations)
//   - Without rate limiting: Brute force in minutes
//   - With 5 attempts/minute: Brute force takes ~160 days
//
// **DoS Protection**:
//   - Prevents overwhelming server with requests
//   - Limits resource consumption per user/IP
//   - Ensures fair resource allocation
//
// **Important**: Rate limit keys should include user ID or IP:
//   - Bad: "mfa" (global limit, one user blocks everyone)
//   - Good: "user:123:mfa" (per-user limit, isolated)
//
// # Edge Cases
//
// **Empty history**: First request is always allowed
//   - No previous attempts exist
//   - Request is recorded and allowed
//
// **Exactly at limit**: If count == maxAttempts, request is rejected
//   - Example: maxAttempts=5, current=5, result=false
//   - This is correct (limit is "up to N", not "N+1")
//
// **All attempts expired**: Old attempts don't count
//   - If all previous attempts are outside window, count=0
//   - Request is allowed (like fresh start)
//
// **Concurrent requests**: First one to acquire lock wins
//   - If 2 requests race to be the "Nth" attempt
//   - Lock ensures only one is recorded as the Nth
//   - Other is rejected as "N+1th"
//
// # Example Usage
//
// **MFA verification** (strict):
//
//	limiter := middleware.GetRateLimiter()
//	userID := "user-123"
//	key := fmt.Sprintf("user:%s:mfa", userID)
//
//	if !limiter.CheckLimit(key, 5, 1*time.Minute) {
//	    return errors.New("too many MFA attempts, please wait")
//	}
//
//	// Proceed with MFA verification
//	if !verifyMFACode(userID, code) {
//	    return errors.New("invalid MFA code")
//	}
//
//	// Success - reset limit
//	limiter.ResetLimit(key)
//
// **API rate limiting** (generous):
//
//	limiter := middleware.GetRateLimiter()
//	userID := c.GetString("user_id")
//	key := fmt.Sprintf("user:%s:api", userID)
//
//	if !limiter.CheckLimit(key, 1000, 1*time.Minute) {
//	    c.JSON(429, gin.H{"error": "rate limit exceeded"})
//	    return
//	}
//
// **Progressive backoff** (escalating):
//
//	limiter := middleware.GetRateLimiter()
//	ip := c.ClientIP()
//
//	// Check 1-minute window (short-term protection)
//	if !limiter.CheckLimit(fmt.Sprintf("ip:%s:1m", ip), 10, 1*time.Minute) {
//	    c.JSON(429, gin.H{"error": "rate limit exceeded (1 min)"})
//	    return
//	}
//
//	// Check 1-hour window (long-term protection)
//	if !limiter.CheckLimit(fmt.Sprintf("ip:%s:1h", ip), 100, 1*time.Hour) {
//	    c.JSON(429, gin.H{"error": "rate limit exceeded (1 hour)"})
//	    return
//	}
//
// # Known Limitations
//
//  1. **In-memory only**: Not distributed across multiple servers
//     - Each API server has independent limits
//     - Attackers can bypass by spreading across servers
//     - Solution: Use Redis for distributed rate limiting
//
//  2. **Lost on restart**: Rate limit state lost when server restarts
//     - Attackers could force restart to reset limits
//     - Solution: Persist to Redis or database
//
//  3. **Memory growth**: Without cleanup, memory usage unbounded
//     - Solution: Automatic cleanup runs every 5 minutes (implemented)
//
//  4. **No burst allowance**: Sliding window is strict
//     - Can't "save up" unused capacity for later burst
//     - Solution: Implement token bucket algorithm instead
//
// See also:
//   - ResetLimit(): Clear rate limit for a key
//   - GetAttempts(): Check current attempt count
func (rl *RateLimiter) CheckLimit(key string, maxAttempts int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Get existing attempts for this key
	attempts, exists := rl.attempts[key]
	if !exists {
		attempts = []time.Time{}
	}

	// Filter out attempts outside the time window
	validAttempts := []time.Time{}
	for _, t := range attempts {
		if now.Sub(t) < window {
			validAttempts = append(validAttempts, t)
		}
	}

	// Check if limit exceeded
	if len(validAttempts) >= maxAttempts {
		// Update with filtered attempts (don't record this request)
		rl.attempts[key] = validAttempts
		return false
	}

	// Record this attempt
	validAttempts = append(validAttempts, now)
	rl.attempts[key] = validAttempts

	return true
}

// ResetLimit clears all attempts for a given key
func (rl *RateLimiter) ResetLimit(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.attempts, key)
}

// GetAttempts returns the number of attempts within the window for a key
func (rl *RateLimiter) GetAttempts(key string, window time.Duration) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now()
	attempts, exists := rl.attempts[key]
	if !exists {
		return 0
	}

	count := 0
	for _, t := range attempts {
		if now.Sub(t) < window {
			count++
		}
	}

	return count
}

// cleanup periodically removes old entries to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()

		for key, attempts := range rl.attempts {
			// Remove entries older than cleanup threshold
			validAttempts := []time.Time{}
			for _, t := range attempts {
				if now.Sub(t) < CleanupThreshold {
					validAttempts = append(validAttempts, t)
				}
			}

			if len(validAttempts) == 0 {
				delete(rl.attempts, key)
			} else {
				rl.attempts[key] = validAttempts
			}
		}

		rl.mu.Unlock()
	}
}
