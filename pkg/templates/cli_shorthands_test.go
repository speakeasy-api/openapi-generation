package templates

import (
	"cmp"
	"encoding/json"
	"fmt"
	"path"
	"testing"

	"github.com/dop251/goja"
	esbuild "github.com/evanw/esbuild/pkg/api"
	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	embedded "github.com/speakeasy-api/openapi-generation/v2/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type retiredShorthand struct {
	Flag string `json:"flag"`
}

type shorthandLock struct {
	Flags   map[string]string           `json:"flags,omitempty"`
	Retired map[string]retiredShorthand `json:"retired,omitempty"`
}

type shorthandInput struct {
	FlagNames []string       `json:"flagNames"`
	Lock      *shorthandLock `json:"lock,omitempty"`
	Pinned    []string       `json:"pinned,omitempty"`
	Preferred []string       `json:"preferred,omitempty"`
	Reserved  []string       `json:"reserved,omitempty"`
}

type shorthandResolution struct {
	Flags   map[string]string           `json:"flags"`
	Retired map[string]retiredShorthand `json:"retired"`
}

// loadCLIIncludes evaluates CLI include files in one runtime, stubbing the
// template registration the generator's engine normally provides.
func loadCLIIncludes(t *testing.T, files ...string) *goja.Runtime {
	t.Helper()

	scripts := make([]string, 0, len(files))
	for _, file := range files {
		scripts = append(scripts, path.Join("cli", file))
	}
	return loadTemplateScripts(t, scripts...)
}

// loadTemplateScripts evaluates template scripts, given relative to the
// templates directory, in one runtime.
func loadTemplateScripts(t *testing.T, scripts ...string) *goja.Runtime {
	t.Helper()

	vm := goja.New()
	noop := func(goja.FunctionCall) goja.Value { return goja.Undefined() }
	require.NoError(t, vm.Set("registerTemplateFunc", noop))
	require.NoError(t, vm.Set("unregisterTemplateFunc", noop))
	require.NoError(t, vm.Set("getOperationModelName", noop))
	for _, script := range scripts {
		scriptPath := path.Join("templates", script)
		source, err := embedded.TemplateFS.ReadFile(scriptPath)
		require.NoError(t, err)

		result := esbuild.Transform(string(source), esbuild.TransformOptions{
			Target: esbuild.ES2015,
			Loader: esbuild.LoaderTS,
		})
		require.Empty(t, result.Errors, scriptPath)

		_, err = vm.RunScript(scriptPath, string(result.Code))
		require.NoError(t, err, scriptPath)
	}
	return vm
}

func callJSON[T any](t *testing.T, vm *goja.Runtime, expression string, args ...any) T {
	t.Helper()

	fn, ok := goja.AssertFunction(vm.Get(expression))
	require.True(t, ok, "%s is not a function", expression)

	parse, ok := goja.AssertFunction(vm.Get("JSON").ToObject(vm).Get("parse"))
	require.True(t, ok)
	values := make([]goja.Value, 0, len(args))
	for _, arg := range args {
		raw, err := json.Marshal(arg)
		require.NoError(t, err)
		value, err := parse(goja.Undefined(), vm.ToValue(string(raw)))
		require.NoError(t, err)
		values = append(values, value)
	}

	result, err := fn(goja.Undefined(), values...)
	require.NoError(t, err)

	stringify, ok := goja.AssertFunction(vm.Get("JSON").ToObject(vm).Get("stringify"))
	require.True(t, ok)
	encoded, err := stringify(goja.Undefined(), result)
	require.NoError(t, err)

	var out T
	require.NoError(t, json.Unmarshal([]byte(encoded.String()), &out))
	return out
}

