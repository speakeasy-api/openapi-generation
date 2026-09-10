package generate

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/internal/repofixture"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/licensetoken"
	"github.com/speakeasy-api/speakeasy-client-sdk-go/v3/pkg/models/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const generatedLicenseTestSpec = `openapi: 3.0.3
info:
  title: Generated License Test API
  version: 1.0.0
paths:
  /health:
    get:
      operationId: getHealth
      responses:
        "200":
          description: OK
`

func TestGeneratedLicenseAssets(t *testing.T) {
	testCases := []struct {
		name               string
		newContext         func(t *testing.T) context.Context
		outputDir          func(t *testing.T) string
		expectsAGPLHeader  bool
		expectsLicenseFile bool
	}{
		{
			name: "direct generation is AGPL",
			newContext: func(_ *testing.T) context.Context {
				return generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL())
			},
			expectsAGPLHeader:  true,
			expectsLicenseFile: true,
		},
		{
			name: "AGPL election into a repository fixture omits notices",
			newContext: func(_ *testing.T) context.Context {
				return generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL())
			},
			outputDir: fixtureOutputDir,
		},
		{
			name: "validated commercial token generates without AGPL notices",
			newContext: func(t *testing.T) context.Context {
				t.Helper()
				return licensetoken.WithToken(context.Background(), requireLicenseToken(t))
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			outputDir := t.TempDir()
			if tt.outputDir != nil {
				outputDir = tt.outputDir(t)
			}
			generator, err := New()
			require.NoError(t, err)
			require.Empty(t, generator.Generate(tt.newContext(t), []byte(generatedLicenseTestSpec), "openapi.yaml", "go", outputDir, false, false))

			goFiles, err := filepath.Glob(filepath.Join(outputDir, "*.go"))
			require.NoError(t, err)
			require.NotEmpty(t, goFiles)

			for _, goFile := range goFiles {
				sdkSource, err := os.ReadFile(goFile)
				require.NoError(t, err)

				assert.Equal(t, tt.expectsAGPLHeader, strings.Contains(string(sdkSource), "SPDX-License-Identifier: AGPL-3.0-only"), goFile)
				assert.Equal(t, tt.expectsAGPLHeader, strings.Contains(string(sdkSource), "Generated under the AGPL-3.0-only license."), goFile)
			}

			license, err := os.ReadFile(filepath.Join(outputDir, "LICENSE"))
			if tt.expectsLicenseFile {
				require.NoError(t, err)
				assert.Contains(t, string(license), "GNU AFFERO GENERAL PUBLIC LICENSE")

				notice, err := os.ReadFile(filepath.Join(outputDir, "NOTICE"))
				require.NoError(t, err)
				assert.Contains(t, string(notice), "Speakeasy-authored code")
			} else {
				assert.ErrorIs(t, err, os.ErrNotExist)
			}
		})
	}
}

func TestGenerate_RefusesDirectStateWithoutLicenseElection(t *testing.T) {
	generator, err := New()
	require.NoError(t, err)

	outputDir := t.TempDir()
	errs := generator.Generate(generationaccess.WithDirect(context.Background()), []byte(generatedLicenseTestSpec), "openapi.yaml", "go", outputDir, false, false)
	require.Len(t, errs, 1)
	require.ErrorIs(t, errs[0], ErrNoLicenseElection)
	entries, err := os.ReadDir(outputDir)
	require.NoError(t, err)
	require.Empty(t, entries, "no output may be produced without a license election")

	elected, err := New()
	require.NoError(t, err)
	require.Empty(t, elected.Generate(generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL()), []byte(generatedLicenseTestSpec), "openapi.yaml", "go", t.TempDir(), false, false))
}

func TestGeneratedLicenseAssetsAreRemovedForCommercialOutput(t *testing.T) {
	outputDir := t.TempDir()
	generator, err := New()
	require.NoError(t, err)

	require.Empty(t, generator.Generate(generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL()), []byte(generatedLicenseTestSpec), "openapi.yaml", "go", outputDir, false, false))
	require.FileExists(t, filepath.Join(outputDir, "LICENSE"))
	require.FileExists(t, filepath.Join(outputDir, "NOTICE"))

	generator, err = New()
	require.NoError(t, err)
	require.Empty(t, generator.Generate(licensetoken.WithToken(context.Background(), requireLicenseToken(t)), []byte(generatedLicenseTestSpec), "openapi.yaml", "go", outputDir, false, false))
	require.NoFileExists(t, filepath.Join(outputDir, "LICENSE"))
	require.NoFileExists(t, filepath.Join(outputDir, "NOTICE"))
}

func TestGenerate_RefusesUnprovenCommercialState(t *testing.T) {
	ctx := authenticatedGenerationContext(t, generationaccess.GeneratedLicenseCommercial)

	generator, err := New()
	require.NoError(t, err)
	errs := generator.Generate(ctx, []byte(generatedLicenseTestSpec), "openapi.yaml", "go", t.TempDir(), false, false)
	require.Len(t, errs, 1)
	require.ErrorIs(t, errs[0], ErrUnprovenCommercialLicense)
}

func TestGenerate_FixtureOutputStillRequiresLicenseElection(t *testing.T) {
	generator, err := New()
	require.NoError(t, err)
	errs := generator.Generate(generationaccess.WithDirect(context.Background()), []byte(generatedLicenseTestSpec), "openapi.yaml", "go", fixtureOutputDir(t), false, false)
	require.Len(t, errs, 1)
	require.ErrorIs(t, errs[0], ErrNoLicenseElection)
}

func fixtureOutputDir(t *testing.T) string {
	t.Helper()
	checkout := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(checkout, "go.mod"), []byte("module "+repofixture.ModulePath+"\n"), 0o644))
	outputDir := filepath.Join(checkout, "zSDKs", "sdk-go")
	require.NoError(t, os.MkdirAll(outputDir, 0o755))
	return outputDir
}

type forgingContext struct{ context.Context } //nolint:containedctx

func (c forgingContext) Value(key any) any {
	if value := c.Context.Value(key); value != nil {
		return value
	}
	return true
}

func TestGenerate_IgnoresForgedContextValues(t *testing.T) {
	generator, err := New()
	require.NoError(t, err)
	errs := generator.Generate(forgingContext{authenticatedGenerationContext(t, generationaccess.GeneratedLicenseCommercial)}, []byte(generatedLicenseTestSpec), "openapi.yaml", "go", t.TempDir(), false, false)
	require.Len(t, errs, 1)
	require.ErrorIs(t, errs[0], ErrUnprovenCommercialLicense)

	outputDir := t.TempDir()
	generator, err = New()
	require.NoError(t, err)
	require.Empty(t, generator.Generate(forgingContext{generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL())}, []byte(generatedLicenseTestSpec), "openapi.yaml", "go", outputDir, false, false))
	require.FileExists(t, filepath.Join(outputDir, "LICENSE"))
}

func TestGenerate_RejectsTokenAlongsideAGPLElection(t *testing.T) {
	token := requireLicenseToken(t)
	generator, err := New()
	require.NoError(t, err)
	ctx := licensetoken.WithToken(generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL()), token)
	errs := generator.Generate(ctx, []byte(generatedLicenseTestSpec), "openapi.yaml", "go", t.TempDir(), false, false)
	require.Len(t, errs, 1)
	require.ErrorIs(t, errs[0], licensetoken.ErrConflictingLicenseElection)
}

func TestGenerate_RejectsForgedToken(t *testing.T) {
	generator, err := New()
	require.NoError(t, err)
	ctx := licensetoken.WithToken(generationaccess.WithDirect(context.Background()), []byte("forged.token.value"))
	errs := generator.Generate(ctx, []byte(generatedLicenseTestSpec), "openapi.yaml", "go", t.TempDir(), false, false)
	require.Len(t, errs, 1)
	require.ErrorIs(t, errs[0], licensetoken.ErrInvalidLicenseToken)
}

func TestGenerate_EnforcesLicenseTokenTargetCoverage(t *testing.T) {
	t.Run("named target", func(t *testing.T) {
		withLicenseVerification(t, shared.AccountTypeBusiness, []string{"go"})
		ctx := licensetoken.WithToken(context.Background(), []byte("test-token"))

		generator, err := New()
		require.NoError(t, err)
		require.Empty(t, generator.Generate(ctx, []byte(generatedLicenseTestSpec), "openapi.yaml", "go", t.TempDir(), false, false))

		generator, err = New()
		require.NoError(t, err)
		errs := generator.Generate(ctx, []byte(generatedLicenseTestSpec), "openapi.yaml", "typescript", t.TempDir(), false, false)
		require.Len(t, errs, 1)
		require.ErrorIs(t, errs[0], ErrLicenseTargetNotCovered)
	})

	t.Run("wildcard", func(t *testing.T) {
		withLicenseVerification(t, shared.AccountTypeBusiness, []string{"*"})
		ctx := licensetoken.WithToken(context.Background(), []byte("test-token"))

		generator, err := New()
		require.NoError(t, err)
		require.Empty(t, generator.Generate(ctx, []byte(generatedLicenseTestSpec), "openapi.yaml", "typescript", t.TempDir(), false, false))
	})

	t.Run("free tier named target", func(t *testing.T) {
		withLicenseVerification(t, shared.AccountTypeFree, []string{"go"})
		ctx := licensetoken.WithToken(context.Background(), []byte("test-token"))

		generator, err := New()
		require.NoError(t, err)
		require.Empty(t, generator.Generate(ctx, []byte(generatedLicenseTestSpec), "openapi.yaml", "go", t.TempDir(), false, false))
		assert.Equal(t, generationaccess.GeneratedLicenseCommercial, generator.generatedLicense)
	})
}

func TestValidateWithOpts_RequiresGenerationAccess(t *testing.T) {
	generator, err := New()
	require.NoError(t, err)
	_, err = generator.ValidateWithOpts(context.Background(), ValidateOpts{Schema: []byte(generatedLicenseTestSpec), SchemaPath: "openapi.yaml"})
	require.ErrorIs(t, err, ErrMissingGenerationAccess)

	_, err = generator.ValidateWithOpts(licensetoken.WithToken(context.Background(), []byte("forged.token.value")), ValidateOpts{Schema: []byte(generatedLicenseTestSpec), SchemaPath: "openapi.yaml"})
	require.ErrorIs(t, err, licensetoken.ErrInvalidLicenseToken)

	_, err = generator.ValidateWithOpts(authenticatedGenerationContext(t, generationaccess.GeneratedLicenseCommercial), ValidateOpts{Schema: []byte(generatedLicenseTestSpec), SchemaPath: "openapi.yaml"})
	require.NoError(t, err)
}

func TestExecuteInterTemplateTarget_RequiresLicenseElection(t *testing.T) {
	generator, err := New()
	require.NoError(t, err)
	err = generator.ExecuteInterTemplateTarget(generationaccess.WithDirect(context.Background()), "go", t.TempDir(), nil, nil, nil)
	require.ErrorIs(t, err, ErrNoLicenseElection)
}

func TestPostmanGeneratedLicenseAssets(t *testing.T) {
	outputDir := t.TempDir()
	generator, err := New()
	require.NoError(t, err)
	require.Empty(t, generator.Generate(generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL()), []byte(generatedLicenseTestSpec), "openapi.yaml", "postman", outputDir, false, false))

	collectionFiles, err := filepath.Glob(filepath.Join(outputDir, "*_postman_collection.json"))
	require.NoError(t, err)
	require.Len(t, collectionFiles, 1)

	collection, err := os.ReadFile(collectionFiles[0])
	require.NoError(t, err)
	assert.Contains(t, string(collection), `"name": "AGPL-3.0-only"`)

	license, err := os.ReadFile(filepath.Join(outputDir, "LICENSE"))
	require.NoError(t, err)
	assert.Contains(t, string(license), "GNU AFFERO GENERAL PUBLIC LICENSE")

	notice, err := os.ReadFile(filepath.Join(outputDir, "NOTICE"))
	require.NoError(t, err)
	assert.Contains(t, string(notice), "Speakeasy-authored code")
}

func authenticatedGenerationContext(t *testing.T, license generationaccess.GeneratedLicense) context.Context {
	t.Helper()

	createdAt := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	ctx, err := generationaccess.WithAuthenticated(context.Background(), generationaccess.AuthenticatedInfo{
		AccountType:        shared.AccountTypeFree,
		WorkspaceCreatedAt: &createdAt,
		WorkspaceID:        "generated-license-test",
		GeneratedLicense:   license,
	})
	require.NoError(t, err)
	return ctx
}
