// Package logger provides structured logging using zerolog.
//
// This package implements production-ready logging with:
// - Structured JSON logging for production (machine-parsable)
// - Pretty console output for development (human-readable)
// - Component-specific loggers (security, websocket, webhook, etc.)
// - Configurable log levels (debug, info, warn, error, fatal)
//
// SECURITY ENHANCEMENT (2025-11-14):
// Added structured logging to replace fmt.Printf and log.Printf calls.
// Benefits:
// - Machine-parsable logs for security monitoring and alerting
// - Consistent log format across all components
// - Automatic field extraction for log aggregation (e.g., Elasticsearch)
// - Performance: Zero-allocation JSON logging in production
//
// Usage:
//   // Initialize once in main()
//   logger.Initialize("info", false) // production: JSON output
//   logger.Initialize("debug", true) // development: pretty output
//
//   // Use component-specific loggers
//   logger.Security().Warn().
//     Str("user_id", "user123").
//     Str("ip", "203.0.113.42").
//     Msg("Failed MFA attempt")
//
//   // Use global logger for general events
//   logger.Log.Info().Msg("Application started")
package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Global logger instance - use this for general application logging.
//
// For component-specific logging, use the helper functions like Security(),
// WebSocket(), etc. to get loggers with pre-configured component tags.
var (
	Log zerolog.Logger
)

// Initialize sets up the global logger with the specified level and output format.
//
// This function should be called once at application startup before any logging occurs.
//
// Parameters:
//   - level: Log level as string ("debug", "info", "warn", "error", "fatal", "panic")
//           Defaults to "info" if invalid level provided
//   - pretty: If true, use human-readable console output (development)
//           If false, use JSON output (production)
//
// Production Configuration:
//   logger.Initialize("info", false)
//   Output: {"level":"info","service":"streamspace-api","time":1699999999,"message":"User logged in"}
//
// Development Configuration:
//   logger.Initialize("debug", true)
//   Output: 10:30:00 INF User logged in service=streamspace-api
//
// Log Levels (from most to least verbose):
//   - debug: Detailed debugging information
//   - info: General informational messages
//   - warn: Warning messages (potential issues)
//   - error: Error messages (handled errors)
//   - fatal: Fatal errors (application exits)
//   - panic: Panic errors (application crashes)
func Initialize(level string, pretty bool) {
	// Parse log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLevel)

	// Configure output format
	if pretty {
		// Pretty console output for development
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	} else {
		// JSON output for production
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}

	// Set global logger
	Log = log.With().
		Str("service", "streamspace-api").
		Logger()

	Log.Info().
		Str("level", logLevel.String()).
		Bool("pretty", pretty).
		Msg("Logger initialized")
}

// GetLogger returns the global logger instance
func GetLogger() *zerolog.Logger {
	return &Log
}

// Security creates a logger for security events
func Security() *zerolog.Logger {
	l := Log.With().Str("component", "security").Logger()
	return &l
}

// WebSocket creates a logger for WebSocket events
func WebSocket() *zerolog.Logger {
	l := Log.With().Str("component", "websocket").Logger()
	return &l
}

// Webhook creates a logger for webhook events
func Webhook() *zerolog.Logger {
	l := Log.With().Str("component", "webhook").Logger()
	return &l
}

// Integration creates a logger for integration events
func Integration() *zerolog.Logger {
	l := Log.With().Str("component", "integration").Logger()
	return &l
}

// Database creates a logger for database events
func Database() *zerolog.Logger {
	l := Log.With().Str("component", "database").Logger()
	return &l
}

// HTTP creates a logger for HTTP request events
func HTTP() *zerolog.Logger {
	l := Log.With().Str("component", "http").Logger()
	return &l
}
