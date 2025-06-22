package testing

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Capture_CapturesStderrOutput(t *testing.T) {
	getStderr := Capture(os.Stderr)
	expected := "error message\n"
	notExpected := "42"

	fmt.Fprint(os.Stderr, expected)

	output := getStderr()

	fmt.Fprint(os.Stderr, notExpected)
	assert.Contains(t, output, expected)
	assert.NotContains(t, output, notExpected)
}

func Test_assertNil_WithPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.Contains(t, fmt.Sprintf("%v", r), "test error")
		} else {
			t.Fatal("Expected panic but none occurred")
		}
	}()

	// This should trigger the panic path in assertNil
	assertNil(errors.New("test error"))
}
