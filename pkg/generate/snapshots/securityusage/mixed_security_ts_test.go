package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestMixedSecurity_Ts(t *testing.T) {
	t.Parallel()

	spec := specMixedSecurity

	genYaml := `typescript:
  packageName: mixed-security-sdk
`

	expectedSnapshotFiles := []string{
		"docs/sdks/sdk/README.md",
	}

	expectedSnapshot := `--- docs/sdks/sdk/README.md ---
# SDK

## Overview

### Available Operations

* [globalSecurity](#globalsecurity)
* [option1Hoisted](#option1hoisted)
* [option2Hoisted](#option2hoisted)
* [option1NotAllowed](#option1notallowed)
* [opLevelMixedAuth](#oplevelmixedauth)

## globalSecurity

### Example Usage

<!-- UsageSnippet language="typescript" operationID="globalSecurity" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `typescript
import { SDK } from "mixed-security-sdk";

const sdk = new SDK({
  security: {
    option1: {
      authA1: "<YOUR_API_KEY_HERE>",
      authA2: {
        username: "",
        password: "",
      },
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
import { SDKCore } from "mixed-security-sdk/core.js";
import { globalSecurity } from "mixed-security-sdk/funcs/global-security.js";

// Use ` + "`" + `SDKCore` + "`" + ` for best tree-shaking performance.
// You can create one instance of it to use across an application.
const sdk = new SDKCore({
  security: {
    option1: {
      authA1: "<YOUR_API_KEY_HERE>",
      authA2: {
        username: "",
        password: "",
      },
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

## option1Hoisted

### Example Usage

<!-- UsageSnippet language="typescript" operationID="option1Hoisted" method="get" path="/opA" -->
` + "`" + `` + "`" + `` + "`" + `typescript
import { SDK } from "mixed-security-sdk";

const sdk = new SDK({
  security: {
    option1: {
      authA1: "<YOUR_API_KEY_HERE>",
      authA2: {
        username: "",
        password: "",
      },
    },
  },
});

async function run() {
  await sdk.option1Hoisted();


}

run();
` + "`" + `` + "`" + `` + "`" + `

### Standalone function

The standalone function version of this method:

` + "`" + `` + "`" + `` + "`" + `typescript
import { SDKCore } from "mixed-security-sdk/core.js";
import { option1Hoisted } from "mixed-security-sdk/funcs/option1-hoisted.js";

// Use ` + "`" + `SDKCore` + "`" + ` for best tree-shaking performance.
// You can create one instance of it to use across an application.
const sdk = new SDKCore({
  security: {
    option1: {
      authA1: "<YOUR_API_KEY_HERE>",
      authA2: {
        username: "",
        password: "",
      },
    },
  },
});

async function run() {
  const res = await option1Hoisted(sdk);
  if (res.ok) {
    const { value: result } = res;
    
  } else {
    console.log("option1Hoisted failed:", res.error);
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

## option2Hoisted

### Example Usage

<!-- UsageSnippet language="typescript" operationID="option2Hoisted" method="get" path="/opB" -->
` + "`" + `` + "`" + `` + "`" + `typescript
import { SDK } from "mixed-security-sdk";

const sdk = new SDK({
  security: {
    option2: {
      authB: "<YOUR_JWT>",
    },
  },
});

async function run() {
  await sdk.option2Hoisted();


}

run();
` + "`" + `` + "`" + `` + "`" + `

### Standalone function

The standalone function version of this method:

` + "`" + `` + "`" + `` + "`" + `typescript
import { SDKCore } from "mixed-security-sdk/core.js";
import { option2Hoisted } from "mixed-security-sdk/funcs/option2-hoisted.js";

// Use ` + "`" + `SDKCore` + "`" + ` for best tree-shaking performance.
// You can create one instance of it to use across an application.
const sdk = new SDKCore({
  security: {
    option2: {
      authB: "<YOUR_JWT>",
    },
  },
});

async function run() {
  const res = await option2Hoisted(sdk);
  if (res.ok) {
    const { value: result } = res;
    
  } else {
    console.log("option2Hoisted failed:", res.error);
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

## option1NotAllowed

### Example Usage

<!-- UsageSnippet language="typescript" operationID="option1NotAllowed" method="get" path="/opBC" -->
` + "`" + `` + "`" + `` + "`" + `typescript
import { SDK } from "mixed-security-sdk";

const sdk = new SDK({
  security: {
    option3: {
      authC: "<YOUR_AUTH_C_HERE>",
    },
  },
});

async function run() {
  await sdk.option1NotAllowed();


}

run();
` + "`" + `` + "`" + `` + "`" + `

### Standalone function

The standalone function version of this method:

` + "`" + `` + "`" + `` + "`" + `typescript
import { SDKCore } from "mixed-security-sdk/core.js";
import { option1NotAllowed } from "mixed-security-sdk/funcs/option1-not-allowed.js";

// Use ` + "`" + `SDKCore` + "`" + ` for best tree-shaking performance.
// You can create one instance of it to use across an application.
const sdk = new SDKCore({
  security: {
    option3: {
      authC: "<YOUR_AUTH_C_HERE>",
    },
  },
});

async function run() {
  const res = await option1NotAllowed(sdk);
  if (res.ok) {
    const { value: result } = res;
    
  } else {
    console.log("option1NotAllowed failed:", res.error);
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

## opLevelMixedAuth

### Example Usage

<!-- UsageSnippet language="typescript" operationID="opLevelMixedAuth" method="get" path="/not/hoisted" -->
` + "`" + `` + "`" + `` + "`" + `typescript
import { SDK } from "mixed-security-sdk";

const sdk = new SDK();

async function run() {
  await sdk.opLevelMixedAuth({
    option1: {
      authA1: "<YOUR_API_KEY_HERE>",
      authB: "<YOUR_JWT>",
    },
  });


}

run();
` + "`" + `` + "`" + `` + "`" + `

### Standalone function

The standalone function version of this method:

` + "`" + `` + "`" + `` + "`" + `typescript
import { SDKCore } from "mixed-security-sdk/core.js";
import { opLevelMixedAuth } from "mixed-security-sdk/funcs/op-level-mixed-auth.js";

// Use ` + "`" + `SDKCore` + "`" + ` for best tree-shaking performance.
// You can create one instance of it to use across an application.
const sdk = new SDKCore();

async function run() {
  const res = await opLevelMixedAuth(sdk, {
    option1: {
      authA1: "<YOUR_API_KEY_HERE>",
      authB: "<YOUR_JWT>",
    },
  });
  if (res.ok) {
    const { value: result } = res;
    
  } else {
    console.log("opLevelMixedAuth failed:", res.error);
  }
}

run();
` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                                                                                                                                      | Type                                                                                                                                                                           | Required                                                                                                                                                                       | Description                                                                                                                                                                    |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| ` + "`" + `security` + "`" + `                                                                                                                                                                     | [operations.OpLevelMixedAuthSecurity](../../models/operations/op-level-mixed-auth-security.md)                                                                                 | :heavy_check_mark:                                                                                                                                                             | The security requirements to use for the request.                                                                                                                              |
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
