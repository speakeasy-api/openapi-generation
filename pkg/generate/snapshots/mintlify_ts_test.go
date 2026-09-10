package snapshots

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

// TestSnapMintlifyMDX_TS asserts the mintlify documentation mode produces
// `.mdx` output with frontmatter, escaped MDX hazards, and rewritten
// intra-doc links. Locks in the contract qynn flagged in the PR review:
// when `docs/` formatting changes upstream, this snapshot diff surfaces
// breakage in the transform.
func TestSnapMintlifyMDX_TS(t *testing.T) {
	t.Parallel()
	spec := `openapi: 3.1.0
info:
  title: Mintlify Snapshot API
  version: 1.0.0
paths:
  /widgets/{id}:
    get:
      operationId: getWidget
      summary: Get a widget by ID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Widget'
components:
  schemas:
    Widget:
      type: object
      description: A widget — uses ` + "`<Map>`" + ` syntax in its description to exercise MDX hazard escaping.
      required: [id, name]
      properties:
        id:
          type: string
        name:
          type: string
`

	genYaml := `generation:
  documentation: mintlify
typescript:
  packageName: mintlifysnap
`

	includeGlobs := []string{
		"docs/sdks/sdk/README.mdx",
		"docs/models/widget.mdx",
	}

	expectedSnapshot := `--- docs/models/widget.mdx ---
---
title: "Widget"
---

A widget — uses ` + "`" + `<Map>` + "`" + ` syntax in its description to exercise MDX hazard escaping.

## Example Usage

` + "`" + `` + "`" + `` + "`" + `typescript
import { Widget } from "mintlifysnap/models";

let value: Widget = {
  id: "<id>",
  name: "<value>",
};
` + "`" + `` + "`" + `` + "`" + `

## Fields

| Field              | Type               | Required           | Description        |
| ------------------ | ------------------ | ------------------ | ------------------ |
| ` + "`" + `id` + "`" + `               | *string*           | :heavy_check_mark: | N/A                |
| ` + "`" + `name` + "`" + `             | *string*           | :heavy_check_mark: | N/A                |

--- docs/sdks/sdk/README.mdx ---
---
title: "SDK"
---

## Overview

### Available Operations

* [getWidget](#getwidget) - Get a widget by ID

## getWidget

Get a widget by ID

### Example Usage

` + "`" + `` + "`" + `` + "`" + `typescript
import { SDK } from "mintlifysnap";

const sdk = new SDK({
  serverURL: "https://api.example.com",
});

async function run() {
  const result = await sdk.getWidget({
    id: "<id>",
  });

  console.log(result);
}

run();
` + "`" + `` + "`" + `` + "`" + `

### Standalone function

The standalone function version of this method:

` + "`" + `` + "`" + `` + "`" + `typescript
import { SDKCore } from "mintlifysnap/core.js";
import { getWidget } from "mintlifysnap/funcs/get-widget.js";

// Use ` + "`" + `SDKCore` + "`" + ` for best tree-shaking performance.
// You can create one instance of it to use across an application.
const sdk = new SDKCore({
  serverURL: "https://api.example.com",
});

async function run() {
  const res = await getWidget(sdk, {
    id: "<id>",
  });
  if (res.ok) {
    const { value: result } = res;
    console.log(result);
  } else {
    console.log("getWidget failed:", res.error);
  }
}

run();
` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                                                                                                                                      | Type                                                                                                                                                                           | Required                                                                                                                                                                       | Description                                                                                                                                                                    |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| ` + "`" + `request` + "`" + `                                                                                                                                                                      | [operations.GetWidgetRequest](../../models/operations/get-widget-request.mdx)                                                                                                   | :heavy_check_mark:                                                                                                                                                             | The request object to use for the request.                                                                                                                                     |
| ` + "`" + `options` + "`" + `                                                                                                                                                                      | RequestOptions                                                                                                                                                                 | :heavy_minus_sign:                                                                                                                                                             | Used to set various options for making HTTP requests.                                                                                                                          |
| ` + "`" + `options.fetchOptions` + "`" + `                                                                                                                                                         | [RequestInit](https://developer.mozilla.org/en-US/docs/Web/API/Request/Request#options)                                                                                        | :heavy_minus_sign:                                                                                                                                                             | Options that are passed to the underlying HTTP request. This can be used to inject extra headers for examples. All ` + "`" + `Request` + "`" + ` options, except ` + "`" + `method` + "`" + ` and ` + "`" + `body` + "`" + `, are allowed. |
| ` + "`" + `options.retries` + "`" + `                                                                                                                                                              | [RetryConfig](../../lib/utils/retryconfig.mdx)                                                                                                                                  | :heavy_minus_sign:                                                                                                                                                             | Enables retrying HTTP requests under certain failure conditions.                                                                                                               |

### Response

**Promise\<[models.Widget](../../models/widget.mdx)\>**

### Errors

| Error Type             | Status Code            | Content Type           |
| ---------------------- | ---------------------- | ---------------------- |
| errors.SDKDefaultError | 4XX, 5XX               | \*/\*                  |

` // end of snapshot

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         spec,
		GenYaml:      genYaml,
		IncludeGlobs: includeGlobs,
		Expected:     expectedSnapshot,
	})
}
