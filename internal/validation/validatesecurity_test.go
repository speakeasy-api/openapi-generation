package validation_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var baseSecurityValidationOpenAPIDocTemplate = `openapi: 3.0.3
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123%s
paths:
  /test:
    get:%s
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
components:
  securitySchemes: %s`

var securityValidationOpenAPIDocTemplate = fmt.Sprintf(baseSecurityValidationOpenAPIDocTemplate, `
security:
  - sec: []`, "", `
    sec:
      %s`)

func Test_ValidateSecurity_HTTP_Errors(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "missing security scheme type",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `scheme: basic`),
			},
			wantErrs: []string{
				"validation error: [line 22:7] validation-required-field - `securityScheme.type` is required",
			},
		},
		{
			name: "invalid security scheme type",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: blah
      scheme: basic`),
			},
			wantErrs: []string{
				"validation error: [line 22:13] validation-allowed-values - securityScheme.type must be one of [`apiKey, http, mutualTLS, oauth2, openIdConnect`]",
			},
		},
		{
			name: "missing http security scheme scheme",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: http`),
			},
			wantErrs: []string{
				"validation error: [line 22:7] validation-required-field - `securityScheme.scheme` is required for type=http",
			},
		},
		{
			name: "invalid http security scheme scheme",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: http
      scheme: blah`),
			},
			wantErrs: []string{
				"validation error: [line 23:15] generator-validate-security - http security scheme `sec` has invalid scheme `blah` requires `basic`, `bearer`, or `custom`",
			},
		},
		{
			name: "http security scheme with in header is warning",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: http
      scheme: bearer
      in: header`),
			},
			wantErrs: []string{
				"validation warn: [line 24:11] validation-allowed-values - securityScheme.in is not used for type=http (only valid for type=apiKey)",
			},
		},
		{
			name: "valid http security scheme scheme - basic",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: http
      scheme: basic`),
			},
			wantErrs: []string{},
		},
		{
			name: "valid http security scheme scheme - Basic",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: http
      scheme: Basic`),
			},
			wantErrs: []string{},
		},
		{
			name: "valid http security scheme scheme - bearer",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: http
      scheme: bearer`),
			},
			wantErrs: []string{},
		},
		{
			name: "valid http security scheme scheme - Bearer",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: http
      scheme: Bearer`),
			},
			wantErrs: []string{},
		},
		{
			name: "valid http security scheme scheme - custom",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: http
      scheme: custom`),
			},
			wantErrs: []string{},
		},
		{
			name: "valid http security scheme scheme - Custom",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: http
      scheme: Custom`),
			},
			wantErrs: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateSecurity{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()
			errStrs := make([]string, len(errs))
			for i, err := range errs {
				errStrs[i] = err.Error()
			}

			if tt.name == "security scheme override collision" {
				require.Len(t, errStrs, 1)
				assert.Contains(t, errStrs[0], "x-speakeasy-name-override collision: \"dup\"")
				return
			}

			assert.ElementsMatch(t, tt.wantErrs, errStrs)
		})
	}
}

func Test_ValidateSecurity_APIKey_Errors(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "missing in keyword in apiKey security scheme",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: apiKey
      name: test`),
			},
			wantErrs: []string{
				"validation error: [line 22:7] validation-required-field - `securityScheme.in` is required for type=apiKey",
			},
		},
		{
			name: "invalid in keyword in apiKey security scheme",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: apiKey
      in: blah
      name: test`),
			},
			wantErrs: []string{
				"validation error: [line 23:11] validation-allowed-values - securityScheme.in must be one of [`header, query, cookie`] for type=apiKey",
			},
		},
		{
			name: "apiKey security scheme missing name",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: apiKey
      in: header`),
			},
			wantErrs: []string{
				"validation error: [line 22:7] validation-required-field - `securityScheme.name` is required for type=apiKey",
			},
		},
		{
			name: "apiKey security scheme Authorization header hint",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: apiKey
      in: header
      name: Authorization`),
			},
			wantErrs: []string{
				"validation hint: [line 24:7] generator-validate-security - apiKey security scheme `sec` uses header `Authorization`; did you mean `type=http scheme=bearer`?",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateSecurity{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()
			errStrs := make([]string, len(errs))
			for i, err := range errs {
				errStrs[i] = err.Error()
			}

			assert.ElementsMatch(t, tt.wantErrs, errStrs)
		})
	}
}

