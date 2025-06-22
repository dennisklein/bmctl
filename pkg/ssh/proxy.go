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
	logger := _logging.FromContext(ctx)
	ctx, cancel := context.WithCancelCause(ctx)
	done := make(chan struct{})
	closeProxy := func() {
		cancel(ErrGracefulShutdown)
		<-done
	}
	cmd := exec.CommandContext(
		ctx,
		"ssh",
		"-N",
		"-o", "ExitOnForwardFailure=yes",
		"-o", "ServerAliveInterval=30",
		"-o", "ServerAliveCountMax=3",
		"-D", strconv.Itoa(port),
		host,
	)
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
	// Monitor the SSH process, generate the done signal and log its status on exit
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
	logger.Debug(fmt.Sprintf("SOCKS5 proxy running at :%d", port))

	cmdLogger := logger.With(slog.String("proxy", fmt.Sprintf(":%d -> %s", port, host)))
	// Log stdout output
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			cmdLogger.Debug(scanner.Text())
		}
	}()
	// Log stderr output
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			cmdLogger.Error(scanner.Text())
		}
	}()

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
func NewProxyDialer(ctx context.Context, proxyHost string) (proxy.Dialer, ProxyCloser, error) {
	if proxyHost == "" {
		return nil, nil, nil
	}

	const proxyPort = 5555 // TODO: Implement a range of ports to chose from if ports are already in use
	proxyAddr := net.JoinHostPort("127.0.0.1", strconv.Itoa(proxyPort))

	closeProxy, err := runSocksProxy(ctx, proxyHost, proxyPort)
	if err != nil {
		if closeProxy != nil {
			closeProxy()
		}
		return nil, nil, fmt.Errorf("failed to start SOCKS5 proxy: %w", err)
	}

	// Wait for the SOCKS5 proxy to be available
	var d net.Dialer
	conn, err := backoff.Retry(
		ctx,
		func() (net.Conn, error) {
			return d.DialContext(ctx, "tcp", proxyAddr)
		},
		backoff.WithBackOff(&backoff.ExponentialBackOff{
			InitialInterval:     10 * time.Millisecond,
			RandomizationFactor: backoff.DefaultRandomizationFactor,
			Multiplier:          backoff.DefaultMultiplier,
			MaxInterval:         1 * time.Second,
		}),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("SOCKS5 proxy not available: %w", err)
	}
	defer conn.Close()

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