func TestResolveOperationShorthands(t *testing.T) {
	t.Parallel()

	noRetired := map[string]retiredShorthand{}

	tests := []struct {
		name        string
		input       shorthandInput
		wantFlags   map[string]string
		wantRetired map[string]retiredShorthand
	}{
		{
			name: "no lock assigns unique initials",
			input: shorthandInput{
				FlagNames: []string{"names", "nick-name", "tag", "owner", "address.city"},
			},
			wantFlags:   map[string]string{"tag": "t"},
			wantRetired: noRetired,
		},
		{
			name: "no lock prefers required path parameters and pins",
			input: shorthandInput{
				FlagNames: []string{"pet-id", "price", "names", "nick-name"},
				Preferred: []string{"pet-id"},
				Pinned:    []string{"names"},
			},
			wantFlags:   map[string]string{"names": "n", "pet-id": "p"},
			wantRetired: noRetired,
		},
		{
			name: "pin beats a required path parameter on the same initial",
			input: shorthandInput{
				FlagNames: []string{"pet-id", "price"},
				Preferred: []string{"pet-id"},
				Pinned:    []string{"price"},
			},
			wantFlags:   map[string]string{"price": "p"},
			wantRetired: noRetired,
		},
		{
			name: "lock keeps a letter when a new flag shares it",
			input: shorthandInput{
				FlagNames: []string{"names", "nick-name", "tag"},
				Lock:      &shorthandLock{Flags: map[string]string{"names": "n", "tag": "t"}},
			},
			wantFlags:   map[string]string{"names": "n", "tag": "t"},
			wantRetired: noRetired,
		},
		{
			name: "lock beats a new required path parameter",
			input: shorthandInput{
				FlagNames: []string{"price", "pet-id"},
				Preferred: []string{"pet-id"},
				Lock:      &shorthandLock{Flags: map[string]string{"price": "p"}},
			},
			wantFlags:   map[string]string{"price": "p"},
			wantRetired: noRetired,
		},
		{
			name: "pin takes a locked letter from another flag",
			input: shorthandInput{
				FlagNames: []string{"names", "nick-name"},
				Lock:      &shorthandLock{Flags: map[string]string{"names": "n"}},
				Pinned:    []string{"nick-name"},
			},
			wantFlags:   map[string]string{"nick-name": "n"},
			wantRetired: noRetired,
		},
		{
			name: "removed flag retires its letter",
			input: shorthandInput{
				FlagNames: []string{"nick-name", "tag"},
				Lock:      &shorthandLock{Flags: map[string]string{"names": "n", "tag": "t"}},
			},
			wantFlags: map[string]string{"tag": "t"},
			wantRetired: map[string]retiredShorthand{
				"n": {Flag: "names"},
			},
		},
		{
			name: "returning flag regains its retired letter",
			input: shorthandInput{
				FlagNames: []string{"names", "nick-name"},
				Lock: &shorthandLock{Retired: map[string]retiredShorthand{
					"n": {Flag: "names"},
				}},
			},
			wantFlags:   map[string]string{"names": "n"},
			wantRetired: noRetired,
		},
		{
			name: "pin reclaims a retired letter",
			input: shorthandInput{
				FlagNames: []string{"nick-name"},
				Lock: &shorthandLock{Retired: map[string]retiredShorthand{
					"n": {Flag: "names"},
				}},
				Pinned: []string{"nick-name"},
			},
			wantFlags:   map[string]string{"nick-name": "n"},
			wantRetired: noRetired,
		},
		{
			name: "retired letter stays reserved for a new flag",
			input: shorthandInput{
				FlagNames: []string{"nick-name"},
				Lock: &shorthandLock{Retired: map[string]retiredShorthand{
					"n": {Flag: "names"},
				}},
			},
			wantFlags: map[string]string{},
			wantRetired: map[string]retiredShorthand{
				"n": {Flag: "names"},
			},
		},
		{
			name: "letter deleted from retired goes to a new flag",
			input: shorthandInput{
				FlagNames: []string{"nick-name", "tag"},
				Lock:      &shorthandLock{Flags: map[string]string{"tag": "t"}},
			},
			wantFlags:   map[string]string{"nick-name": "n", "tag": "t"},
			wantRetired: noRetired,
		},
		{
			name: "letter that became reserved is dropped",
			input: shorthandInput{
				FlagNames: []string{"after", "tag"},
				Lock:      &shorthandLock{Flags: map[string]string{"after": "a", "tag": "t"}},
				Reserved:  []string{"a"},
			},
			wantFlags:   map[string]string{"tag": "t"},
			wantRetired: noRetired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			vm := loadCLIIncludes(t, "includes/flags.ts", "includes/metadata.ts")
			got := callJSON[shorthandResolution](t, vm, "resolveOperationShorthands", tt.input)
			assert.Equal(t, tt.wantFlags, got.Flags)
			assert.Equal(t, tt.wantRetired, got.Retired)
		})
	}
}

func TestResolveOperationShorthandsIgnoresMalformedLockEntries(t *testing.T) {
	t.Parallel()

	vm := loadCLIIncludes(t, "includes/flags.ts", "includes/metadata.ts")
	got := callJSON[shorthandResolution](t, vm, "resolveOperationShorthands", map[string]any{
		"flagNames": []string{"nick-name", "tag", "price", "count", "size", "every", "uri"},
		"lock": map[string]any{
			"flags": map[string]any{
				"count": 5,
				"size":  "ss",
				"every": "é",
			},
			"retired": map[string]any{
				"n":  nil,
				"t":  map[string]any{},
				"p":  "price",
				"cc": map[string]any{"flag": "count"},
				"ü":  map[string]any{"flag": "uri"},
			},
		},
	})
	assert.Equal(t, map[string]string{"count": "c", "every": "e", "nick-name": "n", "price": "p", "size": "s", "tag": "t", "uri": "u"}, got.Flags)
	assert.Equal(t, map[string]retiredShorthand{}, got.Retired)

	got = callJSON[shorthandResolution](t, vm, "resolveOperationShorthands", map[string]any{
		"flagNames": []string{"names", "tag"},
		"lock": map[string]any{
			"flags":   "xy",
			"retired": []string{"n"},
		},
	})
	assert.Equal(t, map[string]string{"names": "n", "tag": "t"}, got.Flags)
	assert.Equal(t, map[string]retiredShorthand{}, got.Retired)
}

