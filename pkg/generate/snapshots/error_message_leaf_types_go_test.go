package snapshots

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
	"github.com/stretchr/testify/require"
)

// Error() reads marked messages of every field representation: optional+nullable
// model fields wrapped in OptionalNullable, required+nullable pointers, and
// non-string enum and integer leaves. A top-level marked field takes precedence
// over a nested marked field and an inferred `message`. Empty string messages
// fall back as if absent. An inferred message is prefixed with a sibling `code`
// when one is present; a marked message is not.
func TestSnapGoErrorMessageLeafTypes(t *testing.T) {
	t.Parallel()
	spec := `openapi: 3.1.0
info:
  title: Test
  version: "1.0"
servers:
  - url: https://api.example.com
paths:
  /widgets:
    get:
      operationId: getWidget
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Cause"
        "400":
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/WrappedCauseError"
        "401":
          description: Unauthorized
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/EnumMessageError"
        "402":
          description: Payment Required
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/IntegerMessageError"
        "403":
          description: Forbidden
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/RequiredNullableMessageError"
        "409":
          description: Conflict
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/PrecedenceError"
        "404":
          description: Not Found
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/NestedCodeError"
        "422":
          description: Unprocessable Entity
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/TopLevelCodeError"
        "429":
          description: Too Many Requests
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/WrappedCodeError"
        "500":
          description: Internal Server Error
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/MarkedMessageCodeError"
components:
  schemas:
    Cause:
      type: object
      properties:
        inner:
          type: object
          nullable: true
          properties:
            message:
              type: string
              nullable: true
              x-speakeasy-error-message: true
    WrappedCauseError:
      type: object
      properties:
        cause:
          $ref: "#/components/schemas/Cause"
    EnumMessageError:
      type: object
      properties:
        status:
          type: string
          enum: [rejected, expired]
          x-speakeasy-error-message: true
    IntegerMessageError:
      type: object
      properties:
        code:
          type: integer
          x-speakeasy-error-message: true
    RequiredNullableMessageError:
      type: object
      required: [message]
      properties:
        message:
          type: string
          nullable: true
          x-speakeasy-error-message: true
    PrecedenceError:
      type: object
      properties:
        message:
          type: string
        detail:
          type: object
          properties:
            reason:
              type: string
              x-speakeasy-error-message: true
        title:
          type: string
          x-speakeasy-error-message: true
    NestedCodeError:
      type: object
      properties:
        error:
          type: object
          properties:
            code:
              type: string
            message:
              type: string
    TopLevelCodeError:
      type: object
      required: [code, message]
      properties:
        code:
          type: integer
        message:
          type: string
    WrappedCodeError:
      type: object
      properties:
        error:
          $ref: "#/components/schemas/WrappedCodeDetail"
    WrappedCodeDetail:
      type: object
      properties:
        code:
          type: string
          nullable: true
        message:
          type: string
          nullable: true
    MarkedMessageCodeError:
      type: object
      properties:
        error:
          type: object
          properties:
            code:
              type: string
            detail:
              type: string
              x-speakeasy-error-message: true`

	genYaml := `go:
  packageName: github.com/example/test
  nullableOptionalWrapper: true
`

	expectedSnapshotFiles := []string{
		"models/apierrors/wrappedcauseerror.go",
		"models/apierrors/enummessageerror.go",
		"models/apierrors/integermessageerror.go",
		"models/apierrors/requirednullablemessageerror.go",
		"models/apierrors/precedenceerror.go",
		"models/apierrors/nestedcodeerror.go",
		"models/apierrors/toplevelcodeerror.go",
		"models/apierrors/wrappedcodeerror.go",
		"models/apierrors/markedmessagecodeerror.go",
	}

	expectedSnapshot := `--- models/apierrors/enummessageerror.go ---
// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
// Generated under the AGPL-3.0-only license.
// SPDX-License-Identifier: AGPL-3.0-only

package apierrors

import (
	"encoding/json"
	"fmt"
	"github.com/example/test/models/components"
)

type EnumMessageError struct {
	Status   *components.Status      ` + "`" + `json:"status,omitzero"` + "`" + `
	HTTPMeta components.HTTPMetadata ` + "`" + `json:"-"` + "`" + `
}

var _ error = &EnumMessageError{}

func (e *EnumMessageError) Error() string {
	if e != nil {
		if e.Status != nil {
			return fmt.Sprint(*e.Status)
		}
	}

	data, _ := json.Marshal(e)
	return string(data)
}


--- models/apierrors/integermessageerror.go ---
// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
// Generated under the AGPL-3.0-only license.
// SPDX-License-Identifier: AGPL-3.0-only

package apierrors

import (
	"encoding/json"
	"fmt"
	"github.com/example/test/models/components"
)

type IntegerMessageError struct {
	Code     *int64                  ` + "`" + `json:"code,omitzero"` + "`" + `
	HTTPMeta components.HTTPMetadata ` + "`" + `json:"-"` + "`" + `
}

var _ error = &IntegerMessageError{}

func (e *IntegerMessageError) Error() string {
	if e != nil {
		if e.Code != nil {
			return fmt.Sprint(*e.Code)
		}
	}

	data, _ := json.Marshal(e)
	return string(data)
}


--- models/apierrors/markedmessagecodeerror.go ---
// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
// Generated under the AGPL-3.0-only license.
// SPDX-License-Identifier: AGPL-3.0-only

package apierrors

import (
	"encoding/json"
	"github.com/example/test/models/components"
)

type MarkedMessageCodeError struct {
	ErrorInfo *components.MarkedMessageCodeErrorError ` + "`" + `json:"error,omitzero"` + "`" + `
	HTTPMeta  components.HTTPMetadata                 ` + "`" + `json:"-"` + "`" + `
}

var _ error = &MarkedMessageCodeError{}

func (e *MarkedMessageCodeError) Error() string {
	if e != nil {
		if v0 := e.ErrorInfo; v0 != nil {
			if v0.Detail != nil && *v0.Detail != "" {
				return *v0.Detail
			}
		}
	}

	data, _ := json.Marshal(e)
	return string(data)
}


--- models/apierrors/nestedcodeerror.go ---
// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
// Generated under the AGPL-3.0-only license.
// SPDX-License-Identifier: AGPL-3.0-only

package apierrors

import (
	"encoding/json"
	"fmt"
	"github.com/example/test/models/components"
)

type NestedCodeError struct {
	ErrorInfo *components.NestedCodeErrorError ` + "`" + `json:"error,omitzero"` + "`" + `
	HTTPMeta  components.HTTPMetadata          ` + "`" + `json:"-"` + "`" + `
}

var _ error = &NestedCodeError{}

func (e *NestedCodeError) Error() string {
	if e != nil {
		if v0 := e.ErrorInfo; v0 != nil {
			if v0.Message != nil && *v0.Message != "" {
				if v0.Code != nil && *v0.Code != "" {
					return fmt.Sprintf("%v: %s", *v0.Code, *v0.Message)
				}
				return *v0.Message
			}
		}
	}

	data, _ := json.Marshal(e)
	return string(data)
}


--- models/apierrors/precedenceerror.go ---
// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
// Generated under the AGPL-3.0-only license.
// SPDX-License-Identifier: AGPL-3.0-only

package apierrors

import (
	"github.com/example/test/models/components"
)

type PrecedenceError struct {
	Message  *string                 ` + "`" + `json:"message,omitzero"` + "`" + `
	Detail   *components.Detail      ` + "`" + `json:"detail,omitzero"` + "`" + `
	Title    *string                 ` + "`" + `json:"title,omitzero"` + "`" + `
	HTTPMeta components.HTTPMetadata ` + "`" + `json:"-"` + "`" + `
}

var _ error = &PrecedenceError{}

func (e *PrecedenceError) Error() string {
	if e.Title == nil || *e.Title == "" {
		return "unknown error"
	}

	return *e.Title
}


--- models/apierrors/requirednullablemessageerror.go ---
// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
// Generated under the AGPL-3.0-only license.
// SPDX-License-Identifier: AGPL-3.0-only

package apierrors

import (
	"github.com/example/test/models/components"
)

type RequiredNullableMessageError struct {
	Message  *string                 ` + "`" + `json:"message"` + "`" + `
	HTTPMeta components.HTTPMetadata ` + "`" + `json:"-"` + "`" + `
}

var _ error = &RequiredNullableMessageError{}

func (e *RequiredNullableMessageError) Error() string {
	if e.Message == nil || *e.Message == "" {
		return "unknown error"
	}

	return *e.Message
}


--- models/apierrors/toplevelcodeerror.go ---
// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
// Generated under the AGPL-3.0-only license.
// SPDX-License-Identifier: AGPL-3.0-only

package apierrors

import (
	"encoding/json"
	"fmt"
	"github.com/example/test/models/components"
)

type TopLevelCodeError struct {
	Code     int64                   ` + "`" + `json:"code"` + "`" + `
	Message  string                  ` + "`" + `json:"message"` + "`" + `
	HTTPMeta components.HTTPMetadata ` + "`" + `json:"-"` + "`" + `
}

var _ error = &TopLevelCodeError{}

func (e *TopLevelCodeError) Error() string {
	if e != nil {
		if e.Message != "" {
			return fmt.Sprintf("%v: %s", e.Code, e.Message)
		}
	}

	data, _ := json.Marshal(e)
	return string(data)
}


--- models/apierrors/wrappedcauseerror.go ---
// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
// Generated under the AGPL-3.0-only license.
// SPDX-License-Identifier: AGPL-3.0-only

package apierrors

import (
	"encoding/json"
	"github.com/example/test/models/components"
)

type WrappedCauseError struct {
	Cause    *components.Cause       ` + "`" + `json:"cause,omitzero"` + "`" + `
	HTTPMeta components.HTTPMetadata ` + "`" + `json:"-"` + "`" + `
}

var _ error = &WrappedCauseError{}

func (e *WrappedCauseError) Error() string {
	if e != nil {
		if v0 := e.Cause; v0 != nil {
			if v1, ok := v0.Inner.Get(); ok && v1 != nil {
				if v2, ok := v1.Message.Get(); ok && v2 != nil && *v2 != "" {
					return *v2
				}
			}
		}
	}

	data, _ := json.Marshal(e)
	return string(data)
}


--- models/apierrors/wrappedcodeerror.go ---
// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
// Generated under the AGPL-3.0-only license.
// SPDX-License-Identifier: AGPL-3.0-only

package apierrors

import (
	"encoding/json"
	"fmt"
	"github.com/example/test/models/components"
)

type WrappedCodeError struct {
	ErrorInfo *components.WrappedCodeDetail ` + "`" + `json:"error,omitzero"` + "`" + `
	HTTPMeta  components.HTTPMetadata       ` + "`" + `json:"-"` + "`" + `
}

var _ error = &WrappedCodeError{}

func (e *WrappedCodeError) Error() string {
	if e != nil {
		if v0 := e.ErrorInfo; v0 != nil {
			if v1, ok := v0.Message.Get(); ok && v1 != nil && *v1 != "" {
				if code, ok := v0.Code.Get(); ok && code != nil && *code != "" {
					return fmt.Sprintf("%v: %s", *code, *v1)
				}
				return *v1
			}
		}
	}

	data, _ := json.Marshal(e)
	return string(data)
}


` // end of snapshot

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         spec,
		GenYaml:      genYaml,
		IncludeGlobs: expectedSnapshotFiles,
		Expected:     expectedSnapshot,
		AfterGenerate: func(t *testing.T, tempDir string) {
			t.Helper()
			runGoErrorMessageLeafTypesContract(t, tempDir)
		},
	})
}

