package snapshots

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoOptionalMethodArgumentsModes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		mode                string
		defaultSignature    string
		defaultConstructor  string
		defaultOptionsType  string
		defaultOptionsSlice string
	}{
		{
			name:                "pointers",
			mode:                "pointers",
			defaultSignature:    "func (s *SDK) ListPets(ctx context.Context, accountID string, limit *int64, cursor optionalnullable.OptionalNullable[string], tags []string, filters map[string]string, opts ...operations.Option)",
			defaultOptionsType:  "[][operations.Option]",
			defaultOptionsSlice: "opts ...operations.Option",
		},
		{
			name:                "shared options",
			mode:                "shared-options",
			defaultSignature:    "func (s *SDK) ListPets(ctx context.Context, accountID string, opts ...operations.Option)",
			defaultConstructor:  "func WithListPetsLimit(limit int64) Option",
			defaultOptionsType:  "[][operations.Option]",
			defaultOptionsSlice: "opts ...operations.Option",
		},
		{
			name:                "method options",
			mode:                "method-options",
			defaultSignature:    "func (s *SDK) ListPets(ctx context.Context, accountID string, opts ...operations.ListPetsOption)",
			defaultConstructor:  "func WithListPetsLimit(limit int64) ListPetsOption",
			defaultOptionsType:  "`[]operations.ListPetsOption`",
			defaultOptionsSlice: "opts ...operations.ListPetsOption",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var compileContracts sync.Once

			snaptest.DoTestSnapshot(t, snaptest.Options{
				Spec: goOptionalMethodArgumentsSpec,
				GenYaml: `go:
  packageName: github.com/example/optionalmethodarguments
  maxMethodParams: 10
  nullableOptionalWrapper: true
  optionalMethodArguments: ` + tt.mode + `
`,
				AfterGenerate: func(t *testing.T, tempDir string) {
					t.Helper()

					sdk := readGeneratedFile(t, tempDir, "sdk.go")
					methodOptions := readGeneratedFile(t, tempDir, "models/operations/method_options.go")
					readme := readGeneratedFile(t, tempDir, "docs/sdks/sdk/README.md")
					listPetsDocs := markdownSection(readme, "## ListPets", "## PointerOperation")

					assert.Contains(t, sdk, tt.defaultSignature)
					assert.Contains(t, sdk, tt.defaultOptionsSlice)
					assert.Contains(t, listPetsDocs, tt.defaultOptionsType)
					if tt.mode == "method-options" {
						assert.NotContains(t, listPetsDocs, "ListPetsOption](")
					}
					if tt.defaultConstructor == "" {
						assert.NotContains(t, methodOptions, "WithListPetsLimit")
						assert.Contains(t, listPetsDocs, "| `limit`")
						assert.Contains(t, sdk, `fmt.Errorf("error applying option: %w", err)`)
						assert.NotContains(t, sdk, "error applying option to PointerOperation")

						accountIndex := strings.Index(listPetsDocs, "| `accountID`")
						limitIndex := strings.Index(listPetsDocs, "| `limit`")
						cursorIndex := strings.Index(listPetsDocs, "| `cursor`")
						tagsIndex := strings.Index(listPetsDocs, "| `tags`")
						filtersIndex := strings.Index(listPetsDocs, "| `filters`")
						assert.True(t,
							accountIndex >= 0 &&
								accountIndex < limitIndex &&
								limitIndex < cursorIndex &&
								cursorIndex < tagsIndex &&
								tagsIndex < filtersIndex,
							"pointer-mode docs should preserve positional method argument order",
						)
					} else {
						assert.Contains(t, methodOptions, tt.defaultConstructor)
						assert.NotContains(t, listPetsDocs, "| `limit`")
						assert.Contains(t, methodOptions, "func WithListPetsCursor(cursor *string)")
						assert.Contains(t, methodOptions, "func WithListPetsTags(tags []string)")
						assert.Contains(t, methodOptions, "func WithListPetsFilters(filters map[string]string)")
						assert.Contains(t, methodOptions, "Maximum number of pets to return.")
						assert.Contains(t, methodOptions, "If omitted, the argument remains unset.")
						assert.Contains(t, sdk, "Optional arguments may be provided with [operations.WithListPets")
					}

					assert.Contains(t, sdk, "func (s *SDK) PointerOperation(ctx context.Context, limit *int64, opts ...operations.Option)")
					assert.Contains(t, sdk, "func (s *SDK) SharedOperationA(ctx context.Context, opts ...operations.Option)")
					assert.Contains(t, sdk, "func (s *SDK) StrictOperation(ctx context.Context, opts ...operations.StrictOperationOption)")
					assert.Contains(t, methodOptions, "func WithSharedOperationALimit(limit int64) Option")
					assert.Contains(t, methodOptions, "func WithStrictOperationLimit(limit int64) StrictOperationOption")
					assert.Contains(t, methodOptions, "func WithStrictOperationTimeoutObj(timeout_ int64) StrictOperationOption")
					assert.Contains(t, methodOptions, "func WithStrictOperationTimeout(timeout time.Duration) StrictOperationOption")

					assert.Contains(t, sdk, "func (s *SDK) SharedCursor(ctx context.Context, limit int64, opts ...operations.Option)")
					assert.Contains(t, sdk, "nextOpts := append([]operations.Option(nil), opts...)")
					assert.Contains(t, sdk, "nextOpts = append(nextOpts, operations.WithSharedCursorCursor(nCVal))")
					assert.Contains(t, sdk, "func (s *SDK) StrictOptionalBody(ctx context.Context, opts ...operations.StrictOptionalBodyOption)")
					assert.Contains(t, sdk, "if request == nil {")
					assert.Contains(t, sdk, "request = &components.PageRequest{}")
					assert.Contains(t, sdk, "nextOpts = append(nextOpts, operations.WithStrictOptionalBodyRequest(*request))")
					assert.Contains(t, sdk, "nextOpts = append(nextOpts, operations.WithStrictURLURLOverride(nextURL))")
					assert.Contains(t, methodOptions, "func WithStrictPollingPolling(")
					assert.Contains(t, sdk, "return s.strictPollingWaitForSuccess(ctx, hookCtx, req, o.Options)")

					compileContracts.Do(func() {
						if tt.mode == "shared-options" {
							runSharedOptionsRuntimeContract(t, tempDir)
						}
						if tt.mode == "method-options" {
							runStrictOptionalBodyPaginationContract(t, tempDir)
							assertStrictOptionsCompileContract(t, tempDir)
						}
					})
				},
			})
		})
	}
}