func TestResolveOperationShorthandsReportsFlagsDisplacedByPins(t *testing.T) {
	t.Parallel()

	type displacedShorthand struct {
		Letter string `json:"letter"`
		Pin    string `json:"pin"`
	}
	type displacedResolution struct {
		Flags     map[string]string             `json:"flags"`
		Displaced map[string]displacedShorthand `json:"displaced"`
	}

	vm := loadCLIIncludes(t, "includes/flags.ts", "includes/metadata.ts")
	got := callJSON[displacedResolution](t, vm, "resolveOperationShorthands", shorthandInput{
		FlagNames: []string{"names", "nick-name", "tag", "title"},
		Lock:      &shorthandLock{Flags: map[string]string{"nick-name": "n", "tag": "t"}},
		Pinned:    []string{"names"},
	})
	assert.Equal(t, map[string]string{"names": "n", "tag": "t"}, got.Flags)
	assert.Equal(t, map[string]displacedShorthand{"nick-name": {Letter: "n", Pin: "names"}}, got.Displaced)

	got = callJSON[displacedResolution](t, vm, "resolveOperationShorthands", shorthandInput{
		FlagNames: []string{"names", "nick-name"},
		Lock:      &shorthandLock{Flags: map[string]string{"names": "n"}},
		Pinned:    []string{"names"},
	})
	assert.Equal(t, map[string]string{"names": "n"}, got.Flags)
	assert.Empty(t, got.Displaced)
}

// shorthandRuntimeHarness replaces the generator bindings and the
// metadata-entry builders, which need a full operation AST, with fakes driven
// by each test operation's flag names and required path parameters. Operation
// enumeration, lock lookup, path parameter preference, declaration matching
// and validation run unchanged.
const shorthandRuntimeHarness = `
var shorthandRecords = [];
var shorthandWarnings = [];
function recordCLIShorthands(operationID, method, resolution, split) {
  shorthandRecords.push({ operationID: operationID, method: method, flags: resolution.flags, retired: resolution.retired, split: split });
}
function logWarning(message) {
  shorthandWarnings.push(message);
}
function sanitizeFieldName(name) {
  return name;
}
function buildOperationMetadataEntries(op) {
  if (!op.TestFlags) return null;
  return {
    entries: op.TestFlags.map(function (name) {
      if (op.TestPathParams.indexOf(name) < 0) return 'FlagName: "' + name + '"';
      return 'FlagName: "' + name + '", FieldPath: "' + sanitizeFieldName(name) + '", Required: true';
    }),
    hadCandidateFields: op.TestFlags.length > 0,
  };
}
function pathParamField(name) {
  return {
    Name: name,
    Optional: false,
    Annotations: {
      Has: function (key) { return key === "param"; },
      Get: function () { return { ParamType: "pathParam" }; },
    },
  };
}
function operationGeneratedFlagNames(op) {
  var names = new Set(op.TestFlags || []);
  if (hasPagination(op)) {
    names.add("all");
    names.add("max-pages");
  }
  return names;
}
function templateBodyFieldPath() {
  return "";
}
function testOperation(spec) {
  var method = spec.method;
  var pathParams = spec.pathParams || [];
  return {
    OriginalID: spec.originalID,
    ID: spec.id || spec.originalID,
    GetID: function () { return this.ID; },
    SerializationMethod: method ? { toString: function () { return method; } } : undefined,
    Request: pathParams.length > 0
      ? { IsRequestBody: false, Field: { Type: { Fields: pathParams.map(pathParamField) } } }
      : { IsRequestBody: true },
    Extensions: spec.paginated ? { Pagination: {} } : undefined,
    TestFlags: spec.flags,
    TestPathParams: pathParams,
  };
}
var testOperations = [];
var context = {
  Global: {
    AST: {
      MainSDK: {
        Operations: [],
        SubSDKs: JSON.parse(testGroups).map(function (group) {
          var operations = group.map(testOperation);
          testOperations.push.apply(testOperations, operations);
          return { Operations: operations, SubSDKs: [] };
        }),
      },
      CLICommands: testManifest,
    },
    Config: { CLIShorthands: testLock },
  },
};
function findTestOperation(originalID, method) {
  var found = testOperations.filter(function (op) {
    return op.OriginalID === originalID && operationShorthandsMethod(op) === method;
  });
  if (found.length !== 1) {
    throw new Error("want one operation " + originalID + " " + method + ", found " + found.length);
  }
  return found[0];
}
function testOperationShorthands(originalID, method) {
  var op = findTestOperation(originalID, method);
  var shorthands = {};
  operationShorthands(op, buildOperationMetadataEntries(op).entries).forEach(function (letter, name) {
    shorthands[name] = letter;
  });
  return shorthands;
}
`

