package generate

import (
	"context"
	"errors"
	"testing"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	generationtelemetry "github.com/speakeasy-api/generation-context/telemetry"
	"github.com/speakeasy-api/speakeasy-client-sdk-go/v3/pkg/models/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateEnrichesSharedTelemetryEvent(t *testing.T) {
	generator, err := New(WithDontWrite())
	require.NoError(t, err)

	ctx := generationaccess.WithDirect(context.Background(), generationaccess.DisableTelemetry(), generationaccess.ElectAGPL())
	var callbackEvent *shared.CliEvent

	err = generationtelemetry.Run(ctx, shared.InteractionTypeTargetGenerate, func(ctx context.Context, event *shared.CliEvent) error {
		callbackEvent = event
		assert.Same(t, event, generationtelemetry.EventFromContext(ctx))

		errs := generator.Generate(ctx, []byte(`openapi: 3.0.3
info:
  title: Telemetry Test
  version: 1.0.0
servers:
  - url: https://example.com
paths:
  /health:
    get:
      operationId: getHealth
      responses:
        "200":
          description: OK
`), "telemetry.yaml", "go", t.TempDir(), false, false)
		require.Empty(t, errs)
		return nil
	})
	require.NoError(t, err)

	require.NotNil(t, callbackEvent)
	require.NotNil(t, callbackEvent.GenerateTarget)
	assert.Equal(t, "go", *callbackEvent.GenerateTarget)
	require.NotNil(t, callbackEvent.LocalCompletedAt)
	require.NotNil(t, callbackEvent.DurationMs)
}

func TestSharedTelemetryLifecycleFinalizesGeneratorCallbackError(t *testing.T) {
	callbackErr := errors.New("generation callback failed")
	ctx := generationaccess.WithDirect(context.Background(), generationaccess.DisableTelemetry(), generationaccess.ElectAGPL())
	var callbackEvent *shared.CliEvent

	err := generationtelemetry.Run(ctx, shared.InteractionTypeTargetGenerate, func(ctx context.Context, event *shared.CliEvent) error {
		callbackEvent = event
		return callbackErr
	})

	require.ErrorIs(t, err, callbackErr)
	require.NotNil(t, callbackEvent)
	assert.False(t, callbackEvent.Success)
	require.NotNil(t, callbackEvent.Error)
	assert.Equal(t, callbackErr.Error(), *callbackEvent.Error)
}
