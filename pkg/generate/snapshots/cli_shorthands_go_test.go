package snapshots

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

const cliShorthandGenYaml = `cli:
  packageName: github.com/example/pet-cli
  cliName: pet-cli
  envVarPrefix: PET
`

// cliShorthandContent is an object request body; a field is a string unless
// suffixed with ":boolean" or ":array".
type cliShorthandContent struct {
	contentType string
	fields      []string
}

func jsonBody(fields ...string) cliShorthandContent {
	return cliShorthandContent{"application/json", fields}
}

func formBody(fields ...string) cliShorthandContent {
	return cliShorthandContent{"application/x-www-form-urlencoded", fields}
}

func multipartBody(fields ...string) cliShorthandContent {
	return cliShorthandContent{"multipart/form-data", fields}
}

type cliShorthandOperation struct {
	path         string
	method       string
	operationID  string
	tag          string
	description  string
	pathParam    string
	query        []string
	body         []cliShorthandContent
	arrayBody    bool
	nameOverride string
	paginated    bool
}

func (o cliShorthandOperation) yaml() string {
	var b strings.Builder
	fmt.Fprintf(&b, "  %s:\n    %s:\n      operationId: %s\n      description: %s\n      tags: [%s]\n      security: []\n",
		o.path, o.method, o.operationID, o.description, o.tag)
	if o.nameOverride != "" {
		fmt.Fprintf(&b, "      x-speakeasy-name-override: %s\n", o.nameOverride)
	}
	if o.pathParam != "" || len(o.query) > 0 || o.paginated {
		b.WriteString("      parameters:\n")
		if o.pathParam != "" {
			fmt.Fprintf(&b, "        - {name: %s, in: path, required: true, schema: {type: string}}\n", o.pathParam)
		}
		for _, name := range o.query {
			fmt.Fprintf(&b, "        - {name: %s, in: query, schema: {type: string}}\n", name)
		}
		if o.paginated {
			b.WriteString("        - {name: offset, in: query, schema: {type: integer}}\n")
		}
	}
	if o.paginated {
		b.WriteString("      x-speakeasy-pagination:\n        type: offsetLimit\n        inputs:\n          - {name: offset, in: parameters, type: offset}\n        outputs:\n          results: $.items\n")
	}
	switch {
	case o.arrayBody:
		b.WriteString("      requestBody:\n        required: true\n        content:\n          application/json:\n            schema: {type: array, items: {type: string}}\n")
	case len(o.body) > 0:
		b.WriteString("      requestBody:\n        required: true\n        content:\n")
		for _, content := range o.body {
			fmt.Fprintf(&b, "          %s:\n            schema:\n              type: object\n              properties:\n", content.contentType)
			for _, field := range content.fields {
				name, fieldType, _ := strings.Cut(field, ":")
				switch fieldType {
				case "":
					fmt.Fprintf(&b, "                %s: {type: string}\n", name)
				case "array":
					fmt.Fprintf(&b, "                %s: {type: array, items: {type: string}}\n", name)
				default:
					fmt.Fprintf(&b, "                %s: {type: %s}\n", name, fieldType)
				}
			}
		}
	}
	b.WriteString("      responses:\n        '200':\n          description: OK\n")
	if o.paginated {
		b.WriteString("          content:\n            application/json:\n              schema:\n                type: object\n                properties:\n                  items: {type: array, items: {type: string}}\n")
	}
	return b.String()
}

func cliShorthandSpec(manifest string, operations ...cliShorthandOperation) string {
	var b strings.Builder
	b.WriteString("openapi: 3.0.3\ninfo:\n  title: Pet API\n  version: 1.0.0\nservers:\n  - url: https://api.example.com\n")
	b.WriteString(manifest)
	b.WriteString("paths:\n")
	for _, op := range operations {
		b.WriteString(op.yaml())
	}
	return b.String()
}

func cliShorthandManifest(pins map[string]string) string {
	var b strings.Builder
	b.WriteString("x-speakeasy-cli-commands:\n  version: 1\n  operations:\n")
	for _, operationID := range slices.Sorted(maps.Keys(pins)) {
		fmt.Fprintf(&b, "    %s:\n      preferShorthand: [%s]\n", operationID, pins[operationID])
	}
	return b.String()
}

type cliShorthandEntry struct {
	Flags   map[string]string `yaml:"flags"`
	Retired map[string]struct {
		Flag string `yaml:"flag"`
	} `yaml:"retired"`
}

// cliShorthandProject generates a CLI into one directory across the explicit
// steps of a test, so each step reads the gen.lock the previous one wrote.
type cliShorthandProject struct {
	t   *testing.T
	dir string
}