type testOperation struct {
	OriginalID string   `json:"originalID"`
	ID         string   `json:"id,omitempty"`
	Method     string   `json:"method,omitempty"`
	Flags      []string `json:"flags"`
	PathParams []string `json:"pathParams,omitempty"`
	Paginated  bool     `json:"paginated,omitempty"`
}

// shorthandFixture is an SDK tree of one sub-SDK per group, the CLI manifest
// operation declarations and the gen.lock cliShorthands section.
type shorthandFixture struct {
	groups       [][]testOperation
	declarations []extensions.CLIOperation
	lock         map[string]any
}

func newShorthandRuntime(t *testing.T, fixture shorthandFixture) *goja.Runtime {
	t.Helper()

	vm := loadTemplateScripts(t,
		"common/common/operations.ts",
		"cli/includes/flags.ts",
		"cli/includes/templating.ts",
		"cli/includes/metadata.ts",
		"cli/includes/intents.ts",
	)
	groups, err := json.Marshal(fixture.groups)
	require.NoError(t, err)
	require.NoError(t, vm.Set("testGroups", string(groups)))
	require.NoError(t, vm.Set("testManifest", &extensions.CLICommandManifest{Operations: fixture.declarations}))
	var lock any
	if fixture.lock != nil {
		lock = fixture.lock
	}
	require.NoError(t, vm.Set("testLock", lock))
	_, err = vm.RunString(shorthandRuntimeHarness)
	require.NoError(t, err)
	return vm
}

func evalJSON[T any](t *testing.T, vm *goja.Runtime, expression string) T {
	t.Helper()

	value, err := vm.RunString("JSON.stringify(" + expression + ")")
	require.NoError(t, err, expression)
	var out T
	require.NoError(t, json.Unmarshal([]byte(value.String()), &out), expression)
	return out
}

// collidingIDLayout is a CreatePet operation split into JSON and multipart
// commands, whose multipart command's generated ID (splitID, by default
// CreatePet_multipart) can equal the operationId of a separate
// CreatePet_multipart operation.
type collidingIDLayout struct {
	suffixedFirst  bool
	splitID        string
	jsonFlags      []string
	multipartFlags []string
	suffixedFlags  []string
	suffixedMethod string
}

func (l collidingIDLayout) groups() [][]testOperation {
	splitID := cmp.Or(l.splitID, "CreatePet_multipart")
	jsonFlags := l.jsonFlags
	if jsonFlags == nil {
		jsonFlags = []string{"bool"}
	}
	split := []testOperation{
		{OriginalID: "CreatePet", Method: "json", Flags: jsonFlags},
		{OriginalID: "CreatePet", ID: splitID, Method: "multipart", Flags: l.multipartFlags},
	}
	suffixed := []testOperation{
		{OriginalID: "CreatePet_multipart", Method: cmp.Or(l.suffixedMethod, "json"), Flags: l.suffixedFlags},
	}
	if l.suffixedFirst {
		return [][]testOperation{suffixed, split}
	}
	return [][]testOperation{split, suffixed}
}

var collidingIDOrders = map[string]bool{
	"suffixed operation first": true,
	"suffixed operation last":  false,
}

func lockFlags(flags map[string]any) map[string]any {
	return map[string]any{"flags": flags}
}

func TestOperationShorthandsMethod(t *testing.T) {
	t.Parallel()

	vm := newShorthandRuntime(t, shorthandFixture{groups: [][]testOperation{{
		{OriginalID: "ListPets"},
		{OriginalID: "CreatePet", Method: "json"},
		{OriginalID: "UpdatePet", Method: "form"},
		{OriginalID: "UploadPet", Method: "multipart"},
	}}})
	got := evalJSON[[]string](t, vm, "testOperations.map(operationShorthandsMethod)")
	assert.Equal(t, []string{"json", "json", "form", "multipart"}, got)
}

