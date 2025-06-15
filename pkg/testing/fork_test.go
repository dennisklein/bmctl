package testing

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_RunForkTest_Success(t *testing.T) {
	if os.Getenv("FORK") == "1" {
		os.Stdout.WriteString("forked stdout\n")
		os.Stderr.WriteString("forked stderr\n")
		os.Exit(0)
	}

	stdout, stderr, err := RunForkTest("Test_RunForkTest_Success")
	require.NoError(t, err)
	assert.Contains(t, stdout, "forked stdout")
	assert.Contains(t, stderr, "forked stderr")
}
