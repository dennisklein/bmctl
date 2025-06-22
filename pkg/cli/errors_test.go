package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ErrSilentExit_Error(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{0, "Exit Code 0"},
		{1, "Exit Code 1"},
		{42, "Exit Code 42"},
		{-1, "Exit Code -1"},
	}

	for _, tt := range tests {
		err := &ErrSilentExit{Code: tt.code}
		require.Equal(t, tt.expected, err.Error())
	}
}