func TestGoMethodOptionsGeneratedTestsCompile(t *testing.T) {
	t.Parallel()

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec: `openapi: 3.1.0
info:
  title: Go method options generated tests
  version: 1.0.0
paths:
  /pets:
    get:
      operationId: listPets
      servers:
        - url: https://api.example.com
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: string
`,
		GenYaml: `generation:
  tests:
    generateTests: true
    generateNewTests: true
go:
  packageName: github.com/example/optionalmethodargumenttests
  imports:
    paths:
      operations: ""
  maxMethodParams: 10
  nullableOptionalWrapper: true
  optionalMethodArguments: method-options
`,
		AfterGenerate: func(t *testing.T, tempDir string) {
			t.Helper()

			testFiles, err := filepath.Glob(filepath.Join(tempDir, "tests", "*_test.go"))
			require.NoError(t, err)
			require.NotEmpty(t, testFiles)

			var generatedTests strings.Builder
			for _, testFile := range testFiles {
				contents, err := os.ReadFile(testFile)
				require.NoError(t, err)
				generatedTests.Write(contents)
			}

			assert.Contains(t, generatedTests.String(), "optionalmethodargumenttests.WithListPetsServerURL(")
			assert.NotContains(t, generatedTests.String(), "optionalmethodargumenttests.WithMethodServerURL(")
		},
	})
}

func TestGoOptionalMethodArgumentsRequireNullableWrapper(t *testing.T) {
	t.Parallel()

	shouldCompile := false
	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec: `openapi: 3.1.0
info:
  title: Nullable optional method arguments
  version: 1.0.0
paths:
  /pets:
    get:
      operationId: listPets
      parameters:
        - name: cursor
          in: query
          required: false
          schema:
            type: [string, "null"]
      responses:
        "200":
          description: OK
`,
		GenYaml: `go:
  packageName: github.com/example/nullableoptionalmethodarguments
  nullableOptionalWrapper: false
  optionalMethodArguments: shared-options
`,
		ShouldCompile: &shouldCompile,
		ExpectErrors: []string{
			"optional method argument cursor is nullable; enable go.nullableOptionalWrapper",
		},
	})
}

func TestGoOptionalMethodArgumentsInvalidConfig(t *testing.T) {
	t.Parallel()

	shouldCompile := false
	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec: `openapi: 3.1.0
info:
  title: Invalid optional method arguments
  version: 1.0.0
paths: {}
`,
		GenYaml: `go:
  packageName: github.com/example/invalidoptionalmethodarguments
  optionalMethodArguments: invalid
`,
		ShouldCompile: &shouldCompile,
		ExpectErrors: []string{
			`optionalMethodArguments must be "pointers", "shared-options", or "method-options"`,
		},
	})
}

