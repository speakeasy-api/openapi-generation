package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestOrSecurity_Ts(t *testing.T) {
	t.Parallel()

	spec := specOrSecurity

	genYaml := `typescript:
  packageName: multi-auth-sdk
`

	expectedSnapshotFiles := []string{
		"docs/sdks/sdk/README.md",
	}

	expectedSnapshot := `--- docs/sdks/sdk/README.md ---
# SDK

## Overview

### Available Operations

* [globalSecurity](#globalsecurity)
* [auth1Hoisted](#auth1hoisted)
* [auth2Hoisted](#auth2hoisted)
* [auth2Preferred](#auth2preferred)
* [opLevelClientCredentials](#oplevelclientcredentials)

## globalSecurity

### Example Usage

<!-- UsageSnippet language="typescript" operationID="globalSecurity" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `typescript
import { SDK } from "multi-auth-sdk";

const sdk = new SDK({
  security: {
    auth1: "<YOUR_API_KEY_HERE>",
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
import { SDKCore } from "multi-auth-sdk/core.js";
import { globalSecurity } from "multi-auth-sdk/funcs/global-security.js";

// Use ` + "`" + `SDKCore` + "`" + ` for best tree-shaking performance.
// You can create one instance of it to use across an application.
const sdk = new SDKCore({
  security: {
    auth1: "<YOUR_API_KEY_HERE>",
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

## auth1Hoisted

### Example Usage

<!-- UsageSnippet language="typescript" operationID="auth1Hoisted" method="get" path="/op1" -->
` + "`" + `` + "`" + `` + "`" + `typescript
import { SDK } from "multi-auth-sdk";

const sdk = new SDK({
  security: {
    auth1: "<YOUR_API_KEY_HERE>",
  },
});

async function run() {
  await sdk.auth1Hoisted();


}

run();
` + "`" + `` + "`" + `` + "`" + `

### Standalone function

The standalone function version of this method:

` + "`" + `` + "`" + `` + "`" + `typescript
import { SDKCore } from "multi-auth-sdk/core.js";
import { auth1Hoisted } from "multi-auth-sdk/funcs/auth1-hoisted.js";

// Use ` + "`" + `SDKCore` + "`" + ` for best tree-shaking performance.
// You can create one instance of it to use across an application.
const sdk = new SDKCore({
  security: {
    auth1: "<YOUR_API_KEY_HERE>",
  },
});

async function run() {
  const res = await auth1Hoisted(sdk);
  if (res.ok) {
    const { value: result } = res;
    
  } else {
    console.log("auth1Hoisted failed:", res.error);
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

## auth2Hoisted

### Example Usage

<!-- UsageSnippet language="typescript" operationID="auth2Hoisted" method="get" path="/op2" -->
` + "`" + `` + "`" + `` + "`" + `typescript
import { SDK } from "multi-auth-sdk";

const sdk = new SDK({
  security: {
    auth2: {
      username: "",
      password: "",
    },
  },
});

async function run() {
  await sdk.auth2Hoisted();


}

run();
` + "`" + `` + "`" + `` + "`" + `

### Standalone function

The standalone function version of this method:

` + "`" + `` + "`" + `` + "`" + `typescript
import { SDKCore } from "multi-auth-sdk/core.js";
import { auth2Hoisted } from "multi-auth-sdk/funcs/auth2-hoisted.js";

// Use ` + "`" + `SDKCore` + "`" + ` for best tree-shaking performance.
// You can create one instance of it to use across an application.
const sdk = new SDKCore({
  security: {
    auth2: {
      username: "",
      password: "",
    },
  },
});

async function run() {
  const res = await auth2Hoisted(sdk);
  if (res.ok) {
    const { value: result } = res;
    
  } else {
    console.log("auth2Hoisted failed:", res.error);
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

## auth2Preferred

### Example Usage

<!-- UsageSnippet language="typescript" operationID="auth2Preferred" method="get" path="/op3" -->
` + "`" + `` + "`" + `` + "`" + `typescript
import { SDK } from "multi-auth-sdk";

const sdk = new SDK({
  security: {
    auth2: {
      username: "",
      password: "",
    },
  },
});

async function run() {
  await sdk.auth2Preferred();


}

run();
` + "`" + `` + "`" + `` + "`" + `

### Standalone function

The standalone function version of this method:

` + "`" + `` + "`" + `` + "`" + `typescript
import { SDKCore } from "multi-auth-sdk/core.js";
import { auth2Preferred } from "multi-auth-sdk/funcs/auth2-preferred.js";

// Use ` + "`" + `SDKCore` + "`" + ` for best tree-shaking performance.
// You can create one instance of it to use across an application.
const sdk = new SDKCore({
  security: {
    auth2: {
      username: "",
      password: "",
    },
  },
});

async function run() {
  const res = await auth2Preferred(sdk);
  if (res.ok) {
    const { value: result } = res;
    
  } else {
    console.log("auth2Preferred failed:", res.error);
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

## opLevelClientCredentials

### Example Usage

<!-- UsageSnippet language="typescript" operationID="opLevelClientCredentials" method="get" path="/op4" -->
` + "`" + `` + "`" + `` + "`" + `typescript
import { SDK } from "multi-auth-sdk";

const sdk = new SDK();

async function run() {
  await sdk.opLevelClientCredentials({
    clientID: "<YOUR_CLIENT_ID_HERE>",
    clientSecret: "<YOUR_CLIENT_SECRET_HERE>",
  });


}

run();
` + "`" + `` + "`" + `` + "`" + `

### Standalone function

The standalone function version of this method:

` + "`" + `` + "`" + `` + "`" + `typescript
import { SDKCore } from "multi-auth-sdk/core.js";
import { opLevelClientCredentials } from "multi-auth-sdk/funcs/op-level-client-credentials.js";

// Use ` + "`" + `SDKCore` + "`" + ` for best tree-shaking performance.
// You can create one instance of it to use across an application.
const sdk = new SDKCore();

async function run() {
  const res = await opLevelClientCredentials(sdk, {
    clientID: "<YOUR_CLIENT_ID_HERE>",
    clientSecret: "<YOUR_CLIENT_SECRET_HERE>",
  });
  if (res.ok) {
    const { value: result } = res;
    
  } else {
    console.log("opLevelClientCredentials failed:", res.error);
  }
}

run();
` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                                                                                                                                      | Type                                                                                                                                                                           | Required                                                                                                                                                                       | Description                                                                                                                                                                    |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| ` + "`" + `security` + "`" + `                                                                                                                                                                     | [operations.OpLevelClientCredentialsSecurity](../../models/operations/op-level-client-credentials-security.md)                                                                 | :heavy_check_mark:                                                                                                                                                             | The security requirements to use for the request.                                                                                                                              |
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
