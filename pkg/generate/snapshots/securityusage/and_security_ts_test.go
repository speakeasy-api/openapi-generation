package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestAndSecurity_Ts(t *testing.T) {
	t.Parallel()

	spec := specAndSecurity

	genYaml := `typescript:
  packageName: and-security-sdk
`

	expectedSnapshotFiles := []string{
		"docs/sdks/sdk/README.md",
	}

	expectedSnapshot := `--- docs/sdks/sdk/README.md ---
# SDK

## Overview

### Available Operations

* [globalSecurity](#globalsecurity)
* [andAuthHoisted](#andauthhoisted)
* [opLevelAndAuth](#oplevelandauth)

## globalSecurity

### Example Usage

<!-- UsageSnippet language="typescript" operationID="globalSecurity" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `typescript
import { SDK } from "and-security-sdk";

const sdk = new SDK({
  security: {
    auth1: "<YOUR_API_KEY_HERE>",
    auth2: {
      username: "",
      password: "",
    },
  },
});

async function run() {
  await sdk.globalSecurity();


}

run();
` + "`" + `` + "`" + `` + "`" + `

### Standalone function

The standalone function version of this method:

` + "`" + `` + "`" + `` + "`" + `typescript
import { SDKCore } from "and-security-sdk/core.js";
import { globalSecurity } from "and-security-sdk/funcs/global-security.js";

// Use ` + "`" + `SDKCore` + "`" + ` for best tree-shaking performance.
// You can create one instance of it to use across an application.
const sdk = new SDKCore({
  security: {
    auth1: "<YOUR_API_KEY_HERE>",
    auth2: {
      username: "",
      password: "",
    },
  },
});

async function run() {
  const res = await globalSecurity(sdk);
  if (res.ok) {
    const { value: result } = res;
    
  } else {
    console.log("globalSecurity failed:", res.error);
  }
}

run();
` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                                                                                                                                      | Type                                                                                                                                                                           | Required                                                                                                                                                                       | Description                                                                                                                                                                    |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| ` + "`" + `options` + "`" + `                                                                                                                                                                      | RequestOptions                                                                                                                                                                 | :heavy_minus_sign:                                                                                                                                                             | Used to set various options for making HTTP requests.                                                                                                                          |
| ` + "`" + `options.fetchOptions` + "`" + `                                                                                                                                                         | [RequestInit](https://developer.mozilla.org/en-US/docs/Web/API/Request/Request#options)                                                                                        | :heavy_minus_sign:                                                                                                                                                             | Options that are passed to the underlying HTTP request. This can be used to inject extra headers for examples. All ` + "`" + `Request` + "`" + ` options, except ` + "`" + `method` + "`" + ` and ` + "`" + `body` + "`" + `, are allowed. |
| ` + "`" + `options.retries` + "`" + `                                                                                                                                                              | [RetryConfig](../../lib/utils/retryconfig.md)                                                                                                                                  | :heavy_minus_sign:                                                                                                                                                             | Enables retrying HTTP requests under certain failure conditions.                                                                                                               |

### Response

**Promise\<void\>**

### Errors

| Error Type             | Status Code            | Content Type           |
| ---------------------- | ---------------------- | ---------------------- |
| errors.SDKDefaultError | 4XX, 5XX               | \*/\*                  |

## andAuthHoisted

### Example Usage

<!-- UsageSnippet language="typescript" operationID="andAuthHoisted" method="get" path="/op1" -->
` + "`" + `` + "`" + `` + "`" + `typescript
import { SDK } from "and-security-sdk";

const sdk = new SDK({
  security: {
    auth1: "<YOUR_API_KEY_HERE>",
    auth2: {
      username: "",
      password: "",
    },
  },
});

async function run() {
  await sdk.andAuthHoisted();


}

run();
` + "`" + `` + "`" + `` + "`" + `

### Standalone function

The standalone function version of this method:

` + "`" + `` + "`" + `` + "`" + `typescript
import { SDKCore } from "and-security-sdk/core.js";
import { andAuthHoisted } from "and-security-sdk/funcs/and-auth-hoisted.js";

// Use ` + "`" + `SDKCore` + "`" + ` for best tree-shaking performance.
// You can create one instance of it to use across an application.
const sdk = new SDKCore({
  security: {
    auth1: "<YOUR_API_KEY_HERE>",
    auth2: {
      username: "",
      password: "",
    },
  },
});

async function run() {
  const res = await andAuthHoisted(sdk);
  if (res.ok) {
    const { value: result } = res;
    
  } else {
    console.log("andAuthHoisted failed:", res.error);
  }
}

run();
` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                                                                                                                                      | Type                                                                                                                                                                           | Required                                                                                                                                                                       | Description                                                                                                                                                                    |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| ` + "`" + `options` + "`" + `                                                                                                                                                                      | RequestOptions                                                                                                                                                                 | :heavy_minus_sign:                                                                                                                                                             | Used to set various options for making HTTP requests.                                                                                                                          |
| ` + "`" + `options.fetchOptions` + "`" + `                                                                                                                                                         | [RequestInit](https://developer.mozilla.org/en-US/docs/Web/API/Request/Request#options)                                                                                        | :heavy_minus_sign:                                                                                                                                                             | Options that are passed to the underlying HTTP request. This can be used to inject extra headers for examples. All ` + "`" + `Request` + "`" + ` options, except ` + "`" + `method` + "`" + ` and ` + "`" + `body` + "`" + `, are allowed. |
| ` + "`" + `options.retries` + "`" + `                                                                                                                                                              | [RetryConfig](../../lib/utils/retryconfig.md)                                                                                                                                  | :heavy_minus_sign:                                                                                                                                                             | Enables retrying HTTP requests under certain failure conditions.                                                                                                               |

### Response

**Promise\<void\>**

### Errors

| Error Type             | Status Code            | Content Type           |
| ---------------------- | ---------------------- | ---------------------- |
| errors.SDKDefaultError | 4XX, 5XX               | \*/\*                  |

## opLevelAndAuth

### Example Usage

<!-- UsageSnippet language="typescript" operationID="opLevelAndAuth" method="get" path="/op2" -->
` + "`" + `` + "`" + `` + "`" + `typescript
import { SDK } from "and-security-sdk";

const sdk = new SDK();

async function run() {
  await sdk.opLevelAndAuth({
    auth2: {
      username: "",
      password: "",
    },
    auth3: "<YOUR_BEARER_TOKEN_HERE>",
  });


}

run();
` + "`" + `` + "`" + `` + "`" + `

### Standalone function

The standalone function version of this method:

` + "`" + `` + "`" + `` + "`" + `typescript
import { SDKCore } from "and-security-sdk/core.js";
import { opLevelAndAuth } from "and-security-sdk/funcs/op-level-and-auth.js";

// Use ` + "`" + `SDKCore` + "`" + ` for best tree-shaking performance.
// You can create one instance of it to use across an application.
const sdk = new SDKCore();

async function run() {
  const res = await opLevelAndAuth(sdk, {
    auth2: {
      username: "",
      password: "",
    },
    auth3: "<YOUR_BEARER_TOKEN_HERE>",
  });
  if (res.ok) {
    const { value: result } = res;
    
  } else {
    console.log("opLevelAndAuth failed:", res.error);
  }
}

run();
` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                                                                                                                                      | Type                                                                                                                                                                           | Required                                                                                                                                                                       | Description                                                                                                                                                                    |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| ` + "`" + `security` + "`" + `                                                                                                                                                                     | [operations.OpLevelAndAuthSecurity](../../models/operations/op-level-and-auth-security.md)                                                                                     | :heavy_check_mark:                                                                                                                                                             | The security requirements to use for the request.                                                                                                                              |
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
