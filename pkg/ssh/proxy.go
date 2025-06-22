package ssh

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os/exec"
	"strconv"
	"time"

	_logging "github.com/GSI-HPC/bmctl/pkg/logging"
	"github.com/cenkalti/backoff/v5"
	"golang.org/x/net/proxy"
)

// ProxyCloser is a function that closes the SSH SOCKS proxy and waits for it to shut down completely.
type ProxyCloser = func()

// ErrGracefulShutdown is returned when the SSH SOCKS proxy is shut down gracefully.
var ErrGracefulShutdown = errors.New("graceful shutdown")

// findAvailablePort finds an available port for the SOCKS proxy.
func findAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("failed to find available port: %w", err)
	}

	defer listener.Close() //nolint:errcheck // Ignore error - this is cleanup

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0, errors.New("failed to get TCP address")
	}

	return addr.Port, nil
}

const (
	// Default SSH connection settings.
	defaultServerAliveInterval = 30
	defaultServerAliveCountMax = 3
)

// Options contains configuration options for SSH connections.
type Options struct {
	ServerAliveInterval int
	ServerAliveCountMax int
}

// DefaultOptions returns default SSH connection options.
func DefaultOptions() Options {
	return Options{
		ServerAliveInterval: defaultServerAliveInterval,
		ServerAliveCountMax: defaultServerAliveCountMax,
	}
}

// createSSHCommand creates the SSH command for SOCKS proxy with appropriate flags.
func createSSHCommand(ctx context.Context, host string, port int, opts Options) *exec.Cmd {
	return exec.CommandContext( //nolint: gosec // G204: Subprocess launched with variable arguments
		ctx,
		"ssh",
		"-N",
		"-o", "ExitOnForwardFailure=yes",
		"-o", "ServerAliveInterval="+strconv.Itoa(opts.ServerAliveInterval),
		"-o", "ServerAliveCountMax="+strconv.Itoa(opts.ServerAliveCountMax),
		"-D", strconv.Itoa(port),
		host,
	)
}

// monitorSSHProcess monitors the SSH process and logs its status on exit.
func monitorSSHProcess(
	ctx context.Context,
	cmd *exec.Cmd,
	done chan struct{},
	logger *slog.Logger,
) {
	go func() {
		defer close(done)

		err := cmd.Wait()
		if err != nil {
			cause := context.Cause(ctx)
			if !errors.Is(cause, ErrGracefulShutdown) {
				logger.Error(fmt.Sprintf("SOCKS5 proxy exited with an error: %v", err))

				return
			}
		}

		logger.Debug("SOCKS5 proxy exited")
	}()
}

// startOutputLoggers starts goroutines to log stdout and stderr from the SSH command.
func startOutputLoggers(stdout, stderr *bufio.Scanner, cmdLogger *slog.Logger) {
	// Log stdout output
	go func() {
		for stdout.Scan() {
			cmdLogger.Debug(stdout.Text())
		}
	}()
	// Log stderr output
	go func() {
		for stderr.Scan() {
			cmdLogger.Error(stderr.Text())
		}
	}()
}

// runSocksProxy starts an SSH SOCKS5 proxy to the specified host on the given local port.
// It runs ssh with dynamic port forwarding (-D flag) to create a SOCKS proxy.
//
// The function returns a ProxyCloser that can be used to gracefully shut down the proxy.
// If the proxy fails to start, it returns an error.
//
// The proxy monitors its SSH connection with keep-alive settings and logs output from
// both stdout and stderr streams.
func runSocksProxy(
	ctx context.Context,
	host string,
	port int,
) (ProxyCloser, error) {
	return runSocksProxyWithOptions(ctx, host, port, DefaultOptions())
}

func runSocksProxyWithOptions(
	ctx context.Context,
	host string,
	port int,
	opts Options,
) (ProxyCloser, error) {
	logger := _logging.FromContext(ctx)
	ctx, cancel := context.WithCancelCause(ctx)
	done := make(chan struct{})
	closeProxy := func() {
		cancel(ErrGracefulShutdown)
		<-done
	}

	cmd := createSSHCommand(ctx, host, port, opts)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel(nil)

		return nil, fmt.Errorf("could not get stdout: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel(nil)

		return nil, fmt.Errorf("could not get stderr: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		cancel(nil)

		return nil, fmt.Errorf("could not start SOCKS5 proxy: %w", err)
	}

	monitorSSHProcess(ctx, cmd, done, logger)
	logger.Debug(fmt.Sprintf("SOCKS5 proxy running at :%d", port))

	cmdLogger := logger.With(slog.String("proxy", fmt.Sprintf(":%d -> %s", port, host)))
	startOutputLoggers(bufio.NewScanner(stdout), bufio.NewScanner(stderr), cmdLogger)

	return closeProxy, nil
}

// NewProxyDialer creates a SOCKS5 proxy dialer that routes connections through an SSH tunnel.
// If proxyHost is empty, it returns nil values indicating no proxy is needed.
//
// The function starts an SSH SOCKS5 proxy on port 5555 locally (remote end is the given proxyHost
// SSH server), waits for it to become available, and returns a proxy.Dialer that can be used to
// establish connections through the proxy.
//
// The caller must call the returned ProxyCloser when done to shut down the SSH tunnel.
//
// Returns:
//   - proxy.Dialer: The SOCKS5 dialer for establishing connections through the proxy
//   - ProxyCloser: A function to gracefully shut down the proxy
//   - error: Any error encountered during proxy setup
func NewProxyDialer( //nolint:ireturn,funlen // Intentionally returns interface from external package
	ctx context.Context,
	proxyHost string,
) (proxy.Dialer, ProxyCloser, error) {
	if proxyHost == "" {
		return nil, func() {}, nil
	}

	proxyPort, err := findAvailablePort()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find available port: %w", err)
	}

	proxyAddr := net.JoinHostPort("127.0.0.1", strconv.Itoa(proxyPort))

	closeProxy, err := runSocksProxy(ctx, proxyHost, proxyPort)
	if err != nil {
		if closeProxy != nil {
			closeProxy()
		}

		return nil, nil, fmt.Errorf("failed to start SOCKS5 proxy: %w", err)
	}

	// Wait for the SOCKS5 proxy to be available
	var dial net.Dialer

	const initialInterval = 10

	conn, err := backoff.Retry(
		ctx,
		func() (net.Conn, error) {
			return dial.DialContext(ctx, "tcp", proxyAddr)
		},
		backoff.WithBackOff(&backoff.ExponentialBackOff{
			InitialInterval:     initialInterval * time.Millisecond,
			RandomizationFactor: backoff.DefaultRandomizationFactor,
			Multiplier:          backoff.DefaultMultiplier,
			MaxInterval:         1 * time.Second,
		}),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("SOCKS5 proxy not available: %w", err)
	}

	defer func() {
		err := conn.Close()
		if err != nil {
			// Log error but don't fail - this is a cleanup operation
			logger := _logging.FromContext(ctx)
			logger.Debug("failed to close proxy connection", "error", err)
		}
	}()

	// Create the SOCKS5 dialer
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		if closeProxy != nil {
			closeProxy()
		}

		return nil, nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
	}

	return dialer, closeProxy, err
}
