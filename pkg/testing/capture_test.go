// SPDX-FileCopyrightText: 2025 GSI Helmholtzzentrum für Schwerionenforschung GmbH <https://www.gsi.de/en/>
//
// SPDX-License-Identifier: LGPL-3.0-or-later

package testing

import (
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
	assert.Contains(t, expected, output)
	assert.NotContains(t, notExpected, output)
}
