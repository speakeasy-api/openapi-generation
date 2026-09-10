package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/internal/generate"
)

const terraformEntityHeader = `openapi: 3.1.0
info:
  title: SpeakeasyMatchTests
  version: 0.1.0
servers:
  - url: http://localhost:35123
`

func TestSpeakeasyMatchTypeMismatch_Errors(t *testing.T) {
	t.Parallel()

	type args struct {
		schema string
	}
	tests := []struct {
		name            string
		args            args
		wantErrContains []string
	}{
		{
			name: "array parameter matching scalar field fails with clear error",
			args: args{
				schema: terraformEntityHeader + `
paths:
  /v0/hostgroups:
    post:
      x-speakeasy-entity-operation: HostGroup#create
      operationId: create-host-group
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/HostGroupRequest'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HostGroupResponse'
  /v0/hostgroups/{id}:
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: string
    get:
      x-speakeasy-entity-operation:
        terraform-resource: HostGroup#read
        terraform-datasource: null
      operationId: get-host-group
      parameters:
        - description: "The IDs of the Host Groups to return"
          in: query
          name: ids
          required: true
          schema:
            items:
              type: string
            type: array
          x-speakeasy-match: id
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HostGroupResponse'
    delete:
      x-speakeasy-entity-operation: HostGroup#delete
      operationId: delete-host-group
      responses:
        '200':
          description: OK
components:
  schemas:
    HostGroupRequest:
      type: object
      properties:
        name:
          type: string
    HostGroupResponse:
      x-speakeasy-entity: HostGroup
      type: object
      properties:
        id:
          type: string
        name:
          type: string
`,
			},
			wantErrContains: []string{
				"x-speakeasy-match type mismatch",
				"parameter type: array",
				"match target",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			errs := testGenerateTerraform(t, []byte(tt.args.schema))
			require.NotEmpty(t, errs, "expected errors but got none")

			// Combine all error messages for checking
			var allErrs strings.Builder
			for _, err := range errs {
				allErrs.WriteString(err.Error())
				allErrs.WriteString("\n")
			}
			combinedErr := allErrs.String()

			for _, wantContains := range tt.wantErrContains {
				assert.Contains(t, combinedErr, wantContains,
					"expected error to contain %q but got: %s", wantContains, combinedErr)
			}
		})
	}
}

func TestSpeakeasyMatchValidCases_Success(t *testing.T) {
	t.Parallel()

	type args struct {
		schema string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "string parameter matching string field succeeds",
			args: args{
				schema: terraformEntityHeader + `
paths:
  /v0/resources:
    post:
      x-speakeasy-entity-operation: Resource#create
      operationId: create-resource
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ResourceRequest'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ResourceResponse'
  /v0/resources/{resourceId}:
    parameters:
      - name: resourceId
        in: path
        required: true
        schema:
          type: string
        x-speakeasy-match: id
    get:
      x-speakeasy-entity-operation: Resource#read
      operationId: get-resource
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ResourceResponse'
    delete:
      x-speakeasy-entity-operation: Resource#delete
      operationId: delete-resource
      responses:
        '200':
          description: OK
components:
  schemas:
    ResourceRequest:
      type: object
      properties:
        name:
          type: string
    ResourceResponse:
      x-speakeasy-entity: Resource
      type: object
      properties:
        id:
          type: string
        name:
          type: string
`,
			},
		},
		{
			name: "body property with x-speakeasy-match to nested entity path succeeds",
			args: args{
				schema: terraformEntityHeader + `
paths:
  /v0/projects/{project_id}:
    parameters:
      - name: project_id
        in: path
        required: true
        schema:
          type: string
        x-speakeasy-match: id
    post:
      x-speakeasy-entity-operation: Project#create
      operationId: create-project
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                name:
                  type: string
                enforce_approval:
                  type: boolean
                  x-speakeasy-match: safety_settings.enforce_approval
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ProjectResponse'
    get:
      x-speakeasy-entity-operation: Project#read
      operationId: get-project
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ProjectResponse'
    delete:
      x-speakeasy-entity-operation: Project#delete
      operationId: delete-project
      responses:
        '200':
          description: OK
components:
  schemas:
    ProjectResponse:
      x-speakeasy-entity: Project
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        safety_settings:
          type: object
          properties:
            enforce_approval:
              type: boolean
`,
			},
		},
		{
			name: "body property with x-speakeasy-match to nested enum entity field succeeds",
			args: args{
				schema: terraformEntityHeader + `
paths:
  /v0/roles/{role_id}:
    parameters:
      - name: role_id
        in: path
        required: true
        schema:
          type: string
        x-speakeasy-match: id
    post:
      x-speakeasy-entity-operation: Role#create
      operationId: create-role
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                name:
                  type: string
                require_approval:
                  type: string
                  x-speakeasy-match: policy_settings.require_approval
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/RoleResponse'
    get:
      x-speakeasy-entity-operation: Role#read
      operationId: get-role
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/RoleResponse'
    delete:
      x-speakeasy-entity-operation: Role#delete
      operationId: delete-role
      responses:
        '200':
          description: OK
components:
  schemas:
    RoleResponse:
      x-speakeasy-entity: Role
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        policy_settings:
          type: object
          properties:
            require_approval:
              type: string
              enum:
                - "off"
                - "warn"
                - "on"
          required:
            - require_approval
      required:
        - id
        - name
        - policy_settings
`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			errs := testGenerateTerraform(t, []byte(tt.args.schema))
			require.Empty(t, errs, "expected no errors but got: %v", errs)
		})
	}
}

func TestSpeakeasyMatchOnRequestBody_Success(t *testing.T) {
	t.Parallel()

	schema := terraformEntityHeader + `
paths:
  /v0/things:
    post:
      x-speakeasy-entity-operation: Thing#create
      operationId: create-thing
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                name:
                  type: string
                filter_id:
                  type: string
                  x-speakeasy-match: id
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ThingResponse'
  /v0/things/{id}:
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: string
    get:
      x-speakeasy-entity-operation: Thing#read
      operationId: get-thing
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ThingResponse'
    delete:
      x-speakeasy-entity-operation: Thing#delete
      operationId: delete-thing
      responses:
        '200':
          description: OK
components:
  schemas:
    ThingResponse:
      x-speakeasy-entity: Thing
      type: object
      properties:
        id:
          type: string
        name:
          type: string
`

	errs := testGenerateTerraform(t, []byte(schema))
	require.Empty(t, errs, "expected no errors but got: %v", errs)
}

func testGenerateTerraform(t *testing.T, contents []byte) []error {
	t.Helper()

	outDir := t.TempDir()

	schemaPath := filepath.Join(outDir, "openapi.yaml")
	if err := os.WriteFile(schemaPath, contents, 0o644); err != nil {
		return []error{err}
	}

	return generate.Generate(generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL()), generate.GenerateOptions{
		OutDir:      outDir,
		Lang:        "terraform",
		SchemaPath:  schemaPath,
		SkipCompile: true,
		Logger:      logger,
	})
}