func newCLIShorthandProject(t *testing.T) *cliShorthandProject {
	t.Helper()
	dir := t.TempDir()
	snaptest.WriteGenYaml(t, dir, cliShorthandGenYaml)
	return &cliShorthandProject{t: t, dir: dir}
}

func (p *cliShorthandProject) run(spec string, opts ...generate.GeneratorOptions) ([]error, []error) {
	p.t.Helper()
	generator, err := generate.New(append([]generate.GeneratorOptions{generate.WithLogger(logging.NewLogger(zapcore.WarnLevel))}, opts...)...)
	require.NoError(p.t, err)
	ctx := generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL())
	errs := generator.Generate(ctx, []byte(spec), "test-schema.yaml", "cli", p.dir, false, false)
	return errs, generator.GetWarnings()
}

// generate returns the warnings of a generation that must succeed.
func (p *cliShorthandProject) generate(spec string, opts ...generate.GeneratorOptions) []string {
	p.t.Helper()
	errs, warnings := p.run(spec, opts...)
	require.Empty(p.t, errs)
	messages := make([]string, 0, len(warnings))
	for _, warning := range warnings {
		messages = append(messages, warning.Error())
	}
	return messages
}

func (p *cliShorthandProject) generateFails(spec, wantErr string) {
	p.t.Helper()
	errs, _ := p.run(spec)
	require.Len(p.t, errs, 1)
	require.ErrorContains(p.t, errs[0], wantErr)
}

func (p *cliShorthandProject) read(elem ...string) string {
	p.t.Helper()
	content, err := os.ReadFile(filepath.Join(append([]string{p.dir}, elem...)...))
	require.NoError(p.t, err)
	return string(content)
}

// usageFlag returns the usage spec of one flag, such as "-n --names <names>",
// from the usage schema entry of one command, keyed by its command path, so
// assertions cannot match another command's flags.
func (p *cliShorthandProject) usageFlag(commandPath, flag string) []string {
	p.t.Helper()
	prefix := strconv.Quote(commandPath) + ":"
	item := regexp.MustCompile(`\{key: "` + regexp.QuoteMeta(flag) + `", text: "\s*flag \\"([^\\]*)\\"`)
	for _, line := range strings.Split(p.read("internal", "usage", "schema.go"), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), prefix) {
			continue
		}
		match := item.FindStringSubmatch(line)
		require.NotNil(p.t, match, "usage flag %s not found on %s", flag, commandPath)
		return strings.Fields(match[1])
	}
	require.FailNow(p.t, "usage entry not found", commandPath)
	return nil
}

// requireShorthands checks each flag's shorthand, "" for none, in both the
// command's flag metadata and its usage entry.
func (p *cliShorthandProject) requireShorthands(commandFile []string, commandPath string, want map[string]string) {
	p.t.Helper()
	command := p.read(append([]string{"internal", "cli"}, commandFile...)...)
	for flag, letter := range want {
		usage := p.usageFlag(commandPath, flag)
		if letter == "" {
			require.Contains(p.t, command, fmt.Sprintf("FlagName: %q", flag), commandPath)
			require.NotContains(p.t, command, fmt.Sprintf("FlagName: %q, Shorthand:", flag), commandPath)
			require.Equal(p.t, "--"+flag, usage[0], commandPath)
			continue
		}
		require.Contains(p.t, command, fmt.Sprintf("FlagName: %q, Shorthand: %q", flag, letter), commandPath)
		require.Equal(p.t, []string{"-" + letter, "--" + flag}, usage[:2], commandPath)
	}
}

func (p *cliShorthandProject) lock() map[string]map[string]cliShorthandEntry {
	p.t.Helper()
	var lock struct {
		CLIShorthands map[string]map[string]cliShorthandEntry `yaml:"cliShorthands"`
	}
	require.NoError(p.t, yaml.Unmarshal([]byte(p.read(".speakeasy", "gen.lock")), &lock))
	return lock.CLIShorthands
}

// lockFlags returns the recorded flags of every operation and method,
// requiring that none retired a letter.
func (p *cliShorthandProject) lockFlags() map[string]map[string]map[string]string {
	p.t.Helper()
	flags := map[string]map[string]map[string]string{}
	for operationID, methods := range p.lock() {
		flags[operationID] = map[string]map[string]string{}
		for method, entry := range methods {
			require.Empty(p.t, entry.Retired, "%s %s", operationID, method)
			flags[operationID][method] = entry.Flags
		}
	}
	return flags
}

