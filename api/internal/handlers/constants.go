// Package handlers defines constants for HTTP handlers.
//
// This file centralizes all "magic numbers" and timeout values to:
// - Make configuration changes easier (single source of truth)
// - Improve code readability (named constants vs. bare numbers)
// - Document the reasoning behind specific values
// - Enable easy tuning for different environments
//
// SECURITY FIX (2025-11-14):
// Extracted all magic numbers to named constants as part of code quality improvements.
// This makes it easier to understand security-critical values like rate limits,
// timeouts, and buffer sizes.
//
// Categories:
// - MFA: Multi-factor authentication limits and timing
// - WebSocket: Connection parameters and buffer sizes
// - Webhook: Retry logic and timeouts
// - Session: Verification and expiry times
package handlers

import "time"

// MFA Constants control multi-factor authentication behavior.
//
// These values balance security (preventing brute force) with usability
// (not frustrating legitimate users).
const (
	// BackupCodesCount is the number of backup codes to generate
	BackupCodesCount = 10

	// BackupCodeLength is the length of each backup code
	BackupCodeLength = 8

	// MFAMaxAttemptsPerMinute is the maximum MFA verification attempts per minute
	MFAMaxAttemptsPerMinute = 5

	// MFARateLimitWindow is the time window for MFA rate limiting
	MFARateLimitWindow = 1 * time.Minute
)

// WebSocket Constants
const (
	// WebSocketPingInterval is how often to send ping messages
	WebSocketPingInterval = 54 * time.Second

	// WebSocketWriteDeadline is the deadline for write operations
	WebSocketWriteDeadline = 10 * time.Second

	// WebSocketReadDeadline is the deadline for read operations
	WebSocketReadDeadline = 60 * time.Second

	// WebSocketBufferSize is the size of the send buffer for each client
	WebSocketBufferSize = 256

	// WebSocketReadBufferSize is the size of the read buffer
	WebSocketReadBufferSize = 1024

	// WebSocketWriteBufferSize is the size of the write buffer
	WebSocketWriteBufferSize = 1024
)

// Webhook Constants
const (
	// WebhookDefaultMaxRetries is the default number of retry attempts
	WebhookDefaultMaxRetries = 3

	// WebhookDefaultRetryDelay is the default delay between retries in seconds
	WebhookDefaultRetryDelay = 60

	// WebhookDefaultBackoffMultiplier is the default exponential backoff multiplier
	WebhookDefaultBackoffMultiplier = 2.0

	// WebhookTimeout is the timeout for webhook HTTP requests
	WebhookTimeout = 10 * time.Second
)

// Session Constants
const (
	// SessionVerificationTimeout is how long a session verification is valid
	SessionVerificationTimeout = 60 * time.Second
)