func readGeneratedFile(t *testing.T, tempDir string, pathParts ...string) string {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(append([]string{tempDir}, pathParts...)...))
	require.NoError(t, err)
	return strings.ReplaceAll(string(contents), "\r\n", "\n")
}

func markdownSection(contents, start, end string) string {
	startIndex := strings.Index(contents, start)
	if startIndex == -1 {
		return ""
	}
	contents = contents[startIndex:]
	if endIndex := strings.Index(contents, end); endIndex != -1 {
		contents = contents[:endIndex]
	}
	return contents
}

func runSharedOptionsRuntimeContract(t *testing.T, tempDir string) {
	t.Helper()

	path := filepath.Join(tempDir, "optional_method_arguments_contract_test.go")
	require.NoError(t, os.WriteFile(path, []byte(`package optionalmethodarguments

import (
    "context"
    "errors"
    "testing"

    "github.com/example/optionalmethodarguments/models/operations"
)

func TestSharedMethodArgumentOptionContract(t *testing.T) {
    client := New()
    _, err := client.SharedOperationB(context.Background(), operations.WithSharedOperationALimit(1))
    if !errors.Is(err, operations.ErrUnsupportedOption) {
        t.Fatalf("expected ErrUnsupportedOption, got %v", err)
    }

    var opts operations.Options
    supported := []string{operations.SupportedOptionSharedOperationALimit}
    if err := operations.WithSharedOperationALimit(1)(&opts, supported...); err != nil {
        t.Fatal(err)
    }
    if err := operations.WithSharedOperationALimit(2)(&opts, supported...); err != nil {
        t.Fatal(err)
    }
    value, ok := opts.GetMethodArgument(operations.SupportedOptionSharedOperationALimit)
    if !ok || *value.(*int64) != 2 {
        t.Fatalf("expected last option value 2, got %v", value)
    }
}
`), 0o644))
	t.Cleanup(func() { _ = os.Remove(path) })

	cmd := exec.Command("go", "test", "-run", "TestSharedMethodArgumentOptionContract", ".")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
	require.NoError(t, os.Remove(path))
}

func runStrictOptionalBodyPaginationContract(t *testing.T, tempDir string) {
	t.Helper()

	path := filepath.Join(tempDir, "optional_body_pagination_contract_test.go")
	require.NoError(t, os.WriteFile(path, []byte(`package optionalmethodarguments

import (
    "context"
    "io"
    "net/http"
    "strings"
    "testing"
)

type optionalBodyPaginationClient struct {
    calls int
}

func (c *optionalBodyPaginationClient) Do(req *http.Request) (*http.Response, error) {
    c.calls++
    return &http.Response{
        StatusCode: http.StatusOK,
        Header: http.Header{
            "Content-Type": []string{"application/json"},
        },
        Body: io.NopCloser(strings.NewReader(`+"`"+`{"numPages":2,"items":["pet"]}`+"`"+`)),
        Request: req,
    }, nil
}

func TestStrictOptionalBodyPaginationWithoutInitialBody(t *testing.T) {
    httpClient := &optionalBodyPaginationClient{}
    client := New(WithClient(httpClient))

    first, err := client.StrictOptionalBody(context.Background())
    if err != nil {
        t.Fatal(err)
    }
    if first.Next == nil {
        t.Fatal("expected next-page function")
    }

    second, err := first.Next()
    if err != nil {
        t.Fatal(err)
    }
    if second == nil {
        t.Fatal("expected second page")
    }
    if httpClient.calls != 2 {
        t.Fatalf("expected 2 requests, got %d", httpClient.calls)
    }
}
`), 0o644))
	t.Cleanup(func() { _ = os.Remove(path) })

	cmd := exec.Command("go", "test", "-run", "TestStrictOptionalBodyPaginationWithoutInitialBody", ".")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
	require.NoError(t, os.Remove(path))
}

func assertStrictOptionsCompileContract(t *testing.T, tempDir string) {
	t.Helper()

	dir := filepath.Join(tempDir, "compilecontract")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	path := filepath.Join(dir, "compile_contract.go")
	require.NoError(t, os.WriteFile(path, []byte(`package compilecontract

import (
    "context"

    sdk "github.com/example/optionalmethodarguments"
    "github.com/example/optionalmethodarguments/models/operations"
    "github.com/example/optionalmethodarguments/retry"
)

func invalid(client *sdk.SDK) {
    _, _ = client.ListPets(context.Background(), "account", operations.WithRetries(retry.Config{}))
}
`), 0o644))

	cmd := exec.Command("go", "test", "./compilecontract")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	require.Error(t, err, "generic operations.Option must not satisfy ListPetsOption")
	assert.Contains(t, string(output), "operations.WithRetries")
	assert.Contains(t, string(output), "operations.ListPetsOption")
	require.NoError(t, os.RemoveAll(dir))
}

