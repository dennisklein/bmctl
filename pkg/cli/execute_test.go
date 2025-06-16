// SPDX-FileCopyrightText: 2025 GSI Helmholtzzentrum für Schwerionenforschung GmbH <https://www.gsi.de/en/>
//
// SPDX-License-Identifier: LGPL-3.0-or-later

package cli

import (
	"context"
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
	exit := Execute(context.Background(), cmd)
	assert.Equal(t, EXIT_SUCCESS, exit)
	assert.Empty(t, getStderr())
}

func Test_ExecuteSilentExit(t *testing.T) {
	getStderr := _testing.Capture(os.Stderr)
	cmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			return &ErrSilentExit{Code: 42}
		},
	}
	exit := Execute(context.Background(), cmd)
	assert.Equal(t, 42, exit)
	assert.Empty(t, getStderr())
}

// FIXME:
// func Test_ExecuteError(t *testing.T) {
//   ctx := context.Background()
// 	getStderr := _testing.Capture(os.Stderr)
// 	cmd := &cobra.Command{
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			return errors.New("fail")
// 		},
// 	}
// 	exit := Execute(ctx, cmd)
// 	assert.Equal(t, EXIT_FAILURE, exit)
// 	assert.Contains(t, getStderr(), "ERROR fail\n")
// }
