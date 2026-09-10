package generate

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const terraformServerVariableProviderAttributesSpec = `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: https://{region}.example.com/{version}/{server_url}/{constructor}
    variables:
      region:
        default: us
      version:
        default: v1
      server_url:
        default: api
      constructor:
        default: ctor
x-speakeasy-globals:
  parameters:
    - name: tenant
      in: query
      schema:
        type: string
security:
  - apiKey: []
paths:
  /pets/{id}:
    get:
      operationId: getPet
      x-speakeasy-entity-operation: Pet#read
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Pet"
components:
  securitySchemes:
    apiKey:
      type: apiKey
      in: header
      name: X-API-Key
  schemas:
    Pet:
      type: object
      x-speakeasy-entity: Pet
      required: [id, name]
      properties:
        id:
          type: string
        name:
          type: string
`

const terraformServerVariableProviderAttributesGenYAML = `configVersion: 2.0.0
generation:
  sdkClassName: SDK
terraform:
  version: 0.0.1
  packageName: testing
  author: hashicorp
  additionalProviderAttributes:
    httpHeaders: http_headers
`

func generateTerraformServerVariableProviderAttributes(t *testing.T, genYAML string) ([]error, string) {
	t.Helper()

	outputDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(outputDir, ".speakeasy"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(outputDir, ".speakeasy", "gen.yaml"), []byte(genYAML), 0o644))

	generator, err := New()
	require.NoError(t, err)

	errs := generator.Generate(generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL()), []byte(terraformServerVariableProviderAttributesSpec), "openapi.yaml", "terraform", outputDir, false, false)
	if len(errs) > 0 {
		return errs, ""
	}

	provider, err := os.ReadFile(filepath.Join(outputDir, "internal", "provider", "provider.go"))
	require.NoError(t, err)

	return nil, string(provider)
}

func TestGenerate_TerraformServerVariableProviderAttributes(t *testing.T) {
	t.Run("unmapped variables keep last definition wins", func(t *testing.T) {
		errs, provider := generateTerraformServerVariableProviderAttributes(t, terraformServerVariableProviderAttributesGenYAML)
		require.Empty(t, errs)

		assert.Equal(t, 1, strings.Count(provider, `"server_url": schema.StringAttribute{`))
		assert.Contains(t, provider, `"constructor": schema.StringAttribute{`)
		assert.Contains(t, provider, `tfsdk:"constructor"`)
		assert.Contains(t, provider, `serverUrlParams["constructor"] = data.Constructor.ValueString()`)
	})

	t.Run("mapped variable renames provider attribute", func(t *testing.T) {
		errs, provider := generateTerraformServerVariableProviderAttributes(t, terraformServerVariableProviderAttributesGenYAML+`  serverVariableProviderAttributes:
    version: api_version
    absent: unused
`)
		require.Empty(t, errs)

		assert.Contains(t, provider, `tfsdk:"api_version"`)
		assert.Contains(t, provider, `"api_version": schema.StringAttribute{`)
		assert.Contains(t, provider, `serverUrlParams["version"] = data.APIVersion.ValueString()`)
		assert.NotContains(t, provider, `"version": schema.StringAttribute{`)
	})

	t.Run("mapped variable sanitizing to an empty attribute name fails", func(t *testing.T) {
		errs, _ := generateTerraformServerVariableProviderAttributes(t, terraformServerVariableProviderAttributesGenYAML+"  serverVariableProviderAttributes:\n    version: \"()\"\n")
		require.NotEmpty(t, errs)

		messages := make([]string, 0, len(errs))
		for _, err := range errs {
			messages = append(messages, err.Error())
		}
		assert.Contains(t, strings.Join(messages, "\n"), `maps server variable "version" to "()", which sanitizes to an empty provider attribute name`)
	})

	t.Run("mapped variable explicitly set to an empty attribute name fails", func(t *testing.T) {
		errs, _ := generateTerraformServerVariableProviderAttributes(t, terraformServerVariableProviderAttributesGenYAML+"  serverVariableProviderAttributes:\n    version: \"\"\n")
		require.NotEmpty(t, errs)

		messages := make([]string, 0, len(errs))
		for _, err := range errs {
			messages = append(messages, err.Error())
		}
		assert.Contains(t, strings.Join(messages, "\n"), `maps server variable "version" to "", which sanitizes to an empty provider attribute name`)
	})

	t.Run("mapped variable colliding with another provider attribute fails", func(t *testing.T) {
		tests := []struct {
			name     string
			mappings string
			want     string
		}{
			{name: "unmapped server variable", mappings: "version: region", want: `maps server variable "version" to provider attribute "region"`},
			{name: "later unmapped server variable", mappings: "region: version", want: `maps server variable "region" to provider attribute "version"`},
			{name: "server_url", mappings: "version: server_url", want: `maps server variable "version" to provider attribute "server_url"`},
			{name: "security", mappings: "version: api_key", want: `maps server variable "version" to provider attribute "api_key"`},
			{name: "global", mappings: "version: tenant", want: `maps server variable "version" to provider attribute "tenant"`},
			{name: "additional provider attribute", mappings: "version: http_headers", want: `maps server variable "version" to provider attribute "http_headers"`},
			{name: "other mapped server variable", mappings: "version: api_version\n    region: api_version", want: `to provider attribute "api_version", which is already used by another provider attribute`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				errs, _ := generateTerraformServerVariableProviderAttributes(t, terraformServerVariableProviderAttributesGenYAML+"  serverVariableProviderAttributes:\n    "+tt.mappings+"\n")
				require.NotEmpty(t, errs)

				var messages []string
				for _, err := range errs {
					messages = append(messages, err.Error())
				}
				assert.Contains(t, strings.Join(messages, "\n"), tt.want)
			})
		}
	})
}
