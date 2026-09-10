package validation_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComprehensiveRecommendedRuleset validates that the comprehensive test spec
// triggers all the expected validation rules in the speakeasy-recommended ruleset.
// This test serves as a regression test to ensure the validation system works correctly.
//
// Rules exercised (22 unique rules):
// - semantic-no-eval-in-markdown
// - semantic-no-script-tags-in-markdown
// - generator-duplicate-tag
// - style-oas3-host-not-example
// - style-operation-tag-defined
// - semantic-duplicated-enum
// - generator-validate-enums
// - generator-missing-error-response
// - generator-missing-examples
// - style-operation-success-response
// - generator-duplicate-inline-schemas
// - generator-pagination
// - generator-duplicate-operation-name
// - validation-operation-id-unique
// - generator-path-params
// - generator-validate-paths
// - semantic-typed-enum
// - generator-duplicate-schema-name
// - semantic-unused-component
// - generator-retries
// - generator-validate-casing
// - generator-duplicate-component-schemas
func TestComprehensiveRecommendedRuleset(t *testing.T) {
	t.Setenv("SPEAKEASY_DEBUG", "true")

	// Read the test spec
	specPath := filepath.Join("testdata", "comprehensive-recommended-ruleset.yaml")
	specBytes, err := os.ReadFile(specPath)
	require.NoError(t, err, "Failed to read test spec")

	// Create validator with speakeasy-recommended ruleset
	v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyRecommended)
	require.NoError(t, err)

	// Validate the spec
	res := validateSpec(v, context.Background(), specBytes, "", types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()

	// Convert errors to strings for comparison
	errStrs := make([]string, len(errs))
	for i, err := range errs {
		errStrs[i] = err.Error()
	}

	// Expected errors grouped by severity
	expectedErrors := []string{
		// ERROR severity
		"validation error: [line 5:16] semantic-no-eval-in-markdown - description contains content with `eval\\(`, forbidden",
		"validation error: [line 5:16] semantic-no-script-tags-in-markdown - description contains content with `<script`, forbidden",
		"validation error: [line 22:5] generator-duplicate-tag - `users` (`Users`) will collide with `users` (`Users`) [line `20`] when converted to class/field name",
		"validation error: [line 39:20] semantic-no-eval-in-markdown - description contains content with `eval\\(`, forbidden",
		"validation error: [line 39:20] semantic-no-script-tags-in-markdown - description contains content with `<script`, forbidden",
		"validation error: [line 140:7] generator-duplicate-operation-name - method name `sdk.CreateItem()` will collide with operation `post /items/create` [line `128`], try using `x-speakeasy-name-override`",
		"validation error: [line 140:20] validation-operation-id-unique - the `post` operation at path `/things/create` contains a duplicate operationId `createItem`",
		"validation error: [line 153:7] generator-path-params - `GET` must define parameter `orderId` as expected by path `/orders/{orderId}`",
		"validation error: [line 161:5] generator-validate-paths - path contains unencoded characters [ `{` ]",
		"validation error: [line 192:15] generator-validate-enums - enum value `two` is not an `integer`",

		// WARN severity
		"validation warn: [line 12:10] style-oas3-host-not-example - server url \"https://example.com/api/v1\" must not point at example.com",
		"validation warn: [line 37:11] style-operation-tag-defined - tag `undefined-tag` for getUser operation is not defined as a global tag",
		"validation warn: [line 50:17] semantic-duplicated-enum - enum contains a duplicate: `profile`",
		"validation warn: [line 72:7] style-operation-success-response - operation `createUser` must define at least a single `2xx` or `3xx` response",
		"validation warn: [line 192:15] semantic-typed-enum - enum value at index `1` does not match schema type [`integer`]",
		"validation warn: [line 194:5] generator-duplicate-schema-name - `user` (`User`) will collide with `User` (`User`) [line `169`] when converted to class name",
		"validation warn: [line 194:5] generator-validate-casing - schema name `user` differs from `User` only by casing and will collide in case-insensitive languages (e.g. C#) and filesystems",
		"validation warn: [line 194:5] semantic-unused-component - `#/components/schemas/user` is potentially unused or has been orphaned",
		"validation warn: [line 198:5] semantic-unused-component - `#/components/schemas/UntypedEnum` is potentially unused or has been orphaned",
		"validation warn: [line 205:5] semantic-unused-component - `#/components/schemas/UnusedSchema` is potentially unused or has been orphaned",
		"validation warn: [line 223:5] semantic-unused-component - `#/components/schemas/Error2` is potentially unused or has been orphaned",

		// HINT severity
		"validation hint: [line 1:1] generator-retries - retries should be configured - consider adding `x-speakeasy-retries` extension globally or per-operation",
		"validation hint: [line 212:5] generator-duplicate-component-schemas - schema `Error1` is structurally identical to `Error2`; consolidate them into a single component referenced via `$ref` to avoid generating redundant types",
		"validation hint: [line 33:11] generator-missing-examples - missing example for parameter. Consider adding an example",
		"validation hint: [line 47:13] generator-missing-examples - missing example for parameter. Consider adding an example",
		"validation hint: [line 52:9] generator-missing-error-response - an error response should be defined for all operations",
		"validation hint: [line 57:17] generator-missing-examples - missing example for `responses`. Consider adding an example",
		"validation hint: [line 68:15] generator-missing-examples - missing example for `requestBody`. Consider adding an example",
		"validation hint: [line 81:9] generator-missing-error-response - an error response should be defined for all operations",
		"validation hint: [line 87:17] generator-duplicate-inline-schemas - `2` duplicates of `object` \"`id`: `string`, `name`: `string`, `price`: `number`\" at lines [87,116]",
		"validation hint: [line 87:17] generator-missing-examples - missing example for `responses`. Consider adding an example",
		"validation hint: [line 98:7] generator-pagination - pagination might be supported by this operation - consider adding `x-speakeasy-pagination` extension",
		"validation hint: [line 103:13] generator-missing-examples - missing example for parameter. Consider adding an example",
		"validation hint: [line 107:13] generator-missing-examples - missing example for parameter. Consider adding an example",
		"validation hint: [line 110:9] generator-missing-error-response - an error response should be defined for all operations",
		"validation hint: [line 116:17] generator-missing-examples - missing example for `responses`. Consider adding an example",
		"validation hint: [line 133:15] generator-missing-examples - missing example for `requestBody`. Consider adding an example",
		"validation hint: [line 145:15] generator-missing-examples - missing example for `requestBody`. Consider adding an example",
		"validation hint: [line 156:9] generator-missing-error-response - an error response should be defined for all operations",
		"validation hint: [line 164:9] generator-missing-error-response - an error response should be defined for all operations",
		"validation hint: [line 170:7] generator-missing-examples - missing example for component. Consider adding an example",
		"validation hint: [line 184:11] generator-validate-enums - enum is nullable but does not contain a `null` value",
		"validation hint: [line 195:7] generator-missing-examples - missing example for component. Consider adding an example",
		"validation hint: [line 199:7] generator-missing-examples - missing example for component. Consider adding an example",
		"validation hint: [line 206:7] generator-missing-examples - missing example for component. Consider adding an example",
		"validation hint: [line 213:7] generator-missing-examples - missing example for component. Consider adding an example",
		"validation hint: [line 224:7] generator-missing-examples - missing example for component. Consider adding an example",
		"validation hint: [line 240:13] generator-missing-examples - missing example for `responses`. Consider adding an example",
	}

	// Assert that we got exactly the expected errors
	assert.ElementsMatch(t, expectedErrors, errStrs, "Validation errors should match expected errors")

	// Also verify we're triggering the expected number of unique rules
	// This is a sanity check that we're exercising a good variety of rules
	t.Logf("Total validation errors: %d", len(errStrs))
}