const goOptionalMethodArgumentsSpec = `openapi: 3.1.0
info:
  title: Go optional method arguments
  version: 1.0.0
servers:
  - url: https://api.example.com
paths:
  /accounts/{accountID}/pets:
    get:
      operationId: listPets
      parameters:
        - name: accountID
          in: path
          required: true
          schema:
            type: string
        - name: limit
          in: query
          required: false
          description: Maximum number of pets to return.
          schema:
            type: integer
            format: int64
        - name: cursor
          in: query
          required: false
          schema:
            type: [string, "null"]
        - name: tags
          in: query
          required: false
          schema:
            type: array
            items:
              type: string
        - name: filters
          in: query
          required: false
          style: deepObject
          explode: true
          schema:
            type: object
            additionalProperties:
              type: string
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: string
  /pointer:
    get:
      operationId: pointerOperation
      x-speakeasy-go-optional-method-arguments: pointers
      parameters:
        - name: limit
          in: query
          required: false
          schema:
            type: integer
            format: int64
      responses:
        "200":
          description: OK
  /shared-a:
    get:
      operationId: sharedOperationA
      x-speakeasy-go-optional-method-arguments: shared-options
      parameters:
        - name: limit
          in: query
          required: false
          schema:
            type: integer
            format: int64
      responses:
        "200":
          description: OK
  /shared-b:
    get:
      operationId: sharedOperationB
      x-speakeasy-go-optional-method-arguments: shared-options
      parameters:
        - name: limit
          in: query
          required: false
          schema:
            type: integer
            format: int64
      responses:
        "200":
          description: OK
  /strict:
    get:
      operationId: strictOperation
      x-speakeasy-go-optional-method-arguments: method-options
      parameters:
        - name: limit
          in: query
          required: false
          schema:
            type: integer
            format: int64
        - name: timeout
          in: query
          required: false
          schema:
            type: integer
            format: int64
      responses:
        "200":
          description: OK
  /shared-cursor:
    get:
      operationId: sharedCursor
      x-speakeasy-go-optional-method-arguments: shared-options
      parameters:
        - name: cursor
          in: query
          required: false
          schema:
            type: string
        - name: limit
          in: query
          required: true
          schema:
            type: integer
            format: int64
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/CursorResponse"
      x-speakeasy-pagination:
        type: cursor
        inputs:
          - name: cursor
            in: parameters
            type: cursor
          - name: limit
            in: parameters
            type: limit
        outputs:
          nextCursor: $.nextCursor
          results: $.items
  /strict-optional-body:
    post:
      operationId: strictOptionalBody
      x-speakeasy-go-optional-method-arguments: method-options
      requestBody:
        required: false
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/PageRequest"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/PageResponse"
      x-speakeasy-pagination:
        type: offsetLimit
        inputs:
          - name: page
            in: requestBody
            type: page
          - name: limit
            in: requestBody
            type: limit
        outputs:
          numPages: $.numPages
          results: $.items
  /strict-url:
    get:
      operationId: strictURL
      x-speakeasy-go-optional-method-arguments: method-options
      parameters:
        - name: attempts
          in: query
          required: true
          schema:
            type: integer
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/URLResponse"
      x-speakeasy-pagination:
        type: url
        outputs:
          nextUrl: $.nextURL
          results: $.items
  /strict-polling:
    get:
      operationId: strictPolling
      x-speakeasy-go-optional-method-arguments: method-options
      parameters:
        - name: requestID
          in: query
          required: true
          schema:
            type: string
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
        "204":
          description: No Content
      x-speakeasy-polling:
        - name: WaitForSuccess
          successCriteria:
            - condition: $statusCode == 200
components:
  schemas:
    CursorResponse:
      type: object
      required: [items]
      properties:
        nextCursor:
          type: string
        items:
          type: array
          items:
            type: string
    PageRequest:
      type: object
      properties:
        page:
          type: [integer, "null"]
          format: int64
        limit:
          type: [integer, "null"]
          format: int64
    PageResponse:
      type: object
      required: [numPages, items]
      properties:
        numPages:
          type: integer
          format: int64
        items:
          type: array
          items:
            type: string
    URLResponse:
      type: object
      required: [items]
      properties:
        nextURL:
          type: string
        items:
          type: array
          items:
            type: string
`
