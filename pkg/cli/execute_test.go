package cli

import (
	"errors"
	"os"
	"testing"

	_testing "github.com/GSI-HPC/bmctl/pkg/testing"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func Test_ExecuteSuccess(t *testing.T) {
	getStderr := _testing.Capture(os.Stderr)
	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {},
	}
	exit := Execute(t.Context(), cmd)
	assert.Equal(t, ExitSuccess, exit)
	assert.Empty(t, getStderr())
}

func Test_ExecuteSilentExit(t *testing.T) {
	getStderr := _testing.Capture(os.Stderr)
	cmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			return &ErrSilentExit{Code: 42}
		},
	}
	exit := Execute(t.Context(), cmd)
	assert.Equal(t, 42, exit)
	assert.Empty(t, getStderr())
}

func Test_ExecuteError(t *testing.T) {
	getStderr := _testing.Capture(os.Stderr)
	cmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("fail")
		},
	}
	exit := Execute(t.Context(), cmd)
	assert.Equal(t, ExitFailure, exit)
	assert.Contains(t, getStderr(), "ERR fail\n")
}
