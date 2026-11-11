package templates

import (
	"encoding/json"
	"path"
	"testing"

	"github.com/dop251/goja"
	esbuild "github.com/evanw/esbuild/pkg/api"
	embedded "github.com/speakeasy-api/openapi-generation/v2/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type retiredShorthand struct {
	Flag    string `json:"flag"`
	FreedIn string `json:"freedIn"`
}

type shorthandLock struct {
	Flags   map[string]string           `json:"flags,omitempty"`
	Retired map[string]retiredShorthand `json:"retired,omitempty"`
}

type shorthandInput struct {
	FlagNames  []string       `json:"flagNames"`
	Lock       *shorthandLock `json:"lock,omitempty"`
	Pinned     []string       `json:"pinned,omitempty"`
	Preferred  []string       `json:"preferred,omitempty"`
	Reserved   []string       `json:"reserved,omitempty"`
	SDKVersion string         `json:"sdkVersion"`
}

type shorthandResolution struct {
	Flags   map[string]string           `json:"flags"`
	Retired map[string]retiredShorthand `json:"retired"`
}

// loadCLIIncludes evaluates CLI include files in one runtime, stubbing the
// template registration the generator's engine normally provides.
func loadCLIIncludes(t *testing.T, files ...string) *goja.Runtime {
	t.Helper()

	vm := goja.New()
	require.NoError(t, vm.Set("registerTemplateFunc", func(goja.FunctionCall) goja.Value { return goja.Undefined() }))
	for _, file := range files {
		scriptPath := path.Join("templates", "cli", file)
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
				FlagNames:  []string{"names", "nick-name", "tag", "owner", "address.city"},
				SDKVersion: "1.0.0",
			},
			wantFlags:   map[string]string{"tag": "t"},
			wantRetired: noRetired,
		},
		{
			name: "no lock prefers required path parameters and pins",
			input: shorthandInput{
				FlagNames:  []string{"pet-id", "price", "names", "nick-name"},
				Preferred:  []string{"pet-id"},
				Pinned:     []string{"names"},
				SDKVersion: "1.0.0",
			},
			wantFlags:   map[string]string{"names": "n", "pet-id": "p"},
			wantRetired: noRetired,
		},
		{
			name: "lock keeps a letter when a new flag shares it",
			input: shorthandInput{
				FlagNames:  []string{"names", "nick-name", "tag"},
				Lock:       &shorthandLock{Flags: map[string]string{"names": "n", "tag": "t"}},
				SDKVersion: "1.1.0",
			},
			wantFlags:   map[string]string{"names": "n", "tag": "t"},
			wantRetired: noRetired,
		},
		{
			name: "lock beats a new required path parameter",
			input: shorthandInput{
				FlagNames:  []string{"price", "pet-id"},
				Preferred:  []string{"pet-id"},
				Lock:       &shorthandLock{Flags: map[string]string{"price": "p"}},
				SDKVersion: "1.1.0",
			},
			wantFlags:   map[string]string{"price": "p"},
			wantRetired: noRetired,
		},
		{
			name: "pin takes a locked letter from another flag",
			input: shorthandInput{
				FlagNames:  []string{"names", "nick-name"},
				Lock:       &shorthandLock{Flags: map[string]string{"names": "n"}},
				Pinned:     []string{"nick-name"},
				SDKVersion: "1.1.0",
			},
			wantFlags:   map[string]string{"nick-name": "n"},
			wantRetired: noRetired,
		},
		{
			name: "removed flag retires its letter",
			input: shorthandInput{
				FlagNames:  []string{"nick-name", "tag"},
				Lock:       &shorthandLock{Flags: map[string]string{"names": "n", "tag": "t"}},
				SDKVersion: "1.4.0",
			},
			wantFlags: map[string]string{"tag": "t"},
			wantRetired: map[string]retiredShorthand{
				"n": {Flag: "names", FreedIn: "1.4.0"},
			},
		},
		{
			name: "returning flag regains its retired letter",
			input: shorthandInput{
				FlagNames: []string{"names", "nick-name"},
				Lock: &shorthandLock{Retired: map[string]retiredShorthand{
					"n": {Flag: "names", FreedIn: "1.4.0"},
				}},
				SDKVersion: "1.5.0",
			},
			wantFlags:   map[string]string{"names": "n"},
			wantRetired: noRetired,
		},
		{
			name: "pin reclaims a retired letter",
			input: shorthandInput{
				FlagNames: []string{"nick-name"},
				Lock: &shorthandLock{Retired: map[string]retiredShorthand{
					"n": {Flag: "names", FreedIn: "1.4.0"},
				}},
				Pinned:     []string{"nick-name"},
				SDKVersion: "1.5.0",
			},
			wantFlags:   map[string]string{"nick-name": "n"},
			wantRetired: noRetired,
		},
		{
			name: "retired letter stays reserved within a major version",
			input: shorthandInput{
				FlagNames: []string{"nick-name"},
				Lock: &shorthandLock{Retired: map[string]retiredShorthand{
					"n": {Flag: "names", FreedIn: "1.4.0"},
				}},
				SDKVersion: "1.9.3",
			},
			wantFlags: map[string]string{},
			wantRetired: map[string]retiredShorthand{
				"n": {Flag: "names", FreedIn: "1.4.0"},
			},
		},
		{
			name: "major bump frees a retired letter",
			input: shorthandInput{
				FlagNames: []string{"nick-name"},
				Lock: &shorthandLock{Retired: map[string]retiredShorthand{
					"n": {Flag: "names", FreedIn: "1.4.0"},
				}},
				SDKVersion: "2.0.0",
			},
			wantFlags:   map[string]string{"nick-name": "n"},
			wantRetired: noRetired,
		},
		{
			name: "pre-1.0 minor bump frees a retired letter",
			input: shorthandInput{
				FlagNames: []string{"nick-name"},
				Lock: &shorthandLock{Retired: map[string]retiredShorthand{
					"n": {Flag: "names", FreedIn: "0.3.0"},
				}},
				SDKVersion: "0.4.0",
			},
			wantFlags:   map[string]string{"nick-name": "n"},
			wantRetired: noRetired,
		},
		{
			name: "pre-1.0 patch bump keeps a retired letter",
			input: shorthandInput{
				FlagNames: []string{"nick-name"},
				Lock: &shorthandLock{Retired: map[string]retiredShorthand{
					"n": {Flag: "names", FreedIn: "0.3.0"},
				}},
				SDKVersion: "0.3.5",
			},
			wantFlags: map[string]string{},
			wantRetired: map[string]retiredShorthand{
				"n": {Flag: "names", FreedIn: "0.3.0"},
			},
		},
		{
			name: "empty version keeps retired letters reserved",
			input: shorthandInput{
				FlagNames: []string{"nick-name", "owner-name"},
				Lock: &shorthandLock{
					Flags: map[string]string{"price": "p"},
					Retired: map[string]retiredShorthand{
						"n": {Flag: "names", FreedIn: "1.4.0"},
					},
				},
				SDKVersion: "",
			},
			wantFlags: map[string]string{},
			wantRetired: map[string]retiredShorthand{
				"n": {Flag: "names", FreedIn: "1.4.0"},
				"p": {Flag: "price", FreedIn: ""},
			},
		},
		{
			name: "known version stamps a retired letter that has no release",
			input: shorthandInput{
				FlagNames: []string{"nick-name"},
				Lock: &shorthandLock{Retired: map[string]retiredShorthand{
					"n": {Flag: "names", FreedIn: ""},
				}},
				SDKVersion: "2.0.0",
			},
			wantFlags: map[string]string{},
			wantRetired: map[string]retiredShorthand{
				"n": {Flag: "names", FreedIn: "2.0.0"},
			},
		},
		{
			name: "letter that became reserved is dropped",
			input: shorthandInput{
				FlagNames:  []string{"after", "tag"},
				Lock:       &shorthandLock{Flags: map[string]string{"after": "a", "tag": "t"}},
				Reserved:   []string{"a"},
				SDKVersion: "1.1.0",
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
