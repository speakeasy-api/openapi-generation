package extensions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func decodeCLIErrorsTest(t *testing.T, source string) (*CLIErrorManifest, error) {
	t.Helper()
	var doc yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte(source), &doc))
	require.NotEmpty(t, doc.Content)
	return DecodeCLIErrorsManifest(doc.Content[0])
}

func requireCLIErrorsDecodeError(t *testing.T, source, contains string) {
	t.Helper()
	_, err := decodeCLIErrorsTest(t, source)
	require.Error(t, err)
	assert.Contains(t, err.Error(), contains)
}

func TestCLIErrorsFullManifest(t *testing.T) {
	manifest, err := decodeCLIErrorsTest(t, `
version: 1
reasons:
  ACCOUNT_KEY_REVOKED:
    type: authentication_error
    hints: Replace the revoked credential
  workspace.suspended:
    type: authorization_error
    hints:
      - Contact the workspace owner
      - Retry after access is restored
  SERVICE_PAUSED:
    hints: Resume the service
  CLI_PROTOCOL:
    hints: Report the response and CLI version
types:
  not_found:
    hints: Check the resource identifier
`)
	require.NoError(t, err)
	assert.Equal(t, 1, manifest.Version)
	require.Len(t, manifest.Reasons, 4)
	first := manifest.Reasons[0]
	first.line = 0
	assert.Equal(t, CLIErrorRule{
		Reason:   "ACCOUNT_KEY_REVOKED",
		Type:     "authentication_error",
		Hints:    []string{"Replace the revoked credential"},
		HasHints: true,
	}, first)
	assert.Equal(t, "workspace.suspended", manifest.Reasons[1].Reason)
	assert.Equal(t, []string{"Contact the workspace owner", "Retry after access is restored"}, manifest.Reasons[1].Hints)
	assert.Empty(t, manifest.Reasons[2].Type, "a rule may declare the hint half alone; its type falls back to later evidence")
	assert.Equal(t, "CLI_PROTOCOL", manifest.Reasons[3].Reason)
	assert.Equal(t, []string{"Report the response and CLI version"}, manifest.Reasons[3].Hints)
	require.Len(t, manifest.Types, 1)
	assert.Equal(t, CLIErrorTypeHints{Type: "not_found", Hints: []string{"Check the resource identifier"}}, manifest.Types[0])
	assert.False(t, manifest.UnwrapErrorArray)
	assert.Empty(t, manifest.Probes)
}

func TestCLIErrorsRequiresVersionAndContent(t *testing.T) {
	requireCLIErrorsDecodeError(t, `reasons: {BROKEN: {type: api_error}}`, "version is required")
	requireCLIErrorsDecodeError(t, `version: "1"
reasons: {BROKEN: {type: api_error}}`, "integer 1")
	requireCLIErrorsDecodeError(t, `version: 2
reasons: {BROKEN: {type: api_error}}`, "unsupported version")
	requireCLIErrorsDecodeError(t, `version: 1`, "at least one of reasonPointer, reasons, probes, types, or unwrapErrorArray")
	requireCLIErrorsDecodeError(t, `version: 1
reasons: {}`, "at least one of reasonPointer, reasons, probes, types, or unwrapErrorArray")
}

