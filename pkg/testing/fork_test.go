// SPDX-FileCopyrightText: 2025 GSI Helmholtzzentrum für Schwerionenforschung GmbH <https://www.gsi.de/en/>
//
// SPDX-License-Identifier: LGPL-3.0-only

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
