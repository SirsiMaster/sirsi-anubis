// Package logging provides structured logging for Sirsi Anubis.
//
// Uses Go 1.21+ log/slog for structured, leveled logging.
// CLI output (user-facing) stays as fmt.Printf in cmd/.
// This package handles diagnostic/debug logging only.
package logging

import (
	"io"
	"log/slog"
	"os"
)

// Level aliases for convenience.
const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

var (
	// logger is the package-level logger instance.
	logger *slog.Logger
	// level is the current log level (can be changed at runtime).
	level = new(slog.LevelVar)
)

func init() {
	// Default: warn level, text format, stderr (not stdout — stdout is for CLI output).
	level.Set(slog.LevelWarn)
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
}

// Init configures the global logger. Call this from main() after flag parsing.
//
//   - verbose: enables debug-level logging
//   - quiet: disables all logging below error
//   - jsonFormat: emits structured JSON instead of text
func Init(verbose, quiet, jsonFormat bool) {
	switch {
	case quiet:
		level.Set(slog.LevelError)
	case verbose:
		level.Set(slog.LevelDebug)
	default:
		level.Set(slog.LevelWarn)
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: level}

	if jsonFormat {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	logger = slog.New(handler)
	slog.SetDefault(logger)
}

// SetOutput redirects log output (useful for testing).
func SetOutput(w io.Writer) {
	logger = slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)
}

// L returns the current logger for direct slog use.
func L() *slog.Logger {
	return logger
}

// --- Convenience functions ---

// Debug logs at debug level.
func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

// Info logs at info level.
func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

// Warn logs at warn level.
func Warn(msg string, args ...any) {
	logger.Warn(msg, args...)
}

// Error logs at error level.
func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}

// With returns a logger with additional context fields.
func With(args ...any) *slog.Logger {
	return logger.With(args...)
}
