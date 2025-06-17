// SPDX-FileCopyrightText: 2025 GSI Helmholtzzentrum für Schwerionenforschung GmbH <https://www.gsi.de/en/>
//
// SPDX-License-Identifier: LGPL-3.0-or-later

package logging

import (
	"log/slog"
	"os"

	"github.com/golang-cz/devslog"
)

var defaultLogger *slog.Logger

func NewLogger(lvl slog.Level) *slog.Logger {
	slogOpts := &slog.HandlerOptions{
		AddSource: true,
		Level:     lvl,
	}
	opts := &devslog.Options{
		HandlerOptions:    slogOpts,
		MaxSlicePrintSize: 4,
		SortKeys:          true,
		NewLineAfterLog:   true,
		StringerFormatter: true,
	}
	return slog.New(devslog.NewHandler(os.Stderr, opts))
}

func Default() *slog.Logger {
  if defaultLogger == nil {
    defaultLogger = NewLogger(slog.LevelInfo)
  }
	return defaultLogger
}
