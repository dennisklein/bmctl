package main

import (
	"os"
	"regexp"
	"testing"

	_testing "github.com/GSI-HPC/bmctl/pkg/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_versionCmd(t *testing.T) {
	getStdout := _testing.Capture(os.Stdout)
	cmd := newVersionCmd()
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Regexp(t, regexp.MustCompile(`(\(devel\))|(v[0-9]+\.[0-9]+\.[0-9]+)`), getStdout())
}
