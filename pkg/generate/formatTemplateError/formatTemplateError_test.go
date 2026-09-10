package formattemplateerror

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatTemplateError1(t *testing.T) {
	input := `GoError: failed to execute template: template: readme/errors.stmpl:6:2: executing "readme/errors.stmpl" at <templateUsageExample .Local.UsageContext "">: error calling templateUsageExample: GoError: failed to execute template: template: readme/snippet.stmpl:2:2: executing "readme/snippet.stmpl" at <templateUsageSnippet .Local>: error calling templateUsageSnippet: GoError: failed to execute template: template: usage/snippet.stmpl:1:3: executing "usage/snippet.stmpl" at <templateImports .RecursiveComputed.UsageImports>: error calling templateImports: Error: Tried to import duplicate: models at templateImports (includes/imports.ts:361:17(47)) at github.com/speakeasy-api/easytemplate.(*Engine).init.(*Engine).init.func1.func6 (native) at github.com/speakeasy-api/easytemplate.(*Engine).init.(*Engine).init.func1.func6 (native) at github.com/speakeasy-api/easytemplate.(*Engine).init.(*Engine).init.func1.func6 (native)`
	actual := FormatTemplateErrorString(input)
	expected := `GoError: failed to execute template:
  template: readme/errors.stmpl:6:2
    executing "readme/errors.stmpl" at <templateUsageExample .Local.UsageContext "">:
    error calling templateUsageExample:
      GoError: failed to execute template:
        template: readme/snippet.stmpl:2:2
          executing "readme/snippet.stmpl" at <templateUsageSnippet .Local>:
          error calling templateUsageSnippet:
            GoError: failed to execute template:
              template: usage/snippet.stmpl:1:3
                executing "usage/snippet.stmpl" at <templateImports .RecursiveComputed.UsageImports>
                error calling templateImports:
                  Error: Tried to import duplicate: models
                  at templateImports (includes/imports.ts:361:17(47))
                  at github.com/speakeasy-api/easytemplate.(*Engine).init.(*Engine).init.func1.func6 (native)
                  at github.com/speakeasy-api/easytemplate.(*Engine).init.(*Engine).init.func1.func6 (native)
                  at github.com/speakeasy-api/easytemplate.(*Engine).init.(*Engine).init.func1.func6 (native)`

	fmt.Println(actual)
	assert.Equal(t, expected, actual)
}

func TestFormatTemplateError2(t *testing.T) {
	input := `GoError: failed to execute template: template: readme/model.stmpl:14:33: executing "readme/model.stmpl" at <templateString "usage/model_snippet.stmpl" .Local>: error calling templateString: failed to execute template: template: usage/model_snippet.stmpl:8:3: executing "usage/model_snippet.stmpl" at <templateModelSnippet $typeDef>: error calling templateModelSnippet: Error: Unexpected scope: shared at getAccessNamespace (includes/imports.ts:487:9(98)) at github.com/speakeasy-api/easytemplate.(*Engine).init.(*Engine).init.func1.func6 (native)`
	actual := FormatTemplateErrorString(input)
	expected := `GoError: failed to execute template:
  template: readme/model.stmpl:14:33
    executing "readme/model.stmpl" at <templateString "usage/model_snippet.stmpl" .Local>:
    error calling templateString:
      failed to execute template:
        template: usage/model_snippet.stmpl:8:3
          executing "usage/model_snippet.stmpl" at <templateModelSnippet $typeDef>
          error calling templateModelSnippet:
            Error: Unexpected scope: shared
            at getAccessNamespace (includes/imports.ts:487:9(98))
            at github.com/speakeasy-api/easytemplate.(*Engine).init.(*Engine).init.func1.func6 (native)`

	fmt.Println(actual)
	assert.Equal(t, expected, actual)
}
