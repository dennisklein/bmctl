package logging

import (
	"log/slog"
	"os"

	"go.abhg.dev/log/silog"
)

var defaultLogger *slog.Logger

func NewLogger(lvl slog.Level) *slog.Logger {
	handler := silog.NewHandler(
		os.Stderr,
		&silog.HandlerOptions{Level: lvl, TimeFormat: "15:04:05.000"},
	)
	return slog.New(handler)
}

func Default() *slog.Logger {
	if defaultLogger == nil {
		defaultLogger = NewLogger(slog.LevelInfo)
	}
	return defaultLogger
}
