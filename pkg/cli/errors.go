package cli

import (
	"fmt"
)

// ErrSilentExit represents an error that causes the application to exit silently with a specific exit code.
type ErrSilentExit struct {
	Code int // Code is the exit code to be used when exiting.
}

// Error implements the error interface for ErrSilentExit.
// It returns a formatted string containing the exit code.
func (e *ErrSilentExit) Error() string {
	return fmt.Sprintf("Exit Code %d", e.Code)
}
