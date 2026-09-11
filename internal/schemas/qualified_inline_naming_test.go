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

// This file covers the `qualified` name resolution mode (config.NameResolutionQualified,
// one rung above `shortest` on the nameResolution ladder): an inline schema (i.e. not
// a $ref and containing no title or x-speakeasy-name-override) is named after its property
// path from the nearest enclosing named schema, e.g.: Order.settings.level -> OrderSettingsLevel.
// The path prefix is applied even without a conflict, so names stay stable as the spec grows.
//
// Scenario matrix:
//
//	prefixed under `qualified`:
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
//
//	rooted at the nearest explicit name:
//	  S13 children of titled/overridden inline objects  TitledParent.settings(title: Preferences).tier
//	      -> PreferencesTier (the explicit name, not the component, is the root)
//
// S14 covers structurally identical inline schemas shared by several parents:
// under qualified they split into per-parent prefixed types instead of merging
// into one bare-named rename-prone type (decision: stability over dedup; see
// TestQualifiedInlineNaming_StructuralDeduplication).
//
// S15 is a prerequisite guard, not a naming scenario: qualified prefixing must
// never run without the `shortest` label machinery it builds on. When the
// toggle was a standalone boolean this needed a runtime test (flag on, Feb 2025
// fixes off -> inert); as a mode it is structural — `qualified` ranks above
// `shortest` on the ladder, so the combination cannot be configured. Reduced to
// the ladder assertion in TestQualifiedInlineNaming_ModeLadder.
const qualifiedInlineNamingSchemasYAML = `Root:
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
    settings:
      title: Preferences
      type: object
      properties:
        tier:
          type: string
          enum:
            - low
            - high
OverriddenParent:
  type: object
  properties:
    status:
      x-speakeasy-name-override: RenamedStatus
      type: string
      enum:
        - open
        - closed
    config:
      x-speakeasy-name-override: RenamedConfig
      type: object
      properties:
        depth:
          type: string
          enum:
            - shallow
            - deep
Redundant:
  type: object
  properties:
    redundantStatus:
      type: string
      enum:
        - one
        - two
    redundantSettings:
      type: object
      properties:
        inner:
          type: string
          enum:
            - a
            - b
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

// resolveInlineNamingSpec is the shared harness for the naming scenarios in this
// file: it walks rootName in yaml under the given mode, scope and pre-seeded
// context stack, runs full name resolution and returns the set of normalized
// registered names. Callers vary only in the spec, mode and entry stack.
func resolveInlineNamingSpec(t *testing.T, yaml, rootName string, mode config.NameResolutionMode, scope ast.Scope, contextStack ast.ContextStack) map[string]bool {
	t.Helper()

	common, err := testutils.SetupTestEnvironment(testutils.TestEnvironmentOptions{
		OpenAPIYAML: testutils.CreateOpenAPIDoc(yaml),
	})
	require.NoError(t, err)

	common.Config.Generation.NameResolution = mode

	schemas := &Schemas{
		Config:    common.Config,
		Target:    common.Target,
		Subsystem: common.Subsystem,
		Namer:     common.Namer,
	}

	rootSchema, exists := common.DocInfo.Doc.GetComponents().GetSchemas().Get(rootName)
	require.True(t, exists)

	ctx := logging.With(context.Background(), logging.NewLogger(zapcore.ErrorLevel))

	params := CreateTestParams(rootSchema, common.DocInfo)
	if scope != "" {
		params.Scope = scope
	}
	params.ContextStack = contextStack
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

// runQualifiedInlineNamingSpec walks the Root schema (cascading through the $ref
// properties so each component is processed with its component context), then
// runs full name resolution and returns the set of normalized registered names.
func runQualifiedInlineNamingSpec(t *testing.T, mode config.NameResolutionMode) map[string]bool {
	t.Helper()

	// The Root wrapper is anonymous; give it a context frame so the legacy
	// (Dec 2023) naming path can name it. Component processing resets the
	// stack, so this does not affect the names under test.
	return resolveInlineNamingSpec(t, qualifiedInlineNamingSchemasYAML, "Root", mode, "", ast.ContextStack{
		{Type: ast.ContextTypeOperation, Identifier: "root"},
	})
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

func TestQualifiedInlineNaming(t *testing.T) {
	names := runQualifiedInlineNamingSpec(t, config.NameResolutionQualified)

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

	// S13: a titled/overridden inline object becomes the naming root for its
	// own unnamed children — the nearest enclosing named schema, not the
	// component further up
	assert.True(t, names["preferences"], "S13: titled inline object keeps its title")
	assert.True(t, names["preferencestier"], "S13: child of titled object is rooted at the title")
	assert.False(t, names["titledparentsettingstier"], "S13: component chain must not bypass the titled root")
	assert.True(t, names["renamedconfig"], "S13: overridden inline object keeps its override")
	assert.True(t, names["renamedconfigdepth"], "S13: child of overridden object is rooted at the override")
	assert.False(t, names["overriddenparentconfigdepth"], "S13: component chain must not bypass the overridden root")

	// S11: no redundant prefix when the name already starts with the parent name
	assert.True(t, names["redundantstatus"], "S11: already-parent-prefixed property keeps its name")
	assert.False(t, names["redundantredundantstatus"], "S11: no doubled parent prefix")

	// S11 nested: the redundant component element is stripped, not the whole
	// chain — children of an already-parent-prefixed object stay qualified
	assert.True(t, names["redundantsettingsinner"], "S11: nested chain keeps qualification after stripping the redundant prefix")
	assert.False(t, names["inner"], "S11: bare leaf name must not survive")
	assert.False(t, names["redundantredundantsettingsinner"], "S11: no doubled parent prefix in nested chains")

	// Parent components themselves are never renamed
	for _, parent := range []string{"order", "account", "titledparent", "overriddenparent", "redundant", "unioned"} {
		assert.True(t, names[parent], "parent component %q should keep its name", parent)
	}
}

// Baseline contrast: under `shortest` (one rung below `qualified`) the same
// spec keeps its bare inline names, proving the prefixes above come from the
// qualified mode alone.
func TestQualifiedInlineNaming_ShortestBaseline(t *testing.T) {
	names := runQualifiedInlineNamingSpec(t, config.NameResolutionShortest)

	// Non-conflicting inline property schemas keep their bare names under shortest
	assert.True(t, names["settings"], "shortest behavior: bare 'settings', got names: %v", names)
	assert.True(t, names["level"], "shortest behavior: bare 'level'")
	assert.True(t, names["tag"], "shortest behavior: singularized bare 'tag' for array items")
	assert.True(t, names["attributes"], "shortest behavior: bare 'attributes'")
	assert.True(t, names["choice"], "shortest behavior: bare 'choice'")
	assert.True(t, names["redundantstatus"], "shortest behavior: bare 'redundantStatus'")

	// The conflicting 'status' pair is renamed reactively via the refName label
	assert.True(t, names["orderstatus"], "reactive rename on conflict")
	assert.True(t, names["accountstatus"], "reactive rename on conflict")

	// Explicit names respected either way
	assert.True(t, names["customstatus"])
	assert.True(t, names["renamedstatus"])
	assert.True(t, names["filecontent"])
}

// S12: inline schemas without a component ancestor (operation requestBody,
// response body, parameter payloads) are untouched by the qualified mode.
// Simulated by pre-seeding the context stack with operation/requestBody frames,
// which is the stack shape those walks produce — no refName frame is present.
func TestQualifiedInlineNaming_NoComponentAncestor(t *testing.T) {
	const schemasYAML = `Payload:
  type: object
  properties:
    status:
      type: string
      enum:
        - pending
        - shipped
