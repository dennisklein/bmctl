package logging

import (
	"context"
	"log/slog"
)

// loggerKey is a private type used as a context key to store logger instances.
// Using a private type prevents key collisions with other packages.
type loggerKey struct{}

// WithLogger adds a logger instance to the given context.
//
// This function creates a new context that carries the provided logger,
// enabling context-aware logging throughout the application call chain.
// The logger can be retrieved later using FromContext.
//
// Parameters:
//   - ctx: The parent context to which the logger will be added
//   - logger: The *slog.Logger instance to store in the context
//
// Returns a new context.Context containing the logger.
//
// Example usage:
//
//	logger := logging.NewLogger(slog.LevelDebug)
//	ctx := logging.WithLogger(context.Background(), logger)
//	// Pass ctx to functions that need the logger
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// FromContext retrieves the logger instance from the given context.
//
// This function extracts a logger that was previously stored in the context
// using WithLogger. If no logger is found in the context, it returns the
// default logger instance, ensuring this function never returns nil.
//
// Parameters:
//   - ctx: The context from which to retrieve the logger
//
// Returns the *slog.Logger instance from the context, or the default logger
// if no logger was found in the context.
//
// Example usage:
//
//	logger := logging.FromContext(ctx)
//	logger.Info("Processing request", "id", requestID)
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return logger
	}

	return Default()
}
