// SPDX-FileCopyrightText: 2025 GSI Helmholtzzentrum für Schwerionenforschung GmbH <https://www.gsi.de/en/>
//
// SPDX-License-Identifier: LGPL-3.0-only

package testing

import (
	"bytes"
	"io"
	"os"
	"syscall"
)

// Capture redirects the provided file's file descriptor (typically os.Stderr)
// to a pipe, allowing all output written to that file descriptor to be captured.
// It returns a function that, when called, restores the original file descriptor,
// closes the pipe, and returns the captured output as a string. This is useful
// for capturing and inspecting output during tests.
func Capture(file *os.File) func() string {
	stderrFd := int(file.Fd())
	movedStderrFd, _ := syscall.Dup(stderrFd)
	r, w, _ := os.Pipe()
	syscall.Dup2(int(w.Fd()), stderrFd)

	return func() string {
		w.Close()
		syscall.Dup2(movedStderrFd, stderrFd)
		syscall.Close(movedStderrFd)

		var buf bytes.Buffer
		io.Copy(&buf, r)
		r.Close()
		return buf.String()
	}
}
