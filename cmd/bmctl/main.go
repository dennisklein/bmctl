// Package main is the entry point for the bmctl command-line tool.
package main

import (
	"errors"
	"fmt"
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

// addBmcClientConfigFlags adds command-line flags for configuring the BMC client.
func addBmcClientConfigFlags(cmd *cobra.Command, cfg *bmc.ClientConfig) {
	cmd.Flags().VarP((*cli.URLFlag)(&cfg.Endpoint), "endpoint", "e", "BMC Endpoint (FQDN or IP)")
	cmd.Flags().StringVarP(&cfg.User, "user", "u", "", "BMC User")
	cmd.Flags().StringVarP(&cfg.Password, "password", "p", "", "BMC Password")
	cmd.Flags().BoolVarP(&cfg.Insecure, "insecure", "k", false, "Ignore validity of BMC TLS Cert")
	cmd.Flags().StringVarP(&cfg.SSHProxy, "ssh-proxy", "J", "", "BMC SSH Proxy")
	cmd.MarkFlagsRequiredTogether("user", "password")

	err := cmd.MarkFlagRequired("endpoint")
	if err != nil {
		panic(err)
	}

	err = cmd.MarkFlagRequired("user")
	if err != nil {
		panic(err)
	}
}

// validateBmcClientConfig validates the BMC client configuration which is initialized by user input.
// Returns an error if any configuration is invalid.
func validateBmcClientConfig(cfg *bmc.ClientConfig) error {
	switch scheme := cfg.Endpoint.Scheme; scheme {
	case "https", "http":
		// valid
	default:
		return errors.New("endpoint must be a valid http(s) URL")
	}

	// Validate string field lengths
	const (
		maxEndpointLength = 253  // RFC 1035 hostname limit
		maxUserLength     = 64   // Common username length limit
		maxPasswordLength = 128  // Reasonable password length limit
		maxSSHProxyLength = 253  // Same as hostname for SSH proxy
	)

	if len(cfg.Endpoint.String()) > maxEndpointLength {
		return fmt.Errorf("endpoint URL too long (max %d characters)", maxEndpointLength)
	}

	if len(cfg.User) > maxUserLength {
		return fmt.Errorf("user too long (max %d characters)", maxUserLength)
	}

	if len(cfg.Password) > maxPasswordLength {
		return fmt.Errorf("password too long (max %d characters)", maxPasswordLength)
	}

	if len(cfg.SSHProxy) > maxSSHProxyLength {
		return fmt.Errorf("ssh-proxy too long (max %d characters)", maxSSHProxyLength)
	}

	return nil
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
