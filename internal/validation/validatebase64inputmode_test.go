package validation_test

import (
	"context"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const base64SpecHeader = `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /test:
    post:
      operationId: test
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/payload'
      responses:
        '200':
          description: OK
components:
  schemas:
`

func Test_ValidateBase64InputMode_Success(t *testing.T) {
	tests := []struct {
		name   string
		schema string
	}{
		{
			name: "format byte with file mode",
			schema: base64SpecHeader + `    payload:
      type: object
      properties:
        data:
          type: string
          format: byte
          x-speakeasy-base64-input-mode: file`,
		},
		{
			name: "contentEncoding base64 with file mode",
			schema: base64SpecHeader + `    payload:
      type: object
      properties:
        data:
          type: string
          contentEncoding: base64
          x-speakeasy-base64-input-mode: file`,
		},
		{
			name: "extension absent",
			schema: base64SpecHeader + `    payload:
      type: object
      properties:
        data:
          type: string
          format: byte`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(
				&config.Configuration{},
				validation.RulesetSpeakeasyRecommended,
				validation.WithFilteredRules([]string{(&validation.ValidateBase64InputMode{}).ID()}),
			)
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.schema), "", types.NewTargetFromTemplate("pythonv2"))
			assert.Empty(t, res.GetValidationErrors())
		})
	}
}

func Test_ValidateBase64InputMode_Errors(t *testing.T) {
	tests := []struct {
		name     string
		schema   string
		wantErrs []string
	}{
		{
			name: "unsupported mode value",
			schema: base64SpecHeader + `    payload:
      type: object
      properties:
        data:
          type: string
          format: byte
          x-speakeasy-base64-input-mode: bytes`,
			wantErrs: []string{
				`validation warn: [line 28:42] generator-validate-base64-input-mode - x-speakeasy-base64-input-mode value "bytes" is not allowed (supported: file)`,
			},
		},
		{
			name: "non-string schema",
			schema: base64SpecHeader + `    payload:
      type: object
      properties:
        data:
          type: integer
          x-speakeasy-base64-input-mode: file`,
			wantErrs: []string{
				"validation warn: [line 26:11] generator-validate-base64-input-mode - x-speakeasy-base64-input-mode requires a type:string schema; extension will be ignored",
			},
		},
		{
			name: "missing format and contentEncoding",
			schema: base64SpecHeader + `    payload:
      type: object
      properties:
        data:
          type: string
          x-speakeasy-base64-input-mode: file`,
			wantErrs: []string{
				"validation warn: [line 27:42] generator-validate-base64-input-mode - x-speakeasy-base64-input-mode requires format:byte or contentEncoding:base64; extension will be ignored",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(
				&config.Configuration{},
				validation.RulesetSpeakeasyRecommended,
				validation.WithFilteredRules([]string{(&validation.ValidateBase64InputMode{}).ID()}),
			)
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.schema), "", types.NewTargetFromTemplate("pythonv2"))
			errs := res.GetValidationErrors()
			errStrs := make([]string, len(errs))
			for i, e := range errs {
				errStrs[i] = e.Error()
			}
			assert.ElementsMatch(t, tt.wantErrs, errStrs)
		})
	}
}