`

	names := resolveInlineNamingSpec(t, schemasYAML, "Payload", config.NameResolutionQualified, ast.ScopeOperations, ast.ContextStack{
		{Type: ast.ContextTypeOperation, Identifier: "createOrder"},
		{Type: ast.ContextTypeRequestBody, Identifier: "requestBody"},
		{Type: ast.ContextTypeRequestMediaType, Identifier: "application/json"},
	})

	assert.True(t, names["status"], "operation-scoped inline schemas are untouched by the qualified mode, got names: %v", names)
}

// S14: structurally identical inline schemas shared by several parents split
// into per-parent types under qualified (decision: stability over dedup).
// Registration IDs include the parent refName frame, so the split needs no
// extra suppression; this locks that behavior in.
func TestQualifiedInlineNaming_StructuralDeduplication(t *testing.T) {
	const schemasYAML = `Root:
  type: object
  properties:
    first:
      $ref: '#/components/schemas/First'
    second:
      $ref: '#/components/schemas/Second'
First:
  type: object
  properties:
    shared:
      type: object
      properties:
        value:
          type: string
Second:
  type: object
  properties:
    shared:
      type: object
      properties:
        value:
          type: string
`

	names := resolveInlineNamingSpec(t, schemasYAML, "Root", config.NameResolutionQualified, "", ast.ContextStack{
		{Type: ast.ContextTypeOperation, Identifier: "root"},
	})

	assert.True(t, names["firstshared"], "each parent gets its own prefixed copy, got names: %v", names)
	assert.True(t, names["secondshared"], "each parent gets its own prefixed copy")
	assert.False(t, names["shared"], "no merged bare-named copy remains")
}

// A first property that equals its parent must keep the qualifier: stripping it
// would fold Parent.parent.status into ParentStatus and collide with
// Parent.status. Only a first segment that leads with the parent as a whole word
// (Redundant.redundantStatus) is redundant.
func TestQualifiedInlineNaming_FirstPropertyEqualsParent(t *testing.T) {
	const schemasYAML = `Root:
  type: object
  properties:
    parent:
      $ref: '#/components/schemas/Parent'