func (p *cliShorthandProject) deleteRetired(operationID, method, letter string) {
	p.t.Helper()
	var lock map[string]any
	require.NoError(p.t, yaml.Unmarshal([]byte(p.read(".speakeasy", "gen.lock")), &lock))
	entry := lock["cliShorthands"].(map[string]any)[operationID].(map[string]any)[method].(map[string]any)
	retired := entry["retired"].(map[string]any)
	require.Contains(p.t, retired, letter)
	delete(retired, letter)
	content, err := yaml.Marshal(lock)
	require.NoError(p.t, err)
	require.NoError(p.t, os.WriteFile(filepath.Join(p.dir, ".speakeasy", "gen.lock"), content, 0o644))
}

func cliShorthandCreatePet(body ...cliShorthandContent) cliShorthandOperation {
	return cliShorthandOperation{
		path:        "/pets",
		method:      "post",
		operationID: "CreatePet",
		tag:         "pets",
		description: "Request content types render as one command each, with separate shorthand histories.",
		body:        body,
	}
}

func cliShorthandSuffixedOperation(body ...cliShorthandContent) cliShorthandOperation {
	return cliShorthandOperation{
		path:        "/notes",
		method:      "post",
		operationID: "CreatePet_multipart",
		tag:         "notes",
		description: "An operationId equal to the generated ID of another operation's multipart command.",
		body:        body,
	}
}

var (
	cliShorthandCreatePetCommand          = []string{"pets", "createpet.go"}
	cliShorthandCreatePetMultipartCommand = []string{"pets", "createpetmultipart.go"}
	cliShorthandSuffixedCommand           = []string{"notes", "createpetmultipart.go"}
)

