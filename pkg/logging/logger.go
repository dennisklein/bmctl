package logging

import (
	"log/slog"
	"os"
	"sync"

	"go.abhg.dev/log/silog"
)

//nolint:gochecknoglobals // Required for thread-safe singleton pattern
var (
	defaultLogger *slog.Logger
	defaultOnce   sync.Once
)

// NewLogger creates a new structured logger with the specified level.
//
// The logger uses a silog handler that outputs to stderr with colored
// terminal formatting and timestamps in HH:MM:SS.mmm format.
//
// Parameters:
//   - lvl: The minimum log level to output (e.g., slog.LevelDebug, slog.LevelInfo)
//
// Returns a configured *slog.Logger instance.
func NewLogger(lvl slog.Level) *slog.Logger {
	handler := silog.NewHandler(
		os.Stderr,
		&silog.HandlerOptions{Level: lvl, TimeFormat: "15:04:05.000"},
	)

	return slog.New(handler)
}

// Default returns the default logger instance, initialized with Info level.
//
// This function is thread-safe and uses sync.Once to ensure the default
// logger is initialized exactly once, even when called concurrently.
//
// The default logger is created with slog.LevelInfo and cannot be
// reconfigured after initialization.
//
// Returns the singleton *slog.Logger instance.
func Default() *slog.Logger {
	defaultOnce.Do(func() {
		defaultLogger = NewLogger(slog.LevelInfo)
	})

	return defaultLogger
}