Parent:
  type: object
  properties:
    status:
      type: string
      enum:
        - a
        - b
    parent:
      type: object
      properties:
        status:
          type: string
          enum:
            - c
            - d
`

	names := resolveInlineNamingSpec(t, schemasYAML, "Root", config.NameResolutionQualified, "", ast.ContextStack{
		{Type: ast.ContextTypeOperation, Identifier: "root"},
	})

	assert.True(t, names["parentstatus"], "Parent.status stays ParentStatus, got names: %v", names)
	assert.True(t, names["parentparent"], "Parent.parent keeps the qualifier")
	assert.True(t, names["parentparentstatus"], "Parent.parent.status keeps the full chain, no collision with Parent.status")
	assert.False(t, names["status"], "no bare leaf survives")
}

// S15: qualified builds on the shortest-mode label machinery; the mode ladder
// makes the prerequisite structural — qualified always implies shortest.
func TestQualifiedInlineNaming_ModeLadder(t *testing.T) {
	assert.True(t, config.NameResolutionQualified.AtLeast(config.NameResolutionShortest))
}

// Legacy is the mode the primary test variant (and most pinned configs) actually
// generate under, yet every other unit test here runs at shortest or above, so
// the !AtLeast(ordered) else-branches in schemas.go / resolution.go would only
// regress in the expensive SDK snaptests. This pins the legacy name set so those
// branches have cheap coverage.
func TestQualifiedInlineNaming_LegacyBaseline(t *testing.T) {
	legacy := runQualifiedInlineNamingSpec(t, config.NameResolutionLegacy)

	// Legacy keeps the first-registered conflicting type's bare name and only
	// renames the loser via the refName label — it never prefixes both sides the
	// way shortest/qualified do.
	assert.True(t, legacy["status"], "legacy: first-registered 'status' keeps its bare name, got: %v", legacy)
	assert.True(t, legacy["orderstatus"], "legacy: the losing 'status' is renamed with its refName label")
	assert.False(t, legacy["accountstatus"], "legacy: only one side of the conflict is prefixed")

	// Inline property schemas are never proactively qualified below shortest.
	for _, bare := range []string{"settings", "level", "attributes", "choice", "redundantstatus"} {
		assert.Truef(t, legacy[bare], "legacy: inline %q keeps its bare name", bare)
	}

	// Explicit names (title/anchor/override) are honored in every mode.
	for _, explicit := range []string{"customstatus", "renamedstatus", "filecontent", "preferences", "renamedconfig"} {
		assert.Truef(t, legacy[explicit], "legacy: explicit name %q respected", explicit)
	}

	// Ordered is the first rung that takes the if-side of the AtLeast(ordered)
	// gates; on this spec it converges on the same names as legacy, so the gate
	// divergence is name-neutral here. Pinning the equality documents that and
	// still fails if either branch regresses independently.
	ordered := runQualifiedInlineNamingSpec(t, config.NameResolutionOrdered)
	assert.Equal(t, legacy, ordered, "legacy and ordered name sets should match on this spec")
}
