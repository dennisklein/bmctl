package bmc

import (
	"context"
	"crypto/tls"
	stdlibx509 "crypto/x509"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/GSI-HPC/bmctl/pkg/logging"
	"github.com/GSI-HPC/bmctl/pkg/ssh"
	"github.com/siderolabs/crypto/x509"
	"github.com/stmcginnis/gofish"
)

// ClientConfig holds configuration for connecting to a BMC endpoint.
type ClientConfig struct {
	Endpoint url.URL
	User     string
	Password string
	Insecure bool
	SSHProxy string
}

// Client provides methods to interact with a BMC.
// not safe for concurrent use.
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

		return nil, fmt.Errorf("failed to create SSH proxy dialer: %w", err)
	}

	httpClient, err := newHTTPClient(cfg.Insecure, dialer)
	if err != nil {
		closeProxy()

		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	logger := logging.FromContext(ctx).
		With(slog.String("bmc_user", cfg.User), slog.String("bmc_endpoint", cfg.Endpoint.String()))

	gofishCfg := gofish.ClientConfig{
		Endpoint:   cfg.Endpoint.String(),
		Username:   cfg.User,
		Password:   cfg.Password,
		BasicAuth:  false,
		HTTPClient: httpClient,
	}

	gofishClient, err := gofish.ConnectContext(ctx, gofishCfg)
	if err != nil {
		closeProxy()

		return nil, fmt.Errorf("failed to connect to BMC %s: %w", cfg.Endpoint.String(), err)
	}

	logger.Debug("BMC connected")

	return &Client{
		closeProxy: closeProxy,
		gofish:     gofishClient,
		logger:     logger,
	}, nil
}

// fileHandler opens and serves the file given by the file path.
func fileHandler(filePath string, _logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := _logger.With(slog.String("https_client", r.RemoteAddr))

		file, err := os.Open(filePath) //nolint:gosec // G304: Potential file inclusion via variable
		if err != nil {
			logger.Error(
				"failed to open file",
				slog.String("file_path", filePath),
				slog.Any("error", err),
			)
			http.Error(w, "File not found", http.StatusNotFound)

			return
		}
		defer file.Close() //nolint:errcheck

		fileInfo, err := file.Stat()
		if err != nil {
			logger.Error(
				"failed to stat file",
				slog.String("file_path", filePath),
				slog.Any("error", err),
			)
			http.Error(w, "Internal server error", http.StatusInternalServerError)

			return
		}

		logger.Debug("serving file ...", slog.String("file_path", filePath))
		http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
		logger.Debug("... finished serving file", slog.String("file_path", filePath))
	}
}

// Boot performs a BMC initiated (out-of-band) virtual media boot.
func (c *Client) Boot(ctx context.Context, img string) error {
	ca, err := c.generateCertificateAuthority()
	if err != nil {
		return err
	}

	serverKeyPair, err := c.generateServerKeyPair(net.ParseIP("127.0.0.1"), "localhost", ca)
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*serverKeyPair.Certificate},
		MinVersion:   tls.VersionTLS12,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", fileHandler(img, c.logger))

	server := &http.Server{
		Addr:              ":8443",
		Handler:           mux,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: 30 * time.Second,
		ErrorLog:          slog.NewLogLogger(c.logger.Handler(), slog.LevelError),
	}

	go func() {
		c.logger.Debug("HTTPS server started", slog.String("addr", server.Addr))

		err := server.ListenAndServeTLS("", "")
		if err != nil && err != http.ErrServerClosed {
			c.logger.Error("HTTPS server error", slog.Any("error", err))
		}
	}()

	<-ctx.Done()
	err = server.Shutdown(ctx)
	c.logger.Debug("HTTPS server stopped", slog.String("addr", server.Addr))

	return err
}

// Close releases any resources held by the Client.
func (c *Client) Close() {
	if c.gofish != nil {
		c.gofish.Logout()
	}

	if c.logger != nil {
		c.logger.Debug("BMC disconnected")
	}

	if c.closeProxy != nil {
		c.closeProxy()
	}
}

// generateCertificateAuthority generates a self-signed certificate authority.
func (c *Client) generateCertificateAuthority() (*x509.CertificateAuthority, error) {
	certAuth, err := x509.NewSelfSignedCertificateAuthority(
		x509.Organization("bmctl"),
		x509.RSA(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create self-signed CA: %w", err)
	}

	c.logger.Debug(
		"Self-signed CA Certificate generated",
		slog.String("cert_pem", string(certAuth.CrtPEM)),
	)

	return certAuth, nil
}

// generateServerKeyPair generates a server key pair for HTTPS communication.
func (c *Client) generateServerKeyPair(
	ipAddr net.IP,
	dnsName string,
	ca *x509.CertificateAuthority,
) (*x509.KeyPair, error) {
	serverKeyPair, err := x509.NewKeyPair(ca,
		x509.IPAddresses([]net.IP{ipAddr}),
		x509.DNSNames([]string{dnsName}),
		x509.CommonName(dnsName),
		x509.Organization("bmctl"),
		x509.KeyUsage(stdlibx509.KeyUsageDigitalSignature|stdlibx509.KeyUsageKeyEncipherment),
		x509.ExtKeyUsage([]stdlibx509.ExtKeyUsage{
			stdlibx509.ExtKeyUsageServerAuth,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate server certificate: %w", err)
	}

	c.logger.Debug(
		"HTTPS Server Certificate generated",
		slog.String("cert_pem", string(serverKeyPair.CrtPEM)),
	)

	return serverKeyPair, nil
}
