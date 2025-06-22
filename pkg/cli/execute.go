package cli

import (
	"context"
	"errors"

	_logging "github.com/GSI-HPC/bmctl/pkg/logging"
	"github.com/spf13/cobra"
)

const (
	// ExitSuccess represents process exit code 0.
	ExitSuccess = 0
	// ExitFailure represents process exit code 1.
	ExitFailure = 1
)

// Execute runs the provided cobra.Command using the given context.
// Disables the cobra facilities for top-level error reporting.
// It returns an exit code based on the command execution result:
//   - ExitSuccess if the command executes without error
//   - The code from ErrSilentExit if that error is returned
//   - ExitFailure for all other errors, after logging them either with
//     the zap logger from the context or fmt as a fallback
func Execute(ctx context.Context, cmd *cobra.Command) int {
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err := cmd.ExecuteContext(ctx)
	if err == nil {
		return ExitSuccess
	}

	var silentExit *ErrSilentExit
	if errors.As(err, &silentExit) {
		return silentExit.Code
	}

	logger := _logging.FromContext(cmd.Context()) //nolint
	logger.Error(err.Error())

	return ExitFailure
}
