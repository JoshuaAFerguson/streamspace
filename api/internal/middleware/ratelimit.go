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

// CheckLimit checks if the rate limit has been exceeded
// Returns true if request is allowed, false if rate limit exceeded
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