func TestSnapCLIShorthands(t *testing.T) {
	t.Parallel()

	t.Run("identity and serialization", func(t *testing.T) {
		t.Parallel()
		p := newCLIShorthandProject(t)

		p.generate(cliShorthandSpec("",
			cliShorthandCreatePet(multipartBody("names:array", "tag:boolean")),
			cliShorthandSuffixedOperation(jsonBody("names")),
		))
		p.requireShorthands(cliShorthandCreatePetCommand, "pets create", map[string]string{"names": "n", "tag": "t"})
		p.requireShorthands(cliShorthandSuffixedCommand, "notes create-pet-multipart", map[string]string{"names": "n"})
		require.Equal(t, map[string]map[string]map[string]string{
			"CreatePet":           {"multipart": {"names": "n", "tag": "t"}},
			"CreatePet_multipart": {"json": {"names": "n"}},
		}, p.lockFlags())

		split := cliShorthandSpec("",
			cliShorthandCreatePet(jsonBody("names:array", "nick_name"), multipartBody("names:array", "nick_name", "tag:boolean")),
			cliShorthandSuffixedOperation(formBody("names", "nick_name")),
		)
		p.generate(split)
		p.requireShorthands(cliShorthandCreatePetCommand, "pets create", map[string]string{"names": "", "nick-name": ""})
		p.requireShorthands(cliShorthandCreatePetMultipartCommand, "pets create-pet-multipart", map[string]string{"names": "n", "nick-name": "", "tag": "t"})
		p.requireShorthands(cliShorthandSuffixedCommand, "notes create-pet-multipart", map[string]string{"names": "n", "nick-name": ""})
		wantLock := map[string]map[string]map[string]string{
			"CreatePet":           {"json": nil, "multipart": {"names": "n", "tag": "t"}},
			"CreatePet_multipart": {"form": {"names": "n"}},
		}
		require.Equal(t, wantLock, p.lockFlags())

		readOutputs := func() []string {
			t.Helper()
			return []string{
				p.read(append([]string{"internal", "cli"}, cliShorthandCreatePetCommand...)...),
				p.read(append([]string{"internal", "cli"}, cliShorthandCreatePetMultipartCommand...)...),
				p.read(append([]string{"internal", "cli"}, cliShorthandSuffixedCommand...)...),
				p.read("internal", "usage", "schema.go"),
			}
		}
		outputs := readOutputs()
		p.generate(split)
		require.Equal(t, outputs, readOutputs())
		require.Equal(t, wantLock, p.lockFlags())

		p.generate(cliShorthandSpec("",
			cliShorthandCreatePet(jsonBody("names:array", "nick_name", "filter")),
			cliShorthandSuffixedOperation(formBody("names", "nick_name")),
		))
		p.requireShorthands(cliShorthandCreatePetCommand, "pets create", map[string]string{"filter": "f", "names": "", "nick-name": ""})
		require.Equal(t, map[string]map[string]map[string]string{
			"CreatePet":           {"json": {"filter": "f"}},
			"CreatePet_multipart": {"form": {"names": "n"}},
		}, p.lockFlags())
	})

	t.Run("pins", func(t *testing.T) {
		t.Parallel()
		p := newCLIShorthandProject(t)
		note := func(fields ...string) cliShorthandOperation {
			op := cliShorthandSuffixedOperation(jsonBody(fields...))
			op.path = "/notes/{note_id}"
			op.pathParam = "note_id"
			return op
		}
		renamed := cliShorthandOperation{
			path:         "/pets/search",
			method:       "get",
			operationID:  "ListPets",
			tag:          "pets",
			description:  "A paginated operation whose method name equals another operation's operationId.",
			nameOverride: "foo",
			paginated:    true,
		}
		foo := cliShorthandOperation{
			path:        "/notes/search",
			method:      "get",
			operationID: "foo",
			tag:         "notes",
			description: "An operation without pagination whose operationId equals another operation's method name.",
			query:       []string{"after", "around"},
		}

		p.generate(cliShorthandSpec("",
			cliShorthandCreatePet(jsonBody("bark", "bool"), multipartBody("nick_name")),
			note("nick_name", "tag"),
			renamed,
			foo,
		))
		require.Equal(t, map[string]map[string]map[string]string{
			"CreatePet":           {"json": nil, "multipart": {"nick-name": "n"}},
			"CreatePet_multipart": {"json": {"note-id": "n", "tag": "t"}},
		}, p.lockFlags())

		warnings := p.generate(cliShorthandSpec(
			cliShorthandManifest(map[string]string{"CreatePet": "bark", "CreatePet_multipart": "names", "foo": "after"}),
			cliShorthandCreatePet(jsonBody("bark", "bool"), multipartBody("names", "nick_name")),
			note("names", "nick_name", "tag"),
			renamed,
			foo,
		))
		p.requireShorthands(cliShorthandCreatePetCommand, "pets create", map[string]string{"bark": "b", "bool": ""})
		p.requireShorthands(cliShorthandCreatePetMultipartCommand, "pets create-pet-multipart", map[string]string{"names": "", "nick-name": "n"})
		p.requireShorthands(cliShorthandSuffixedCommand, "notes create-pet-multipart", map[string]string{"names": "n", "nick-name": "", "note-id": "", "tag": "t"})
		p.requireShorthands([]string{"notes", "foo.go"}, "notes foo", map[string]string{"after": "a", "around": ""})
		var pinWarnings []string
		for _, warning := range warnings {
			if strings.Contains(warning, "preferShorthand") {
				pinWarnings = append(pinWarnings, warning)
			}
		}
		require.Len(t, pinWarnings, 1)
		require.Contains(t, pinWarnings[0], `operation "CreatePet_multipart" (json request): preferShorthand "names" takes -n from "note-id"`)
		require.Equal(t, map[string]map[string]map[string]string{
			"CreatePet":           {"json": {"bark": "b"}, "multipart": {"nick-name": "n"}},
			"CreatePet_multipart": {"json": {"names": "n", "tag": "t"}},
			"foo":                 {"json": {"after": "a"}},
		}, p.lockFlags())

		p.generateFails(cliShorthandSpec(
			cliShorthandManifest(map[string]string{"CreatePet_multipart": "photo"}),
			cliShorthandCreatePet(jsonBody("bark", "bool"), multipartBody("names", "nick_name", "photo")),
			note("names", "nick_name", "tag"),
			renamed,
			foo,
		), `operation "CreatePet_multipart" preferShorthand "photo" does not match any generated flag on that operation`)
	})

	t.Run("usage generation", func(t *testing.T) {
		t.Parallel()
		p := newCLIShorthandProject(t)

		p.generate(cliShorthandSpec("", cliShorthandCreatePet(jsonBody("names"))))
		require.Equal(t, map[string]map[string]map[string]string{"CreatePet": {"json": {"names": "n"}}}, p.lockFlags())

		p.generate(cliShorthandSpec("", cliShorthandCreatePet(jsonBody("names", "nick_name", "tag"))), generate.WithOutputUsageGroup("all"))
		p.requireShorthands(cliShorthandCreatePetCommand, "pets create", map[string]string{"names": "n", "nick-name": "", "tag": "t"})
		recorded := p.lock()
		require.Equal(t, map[string]string{"names": "n", "tag": "t"}, recorded["CreatePet"]["json"].Flags)

		p.generate(cliShorthandSpec("", cliShorthandCreatePet(jsonBody("nick_name"))), generate.WithUsageSnippetArgsGenerateAll())
		require.Equal(t, recorded, p.lock())
	})

	t.Run("lost inputs", func(t *testing.T) {
		t.Parallel()
		p := newCLIShorthandProject(t)
		listPets := func(query ...string) cliShorthandOperation {
			return cliShorthandOperation{
				path:        "/pets",
				method:      "get",
				operationID: "ListPets",
				tag:         "pets",
				description: "Query parameters that can all be removed, leaving the operation without inputs.",
				query:       query,
			}
		}
		createPet := func(body ...cliShorthandContent) cliShorthandOperation {
			op := cliShorthandCreatePet(body...)
			op.path = "/pets/new"
			return op
		}

		p.generate(cliShorthandSpec("", listPets("names"), createPet(jsonBody("names", "tag"))))
		require.Equal(t, map[string]map[string]map[string]string{
			"ListPets":  {"json": {"names": "n"}},
			"CreatePet": {"json": {"names": "n", "tag": "t"}},
		}, p.lockFlags())

		wholeBody := createPet()
		wholeBody.arrayBody = true
		p.generate(cliShorthandSpec("", listPets(), wholeBody))
		lock := p.lock()
		require.Empty(t, lock["ListPets"]["json"].Flags)
		require.Equal(t, "names", lock["ListPets"]["json"].Retired["n"].Flag)
		require.Empty(t, lock["CreatePet"]["json"].Flags)
		require.Equal(t, "names", lock["CreatePet"]["json"].Retired["n"].Flag)
		require.Equal(t, "tag", lock["CreatePet"]["json"].Retired["t"].Flag)

		p.deleteRetired("CreatePet", "json", "n")
		p.generate(cliShorthandSpec("", listPets("nick_name"), createPet(jsonBody("nick_name", "tag"))))
		p.requireShorthands([]string{"pets", "listpets.go"}, "pets list", map[string]string{"nick-name": ""})
		p.requireShorthands(cliShorthandCreatePetCommand, "pets create", map[string]string{"nick-name": "n", "tag": "t"})
		lock = p.lock()
		require.Empty(t, lock["ListPets"]["json"].Flags)
		require.Equal(t, "names", lock["ListPets"]["json"].Retired["n"].Flag)
		require.Equal(t, map[string]string{"nick-name": "n", "tag": "t"}, lock["CreatePet"]["json"].Flags)
		require.Empty(t, lock["CreatePet"]["json"].Retired)
	})

	t.Run("pin starting with a on a paginated operation fails generation", func(t *testing.T) {
		t.Parallel()
		newCLIShorthandProject(t).generateFails(cliShorthandPaginatedSpec, `operation "ListPets" preferShorthand "after" cannot take -a`)
	})

	t.Run("intent shorthand collides with a pinned flag", func(t *testing.T) {
		t.Parallel()
		newCLIShorthandProject(t).generateFails(cliShorthandIntentSpec,
			`command "create" declares shorthand "-s" on flag "--summary", which collides with the auto-assigned shorthand of generated flag "--sort"`)
	})
}

