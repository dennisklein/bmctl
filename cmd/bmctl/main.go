// Package main is the entry point for the bmctl command-line tool.
package main

import (
	"log/slog"
	"os"

	"github.com/GSI-HPC/bmctl/pkg/bmc"
	"github.com/GSI-HPC/bmctl/pkg/cli"
	"github.com/GSI-HPC/bmctl/pkg/logging"
	"github.com/spf13/cobra"
)

func newRootCmd(showDebug *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bmctl",
		Short: "Out-of-band datacenter device management via the BMC interface",
		Long:  ``,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var logger *slog.Logger
			if *showDebug {
				logger = logging.NewLogger(slog.LevelDebug)
			} else {
				logger = logging.NewLogger(slog.LevelInfo)
			}
			ctx := logging.WithLogger(cmd.Context(), logger)
			parent := cmd
			for parent != nil {
				parent.SetContext(ctx)
				parent = parent.Parent()
			}
		},
	}
	cmd.PersistentFlags().BoolVarP(showDebug, "debug", "d", false, "show debug logs")

	return cmd
}

func addBmcClientConfigFlags(cmd *cobra.Command, cfg *bmc.ClientConfig) {
	cmd.Flags().StringVarP(&cfg.Endpoint, "endpoint", "e", "", "BMC Endpoint (DNS Name or IP)")
	cmd.Flags().StringVarP(&cfg.User, "user", "u", "", "BMC User")
	cmd.Flags().StringVarP(&cfg.Password, "password", "p", "", "BMC Password")
	cmd.Flags().BoolVarP(&cfg.Insecure, "insecure", "k", false, "Ignore validity of BMC TLS Cert")
	cmd.Flags().StringVarP(&cfg.SSHProxy, "ssh-proxy", "J", "", "BMC SSH Proxy")
	cmd.MarkFlagsRequiredTogether("user", "password")

	if err := cmd.MarkFlagRequired("endpoint"); err != nil {
		panic(err)
	}

	if err := cmd.MarkFlagRequired("user"); err != nil {
		panic(err)
	}
}

func main() {
	var (
		showDebug       bool
		bmcClientConfig bmc.ClientConfig
	)

	ctx := cli.SignalContext()

	rootCmd := newRootCmd(&showDebug)
	rootCmd.AddCommand(newBootCmd(&bmcClientConfig))
	rootCmd.AddCommand(newVersionCmd())

	os.Exit(cli.Execute(ctx, rootCmd))
}