func TestOperationRequestMethods(t *testing.T) {
	t.Parallel()

	for name, suffixedFirst := range collidingIDOrders {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			vm := newShorthandRuntime(t, shorthandFixture{
				groups: collidingIDLayout{suffixedFirst: suffixedFirst, multipartFlags: []string{"names"}, suffixedFlags: []string{"names"}}.groups(),
			})
			got := evalJSON[map[string][]string](t, vm, `{
				CreatePet: Array.from(operationRequestMethods("CreatePet")).sort(),
				CreatePet_multipart: Array.from(operationRequestMethods("CreatePet_multipart")).sort(),
			}`)
			assert.Equal(t, map[string][]string{
				"CreatePet":           {"json", "multipart"},
				"CreatePet_multipart": {"json"},
			}, got)
		})
	}
}

func TestRecordedOperationShorthands(t *testing.T) {
	t.Parallel()

	splitCreatePet := [][]testOperation{{
		{OriginalID: "CreatePet", Method: "json", Flags: []string{"names"}},
		{OriginalID: "CreatePet", ID: "CreatePet_multipart", Method: "multipart", Flags: []string{"names"}},
	}}
	singleCreatePet := func(method string) [][]testOperation {
		return [][]testOperation{{{OriginalID: "CreatePet", Method: method, Flags: []string{"names"}}}}
	}
	names := lockFlags(map[string]any{"names": "n"})
	nova := lockFlags(map[string]any{"nova": "n"})

	type lookup struct {
		originalID string
		method     string
		want       map[string]any
	}
	type lookupCase struct {
		name    string
		groups  [][]testOperation
		lock    map[string]any
		lookups []lookup
	}
	collidingCases := make([]lookupCase, 0, len(collidingIDOrders))
	for orderName, suffixedFirst := range collidingIDOrders {
		collidingCases = append(collidingCases, lookupCase{
			name:   orderName + " colliding generated ID keeps each operation's history",
			groups: collidingIDLayout{suffixedFirst: suffixedFirst, multipartFlags: []string{"names"}, suffixedFlags: []string{"names"}, suffixedMethod: "form"}.groups(),
			lock: map[string]any{
				"CreatePet":           map[string]any{"multipart": nova},
				"CreatePet_multipart": map[string]any{"json": names},
			},
			lookups: []lookup{{"CreatePet", "multipart", nova}, {"CreatePet_multipart", "form", names}},
		})
	}
	tests := append([]lookupCase{
		{
			name:    "split command reads its own method",
			groups:  splitCreatePet,
			lock:    map[string]any{"CreatePet": map[string]any{"json": names, "multipart": nova}},
			lookups: []lookup{{"CreatePet", "json", names}, {"CreatePet", "multipart", nova}},
		},
		{
			name:    "split command does not inherit another method",
			groups:  splitCreatePet,
			lock:    map[string]any{"CreatePet": map[string]any{"json": names}},
			lookups: []lookup{{"CreatePet", "json", names}, {"CreatePet", "multipart", nil}},
		},
		{
			name:    "single command inherits the only recorded method",
			groups:  singleCreatePet("form"),
			lock:    map[string]any{"CreatePet": map[string]any{"json": names}},
			lookups: []lookup{{"CreatePet", "form", names}},
		},
		{
			name:    "single command does not inherit from several recorded methods",
			groups:  singleCreatePet("form"),
			lock:    map[string]any{"CreatePet": map[string]any{"json": names, "multipart": nova}},
			lookups: []lookup{{"CreatePet", "form", nil}},
		},
		{
			name:    "recorded empty entry wins over a sibling method",
			groups:  singleCreatePet("json"),
			lock:    map[string]any{"CreatePet": map[string]any{"json": map[string]any{}, "multipart": names}},
			lookups: []lookup{{"CreatePet", "json", map[string]any{}}},
		},
		{
			name:    "no lock",
			groups:  singleCreatePet("json"),
			lookups: []lookup{{"CreatePet", "json", nil}},
		},
		{
			name: "operation IDs naming inherited object properties",
			groups: [][]testOperation{{
				{OriginalID: "constructor", Method: "json", Flags: []string{"names"}},
				{OriginalID: "toString", Method: "json"},
			}},
			lock:    map[string]any{"constructor": map[string]any{"json": names}},
			lookups: []lookup{{"constructor", "json", names}, {"toString", "json", nil}},
		},
	}, collidingCases...)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			vm := newShorthandRuntime(t, shorthandFixture{groups: tt.groups, lock: tt.lock})
			for _, l := range tt.lookups {
				got := evalJSON[map[string]any](t, vm, fmt.Sprintf("recordedOperationShorthands(findTestOperation(%q, %q), %q)", l.originalID, l.method, l.method))
				assert.Equal(t, l.want, got, "%s %s", l.originalID, l.method)
			}
		})
	}
}

