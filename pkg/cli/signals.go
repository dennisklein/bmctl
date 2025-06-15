// SPDX-FileCopyrightText: 2025 GSI Helmholtzzentrum für Schwerionenforschung GmbH <https://www.gsi.de/en/>
//
// SPDX-License-Identifier: LGPL-3.0-only

package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"golang.org/x/sys/unix"
)

var terminationSignals = []os.Signal{unix.SIGTERM, unix.SIGINT}

// SignalContext returns a context that is canceled when a termination signal is received.
// It listens for SIGTERM and SIGINT. On the first signal, the context is canceled with
// an error describing the signal. If a second signal is received, the process exits immediately.
// This allows for graceful shutdown on the first signal and forced termination on the second.
func SignalContext() context.Context {
	const limit = 2
	signals := make(chan os.Signal, limit)
	signal.Notify(signals, terminationSignals...)

	ctx, cancel := context.WithCancelCause(context.Background())

	go func() {
		retries := 0
		for range signals {
			retries++
			err := fmt.Errorf("got %d interrupts, forcing shutdown", retries)
			cancel(err)
			if retries >= limit {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(EXIT_FAILURE)
			}
		}
	}()

	return ctx
}
