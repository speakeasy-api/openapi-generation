package validation_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_OperationValidity_Correctness(t *testing.T) {
	tests := []struct {
		name           string
		schema         string
		valid, invalid []string
	}{
		{
			name: "simple valid operation in valid spec",
			schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /test:
    get:
      operationId: getValid
      responses:
        '200':
          description: OK`,
			valid: []string{"getValid"},
		},
		{
			name: "mixed invalid operation in simple error",
			schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /valid:
    get:
      operationId: getValid
      responses:
        '200':
          description: OK
  /invalid:
    get:
      operationId: getInvalid
      parameters:
        - name:
          in: query
          schema:
            type: string`,
			valid:   []string{"getValid"},
			invalid: []string{"getInvalid"},
		},
		{
			name: "mixed invalid operation in referenced error",
			schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /valid:
    get:
      operationId: getValid
      responses:
        '200':
          description: OK
  /invalid:
    get:
      operationId: getInvalid
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/test'
components:
  schemas:
    test:
      type: array
      items: true`,
			valid:   []string{"getValid"},
			invalid: []string{"getInvalid"},
		},
		{
			name: "mixed invalid operations in referenced error",
			schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /valid:
    get:
      operationId: getValid
      responses:
        '200':
          description: OK
  /invalid:
    get:
      operationId: getInvalid
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/test'
  /invalid2:
    get:
      operationId: getInvalid2
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/test'
components:
  schemas:
    test:
      type: array
      items: true`,
			valid:   []string{"getValid"},
			invalid: []string{"getInvalid", "getInvalid2"},
		},
		{
			name: "circular reference valid operations",
			schema: `openapi: 3.1.0
info:
  title: Test
  version: 0.0.1
servers:
  - url: http://localhost:35123
paths:
  /valid:
    get:
      operationId: getValid
      responses:
        '200':
          description: OK
  /valid2:
    get:
      operationId: getValid2
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/test'
components:
  schemas:
    test:
      type: object
      properties:
        children:
          $ref: '#/components/schemas/test'`,
			valid:   []string{"getValid", "getValid2"},
			invalid: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithParseValidOperations())
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.schema), "", types.NewTargetFromTemplate("go"))
			fmt.Println(errors.Join(res.GetValidationErrors()...))

			assert.ElementsMatch(t, tt.valid, res.GetValidOperations())
			assert.ElementsMatch(t, tt.invalid, res.GetInvalidOperations())
		})
	}
}