func Test_ValidateSecurity_OAuth2_Errors(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "missing flows keyword in oauth2 security scheme",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: oauth2`),
			},
			wantErrs: []string{
				"validation error: [line 22:7] validation-required-field - `securityScheme.flows` is required for type=oauth2",
			},
		},
		{
			name: "oauth2 security scheme missing any flows",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: oauth2
      flows: {}`),
			},
			wantErrs: []string{
				"validation error: [line 23:14] generator-validate-security - oauth2 security scheme `sec` is missing a flow requires `implicit`, `password`, `clientCredentials` or `authorizationCode`",
			},
		},
		{
			name: "implicit flow missing authorizationUrl",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: oauth2
      flows:
        implicit:
          refreshUrl: http://blah.com
          scopes: {}`),
			},
			wantErrs: []string{
				"validation error: [line 25:11] validation-required-field - oAuthFlow.authorizationUrl is required for type=implicit",
			},
		},
		{
			name: "implicit flow empty authorizationUrl",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: oauth2
      flows:
        implicit:
          authorizationUrl: ""
          refreshUrl: http://blah.com
          scopes: {}`),
			},
			wantErrs: []string{
				"validation error: [line 25:29] validation-required-field - oAuthFlow.authorizationUrl is required for type=implicit",
			},
		},
		{
			name: "implicit flow invalid authorizationUrl",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: oauth2
      flows:
        implicit:
          authorizationUrl: http:// blah.
          refreshUrl: http://blah.com
          scopes: {}`),
			},
			wantErrs: []string{
				"validation error: [line 25:29] validation-invalid-format - oAuthFlow.authorizationUrl is not a valid uri: parse \"http:// blah.\": invalid character \" \" in host name",
			},
		},
		{
			name: "implicit flow invalid refreshUrl",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: oauth2
      flows:
        implicit:
          authorizationUrl: http://blah.com
          refreshUrl: http:// blah.
          scopes: {}`),
			},
			wantErrs: []string{
				"validation error: [line 26:23] validation-invalid-format - oAuthFlow.refreshUrl is not a valid uri: parse \"http:// blah.\": invalid character \" \" in host name",
			},
		},
		{
			name: "implicit flow missing scopes",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: oauth2
      flows:
        implicit:
          authorizationUrl: http://blah.com
          refreshUrl: http://blah.com`),
			},
			wantErrs: []string{
				"validation error: [line 25:11] validation-required-field - oAuthFlow.scopes is required (empty map is allowed)",
			},
		},
		{
			name: "password flow invalid",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: oauth2
      flows:
        password: {}`),
			},
			wantErrs: []string{
				"validation error: [line 24:19] validation-required-field - oAuthFlow.scopes is required (empty map is allowed)",
				"validation warn: [line 24:19] validation-required-field - oAuthFlow.tokenUrl is required for type=password",
			},
		},
		{
			name: "clientCredentials flow invalid",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: oauth2
      flows:
        clientCredentials: {}`),
			},
			wantErrs: []string{
				"validation error: [line 24:28] validation-required-field - oAuthFlow.scopes is required (empty map is allowed)",
				"validation warn: [line 24:28] validation-required-field - oAuthFlow.tokenUrl is required for type=clientCredentials",
			},
		},
		{
			name: "authorizationCode flow invalid",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: oauth2
      flows:
        authorizationCode: {}`),
			},
			wantErrs: []string{
				"validation error: [line 24:28] validation-required-field - oAuthFlow.authorizationUrl is required for type=authorizationCode",
				"validation error: [line 24:28] validation-required-field - oAuthFlow.scopes is required (empty map is allowed)",
				"validation warn: [line 24:28] validation-required-field - oAuthFlow.tokenUrl is required for type=authorizationCode",
			},
		},
		{
			name: "x-speakeasy-overridable-scopes extension only supported for clientCredentials",
			args: args{
				schema: fmt.Sprintf(baseSecurityValidationOpenAPIDocTemplate, `
security:
  - oauth2Password: [read, write]
  - clientCredentials: [read, write]`, "", `
    oauth2Password:
      type: oauth2
      flows:
        password:
          tokenUrl: http://localhost:35456/oauth2/token
          x-speakeasy-overridable-scopes: true
          scopes:
            read: Read access
            write: Write access
    clientCredentials:
      type: oauth2
      flows:
        clientCredentials:
          tokenUrl: http://localhost:35456/oauth2/token
          x-speakeasy-overridable-scopes: true
          scopes:
            read: Read access
            write: Write access`),
			},
			wantErrs: []string{
				"validation warn: [line 26:11] generator-validate-security - security scheme `oauth2Password` uses `x-speakeasy-overridable-scopes` extension which is not currently supported for OAuth2 `password` flow",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateSecurity{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()
			errStrs := make([]string, len(errs))
			for i, err := range errs {
				errStrs[i] = err.Error()
			}

			assert.ElementsMatch(t, tt.wantErrs, errStrs)
		})
	}
}

func Test_ValidateSecurity_OpenIdConnect_Errors(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "missing openIdConnectUrl",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: openIdConnect`),
			},
			wantErrs: []string{
				"validation error: [line 22:7] validation-required-field - `securityScheme.openIdConnectUrl` is required for type=openIdConnect",
			},
		},
		{
			name: "empty openIdConnectUrl",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: openIdConnect
      openIdConnectUrl: ""`),
			},
			wantErrs: []string{
				"validation error: [line 23:25] validation-required-field - `securityScheme.openIdConnectUrl` is required for type=openIdConnect",
			},
		},
		{
			name: "invalid openIdConnectUrl",
			args: args{
				schema: fmt.Sprintf(securityValidationOpenAPIDocTemplate, `type: openIdConnect
      openIdConnectUrl: http:// blah.`),
			},
			wantErrs: []string{
				"validation error: [line 23:25] validation-invalid-format - `securityScheme.openIdConnectUrl` is not a valid uri: parse \"http:// blah.\": invalid character \" \" in host name",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateSecurity{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()
			errStrs := make([]string, len(errs))
			for i, err := range errs {
				errStrs[i] = err.Error()
			}

			assert.ElementsMatch(t, tt.wantErrs, errStrs)
		})
	}
}

