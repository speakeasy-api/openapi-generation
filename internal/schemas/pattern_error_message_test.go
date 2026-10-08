package schemas

import (
	"context"
	"fmt"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrimitiveUnionPatternErrorMessage(t *testing.T) {
	t.Parallel()
	for _, union := range []string{"oneOf", "anyOf"} {
		for _, tc := range []struct{ name, secondPattern, wantPattern, wantMessage string }{
			{name: "changed pattern", secondPattern: "^[0-9]+$", wantPattern: "(^[a-z]+$|^[0-9]+$)"},
			{name: "equal pattern", secondPattern: "^[a-z]+$", wantPattern: "^[a-z]+$", wantMessage: "Use letters."},
		} {
			t.Run(union+"/"+tc.name, func(t *testing.T) {
				t.Parallel()
				doc := testutils.CreateOpenAPIDoc(fmt.Sprintf(`TestSchema:
  %s:
    - type: string
      pattern: '^[a-z]+$'
      x-speakeasy-pattern-error-message: Use letters.
    - type: string
      pattern: '%s'
      x-speakeasy-pattern-error-message: Use digits.
`, union, tc.secondPattern))
				common, err := testutils.SetupTestEnvironment(testutils.TestEnvironmentOptions{OpenAPIYAML: doc, MockFeatures: testutils.NewMockFeaturesConfig().WithSupportAllFeatures()})
				require.NoError(t, err)
				schemas := &Schemas{Config: common.Config, Target: common.Target, Subsystem: common.Subsystem, Namer: common.Namer}
				schema := common.DocInfo.Doc.GetComponents().GetSchemas().GetOrZero("TestSchema")
				field, err := schemas.HandleSchema(context.Background(), Params{Schema: schema, DocInfo: common.DocInfo, MaxDepth: 10, TypeDefCache: map[string]*ast.TypeDef{}, SerializationMethod: ast.SerializationMethodJSON, Scope: ast.ScopeShared})
				require.NoError(t, err)
				assert.Equal(t, tc.wantPattern, *field.Type.Validations.Pattern)
				assert.Equal(t, tc.wantMessage, field.Type.Extensions.PatternErrorMessage)
				members := schema.GetSchema().GetOneOf()
				if union == "anyOf" {
					members = schema.GetSchema().GetAnyOf()
				}
				message, err := common.Subsystem.Extensions.HandlePatternErrorMessageExtension(members[0].GetSchema().GetExtensions())
				require.NoError(t, err)
				assert.Equal(t, "Use letters.", message)
			})
		}
	}
}
