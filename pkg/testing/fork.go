package testing

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// RunForkTest runs a fork test that may crash or exit.
// It executes the specified test in a separate process to isolate crashes or exits.
// Returns stdout, stderr, and any error from the forked process.
func RunForkTest(testName string) (string, string, error) {
	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%v", testName)) //nolint:gosec
	cmd.Env = append(os.Environ(), "FORK=1")

	var stdoutChild, stderrChild bytes.Buffer

	cmd.Stdout = &stdoutChild
	cmd.Stderr = &stderrChild

	err := cmd.Run()

	return stdoutChild.String(), stderrChild.String(), err
}