func TestOperationShorthandsKeepsOperationsApart(t *testing.T) {
	t.Parallel()

	type record struct {
		OperationID string            `json:"operationID"`
		Method      string            `json:"method"`
		Flags       map[string]string `json:"flags"`
		Split       bool              `json:"split"`
	}
	resolve := func(t *testing.T, vm *goja.Runtime, operations [][2]string) map[string]map[string]string {
		t.Helper()
		got := map[string]map[string]string{}
		for _, op := range operations {
			got[op[0]+" "+op[1]] = evalJSON[map[string]string](t, vm, fmt.Sprintf("testOperationShorthands(%q, %q)", op[0], op[1]))
		}
		return got
	}

	for _, splitID := range []string{"CreatePet_multipart", "CreatePetMultipart"} {
		for _, suffixedMethod := range []string{"json", "multipart"} {
			for orderName, suffixedFirst := range collidingIDOrders {
				t.Run(fmt.Sprintf("%s split command ID %s suffixed operation %s", orderName, splitID, suffixedMethod), func(t *testing.T) {
					t.Parallel()
					collidingFlags := []string{"names", "nick-name"}
					vm := newShorthandRuntime(t, shorthandFixture{
						groups: collidingIDLayout{
							suffixedFirst:  suffixedFirst,
							splitID:        splitID,
							jsonFlags:      collidingFlags,
							multipartFlags: collidingFlags,
							suffixedFlags:  collidingFlags,
							suffixedMethod: suffixedMethod,
						}.groups(),
						lock: map[string]any{
							"CreatePet": map[string]any{
								"json":      lockFlags(map[string]any{"nick-name": "n"}),
								"multipart": lockFlags(map[string]any{"names": "n"}),
							},
							"CreatePet_multipart": map[string]any{
								suffixedMethod: lockFlags(map[string]any{"nick-name": "n"}),
							},
						},
					})
					operations := [][2]string{{"CreatePet", "json"}, {"CreatePet", "multipart"}, {"CreatePet_multipart", suffixedMethod}}
					want := map[string]map[string]string{
						"CreatePet json":                        {"nick-name": "n"},
						"CreatePet multipart":                   {"names": "n"},
						"CreatePet_multipart " + suffixedMethod: {"nick-name": "n"},
					}
					assert.Equal(t, want, resolve(t, vm, operations))
					assert.Equal(t, want, resolve(t, vm, operations))
					assert.ElementsMatch(t, []record{
						{OperationID: "CreatePet", Method: "json", Flags: map[string]string{"nick-name": "n"}, Split: true},
						{OperationID: "CreatePet", Method: "multipart", Flags: map[string]string{"names": "n"}, Split: true},
						{OperationID: "CreatePet_multipart", Method: suffixedMethod, Flags: map[string]string{"nick-name": "n"}, Split: false},
					}, evalJSON[[]record](t, vm, "shorthandRecords"))
				})
			}
		}
	}

	t.Run("operation IDs naming inherited object properties", func(t *testing.T) {
		t.Parallel()
		vm := newShorthandRuntime(t, shorthandFixture{
			groups: [][]testOperation{{
				{OriginalID: "constructor", Method: "json", Flags: []string{"names", "nick-name"}},
				{OriginalID: "toString", Method: "json", Flags: []string{}},
			}},
			lock: map[string]any{"constructor": map[string]any{"json": lockFlags(map[string]any{"names": "n"})}},
		})
		assert.Equal(t, map[string]map[string]string{
			"constructor json": {"names": "n"},
			"toString json":    {},
		}, resolve(t, vm, [][2]string{{"constructor", "json"}, {"toString", "json"}}))
	})
}

func TestRecordOperationShorthandsRetiresOperationsWithoutFlags(t *testing.T) {
	t.Parallel()

	vm := newShorthandRuntime(t, shorthandFixture{
		groups: [][]testOperation{{{OriginalID: "ListPets"}}},
		lock:   map[string]any{"ListPets": map[string]any{"json": lockFlags(map[string]any{"names": "n"})}},
	})
	_, err := vm.RunString(`recordOperationShorthands(findTestOperation("ListPets", "json"))`)
	require.NoError(t, err)
	got := evalJSON[[]shorthandResolution](t, vm, "shorthandRecords")
	assert.Equal(t, []shorthandResolution{{
		Flags:   map[string]string{},
		Retired: map[string]retiredShorthand{"n": {Flag: "names"}},
	}}, got)
}