func Test_DefinedSecuritySchemes(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
	}{
		{
			name: "security schemes all defined",
			args: args{
				schema: fmt.Sprintf(baseSecurityValidationOpenAPIDocTemplate, `
security:
  - sec: []
  - sec2: []`, `
      security:
        - sec3: []`, `
    sec:
      type: apiKey
      name: api_key
      in: header
    sec2:
      type: apiKey
      name: api_key
      in: header
    sec3:
      type: apiKey
      name: api_key
      in: header`),
			},
			wantErrs: []string{},
		},
		{
			name: "security schemes defined via x-speakeasy-name-override",
			args: args{
				schema: fmt.Sprintf(baseSecurityValidationOpenAPIDocTemplate, `
security:
  - MyAuth: []`, ``, `
    sec:
      type: apiKey
      name: api_key
      in: header
      x-speakeasy-name-override: MyAuth`),
			},
			wantErrs: []string{},
		},
		{
			name: "security scheme override collision",
			args: args{
				schema: fmt.Sprintf(baseSecurityValidationOpenAPIDocTemplate, `
security:
  - sec: []`, ``, `
    sec:
      type: apiKey
      name: api_key
      in: header
      x-speakeasy-name-override: dup
    sec2:
      type: apiKey
      name: api_key
      in: header
      x-speakeasy-name-override: dup`),
			},
			wantErrs: []string{
				"validation error: [line 27:7] generator-validate-security - x-speakeasy-name-override collision: \"dup\" is already used by security scheme \"sec\"",
			},
		},
		{
			name: "security scheme override no collision - different namespaces",
			args: args{
				schema: fmt.Sprintf(baseSecurityValidationOpenAPIDocTemplate, `
security:
  - sec: []`, ``, `
    sec:
      type: http
      scheme: bearer
      x-speakeasy-name-override: bearerAuth
      x-speakeasy-model-namespace: pets
    sec2:
      type: http
      scheme: bearer
      x-speakeasy-name-override: bearerAuth
      x-speakeasy-model-namespace: orders`),
			},
			wantErrs: []string{},
		},
		{
			name: "security scheme override collision - same namespace",
			args: args{
				schema: fmt.Sprintf(baseSecurityValidationOpenAPIDocTemplate, `
security:
  - sec: []`, ``, `
    sec:
      type: http
      scheme: bearer
      x-speakeasy-name-override: bearerAuth
      x-speakeasy-model-namespace: pets
    sec2:
      type: http
      scheme: bearer
      x-speakeasy-name-override: bearerAuth
      x-speakeasy-model-namespace: pets`),
			},
			wantErrs: []string{
				"validation error: [line 27:7] generator-validate-security - x-speakeasy-name-override collision: \"bearerAuth\" is already used by security scheme \"sec\"",
			},
		},
		{
			name: "security scheme override no collision - one has namespace other does not",
			args: args{
				schema: fmt.Sprintf(baseSecurityValidationOpenAPIDocTemplate, `
security:
  - sec: []`, ``, `
    sec:
      type: http
      scheme: bearer
      x-speakeasy-name-override: bearerAuth
      x-speakeasy-model-namespace: pets
    sec2:
      type: http
      scheme: bearer
      x-speakeasy-name-override: bearerAuth`),
			},
			wantErrs: []string{},
		},
		{
			name: "security scheme override conflicts with existing scheme key",
			args: args{
				schema: fmt.Sprintf(baseSecurityValidationOpenAPIDocTemplate, `
security:
  - sec: []`, ``, `
    sec:
      type: apiKey
      name: api_key
      in: header
    Duplicate:
      type: apiKey
      name: dup_key
      in: header
    sec2:
      type: apiKey
      name: api_key_2
      in: header
      x-speakeasy-name-override: Duplicate`),
			},
			wantErrs: []string{
				"validation error: [line 30:7] generator-validate-security - x-speakeasy-name-override collision: \"Duplicate\" conflicts with existing security scheme key",
			},
		},
		{
			name: "global security scheme missing",
			args: args{
				schema: fmt.Sprintf(baseSecurityValidationOpenAPIDocTemplate, `
security:
  - sec: []
  - sec2: []`, `
      security:
        - sec3: []`, `
    sec:
      type: apiKey
      name: api_key
      in: header
    sec3:
      type: apiKey
      name: api_key
      in: header`),
			},
			wantErrs: []string{
				"validation warn: [line 9:5] generator-validate-security - security scheme `sec2` not found, assuming `type: apiKey`, `in: header`, `name: Authorization`",
			},
		},
		{
			name: "operation security scheme missing",
			args: args{
				schema: fmt.Sprintf(baseSecurityValidationOpenAPIDocTemplate, `
security:
  - sec: []
  - sec2: []`, `
      security:
        - sec3: []`, `
    sec:
      type: apiKey
      name: api_key
      in: header
    sec2:
      type: apiKey
      name: api_key
      in: header`),
			},
			wantErrs: []string{
				"validation warn: [line 14:11] generator-validate-security - security scheme `sec3` not found, assuming `type: apiKey`, `in: header`, `name: Authorization`",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithFilteredRules([]string{(&validation.ValidateSecurity{}).ID()}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()
			errStrs := make([]string, len(errs))
			for i, err := range errs {
				errStrs[i] = err.Error()
			}

			assert.ElementsMatch(t, tt.wantErrs, errStrs)
		})
	}
}
