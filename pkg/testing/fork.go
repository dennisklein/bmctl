package testing

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// Run a fork test that may crash or exit.
func RunForkTest(testName string) (string, string, error) {
	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%v", testName))
	cmd.Env = append(os.Environ(), "FORK=1")

	var stdoutChild, stderrChild bytes.Buffer
	cmd.Stdout = &stdoutChild
	cmd.Stderr = &stderrChild

	err := cmd.Run()

	return stdoutChild.String(), stderrChild.String(), err
}
