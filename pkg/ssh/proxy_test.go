package ssh

import (
	"net"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_defaultOptions(t *testing.T) {
	opts := DefaultOptions()
	assert.Equal(t, 30, opts.ServerAliveInterval)
	assert.Equal(t, 3, opts.ServerAliveCountMax)
}

func Test_findAvailablePort(t *testing.T) {
	port, err := findAvailablePort()
	require.NoError(t, err)
	assert.Positive(t, port)
	assert.Less(t, port, 65536)

	// Test that the port is actually available
	listener, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
	require.NoError(t, err)

	defer func() {
		require.NoError(t, listener.Close())
	}()
}

func Test_createSSHCommand(t *testing.T) {
	ctx := t.Context()
	host := "test.example.com"
	port := 5555
	opts := Options{
		ServerAliveInterval: 30,
		ServerAliveCountMax: 3,
	}

	cmd := createSSHCommand(ctx, host, port, opts)
	assert.Contains(t, cmd.Path, "ssh")

	expectedArgs := []string{
		"ssh",
		"-N",
		"-o", "ExitOnForwardFailure=yes",
		"-o", "ServerAliveInterval=30",
		"-o", "ServerAliveCountMax=3",
		"-D", "5555",
		"test.example.com",
	}
	assert.Equal(t, expectedArgs, cmd.Args)
}

func Test_createSSHCommand_WithCustomOptions(t *testing.T) {
	ctx := t.Context()
	host := "test.example.com"
	port := 8080
	opts := Options{
		ServerAliveInterval: 60,
		ServerAliveCountMax: 5,
	}

	cmd := createSSHCommand(ctx, host, port, opts)

	expectedArgs := []string{
		"ssh",
		"-N",
		"-o", "ExitOnForwardFailure=yes",
		"-o", "ServerAliveInterval=60",
		"-o", "ServerAliveCountMax=5",
		"-D", "8080",
		"test.example.com",
	}
	assert.Equal(t, expectedArgs, cmd.Args)
}

func Test_NewProxyDialer_EmptyHost(t *testing.T) {
	ctx := t.Context()
	dialer, closer, err := NewProxyDialer(ctx, "")

	require.NoError(t, err)
	assert.Nil(t, dialer)
	assert.Nil(t, closer)
}