func TestOperationShorthandsPreferRequiredPathParameters(t *testing.T) {
	t.Parallel()

	updatePet := [][]testOperation{{{OriginalID: "UpdatePet", Method: "json", Flags: []string{"price", "pet-id"}, PathParams: []string{"pet-id"}}}}
	tests := map[string]struct {
		declarations []extensions.CLIOperation
		want         map[string]string
	}{
		"required path parameter wins its initial": {
			want: map[string]string{"pet-id": "p"},
		},
		"pin beats a required path parameter": {
			declarations: []extensions.CLIOperation{{OperationID: "UpdatePet", PreferShorthand: []string{"price"}}},
			want:         map[string]string{"price": "p"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			vm := newShorthandRuntime(t, shorthandFixture{groups: updatePet, declarations: tt.declarations})
			assert.Equal(t, tt.want, evalJSON[map[string]string](t, vm, `testOperationShorthands("UpdatePet", "json")`))
		})
	}
}

func TestOperationShorthandsWarnsOnceForDisplacedRecordedFlag(t *testing.T) {
	t.Parallel()

	vm := newShorthandRuntime(t, shorthandFixture{
		groups:       collidingIDLayout{multipartFlags: []string{"names", "nick-name"}, suffixedFlags: []string{"nick-name"}}.groups(),
		declarations: []extensions.CLIOperation{{OperationID: "CreatePet", PreferShorthand: []string{"names"}}},
		lock: map[string]any{
			"CreatePet":           map[string]any{"multipart": lockFlags(map[string]any{"nick-name": "n"})},
			"CreatePet_multipart": map[string]any{"json": lockFlags(map[string]any{"nick-name": "n"})},
		},
	})
	for range 2 {
		assert.Equal(t, map[string]string{"names": "n"}, evalJSON[map[string]string](t, vm, `testOperationShorthands("CreatePet", "multipart")`))
		assert.Equal(t, map[string]string{"bool": "b"}, evalJSON[map[string]string](t, vm, `testOperationShorthands("CreatePet", "json")`))
		assert.Equal(t, map[string]string{"nick-name": "n"}, evalJSON[map[string]string](t, vm, `testOperationShorthands("CreatePet_multipart", "json")`))
	}
	warnings := evalJSON[[]string](t, vm, "shorthandWarnings")
	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], `operation "CreatePet" (multipart request): preferShorthand "names" takes -n from "nick-name"`)
}

func TestCLIDeclaredOperations(t *testing.T) {
	t.Parallel()

	for orderName, suffixedFirst := range collidingIDOrders {
		t.Run(orderName, func(t *testing.T) {
			t.Parallel()
			groups := collidingIDLayout{suffixedFirst: suffixedFirst, multipartFlags: []string{"names"}, suffixedFlags: []string{"names"}}.groups()
			groups = append(groups,
				[]testOperation{{OriginalID: "ListPets", ID: "foo", Flags: []string{"offset"}, Paginated: true}},
				[]testOperation{{OriginalID: "foo", Flags: []string{"after", "around"}}},
			)
			vm := newShorthandRuntime(t, shorthandFixture{groups: groups})
			got := evalJSON[map[string][]string](t, vm, `(function () {
				var describe = function (op) { return op.OriginalID + " " + operationShorthandsMethod(op); };
				return {
					CreatePet: cliDeclaredOperations("CreatePet").map(describe).sort(),
					CreatePet_multipart: cliDeclaredOperations("CreatePet_multipart").map(describe),
					foo: cliDeclaredOperations("foo").map(describe),
					ListPets: cliDeclaredOperations("ListPets").map(describe),
				};
			})()`)
			assert.Equal(t, map[string][]string{
				"CreatePet":           {"CreatePet json", "CreatePet multipart"},
				"CreatePet_multipart": {"CreatePet_multipart json"},
				"foo":                 {"foo json"},
				"ListPets":            {"ListPets json"},
			}, got)
		})
	}
}

