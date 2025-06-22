package testing

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_RunForkTest_Success(t *testing.T) {
	if os.Getenv("FORK") == "1" {
		_, err := os.Stdout.WriteString("forked stdout\n")
		require.NoError(t, err)
		_, err = os.Stderr.WriteString("forked stderr\n")
		require.NoError(t, err)
		os.Exit(0)
	}

	stdout, stderr, err := RunForkTest("Test_RunForkTest_Success")
	require.NoError(t, err)
	assert.Contains(t, stdout, "forked stdout")
	assert.Contains(t, stderr, "forked stderr")
}

func Test_RunForkTest_Failure(t *testing.T) {
	if os.Getenv("FORK") == "1" {
		_, err := os.Stderr.WriteString("forked error\n")
		require.NoError(t, err)
		os.Exit(1) // Exit with error code
	}

	stdout, stderr, err := RunForkTest("Test_RunForkTest_Failure")
	require.Error(t, err) // Should have error due to exit code 1
	assert.Contains(t, stderr, "forked error")
	assert.Empty(t, stdout)
}
