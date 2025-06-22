package cli

import (
	"fmt"
)

// SilentExitError represents an error that causes the application to exit silently with a specific exit code.
type SilentExitError struct {
	Code int // Code is the exit code to be used when exiting.
}

// Error implements the error interface for SilentExitError.
// It returns a formatted string containing the exit code.
func (e *SilentExitError) Error() string {
	return fmt.Sprintf("Exit Code %d", e.Code)
}
