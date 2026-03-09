package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitTracer_EmptyEndpoint(t *testing.T) {
	tp, err := InitTracer(context.Background(), "")
	require.NoError(t, err)
	assert.NotNil(t, tp)
}
