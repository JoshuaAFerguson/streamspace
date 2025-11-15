// Package middleware defines constants for middleware components.
//
// This file centralizes configuration values for:
// - Rate limiting (prevent brute force and DoS attacks)
// - CSRF protection (prevent cross-site request forgery)
// - Request size limits (prevent payload DoS attacks)
//
// SECURITY FIX (2025-11-14):
// Extracted all magic numbers to improve code maintainability and security auditing.
package middleware

import "time"

// Rate Limiting Constants control the in-memory rate limiter.
//
// The rate limiter prevents brute force attacks by limiting requests per time window.
// Cleanup runs periodically to prevent memory leaks from abandoned rate limit entries.
const (
	// DefaultMaxAttempts is the default maximum number of attempts allowed
	DefaultMaxAttempts = 5

	// DefaultRateLimitWindow is the default time window for rate limiting
	DefaultRateLimitWindow = 1 * time.Minute

	// CleanupInterval is how often the rate limiter cleans up old entries
	CleanupInterval = 5 * time.Minute

	// CleanupThreshold is the age threshold for removing old entries
	CleanupThreshold = 10 * time.Minute
)