func TestCLIErrorsProbes(t *testing.T) {
	manifest, err := decodeCLIErrorsTest(t, `version: 1
reasonPointer: $.details[*].reason
reasons:
  ACCOUNT_KEY_REVOKED:
    type: authentication_error
  TARGET_MISSING:
    hints: This hint applies only to a detail-carrier reason
probes:
  - pointer: $.status
    reasons:
      ACCESS_DENIED:
        type: authorization_error
      TARGET_MISSING:
        type: not_found
  - pointer: $['error-state']
unwrapErrorArray: true`)
	require.NoError(t, err)
	assert.True(t, manifest.UnwrapErrorArray)
	require.Len(t, manifest.Probes, 2)
	assert.Equal(t, "$.status", manifest.Probes[0].Pointer)
	assert.Equal(t, []CLIErrorPointerSegment{{Field: "status"}}, manifest.Probes[0].Segments)
	require.Len(t, manifest.Probes[0].Reasons, 2)
	assert.Equal(t, "ACCESS_DENIED", manifest.Probes[0].Reasons[0].Reason)
	assert.Equal(t, "authorization_error", manifest.Probes[0].Reasons[0].Type)
	assert.Equal(t, "not_found", manifest.Probes[0].Reasons[1].Type,
		"the same reason value may carry a type at one carrier and hints alone at another")
	assert.Equal(t, "$['error-state']", manifest.Probes[1].Pointer,
		"a probe may declare a pointer alone; a declared carrier promotes its reason codes verbatim")
	assert.Empty(t, manifest.Probes[1].Reasons)
}

func TestCLIErrorsQuotedCarrierIsNotADuplicate(t *testing.T) {
	// $['error.reason'] names one member literally called "error.reason";
	// $.error.reason names the reason member nested under error. A naive
	// dot-join canonicalization conflates the two carriers and rejects a
	// legitimate manifest as a duplicate.
	manifest, err := decodeCLIErrorsTest(t, `version: 1
reasonPointer: $.error.reason
probes:
  - pointer: $['error.reason']`)
	require.NoError(t, err)
	require.Len(t, manifest.Probes, 1)
	assert.Equal(t, []CLIErrorPointerSegment{{Field: "error.reason"}}, manifest.Probes[0].Segments)

	// A quoted member literally containing "[*]" is distinct from a
	// wildcard fan-out over an array of that name.
	manifest, err = decodeCLIErrorsTest(t, `version: 1
reasonPointer: $.details[*].reason
probes:
  - pointer: $['details[*].reason']`)
	require.NoError(t, err)
	require.Len(t, manifest.Probes, 1)

	// Two probes whose spellings differ only by quoting stay distinct from
	// each other, too.
	manifest, err = decodeCLIErrorsTest(t, `version: 1
probes:
  - pointer: $.a.b
  - pointer: $['a.b']`)
	require.NoError(t, err)
	require.Len(t, manifest.Probes, 2)

	// Equivalent spellings of the same carrier still collide: duplicate
	// detection keeps firing across dot and bracket forms.
	requireCLIErrorsDecodeError(t, `version: 1
probes:
  - pointer: $.a.b
  - pointer: $['a']['b']`, "duplicates probes[0]")
}

func TestCLIErrorsProbeValidation(t *testing.T) {
	requireCLIErrorsDecodeError(t, `version: 1
probes: {}`, "must be a sequence of probe declarations")
	requireCLIErrorsDecodeError(t, `version: 1
probes: []`, "at least one probe declaration")
	requireCLIErrorsDecodeError(t, `version: 1
probes:
  - reasons:
      DENIED:
        type: authorization_error`, "must declare pointer")
	requireCLIErrorsDecodeError(t, `version: 1
probes:
  - pointer: $.status
    pointers: {}`, `unknown key "pointers"`)
	requireCLIErrorsDecodeError(t, `version: 1
probes:
  - pointer: $.details[*]`, "must end at the named member")
	requireCLIErrorsDecodeError(t, `version: 1
probes:
  - pointer: $.reason`, "duplicates the default primary carrier $.reason")
	requireCLIErrorsDecodeError(t, `version: 1
reasonPointer: $.details[*].reason
probes:
  - pointer: $['details'][*].reason`, "duplicates the declared reasonPointer")
	requireCLIErrorsDecodeError(t, `version: 1
probes:
  - pointer: $.status
  - pointer: $['status']`, "duplicates probes[0]")
	requireCLIErrorsDecodeError(t, `version: 1
probes:
  - pointer: $.status
    reasons:
      CLI_PROTOCOL:
        hints: Fix it`, "runtime-owned; declare its hints under the top-level reasons")
	requireCLIErrorsDecodeError(t, `version: 1
reasons:
  TARGET_MISSING:
    type: validation_error
probes:
  - pointer: $.status
    reasons:
      TARGET_MISSING:
        type: not_found`, "a reason value must map to a single error type across carriers")
	requireCLIErrorsDecodeError(t, `version: 1
probes:
  - pointer: $.status
    reasons:
      TARGET_MISSING:
        type: not_found
  - pointer: $.state
    reasons:
      TARGET_MISSING:
        type: validation_error`, "a reason value must map to a single error type across carriers")
	requireCLIErrorsDecodeError(t, `version: 1
unwrapErrorArray: yes please`, "unwrapErrorArray must be true or false")
}

