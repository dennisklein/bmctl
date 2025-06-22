package testing

import (
	"bytes"
	"io"
	"os"
	"syscall"
)

func assertNil(err error) {
	if err != nil {
		panic(err)
	}
}

// Capture redirects the provided file's file descriptor (typically os.Stderr)
// to a pipe, allowing all output written to that file descriptor to be captured.
// It returns a function that, when called, restores the original file descriptor,
// closes the pipe, and returns the captured output as a string. This is useful
// for capturing and inspecting output during tests.
func Capture(file *os.File) func() string {
	stderrFd := int(file.Fd())
	movedStderrFd, err := syscall.Dup(stderrFd)
	assertNil(err)
	read, write, err := os.Pipe()
	assertNil(err)
	err = syscall.Dup2(int(write.Fd()), stderrFd)
	assertNil(err)

	return func() string {
		err = write.Close()
		assertNil(err)
		err = syscall.Dup2(movedStderrFd, stderrFd)
		assertNil(err)
		err = syscall.Close(movedStderrFd)
		assertNil(err)

		var buf bytes.Buffer
		_, err = io.Copy(&buf, read)
		assertNil(err)
		err = read.Close()
		assertNil(err)

		return buf.String()
	}
}
