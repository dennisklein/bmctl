// SPDX-FileCopyrightText: 2025 GSI Helmholtzzentrum für Schwerionenforschung GmbH <https://www.gsi.de/en/>
//
// SPDX-License-Identifier: LGPL-3.0-only

package main

import (
	"log/slog"
	"os"

	"github.com/GSI-HPC/bmctl/pkg/cli"
	_logging "github.com/GSI-HPC/bmctl/pkg/logging"
	"github.com/spf13/cobra"
)

var showDebug = false

func logLevel() slog.Level {
	if showDebug {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}

func setupLogging(cmd *cobra.Command, args []string) {
	opts := &slog.HandlerOptions{Level: logLevel()}
	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)
	ctx := _logging.WithLogger(cmd.Context(), logger)
	parent := cmd
	for parent != nil {
		parent.SetContext(ctx)
		parent = parent.Parent()
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "bmctl",
		Short:            "Out-of-band datacenter device management via the BMC interface",
		Long:             ``,
		PersistentPreRun: setupLogging,
	}
	cmd.PersistentFlags().BoolVarP(&showDebug, "debug", "d", false, "show debug logs")
	return cmd
}

func main() {
	ctx := cli.SignalContext()

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newVersionCmd())

	os.Exit(cli.Execute(ctx, rootCmd))
}
