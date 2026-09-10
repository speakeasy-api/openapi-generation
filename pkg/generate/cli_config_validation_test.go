package generate

import (
	"context"
	"testing"

	pkgerrors "github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConfigCLIReleaseDistributionValidation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		cfg    map[string]any
		errMsg string
	}{
		"homebrew requires generateRelease": {
			cfg: map[string]any{
				"packageName":     "github.com/example/petstore-cli",
				"generateRelease": false,
				"distribution": map[string]any{
					"homebrew": map[string]any{
						"enabled": true,
						"tap":     "example/homebrew-petstore",
					},
				},
			},
			errMsg: "distribution.homebrew.enabled requires generateRelease to be true",
		},
		"homebrew requires tap": {
			cfg: map[string]any{
				"packageName":     "github.com/example/petstore-cli",
				"generateRelease": true,
				"distribution": map[string]any{
					"homebrew": map[string]any{
						"enabled": true,
					},
				},
			},
			errMsg: "distribution.homebrew.tap is required when distribution.homebrew.enabled is true",
		},
		"homebrew requires github repo": {
			cfg: map[string]any{
				"packageName":     "example/petstore-cli",
				"generateRelease": true,
				"distribution": map[string]any{
					"homebrew": map[string]any{
						"enabled": true,
						"tap":     "example/homebrew-petstore",
					},
				},
			},
			errMsg: "distribution.homebrew.enabled requires repoURL or packageName to resolve to a GitHub repository",
		},
		"homebrew requires owner repo format": {
			cfg: map[string]any{
				"packageName":     "github.com/example/petstore-cli",
				"generateRelease": true,
				"distribution": map[string]any{
					"homebrew": map[string]any{
						"enabled": true,
						"tap":     "example/homebrew-petstore/extra",
					},
				},
			},
			errMsg: "distribution.homebrew.tap must use the format owner/repo",
		},
		"homebrew requires prefixed repository": {
			cfg: map[string]any{
				"packageName":     "github.com/example/petstore-cli",
				"generateRelease": true,
				"distribution": map[string]any{
					"homebrew": map[string]any{
						"enabled": true,
						"tap":     "example/petstore",
					},
				},
			},
			errMsg: "distribution.homebrew.tap repository name must start with 'homebrew-'",
		},
		"winget requires generateRelease": {
			cfg: map[string]any{
				"packageName":     "github.com/example/petstore-cli",
				"generateRelease": false,
				"distribution": map[string]any{
					"winget": map[string]any{
						"enabled":           true,
						"publisher":         "Example Inc.",
						"publisherUrl":      "https://example.com",
						"repositoryOwner":   "example",
						"packageIdentifier": "Example.Petstore",
						"license":           "MIT",
					},
				},
			},
			errMsg: "distribution.winget.enabled requires generateRelease to be true",
		},
		"winget requires publisher": {
			cfg: map[string]any{
				"packageName":     "github.com/example/petstore-cli",
				"generateRelease": true,
				"distribution": map[string]any{
					"winget": map[string]any{
						"enabled":           true,
						"publisherUrl":      "https://example.com",
						"repositoryOwner":   "example",
						"packageIdentifier": "Example.Petstore",
						"license":           "MIT",
					},
				},
			},
			errMsg: "distribution.winget.publisher is required when distribution.winget.enabled is true",
		},
		"winget requires repository owner": {
			cfg: map[string]any{
				"packageName":     "github.com/example/petstore-cli",
				"generateRelease": true,
				"distribution": map[string]any{
					"winget": map[string]any{
						"enabled":           true,
						"publisher":         "Example Inc.",
						"publisherUrl":      "https://example.com",
						"packageIdentifier": "Example.Petstore",
						"license":           "MIT",
					},
				},
			},
			errMsg: "distribution.winget.repositoryOwner is required when distribution.winget.enabled is true",
		},
		"winget requires package identifier": {
			cfg: map[string]any{
				"packageName":     "github.com/example/petstore-cli",
				"generateRelease": true,
				"distribution": map[string]any{
					"winget": map[string]any{
						"enabled":         true,
						"publisher":       "Example Inc.",
						"publisherUrl":    "https://example.com",
						"repositoryOwner": "example",
						"license":         "MIT",
					},
				},
			},
			errMsg: "distribution.winget.packageIdentifier is required when distribution.winget.enabled is true",
		},
		"winget requires publisher url": {
			cfg: map[string]any{
				"packageName":     "github.com/example/petstore-cli",
				"generateRelease": true,
				"distribution": map[string]any{
					"winget": map[string]any{
						"enabled":           true,
						"publisher":         "Example Inc.",
						"repositoryOwner":   "example",
						"packageIdentifier": "Example.Petstore",
						"license":           "MIT",
					},
				},
			},
			errMsg: "distribution.winget.publisherUrl is required when distribution.winget.enabled is true",
		},
		"winget requires license": {
			cfg: map[string]any{
				"packageName":     "github.com/example/petstore-cli",
				"generateRelease": true,
				"distribution": map[string]any{
					"winget": map[string]any{
						"enabled":           true,
						"publisher":         "Example Inc.",
						"publisherUrl":      "https://example.com",
						"repositoryOwner":   "example",
						"packageIdentifier": "Example.Petstore",
					},
				},
			},
			errMsg: "distribution.winget.license is required when distribution.winget.enabled is true",
		},
		"winget requires valid publisher url": {
			cfg: map[string]any{
				"packageName":     "github.com/example/petstore-cli",
				"generateRelease": true,
				"distribution": map[string]any{
					"winget": map[string]any{
						"enabled":           true,
						"publisher":         "Example Inc.",
						"publisherUrl":      "not-a-url",
						"repositoryOwner":   "example",
						"packageIdentifier": "Example.Petstore",
						"license":           "MIT",
					},
				},
			},
			errMsg: "distribution.winget.publisherUrl must be an absolute http or https URL",
		},
		"winget requires valid repository owner": {
			cfg: map[string]any{
				"packageName":     "github.com/example/petstore-cli",
				"generateRelease": true,
				"distribution": map[string]any{
					"winget": map[string]any{
						"enabled":           true,
						"publisher":         "Example Inc.",
						"publisherUrl":      "https://example.com",
						"repositoryOwner":   "example/org",
						"packageIdentifier": "Example.Petstore",
						"license":           "MIT",
					},
				},
			},
			errMsg: "distribution.winget.repositoryOwner must be a valid GitHub owner name",
		},
		"winget requires github repo": {
			cfg: map[string]any{
				"packageName":     "example/petstore-cli",
				"generateRelease": true,
				"distribution": map[string]any{
					"winget": map[string]any{
						"enabled":           true,
						"publisher":         "Example Inc.",
						"publisherUrl":      "https://example.com",
						"repositoryOwner":   "example",
						"packageIdentifier": "Example.Petstore",
						"license":           "MIT",
					},
				},
			},
			errMsg: "distribution.winget.enabled requires repoURL or packageName to resolve to a GitHub repository",
		},
		"nfpm requires generateRelease": {
			cfg: map[string]any{
				"packageName":     "github.com/example/petstore-cli",
				"generateRelease": false,
				"distribution": map[string]any{
					"nfpm": map[string]any{
						"enabled":    true,
						"formats":    "deb,rpm",
						"maintainer": "Example Inc. <dev@example.com>",
						"license":    "MIT",
					},
				},
			},
			errMsg: "distribution.nfpm.enabled requires generateRelease to be true",
		},
		"nfpm requires maintainer": {
			cfg: map[string]any{
				"packageName":     "github.com/example/petstore-cli",
				"generateRelease": true,
				"distribution": map[string]any{
					"nfpm": map[string]any{
						"enabled": true,
					},
				},
			},
			errMsg: "distribution.nfpm.maintainer is required when distribution.nfpm.enabled is true",
		},
		"nfpm requires license": {
			cfg: map[string]any{
				"packageName":     "github.com/example/petstore-cli",
				"generateRelease": true,
				"distribution": map[string]any{
					"nfpm": map[string]any{
						"enabled":    true,
						"formats":    "deb,rpm",
						"maintainer": "Example Inc. <dev@example.com>",
					},
				},
			},
			errMsg: "distribution.nfpm.license is required when distribution.nfpm.enabled is true",
		},
		"nfpm requires formats": {
			cfg: map[string]any{
				"packageName":     "github.com/example/petstore-cli",
				"generateRelease": true,
				"distribution": map[string]any{
					"nfpm": map[string]any{
						"enabled":    true,
						"maintainer": "Example Inc. <dev@example.com>",
						"license":    "MIT",
					},
				},
			},
			errMsg: "distribution.nfpm.formats is required when distribution.nfpm.enabled is true",
		},
		"nfpm rejects unsupported format": {
			cfg: map[string]any{
				"packageName":     "github.com/example/petstore-cli",
				"generateRelease": true,
				"distribution": map[string]any{
					"nfpm": map[string]any{
						"enabled":    true,
						"formats":    "deb,msi",
						"maintainer": "Example Inc. <dev@example.com>",
						"license":    "MIT",
					},
				},
			},
			errMsg: "distribution.nfpm.formats contains unsupported format 'msi'. Allowed values: deb, rpm, apk",
		},
		"nfpm rejects whitespace format": {
			cfg: map[string]any{
				"packageName":     "github.com/example/petstore-cli",
				"generateRelease": true,
				"distribution": map[string]any{
					"nfpm": map[string]any{
						"enabled":    true,
						"formats":    " ",
						"maintainer": "Example Inc. <dev@example.com>",
						"license":    "MIT",
					},
				},
			},
			errMsg: "distribution.nfpm.formats is required when distribution.nfpm.enabled is true",
		},
		"nfpm ignores empty segments and rejects remaining invalid format": {
			cfg: map[string]any{
				"packageName":     "github.com/example/petstore-cli",
				"generateRelease": true,
				"distribution": map[string]any{
					"nfpm": map[string]any{
						"enabled":    true,
						"formats":    ",deb,msi,",
						"maintainer": "Example Inc. <dev@example.com>",
						"license":    "MIT",
					},
				},
			},
			errMsg: "distribution.nfpm.formats contains unsupported format 'msi'. Allowed values: deb, rpm, apk",
		},
	}

	g, err := New()
	require.NoError(t, err)

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := g.validateConfig(context.Background(), newCLITestConfig(tc.cfg), "cli")
			require.Error(t, err)

			var validationErr *pkgerrors.ValidationError
			require.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "validation error: gen.yaml validation error for cli: "+tc.errMsg, err.Error())
		})
	}
}