const cliShorthandPaginatedSpec = `openapi: 3.0.3
info:
  title: Pet API
  version: 1.0.0
servers:
  - url: https://api.example.com
x-speakeasy-cli-commands:
  version: 1
  operations:
    ListPets:
      preferShorthand: [after]
paths:
  /pets:
    get:
      operationId: ListPets
      description: A paginated operation reserves -a for --all, so a pin starting with a cannot take it.
      tags: [pets]
      security: []
      parameters:
        - {name: after, in: query, schema: {type: string}}
      x-speakeasy-pagination:
        type: cursor
        inputs:
          - {name: after, in: parameters, type: cursor}
        outputs:
          nextCursor: $.next
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  next: {type: string}
                  pets: {type: array, items: {type: string}}
`

const cliShorthandIntentSpec = `openapi: 3.0.3
info:
  title: Pet API
  version: 1.0.0
servers:
  - url: https://api.example.com
x-speakeasy-cli-commands:
  version: 1
  categories: [Create]
  operations:
    CreatePet:
      preferShorthand: [sort]
  commands:
    create:
      category: Create
      summary: Create a pet
      op: CreatePet
      args:
        name:
          to: $.name
          variadic: true
          summary: Pet name
      flags:
        summary:
          to: $.summary
          shorthand: s
          summary: Pet summary
paths:
  /pets:
    post:
      operationId: CreatePet
      description: Two query parameters share an initial letter; the pinned one keeps it, so an intent flag cannot reuse it.
      tags: [pets]
      security: []
      parameters:
        - {name: sort, in: query, schema: {type: string}}
        - {name: size, in: query, schema: {type: string}}
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [name]
              properties:
                name: {type: string}
                summary: {type: string}
      responses:
        '200':
          description: OK
`
