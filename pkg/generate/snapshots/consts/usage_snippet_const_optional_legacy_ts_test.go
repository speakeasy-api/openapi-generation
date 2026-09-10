package consts

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestSnapTsUsageSnippetConstOptionalLegacy(t *testing.T) {
	t.Parallel()
	spec := `openapi: 3.1.0
info:
  title: DefaultsAndConsts API
  version: 1.0.0
paths:
  /test:
    post:
      operationId: testConst
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ConstModel'
      responses:
        '200':
          description: OK
components:
  schemas:
    ConstModel:
      type: object
      required:
        - requiredConstString
        - requiredConstNumber
        - requiredConstBoolean
        - requiredConstEnum
        - requiredConstDate
        - requiredConstDateTime
        - requiredConstBigint
        - requiredConstDecimal
        - requiredDefaultString
        - requiredDefaultNumber
        - requiredDefaultBoolean
        - requiredDefaultEnum
        - requiredDefaultDate
        - requiredDefaultDateTime
        - requiredDefaultBigint
        - requiredDefaultDecimal
        - requiredConstDefaultString
        - requiredConstDefaultNumber
        - requiredConstDefaultBoolean
        - requiredConstDefaultEnum
        - requiredConstDefaultDate
        - requiredConstDefaultDateTime
        - requiredConstDefaultBigint
        - requiredConstDefaultDecimal
      properties:
        requiredConstString:
          type: string
          const: fixed-string
        requiredConstNumber:
          type: number
          const: 42.5
        requiredConstBoolean:
          type: boolean
          const: true
        requiredConstEnum:
          type: string
          enum: ["A", "B"]
          const: "A"
        requiredConstDate:
          type: string
          format: date
          const: "2023-12-25"
        requiredConstDateTime:
          type: string
          format: date-time
          const: "2023-12-25T10:30:00Z"
        requiredConstBigint:
          type: integer
          format: bigint
          const: 9007199254740991
        requiredConstDecimal:
          type: number
          format: decimal
          const: 123.456789
        optionalConstString:
          type: string
          const: "optional-const-string"
        optionalConstNumber:
          type: number
          const: 123.45
        optionalConstBoolean:
          type: boolean
          const: false
        optionalConstEnum:
          type: string
          enum: ["A", "B"]
          const: "A"
        optionalConstDate:
          type: string
          format: date
          const: "2025-06-15"
        optionalConstDateTime:
          type: string
          format: date-time
          const: "2025-06-15T15:30:00Z"
        optionalConstBigint:
          type: integer
          format: bigint
          const: 987654321098765432
        optionalConstDecimal:
          type: number
          format: decimal
          const: 456.123789
        requiredDefaultString:
          type: string
          default: "default-string"
        requiredDefaultNumber:
          type: number
          default: 99.9
        requiredDefaultBoolean:
          type: boolean
          default: false
        requiredDefaultEnum:
          type: string
          enum: ["A", "B"]
          default: "A"
        requiredDefaultDate:
          type: string
          format: date
          default: "2024-01-01"
        requiredDefaultDateTime:
          type: string
          format: date-time
          default: "2024-01-01T00:00:00Z"
        requiredDefaultBigint:
          type: integer
          format: bigint
          default: 1234567890123456789
        requiredDefaultDecimal:
          type: number
          format: decimal
          default: 999.999999
        optionalDefaultString:
          type: string
          default: "optional-default"
        optionalDefaultNumber:
          type: number
          default: 77.7
        optionalDefaultBoolean:
          type: boolean
          default: true
        optionalDefaultEnum:
          type: string
          enum: ["A", "B"]
          default: "A"
        optionalDefaultDate:
          type: string
          format: date
          default: "2024-12-31"
        optionalDefaultDateTime:
          type: string
          format: date-time
          default: "2024-12-31T23:59:59Z"
        optionalDefaultBigint:
          type: integer
          format: bigint
          default: 555666777888999
        optionalDefaultDecimal:
          type: number
          format: decimal
          default: 111.222333
        optionalString:
          type: string
        optionalNumber:
          type: number
        optionalBoolean:
          type: boolean
        optionalEnum:
          type: string
          enum: ["A", "B"]
        optionalDate:
          type: string
          format: date
        optionalDateTime:
          type: string
          format: date-time
        optionalBigint:
          type: integer
          format: bigint
        optionalDecimal:
          type: number
          format: decimal
        requiredConstDefaultString:
          type: string
          const: "fixed-string"
          default: "fixed-string"
        requiredConstDefaultNumber:
          type: number
          const: 42.5
          default: 42.5
        requiredConstDefaultBoolean:
          type: boolean
          const: true
          default: true
        requiredConstDefaultEnum:
          type: string
          enum: ["A", "B"]
          const: "A"
          default: "A"
        requiredConstDefaultDate:
          type: string
          format: date
          const: "2023-12-25"
          default: "2023-12-25"
        requiredConstDefaultDateTime:
          type: string
          format: date-time
          const: "2023-12-25T10:30:00Z"
          default: "2023-12-25T10:30:00Z"
        requiredConstDefaultBigint:
          type: integer
          format: bigint
          const: 9007199254740991
          default: 9007199254740991
        requiredConstDefaultDecimal:
          type: number
          format: decimal
          const: 123.456789
          default: 123.456789
        optionalConstDefaultString:
          type: string
          const: "optional-const-string"
          default: "optional-const-string"
        optionalConstDefaultNumber:
          type: number
          const: 123.45
          default: 123.45
        optionalConstDefaultBoolean:
          type: boolean
          const: false
          default: false
        optionalConstDefaultEnum:
          type: string
          enum: ["A", "B"]
          const: "A"
          default: "A"
        optionalConstDefaultDate:
          type: string
          format: date
          const: "2025-06-15"
          default: "2025-06-15"
        optionalConstDefaultDateTime:
          type: string
          format: date-time
          const: "2025-06-15T15:30:00Z"
          default: "2025-06-15T15:30:00Z"
        optionalConstDefaultBigint:
          type: integer
          format: bigint
          const: 987654321098765432
          default: 987654321098765432
        optionalConstDefaultDecimal:
          type: number
          format: decimal
          const: 456.123789
          default: 456.123789`

	genYaml := `typescript:
  packageName: defaultsandconsts
  constFieldsAlwaysOptional: false
`

	// expectedSnapshotFiles defines which generated files to include in the snapshot
	expectedSnapshotFiles := []string{
		"docs/sdks/sdk/README.md",
	}

	expectedSnapshot := `--- docs/sdks/sdk/README.md ---
# SDK

## Overview

### Available Operations

* [testConst](#testconst)

## testConst

### Example Usage

<!-- UsageSnippet language="typescript" operationID="testConst" method="post" path="/test" -->
` + "`" + `` + "`" + `` + "`" + `typescript
import { SDK } from "defaultsandconsts";
import { Decimal } from "defaultsandconsts/types";

const sdk = new SDK({
  serverURL: "https://api.example.com",
});

async function run() {
  await sdk.testConst({
    requiredConstString: "fixed-string",
    requiredConstNumber: 42.5,
    requiredConstBoolean: true,
    requiredConstEnum: "A",
    requiredConstDate: new Date("2023-12-25"),
    requiredConstDateTime: new Date("2023-12-25T10:30:00Z"),
    requiredConstBigint: 9007199254740991,
    requiredConstDecimal: new Decimal("123.456789"),
  });


}

run();
` + "`" + `` + "`" + `` + "`" + `

### Standalone function

The standalone function version of this method:

` + "`" + `` + "`" + `` + "`" + `typescript
import { SDKCore } from "defaultsandconsts/core.js";
import { testConst } from "defaultsandconsts/funcs/test-const.js";
import { Decimal } from "defaultsandconsts/types";

// Use ` + "`" + `SDKCore` + "`" + ` for best tree-shaking performance.
// You can create one instance of it to use across an application.
const sdk = new SDKCore({
  serverURL: "https://api.example.com",
});

async function run() {
  const res = await testConst(sdk, {
    requiredConstString: "fixed-string",
    requiredConstNumber: 42.5,
    requiredConstBoolean: true,
    requiredConstEnum: "A",
    requiredConstDate: new Date("2023-12-25"),
    requiredConstDateTime: new Date("2023-12-25T10:30:00Z"),
    requiredConstBigint: 9007199254740991,
    requiredConstDecimal: new Decimal("123.456789"),
  });
  if (res.ok) {
    const { value: result } = res;
    
  } else {
    console.log("testConst failed:", res.error);
  }
}

run();
` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                                                                                                                                      | Type                                                                                                                                                                           | Required                                                                                                                                                                       | Description                                                                                                                                                                    |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| ` + "`" + `request` + "`" + `                                                                                                                                                                      | [models.ConstModel](../../models/const-model.md)                                                                                                                               | :heavy_check_mark:                                                                                                                                                             | The request object to use for the request.                                                                                                                                     |
| ` + "`" + `options` + "`" + `                                                                                                                                                                      | RequestOptions                                                                                                                                                                 | :heavy_minus_sign:                                                                                                                                                             | Used to set various options for making HTTP requests.                                                                                                                          |
| ` + "`" + `options.fetchOptions` + "`" + `                                                                                                                                                         | [RequestInit](https://developer.mozilla.org/en-US/docs/Web/API/Request/Request#options)                                                                                        | :heavy_minus_sign:                                                                                                                                                             | Options that are passed to the underlying HTTP request. This can be used to inject extra headers for examples. All ` + "`" + `Request` + "`" + ` options, except ` + "`" + `method` + "`" + ` and ` + "`" + `body` + "`" + `, are allowed. |
| ` + "`" + `options.retries` + "`" + `                                                                                                                                                              | [RetryConfig](../../lib/utils/retryconfig.md)                                                                                                                                  | :heavy_minus_sign:                                                                                                                                                             | Enables retrying HTTP requests under certain failure conditions.                                                                                                               |

### Response

**Promise\<void\>**

### Errors

| Error Type             | Status Code            | Content Type           |
| ---------------------- | ---------------------- | ---------------------- |
| errors.SDKDefaultError | 4XX, 5XX               | \*/\*                  |

` // end of snapshot

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         spec,
		GenYaml:      genYaml,
		IncludeGlobs: expectedSnapshotFiles,
		Expected:     expectedSnapshot,
	})
}