func TestValidateCLIOperationDeclarations(t *testing.T) {
	t.Parallel()

	splitCreatePet := [][]testOperation{{
		{OriginalID: "CreatePet", Method: "json", Flags: []string{"bark", "bool"}},
		{OriginalID: "CreatePet", ID: "CreatePet_multipart", Method: "multipart", Flags: []string{"num"}},
	}}
	sortAndSize := [][]testOperation{{{OriginalID: "CreatePet", Method: "json", Flags: []string{"name", "sort", "size"}}}}
	summaryFlag := []extensions.CLICommandInput{{Name: "summary", Shorthand: "s"}}

	type declarationCase struct {
		name         string
		groups       [][]testOperation
		declarations []extensions.CLIOperation
		wantErr      string
	}
	collidingCases := make([]declarationCase, 0, 2*len(collidingIDOrders))
	for orderName, suffixedFirst := range collidingIDOrders {
		groups := collidingIDLayout{suffixedFirst: suffixedFirst, multipartFlags: []string{"tag"}, suffixedFlags: []string{"names", "nick-name"}}.groups()
		collidingCases = append(collidingCases,
			declarationCase{
				name:         orderName + " pin on the declared operation",
				groups:       groups,
				declarations: []extensions.CLIOperation{{OperationID: "CreatePet_multipart", PreferShorthand: []string{"names"}}},
			},
			declarationCase{
				name:         orderName + " pin on a flag of the command sharing its generated ID",
				groups:       groups,
				declarations: []extensions.CLIOperation{{OperationID: "CreatePet_multipart", PreferShorthand: []string{"tag"}}},
				wantErr:      `operation "CreatePet_multipart" preferShorthand "tag" does not match any generated flag on that operation`,
			},
		)
	}
	tests := append([]declarationCase{
		{
			name:         "pins on flags of different content-type variants",
			groups:       splitCreatePet,
			declarations: []extensions.CLIOperation{{OperationID: "CreatePet", PreferShorthand: []string{"bool", "num"}}},
		},
		{
			name:         "pin on no content-type variant",
			groups:       splitCreatePet,
			declarations: []extensions.CLIOperation{{OperationID: "CreatePet", PreferShorthand: []string{"missing"}}},
			wantErr:      `operation "CreatePet" preferShorthand "missing" does not match any generated flag on that operation`,
		},
		{
			name:         "undeclared operation",
			groups:       splitCreatePet,
			declarations: []extensions.CLIOperation{{OperationID: "CreatePet_multipart"}},
			wantErr:      `operation "CreatePet_multipart" is not part of the generated CLI surface`,
		},
		{
			name:         "pin starting with a on a paginated operation",
			groups:       [][]testOperation{{{OriginalID: "ListPets", Flags: []string{"after"}, Paginated: true}}},
			declarations: []extensions.CLIOperation{{OperationID: "ListPets", PreferShorthand: []string{"after"}}},
			wantErr:      `operation "ListPets" preferShorthand "after" cannot take -a`,
		},
		{
			name: "pin applies to the operationId, not a name override matching it",
			groups: [][]testOperation{
				{{OriginalID: "ListPets", ID: "foo", Flags: []string{"offset"}, Paginated: true}},
				{{OriginalID: "foo", Flags: []string{"after", "around"}}},
			},
			declarations: []extensions.CLIOperation{{OperationID: "foo", PreferShorthand: []string{"after"}}},
		},
		{
			name:         "declared shorthand of colliding generated flags",
			groups:       sortAndSize,
			declarations: []extensions.CLIOperation{{OperationID: "CreatePet", Flags: summaryFlag}},
		},
		{
			name: "declared shorthand collides with an auto-assigned letter of the second variant",
			groups: [][]testOperation{{
				{OriginalID: "CreatePet", Method: "json", Flags: []string{"bool"}},
				{OriginalID: "CreatePet", ID: "CreatePet_multipart", Method: "multipart", Flags: []string{"names"}},
			}},
			declarations: []extensions.CLIOperation{{OperationID: "CreatePet", Flags: []extensions.CLICommandInput{{Name: "extra", Shorthand: "n"}}}},
			wantErr:      `declares shorthand "-n" on flag "--extra", which collides with the auto-assigned shorthand of generated flag "--names"`,
		},
		{
			name: "declared shorthand collides with an auto-assigned letter of the first variant",
			groups: [][]testOperation{{
				{OriginalID: "CreatePet", Method: "json", Flags: []string{"bool"}},
				{OriginalID: "CreatePet", ID: "CreatePet_multipart", Method: "multipart", Flags: []string{"names"}},
			}},
			declarations: []extensions.CLIOperation{{OperationID: "CreatePet", Flags: []extensions.CLICommandInput{{Name: "extra", Shorthand: "b"}}}},
			wantErr:      `declares shorthand "-b" on flag "--extra", which collides with the auto-assigned shorthand of generated flag "--bool"`,
		},
		{
			name:         "declared shorthand collides with a pinned flag",
			groups:       sortAndSize,
			declarations: []extensions.CLIOperation{{OperationID: "CreatePet", PreferShorthand: []string{"sort"}, Flags: summaryFlag}},
			wantErr:      `declares shorthand "-s" on flag "--summary", which collides with the auto-assigned shorthand of generated flag "--sort"`,
		},
	}, collidingCases...)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			vm := newShorthandRuntime(t, shorthandFixture{groups: tt.groups, declarations: tt.declarations})
			_, err := vm.RunString("validateCLIOperationDeclarations(context.Global.AST.CLICommands)")
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}
