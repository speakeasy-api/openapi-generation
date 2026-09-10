package analytics

import (
	"context"
	"testing"

	"github.com/posthog/posthog-go"
	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/stretchr/testify/assert"
)

type recordingClient struct {
	captures int
}

func (c *recordingClient) Enqueue(posthog.Message) error {
	c.captures++
	return nil
}

func (*recordingClient) Close() error {
	return nil
}

func TestRecordUsageRequiresTelemetryEligibleAccessState(t *testing.T) {
	t.Setenv("SPEAKEASY_DISABLE_TELEMETRY", "")

	testCases := []struct {
		name     string
		direct   bool
		disabled bool
		want     int
	}{
		{name: "missing state", want: 0},
		{name: "direct state", direct: true, want: 1},
		{name: "direct telemetry opt out", direct: true, disabled: true, want: 0},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.direct {
				if tt.disabled {
					ctx = generationaccess.WithDirect(ctx, generationaccess.DisableTelemetry(), generationaccess.ElectAGPL())
				} else {
					ctx = generationaccess.WithDirect(ctx, generationaccess.ElectAGPL())
				}
			}

			client := &recordingClient{}

			RecordUsage(ctx, "generate", nil, nil, &Data{}, Config{PostHog: client})

			assert.Equal(t, tt.want, client.captures)
		})
	}
}