func TestValidateConfigCLIReleaseDistributionValidationSuccess(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]any{
		"homebrew valid": {
			"packageName":     "github.com/example/petstore-cli",
			"generateRelease": true,
			"distribution": map[string]any{
				"homebrew": map[string]any{
					"enabled": true,
					"tap":     "example/homebrew-petstore",
				},
			},
		},
		"winget valid": {
			"packageName":     "github.com/example/petstore-cli",
			"generateRelease": true,
			"distribution": map[string]any{
				"winget": map[string]any{
					"enabled":           true,
					"publisher":         "Example Inc.",
					"publisherUrl":      "https://example.com",
					"repositoryOwner":   "example",
					"packageIdentifier": "Example.Petstore",
					"license":           "MIT",
				},
			},
		},
		"nfpm valid": {
			"packageName":     "github.com/example/petstore-cli",
			"generateRelease": true,
			"distribution": map[string]any{
				"nfpm": map[string]any{
					"enabled":    true,
					"formats":    "deb,rpm",
					"maintainer": "Example Inc. <dev@example.com>",
					"license":    "MIT",
				},
			},
		},
	}

	g, err := New()
	require.NoError(t, err)

	for name, cfg := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.NoError(t, g.validateConfig(context.Background(), newCLITestConfig(cfg), "cli"))
		})
	}
}

func newCLITestConfig(cliCfg map[string]any) config.Config {
	baseConfig := &config.Configuration{
		Languages: map[string]config.LanguageConfig{
			"cli": {
				Cfg: cliCfg,
			},
		},
	}

	return config.Config{Config: baseConfig}
}
