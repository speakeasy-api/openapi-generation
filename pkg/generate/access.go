package generate

import (
	"context"
	"errors"
	"fmt"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/internal/repofixture"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/licensetoken"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
)

var ErrMissingGenerationAccess = errors.New("generation access state is required; use generation-context/access.WithDirect or WithAuthenticated")

var ErrNoLicenseElection = errors.New(
	"no license elected for generated output: pass --license agpl-3.0-only (or set SPEAKEASY_GENERATED_LICENSE=agpl-3.0-only, or use generation-context/access.ElectAGPL) to accept AGPL-3.0-only licensing, or supply a commercial license token via --license-token, SPEAKEASY_LICENSE_TOKEN, or licensetoken.WithToken")

var ErrUnprovenCommercialLicense = errors.New(
	"commercial license requires a validated license token: attach one with licensetoken.WithToken (or --license-token / SPEAKEASY_LICENSE_TOKEN); authenticated state asserting a commercial license is not proof by itself")

var ErrLicenseTargetNotCovered = errors.New(
	"the license token does not cover this target: elect AGPL-3.0-only for it or obtain a token that covers it")

var resolveLicenseToken = licensetoken.ResolveGenerationAccess

func resolveGenerationAccess(ctx context.Context, target string, requireCommercialProof bool) (context.Context, error) {
	ctx, verification, err := resolveLicenseToken(ctx)
	if err != nil {
		return ctx, err
	}

	state, ok := generationaccess.StateFromContext(ctx)
	if !ok {
		return ctx, ErrMissingGenerationAccess
	}
	if requireCommercialProof && state.GeneratedLicense() == generationaccess.GeneratedLicenseCommercial {
		if verification == nil {
			return ctx, ErrUnprovenCommercialLicense
		}
		coverageTarget := target
		if resolved, err := GetTargetFromTargetString(target); err == nil {
			coverageTarget = resolved.Target
		}
		if !verification.Covers(coverageTarget) {
			return ctx, fmt.Errorf("%w: target %q, token covers %s", ErrLicenseTargetNotCovered, coverageTarget, verification.Targets)
		}
	}

	return ctx, nil
}

func (g *Generator) establishLicense(ctx context.Context, target, outDir string) (context.Context, error) {
	ctx, err := resolveGenerationAccess(ctx, target, true)
	if err != nil {
		return ctx, err
	}
	state, _ := generationaccess.StateFromContext(ctx)
	if state.Mode() == generationaccess.ModeDirect && state.GeneratedLicense() == "" {
		return ctx, ErrNoLicenseElection
	}
	g.generatedLicense = state.GeneratedLicense()
	if isFixtureOutput(g.log, outDir) {
		g.generatedLicense = ""
		options := []generationaccess.DirectOption{generationaccess.ElectAGPL()}
		if state.TelemetryDisabled() {
			options = append(options, generationaccess.DisableTelemetry())
		}
		ctx = generationaccess.WithDirect(ctx, options...)
	}
	g.licenseResolved = true
	return ctx, nil
}

func isFixtureOutput(log logging.Logger, outDir string) bool {
	if outDir == "" {
		return false
	}
	fixture, err := repofixture.DetectFixtureOutput(outDir)
	if err != nil {
		log.Warn("could not determine whether the output directory is a repository fixture: " + err.Error())
		return false
	}
	if fixture {
		log.Info("output directory is a repository fixture: license notices are omitted")
	}
	return fixture
}
