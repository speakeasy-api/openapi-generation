package schemas

import (
	"context"
	"strings"
	"testing"
	"unicode"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/namer"
	"github.com/speakeasy-api/openapi-generation/v2/internal/testutils"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

// TestEagerInlineNaming covers the eager inline naming strategy: inline schemas
// that are (possibly nested) object properties of a component are preemptively
// prefixed with their parent component name, so that names stay stable when
// later spec additions introduce conflicts.
//
// Scenario matrix:
//
//	prefixed when enabled:
//	  S1  enum property of a component                  Order.status          -> OrderStatus
//	  S2  inline object property of a component         Order.settings      -> OrderSettings
//	  S3  nested inline property (full chain)           Order.settings.level -> OrderSettingsLevel
//	  S4  inline object in array items of a property    Order.tags[]        -> OrderTags
//	  S5  inline object in additionalProperties         Order.attributes{}  -> OrderAttributes
//	  S6  oneOf union named after a property            Unioned.choice            -> UnionedChoice
//	  S7  conflicting property names across components  Order.status / Account.status
//	      (same final names as the reactive strategy, no double prefix)
//
//	never prefixed:
//	  S8  $ref properties (named components)            Order.file -> FileContent
//	  S9  title-named inline schemas                    TitledParent.status (title: CustomStatus)
//	  S10 x-speakeasy-name-override inline schemas      OverriddenParent.status (-> RenamedStatus)
//	  S11 redundant prefix (name already starts with    Redundant.redundantStatus -> RedundantStatus
//	      the parent component name)
//	  S12 inline schemas without a component ancestor   operation requestBody/response/parameter
//	      (no refName frame; covered by the pre-seeded operation stack test below)
//	  S13 structurally deduplicated inline schemas shared by several parents (open decision, skipped)
//	  S14 flag has no effect without nameResolutionFeb2025
const eagerInlineNamingSchemasYAML = `Root:
  type: object
  properties:
    order:
      $ref: '#/components/schemas/Order'
    account:
      $ref: '#/components/schemas/Account'
    titledParent:
      $ref: '#/components/schemas/TitledParent'
    overriddenParent:
      $ref: '#/components/schemas/OverriddenParent'
    redundant:
      $ref: '#/components/schemas/Redundant'
    unioned:
      $ref: '#/components/schemas/Unioned'
Order:
  type: object
  properties:
    status:
      type: string
      enum:
        - pending
        - shipped
    settings:
      type: object
      properties:
        level:
          type: string
          enum:
            - low
            - high
    tags:
      type: array
      items:
        type: object
        properties:
          key:
            type: string
    attributes:
      type: object
      additionalProperties:
        type: object
        properties:
          value:
            type: string
    file:
      $ref: '#/components/schemas/FileContent'
FileContent:
  type: object
  properties:
    path:
      type: string
Account:
  type: object
  properties:
    status:
      type: string
      enum:
        - basic
        - premium
TitledParent:
  type: object
  properties:
    status:
      title: CustomStatus
      type: string
      enum:
        - active
        - inactive
OverriddenParent:
  type: object
  properties:
    status:
      x-speakeasy-name-override: RenamedStatus
      type: string
      enum:
        - open
        - closed
Redundant:
  type: object
  properties:
    redundantStatus:
      type: string
      enum:
        - one
        - two
Unioned:
  type: object
  properties:
    choice:
      oneOf:
        - type: object
          properties:
            a:
              type: string
        - type: string
`

// runEagerInlineNamingSpec walks the Root schema (cascading through the $ref
// properties so each component is processed with its component context), then
// runs full name resolution and returns the set of normalized registered names.
func runEagerInlineNamingSpec(t *testing.T, eager bool, fixes *config.Fixes) map[string]bool {
	t.Helper()

	common, err := testutils.SetupTestEnvironment(testutils.TestEnvironmentOptions{
		OpenAPIYAML: testutils.CreateOpenAPIDoc(eagerInlineNamingSchemasYAML),
		Fixes:       fixes,
	})
	require.NoError(t, err)

	common.Config.EagerInlineNaming = eager

	schemas := &Schemas{
		Config:    common.Config,
		Target:    common.Target,
		Subsystem: common.Subsystem,
		Namer:     common.Namer,
	}

	rootSchema, exists := common.DocInfo.Doc.GetComponents().GetSchemas().Get("Root")
	require.True(t, exists)

	ctx := logging.With(context.Background(), logging.NewLogger(zapcore.ErrorLevel))

	params := CreateTestParams(rootSchema, common.DocInfo)
	// The Root wrapper is anonymous; give it a context frame so the legacy
	// (Dec 2023) naming path can name it. Component processing resets the
	// stack, so this does not affect the names under test.
	params.ContextStack = ast.ContextStack{
		{Type: ast.ContextTypeOperation, Identifier: "root"},
	}
	_, err = schemas.HandleSchema(ctx, params)
	require.NoError(t, err)

	resolver, err := namer.NewResolver(common.Subsystem, nil)
	require.NoError(t, err)

	resolver.ResolveNames(ctx, common.Subsystem.Register, testImportConfig(), false, false)

	names := map[string]bool{}
	for _, td := range common.Subsystem.Register.AllTypes() {
		names[normalizeTestName(td.Name)] = true
	}
	return names
}

func testImportConfig() configuration.ImportConfig {
	return configuration.ImportConfig{
		Option: configuration.ImportOptionOpenAPI,
		Paths: map[string]string{
			string(ast.ScopeShared):     "models/shared",
			string(ast.ScopeOperations): "models/operations",
			string(ast.ScopeErrors):     "models/errors",
			string(ast.ScopeWebhooks):   "models/webhooks",
			string(ast.ScopeCallbacks):  "models/callbacks",
		},
	}
}

// normalizeTestName lowercases and strips non-alphanumerics so assertions hold
// regardless of the raw separator convention ("Order_status" == "OrderStatus").
func normalizeTestName(name string) string {
	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

func TestEagerInlineNaming_Enabled(t *testing.T) {
	names := runEagerInlineNamingSpec(t, true, nil)

	// S1: enum property of a component
	assert.True(t, names["orderstatus"], "S1: Order.status should be prefixed, got names: %v", names)
	assert.False(t, names["status"], "S1: bare 'status' should no longer exist")

	// S2: inline object property of a component
	assert.True(t, names["ordersettings"], "S2: Order.settings should be prefixed")
	assert.False(t, names["settings"], "S2: bare 'settings' should no longer exist")

	// S3: nested inline property carries the full chain
	assert.True(t, names["ordersettingslevel"], "S3: Order.settings.level should carry the full chain")
	assert.False(t, names["level"], "S3: bare 'level' should no longer exist")

	// S4: inline object in array items of a component property (items are singularized)
	assert.True(t, names["ordertag"], "S4: Order.tags items should be prefixed")
	assert.False(t, names["tag"], "S4: bare 'tag' should no longer exist")

	// S5: inline object in additionalProperties of a component property
	assert.True(t, names["orderattributes"], "S5: Order.attributes values should be prefixed")
	assert.False(t, names["attributes"], "S5: bare 'attributes' should no longer exist")

	// S6: oneOf union named after a property; the variant named after the union
	// ("ChoiceUnion" today) follows the union's new name
	assert.True(t, names["unionedchoice"], "S6: Unioned.choice union should be prefixed")
	assert.False(t, names["choice"], "S6: bare 'choice' should no longer exist")
	assert.False(t, names["choiceunion"], "S6: union-derived variant names follow the prefixed union name")

	// S7: names for conflicting properties match what the reactive strategy
	// would have produced once the conflict appeared
	assert.True(t, names["accountstatus"], "S7: Account.status should be prefixed")
	assert.False(t, names["orderorderstatus"], "S7: no double prefixing")

	// S8: $ref properties keep their component name
	assert.True(t, names["filecontent"], "S8: referenced components keep their name")
	assert.False(t, names["orderfilecontent"], "S8: referenced components are never prefixed")
	assert.False(t, names["orderfile"], "S8: referenced components are never prefixed")

	// S9: explicit titles are respected as-is
	assert.True(t, names["customstatus"], "S9: title-named schemas keep their title")
	assert.False(t, names["titledparentcustomstatus"], "S9: title-named schemas are not prefixed")

	// S10: x-speakeasy-name-override is respected as-is
	assert.True(t, names["renamedstatus"], "S10: overridden schemas keep their override")
	assert.False(t, names["overriddenparentrenamedstatus"], "S10: overridden schemas are not prefixed")

	// S11: no redundant prefix when the name already starts with the parent name
	assert.True(t, names["redundantstatus"], "S11: already-parent-prefixed property keeps its name")
	assert.False(t, names["redundantredundantstatus"], "S11: no doubled parent prefix")

	// Parent components themselves are never renamed
	for _, parent := range []string{"order", "account", "titledparent", "overriddenparent", "redundant", "unioned"} {
		assert.True(t, names[parent], "parent component %q should keep its name", parent)
	}
}

func TestEagerInlineNaming_Disabled(t *testing.T) {
	names := runEagerInlineNamingSpec(t, false, nil)

	// Non-conflicting inline property schemas keep their bare names today
	assert.True(t, names["settings"], "current behavior: bare 'settings', got names: %v", names)
	assert.True(t, names["level"], "current behavior: bare 'level'")
	assert.True(t, names["tag"], "current behavior: singularized bare 'tag' for array items")
	assert.True(t, names["attributes"], "current behavior: bare 'attributes'")
	assert.True(t, names["choice"], "current behavior: bare 'choice'")
	assert.True(t, names["redundantstatus"], "current behavior: bare 'redundantStatus'")

	// The conflicting 'status' pair is renamed reactively via the refName label
	assert.True(t, names["orderstatus"], "reactive rename on conflict")
	assert.True(t, names["accountstatus"], "reactive rename on conflict")

	// Explicit names respected either way
	assert.True(t, names["customstatus"])
	assert.True(t, names["renamedstatus"])
	assert.True(t, names["filecontent"])
}

// S12: inline schemas without a component ancestor (operation requestBody,
// response body, parameter payloads) are untouched by the flag. Simulated by
// pre-seeding the context stack with operation/requestBody frames, which is the
// stack shape those walks produce — no refName frame is present.
func TestEagerInlineNaming_NoComponentAncestor(t *testing.T) {
	const schemasYAML = `Payload:
  type: object
  properties:
    status:
      type: string
      enum:
        - pending
        - shipped
`

	common, err := testutils.SetupTestEnvironment(testutils.TestEnvironmentOptions{
		OpenAPIYAML: testutils.CreateOpenAPIDoc(schemasYAML),
	})
	require.NoError(t, err)

	common.Config.EagerInlineNaming = true

	schemas := &Schemas{
		Config:    common.Config,
		Target:    common.Target,
		Subsystem: common.Subsystem,
		Namer:     common.Namer,
	}

	payload, exists := common.DocInfo.Doc.GetComponents().GetSchemas().Get("Payload")
	require.True(t, exists)

	ctx := logging.With(context.Background(), logging.NewLogger(zapcore.ErrorLevel))

	params := CreateTestParams(payload, common.DocInfo)
	params.Scope = ast.ScopeOperations
	params.ContextStack = ast.ContextStack{
		{Type: ast.ContextTypeOperation, Identifier: "createOrder"},
		{Type: ast.ContextTypeRequestBody, Identifier: "requestBody"},
		{Type: ast.ContextTypeRequestMediaType, Identifier: "application/json"},
	}
	_, err = schemas.HandleSchema(ctx, params)
	require.NoError(t, err)

	resolver, err := namer.NewResolver(common.Subsystem, nil)
	require.NoError(t, err)

	resolver.ResolveNames(ctx, common.Subsystem.Register, testImportConfig(), false, false)

	names := map[string]bool{}
	for _, td := range common.Subsystem.Register.AllTypes() {
		names[normalizeTestName(td.Name)] = true
	}

	assert.True(t, names["status"], "operation-scoped inline schemas are untouched by the flag, got names: %v", names)
}

// S13: a structurally deduplicated inline schema shared by several parents has
// no single parent to prefix with. Desired behavior is an open decision:
// either keep the merged type with its bare (rename-prone) name, or stop
// deduplicating under the flag so each parent gets its own stably named copy.
func TestEagerInlineNaming_StructuralDeduplication(t *testing.T) {
	t.Skip("pending decision: keep merged bare name vs split per parent under the flag")
}

// S14: the flag builds on the Feb 2025 label machinery and must be inert without it.
func TestEagerInlineNaming_RequiresNameResolutionFeb2025(t *testing.T) {
	fixes := testutils.DefaultFixes()
	fixes.NameResolutionFeb2025 = false
	fixes.NameResolutionDec2023 = true

	names := runEagerInlineNamingSpec(t, true, fixes)

	assert.True(t, names["settings"], "flag is inert without nameResolutionFeb2025, got names: %v", names)
	assert.False(t, names["ordersettings"], "flag is inert without nameResolutionFeb2025")
}
