// SPDX-FileCopyrightText: 2025 GSI Helmholtzzentrum für Schwerionenforschung GmbH <https://www.gsi.de/en/>
//
// SPDX-License-Identifier: LGPL-3.0-or-later

package main

import (
	"log/slog"
	"os"

	"github.com/GSI-HPC/bmctl/pkg/bmc"
	"github.com/GSI-HPC/bmctl/pkg/cli"
	"github.com/GSI-HPC/bmctl/pkg/logging"
	"github.com/spf13/cobra"
)

var (
	showDebug       bool
	bmcClientConfig bmc.ClientConfig
)

func logLevel() slog.Level {
	if showDebug {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}

func setupLogging(cmd *cobra.Command, args []string) {
	logger := logging.NewLogger(logLevel())
	ctx := logging.WithLogger(cmd.Context(), logger)
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

func addBmcClientConfigFlags(cmd *cobra.Command) {
  cmd.Flags().StringVarP(&bmcClientConfig.Endpoint, "endpoint", "e", "", "BMC Endpoint (DNS Name or IP)")
  cmd.Flags().StringVarP(&bmcClientConfig.User, "user", "u", "", "BMC User")
  cmd.Flags().StringVarP(&bmcClientConfig.Password, "password", "p", "", "BMC Password")
  cmd.Flags().BoolVarP(&bmcClientConfig.Insecure, "insecure", "k", false, "Ignore validity of BMC TLS Certificate")
  cmd.Flags().StringVarP(&bmcClientConfig.SshProxy, "ssh-proxy", "J", "", "BMC SSH Proxy")
	cmd.MarkFlagRequired("endpoint")
	cmd.MarkFlagRequired("user")
	cmd.MarkFlagsRequiredTogether("user", "password")
}

func main() {
	ctx := cli.SignalContext()

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newBootCmd())
	rootCmd.AddCommand(newVersionCmd())

	os.Exit(cli.Execute(ctx, rootCmd))
}