func TestCLIErrorsUnwrapAloneIsDeclarable(t *testing.T) {
	manifest, err := decodeCLIErrorsTest(t, `version: 1
unwrapErrorArray: true`)
	require.NoError(t, err)
	assert.True(t, manifest.UnwrapErrorArray)
}

func TestCLIErrorsReasonPointer(t *testing.T) {
	manifest, err := decodeCLIErrorsTest(t, `version: 1
reasonPointer: $.reason
reasons:
  INSUFFICIENT_FUNDS:
    type: validation_error`)
	require.NoError(t, err)
	assert.Equal(t, "$.reason", manifest.ReasonPointer)
	assert.Equal(t, []CLIErrorPointerSegment{{Field: "reason"}}, manifest.ReasonSegments)

	manifest, err = decodeCLIErrorsTest(t, `version: 1
reasonPointer: $.details[*].reason
reasons:
  ACCOUNT_FROZEN:
    type: authorization_error`)
	require.NoError(t, err)
	assert.Equal(t, "$.details[*].reason", manifest.ReasonPointer)
	assert.Equal(t, []CLIErrorPointerSegment{
		{Field: "details"},
		{IsWild: true},
		{Field: "reason"},
	}, manifest.ReasonSegments)

	manifest, err = decodeCLIErrorsTest(t, `version: 1
reasonPointer: $['error-code']`)
	require.NoError(t, err)
	assert.Equal(t, []CLIErrorPointerSegment{{Field: "error-code"}}, manifest.ReasonSegments)
	assert.Empty(t, manifest.Reasons, "a declared carrier alone opts into surfacing its reason codes")
}

func TestCLIErrorsReasonPointerGrammar(t *testing.T) {
	pointerError := func(pointer, contains string) {
		t.Helper()
		requireCLIErrorsDecodeError(t, "version: 1\nreasonPointer: \""+pointer+"\"", contains)
	}
	pointerError("", "reasonPointer is empty")
	pointerError("/reason", "JSON Pointer syntax")
	pointerError("reason", "must start with $")
	pointerError("$", "addresses the error object root")
	pointerError("$..reason", "recursive descent")
	pointerError("$.*", "dot wildcard")
	pointerError("$.details[0].reason", "spell the segment as [*]")
	pointerError("$.details[?(@.reason)].reason", "filter expression")
	pointerError("$.details[1:2].reason", "slice")
	pointerError("$.details['a','b']", "union selector")
	pointerError("$.details[*]", "must end at the named member")
	pointerError("$[*].reason", "begins with a [*] fan-out")
	pointerError("$.details[*.reason", "unterminated bracket segment")
	pointerError("$.reason.", "dangling dot")
	pointerError("$.details[*]reason", "unexpected trailing characters")
	pointerError("$.bad name", "not a plain identifier")
}

func TestCLIErrorsStrictKeysAndTypes(t *testing.T) {
	requireCLIErrorsDecodeError(t, `version: 1
reason: {BROKEN: {type: api_error}}`, `unknown key "reason"`)
	requireCLIErrorsDecodeError(t, `version: 1
reasons:
  BROKEN:
    typo: value`, `unknown key "typo"`)
	requireCLIErrorsDecodeError(t, `version: 1
reasons:
  BROKEN:
    type: authentcation_error`, "authentication_error")
	requireCLIErrorsDecodeError(t, `version: 1
types:
  authentcation_error:
    hints: Fix it`, "authentication_error")
	requireCLIErrorsDecodeError(t, `version: 1
types:
  not_found:
    type: not_found`, `unknown key "type"`)
	requireCLIErrorsDecodeError(t, `version: 1
types:
  not_found: {}`, "must declare hints")
}

