// SPDX-FileCopyrightText: 2025 GSI Helmholtzzentrum für Schwerionenforschung GmbH <https://www.gsi.de/en/>
//
// SPDX-License-Identifier: LGPL-3.0-only

package cli

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	_testing "github.com/GSI-HPC/bmctl/pkg/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SignalContext_CancelOnSignal(t *testing.T) {
	ctx := SignalContext()
	p, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	err = p.Signal(syscall.SIGINT)
	require.NoError(t, err)

	select {
	case <-ctx.Done():
		// Success: context cancelled
	case <-time.After(1 * time.Second):
		assert.Fail(t, "context was not cancelled after signal")
	}
}

func Test_SignalContext_ForceShutdown(t *testing.T) {
	if os.Getenv("FORK") == "1" {
		ctx := SignalContext()
		p, err := os.FindProcess(os.Getpid())
		require.NoError(t, err)
		err = p.Signal(syscall.SIGINT)
		require.NoError(t, err)
		time.Sleep(100 * time.Millisecond)
		err = p.Signal(syscall.SIGINT)
		require.NoError(t, err)
		time.Sleep(100 * time.Millisecond)
		<-ctx.Done()
	}

	stdout, stderr, err := _testing.RunForkTest("Test_SignalContext_ForceShutdown")
	exiterr, ok := err.(*exec.ExitError)
	assert.True(t, ok)
	assert.Equal(t, exiterr.ExitCode(), EXIT_FAILURE)
	assert.Contains(t, stderr, "got 2 interrupts, forcing shutdown")
	assert.Empty(t, stdout)
}
