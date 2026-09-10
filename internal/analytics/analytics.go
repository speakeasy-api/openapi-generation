package analytics

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"github.com/posthog/posthog-go"
	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
)

const maxValidationEvents = 20

type telemetryClient interface {
	Enqueue(posthog.Message) error
	Close() error
}

type Config struct {
	PostHog         telemetryClient
	CustomerID      string
	WorkspaceID     string
	Language        string
	Template        string
	RunLocation     string
	GenVersion      string
	CLIVersion      string
	FeatureTracking map[string]bool
	ConfigTracking  map[string]any
	GenIgnoreUsed   bool
}

type Data struct {
	ServerURL      string
	SupportEmail   string
	SupportURL     string
	DocTitle       string
	OpenAPIVersion string
}

func RecordUsage(ctx context.Context, eventType string, errs, warnings []error, data *Data, cfg Config) {
	state, ok := generationaccess.StateFromContext(ctx)
	if os.Getenv("SPEAKEASY_DISABLE_TELEMETRY") == "true" || !ok || state.TelemetryDisabled() {
		return
	}

	p := cfg.PostHog

	if p == nil {
		var err error
		p, err = posthog.NewWithConfig("phc_hiYSF5Axu49I1xs4Z5BG8KCI3PGNLM8ERRs7eocmfX9", posthog.Config{
			Endpoint: "https://metrics.speakeasy.com",
		})
		if err != nil {
			logging.From(ctx).Error(fmt.Sprintf("failed to create posthog client: %v", err))
			return
		}
		defer func() {
			_ = p.Close()
		}()
	}

	runID := uuid.New().String()

	distinctID := cfg.CustomerID
	if distinctID == "" {
		distinctID = runID
	}

	props := posthog.NewProperties().
		Set("language", cfg.Language).
		Set("template", cfg.Template).
		Set("success", len(errs) == 0).
		Set("os", runtime.GOOS).
		Set("arch", runtime.GOARCH).
		Set("run_location", cfg.RunLocation).
		Set("server_url", data.ServerURL).
		Set("support_email", data.SupportEmail).
		Set("support_url", data.SupportURL).
		Set("doc_title", data.DocTitle).
		Set("run_id", runID).
		Set("customer_id", cfg.CustomerID).
		Set("workspace_id", cfg.WorkspaceID).
		Set("gen_version", cfg.GenVersion).
		Set("cli_version", cfg.CLIVersion).
		Set("gen_ignore_used", cfg.GenIgnoreUsed)

	if eventType == "generate" {
		for k, v := range cfg.FeatureTracking {
			props.Set("feature_"+k, v)
		}

		for k, v := range cfg.ConfigTracking {
			props.Set("cfg_"+k, v)
		}
	}

	if len(errs) > 0 {
		errStrs := []string{}

		for i, err := range errs {
			if errors.Is(err, errors.ErrOpenAPIV2) {
				data.OpenAPIVersion = "2.0"
			}
			if i > maxValidationEvents {
				continue
			}

			errStrs = append(errStrs, err.Error())
		}

		props.Set("error", strings.Join(errStrs, "\n"))
	}

	props.Set("openapi_version", data.OpenAPIVersion)

	_ = p.Enqueue(posthog.Capture{
		DistinctId: distinctID,
		Event:      "openapi_sdk_gen_" + eventType,
		Properties: props,
	})

	if len(warnings) > maxValidationEvents {
		warnings = warnings[:maxValidationEvents]
	}
	if len(errs) > maxValidationEvents {
		errs = errs[:maxValidationEvents]
	}

	setWarningsOrErrsAndEnqueue(p, warnings, cfg.Language, runID, distinctID)
	setWarningsOrErrsAndEnqueue(p, errs, cfg.Language, runID, distinctID)
}

func setWarningsOrErrsAndEnqueue(p telemetryClient, errs []error, language, runID, distinctID string) {
	if len(errs) == 0 {
		return
	}

	var eventType string
	if vErr := errors.GetValidationErr(errs[0]); vErr != nil {
		if vErr.Severity == errors.SeverityWarn {
			eventType = "warning"
		} else {
			eventType = "error"
		}
	}
	properties := posthog.NewProperties().
		Set("language", language).
		Set("run_id", runID)

	supportedCount := 0
	unsupportedCount := 0
	for _, err := range errs {
		var supportPrefix string
		var errCount int
		if errors.IsUnsupported(err) {
			unsupportedCount++
			supportPrefix = "unsupported"
			errCount = unsupportedCount
		} else {
			supportedCount++
			supportPrefix = "supported"
			errCount = supportedCount
		}
		properties.Set(fmt.Sprintf("%s_%s_%d", supportPrefix, eventType, errCount), err.Error())
	}

	_ = p.Enqueue(posthog.Capture{
		DistinctId: distinctID,
		Event:      fmt.Sprintf("openapi_sdk_gen_%ss", eventType),
		Properties: properties,
	})
}