func TestCLIErrorsReasonKeysAreExactStrings(t *testing.T) {
	manifest, err := decodeCLIErrorsTest(t, `version: 1
reasons:
  lower-case/reason.v2:
    type: api_error`)
	require.NoError(t, err)
	assert.Equal(t, "lower-case/reason.v2", manifest.Reasons[0].Reason)

	requireCLIErrorsDecodeError(t, `version: 1
reasons:
  " padded ":
    type: api_error`, "leading or trailing whitespace")
	requireCLIErrorsDecodeError(t, "version: 1\nreasons:\n  \"line\\nbreak\":\n    type: api_error", "control characters")
	requireCLIErrorsDecodeError(t, `version: 1
reasons:
  BROKEN: {}`, "type, hints, or both")
}

func TestCLIErrorsRuntimeAndUnsupportedTypesAreDeclarable(t *testing.T) {
	for _, errorType := range []string{"runtime_error", "unsupported_error", "async_failed", "async_timeout", "async_unknown_state"} {
		manifest, err := decodeCLIErrorsTest(t, "version: 1\ntypes:\n  "+errorType+":\n    hints: Fix it\nreasons:\n  CUSTOM_REASON:\n    type: "+errorType)
		require.NoError(t, err)
		require.Len(t, manifest.Types, 1)
		assert.Equal(t, errorType, manifest.Types[0].Type)
		require.Len(t, manifest.Reasons, 1)
		assert.Equal(t, errorType, manifest.Reasons[0].Type)
	}
}

func TestCLIErrorsRuntimeReasonNamespace(t *testing.T) {
	requireCLIErrorsDecodeError(t, `version: 1
reasons:
  CLI_UNKNOWN:
    hints: Fix it`, "not a reason the generated runtime emits")
	requireCLIErrorsDecodeError(t, `version: 1
reasons:
  CLI_PROTOCOL:
    type: protocol_error
    hints: Fix it`, "may declare hints only")
	requireCLIErrorsDecodeError(t, `version: 1
reasons:
  CLI_PROTOCOL: {}`, "must declare hints")
}

func TestCLIErrorsHintValidation(t *testing.T) {
	requireCLIErrorsDecodeError(t, `version: 1
reasons:
  BROKEN:
    type: api_error
    hints: "  "`, "leading or trailing whitespace")
	requireCLIErrorsDecodeError(t, `version: 1
reasons:
  BROKEN:
    type: api_error
    hints: []`, "at least one line")
	requireCLIErrorsDecodeError(t, "version: 1\nreasons:\n  BROKEN:\n    type: api_error\n    hints: \"first\\nsecond\"", "control characters")
}

func TestCLIErrorsYAMLStrictnessAndDeterminism(t *testing.T) {
	requireCLIErrorsDecodeError(t, `version: 1
reasons:
  BROKEN:
    type: api_error
    type: server_error`, "duplicate key")
	requireCLIErrorsDecodeError(t, `version: 1
reasons:
  BROKEN:
    <<: &shared
      type: api_error
    type: server_error`, "duplicate key")

	manifest, err := decodeCLIErrorsTest(t, `version: 1
reasons:
  FIRST: &rule
    type: api_error
    hints: First hint
  SECOND: *rule
types:
  server_error:
    hints: Retry later`)
	require.NoError(t, err)
	assert.Equal(t, []string{"FIRST", "SECOND"}, []string{manifest.Reasons[0].Reason, manifest.Reasons[1].Reason})
	assert.Equal(t, manifest.Reasons[0].Hints, manifest.Reasons[1].Hints)
}
