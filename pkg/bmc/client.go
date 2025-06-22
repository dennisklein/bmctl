package bmc

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/GSI-HPC/bmctl/pkg/logging"
	"github.com/GSI-HPC/bmctl/pkg/ssh"
	"github.com/stmcginnis/gofish"
)

// ClientConfig holds configuration for connecting to a BMC endpoint.
type ClientConfig struct {
	Endpoint string
	User     string
	Password string
	Insecure bool
	SSHProxy string
}

// Client provides methods to interact with a BMC.
type Client struct {
	closeProxy ssh.ProxyCloser
	gofish     *gofish.APIClient
	logger     *slog.Logger
}

// NewClient creates a new Client with the given configuration.
func NewClient(ctx context.Context, cfg ClientConfig) (*Client, error) {
	dialer, closeProxy, err := ssh.NewProxyDialer(ctx, cfg.SSHProxy)
	if err != nil {
		closeProxy()

		return nil, err
	}

	httpClient, err := newHTTPClient(cfg.Insecure, dialer)
	if err != nil {
		closeProxy()

		return nil, err
	}

	gofishCfg := gofish.ClientConfig{
		Endpoint:   "https://" + cfg.Endpoint,
		Username:   cfg.User,
		Password:   cfg.Password,
		BasicAuth:  false,
		HTTPClient: httpClient,
	}

	gofishClient, err := gofish.ConnectContext(ctx, gofishCfg)
	if err != nil {
		closeProxy()

		return nil, err
	}

	logger := logging.FromContext(ctx).
		With(slog.String("bmc_client", fmt.Sprintf("%s@%s", cfg.User, cfg.Endpoint)))
	logger.Debug("BMC connected")

	return &Client{
		closeProxy: closeProxy,
		gofish:     gofishClient,
		logger:     logger,
	}, nil
}

// Boot performs a BMC initiated virtual media boot.
func (c *Client) Boot(ctx context.Context, img string) error {
	c.logger.Info(c.gofish.GetService().RedfishVersion)

	return nil
}

// Close releases any resources held by the Client.
func (c *Client) Close() {
	c.gofish.Logout()
	c.logger.Debug("BMC disconnected")
	c.closeProxy()
}
