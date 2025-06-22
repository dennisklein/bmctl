package logging

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_WithLoggerAndFromContext(t *testing.T) {
	ctx := t.Context()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	ctxWithLogger := WithLogger(ctx, logger)
	retrieved := FromContext(ctxWithLogger)
	require.Same(t, logger, retrieved)
}

func Test_FromContextReturnsDefault(t *testing.T) {
	ctx := t.Context()
	retrieved := FromContext(ctx)
	require.Same(t, Default(), retrieved)
}

func Test_WithLoggerIsContextSpecific(t *testing.T) {
	ctx := t.Context()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	ctxWithLogger := WithLogger(ctx, logger)
	ctxWithoutLogger := ctx

	assert.Same(t, logger, FromContext(ctxWithLogger))
	assert.Same(t, Default(), FromContext(ctxWithoutLogger))
}