func runGoErrorMessageLeafTypesContract(t *testing.T, tempDir string) {
	t.Helper()

	path := filepath.Join(tempDir, "models", "apierrors", "error_message_contract_test.go")
	require.NoError(t, os.WriteFile(path, []byte(`package apierrors_test

import (
	"encoding/json"
	"testing"

	"github.com/example/test/models/apierrors"
)

func assertErrorMessage(t *testing.T, e error, body, want string) {
	t.Helper()
	if err := json.Unmarshal([]byte(body), e); err != nil {
		t.Fatalf("unmarshal %s: %v", body, err)
	}
	if got := e.Error(); got != want {
		t.Errorf("%s: got %q, want %q", body, got, want)
	}
}

func assertJSONFallback(t *testing.T, e error, body string) {
	t.Helper()
	if err := json.Unmarshal([]byte(body), e); err != nil {
		t.Fatalf("unmarshal %s: %v", body, err)
	}
	want, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("marshal %s: %v", body, err)
	}
	if got := e.Error(); got != string(want) {
		t.Errorf("%s: got %q, want JSON fallback %q", body, got, want)
	}
}

func TestErrorMessageLeafTypes(t *testing.T) {
	assertErrorMessage(t, &apierrors.WrappedCauseError{}, `+"`"+`{"cause":{"inner":{"message":"wrapped"}}}`+"`"+`, "wrapped")
	for _, body := range []string{
		`+"`"+`{"cause":{"inner":{"message":""}}}`+"`"+`,
		`+"`"+`{"cause":{"inner":{"message":null}}}`+"`"+`,
		`+"`"+`{"cause":{"inner":{}}}`+"`"+`,
		`+"`"+`{"cause":{"inner":null}}`+"`"+`,
		`+"`"+`{"cause":{}}`+"`"+`,
		`+"`"+`{}`+"`"+`,
	} {
		assertJSONFallback(t, &apierrors.WrappedCauseError{}, body)
	}

	assertErrorMessage(t, &apierrors.EnumMessageError{}, `+"`"+`{"status":"expired"}`+"`"+`, "expired")
	assertJSONFallback(t, &apierrors.EnumMessageError{}, `+"`"+`{}`+"`"+`)

	assertErrorMessage(t, &apierrors.IntegerMessageError{}, `+"`"+`{"code":42}`+"`"+`, "42")
	assertJSONFallback(t, &apierrors.IntegerMessageError{}, `+"`"+`{}`+"`"+`)

	assertErrorMessage(t, &apierrors.RequiredNullableMessageError{}, `+"`"+`{"message":"required"}`+"`"+`, "required")
	assertErrorMessage(t, &apierrors.RequiredNullableMessageError{}, `+"`"+`{"message":null}`+"`"+`, "unknown error")
	assertErrorMessage(t, &apierrors.RequiredNullableMessageError{}, `+"`"+`{"message":""}`+"`"+`, "unknown error")

	assertErrorMessage(t, &apierrors.PrecedenceError{}, `+"`"+`{"message":"inferred","detail":{"reason":"nested"},"title":"top"}`+"`"+`, "top")
	assertErrorMessage(t, &apierrors.PrecedenceError{}, `+"`"+`{"message":"inferred","title":""}`+"`"+`, "unknown error")

	assertErrorMessage(t, &apierrors.NestedCodeError{}, `+"`"+`{"error":{"code":"rate_limit","message":"slow down"}}`+"`"+`, "rate_limit: slow down")
	assertErrorMessage(t, &apierrors.NestedCodeError{}, `+"`"+`{"error":{"code":"","message":"slow down"}}`+"`"+`, "slow down")
	assertErrorMessage(t, &apierrors.NestedCodeError{}, `+"`"+`{"error":{"message":"slow down"}}`+"`"+`, "slow down")
	assertJSONFallback(t, &apierrors.NestedCodeError{}, `+"`"+`{"error":{"code":"rate_limit","message":""}}`+"`"+`)
	assertJSONFallback(t, &apierrors.NestedCodeError{}, `+"`"+`{"error":{"code":"rate_limit"}}`+"`"+`)

	assertErrorMessage(t, &apierrors.TopLevelCodeError{}, `+"`"+`{"code":409,"message":"conflict"}`+"`"+`, "409: conflict")
	assertJSONFallback(t, &apierrors.TopLevelCodeError{}, `+"`"+`{"code":409,"message":""}`+"`"+`)

	assertErrorMessage(t, &apierrors.WrappedCodeError{}, `+"`"+`{"error":{"code":"quota","message":"exhausted"}}`+"`"+`, "quota: exhausted")
	assertErrorMessage(t, &apierrors.WrappedCodeError{}, `+"`"+`{"error":{"code":null,"message":"exhausted"}}`+"`"+`, "exhausted")
	assertJSONFallback(t, &apierrors.WrappedCodeError{}, `+"`"+`{"error":{"code":"quota","message":""}}`+"`"+`)

	assertErrorMessage(t, &apierrors.MarkedMessageCodeError{}, `+"`"+`{"error":{"code":"denied","detail":"not allowed"}}`+"`"+`, "not allowed")
	assertJSONFallback(t, &apierrors.MarkedMessageCodeError{}, `+"`"+`{"error":{"code":"denied","detail":""}}`+"`"+`)
}
`), 0o644))
	t.Cleanup(func() { _ = os.Remove(path) })

	cmd := exec.Command("go", "test", "-count=1", "-run", "TestErrorMessageLeafTypes", "./models/apierrors")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
	require.NoError(t, os.Remove(path))
}
