# openapi

Developer-friendly & type-safe Typescript SDK specifically catered to leverage *openapi* API.

[![Built by Speakeasy](https://img.shields.io/badge/Built_by-SPEAKEASY-374151?style=for-the-badge&labelColor=f3f4f6)](https://www.speakeasy.com/?utm_source=openapi&utm_campaign=typescript)
[![License: MIT](https://img.shields.io/badge/LICENSE_//_MIT-3b5bdb?style=for-the-badge&labelColor=eff6ff)](https://opensource.org/licenses/MIT)


<br /><br />
> [!IMPORTANT]
> This SDK is not yet ready for production use. Delete this section before > publishing to a package manager.

<!-- Start Summary [summary] -->
## Summary

SDK Review: A test document for reviewing the SDK.

This document will show case as many of our features as possible in as little operations/models as possible.
This will then generate a SDK that we can more easily review than the test SDKs based on uber.yaml spec.

For more information about the API: [Speakeasy Docs](https://speakeasy.com/docs)
<!-- End Summary [summary] -->

<!-- Start Table of Contents [toc] -->
## Table of Contents
<!-- $toc-max-depth=2 -->
* [openapi](#openapi)
  * [SDK Installation](#sdk-installation)
  * [Requirements](#requirements)
  * [SDK Example Usage](#sdk-example-usage)
  * [Authentication](#authentication)
  * [Available Resources and Operations](#available-resources-and-operations)
  * [Standalone functions](#standalone-functions)
  * [React hooks with TanStack Query](#react-hooks-with-tanstack-query)
  * [Global Parameters](#global-parameters)
  * [Server-sent event streaming](#server-sent-event-streaming)
  * [Pagination](#pagination)
  * [File uploads](#file-uploads)
  * [Retries](#retries)
  * [Error Handling](#error-handling)
  * [Server Selection](#server-selection)
  * [Custom HTTP Client](#custom-http-client)
  * [Debugging](#debugging)
* [Development](#development)
  * [Maturity](#maturity)
  * [Contributions](#contributions)

<!-- End Table of Contents [toc] -->

<!-- Start SDK Installation [installation] -->
## SDK Installation

> [!TIP]
> To finish publishing your SDK to npm and others you must [run your first generation action](https://www.speakeasy.com/docs/github-setup#step-by-step-guide).


The SDK can be installed with either [npm](https://www.npmjs.com/), [pnpm](https://pnpm.io/), [bun](https://bun.sh/) or [yarn](https://classic.yarnpkg.com/en/) package managers.

### NPM

```bash
npm add https://github.com/speakeasy-sdks/test-sdk
# Install optional peer dependencies if you plan to use React hooks
npm add @tanstack/react-query react react-dom
```

### PNPM

```bash
pnpm add https://github.com/speakeasy-sdks/test-sdk
# Install optional peer dependencies if you plan to use React hooks
pnpm add @tanstack/react-query react react-dom
```

### Bun

```bash
bun add https://github.com/speakeasy-sdks/test-sdk
# Install optional peer dependencies if you plan to use React hooks
bun add @tanstack/react-query react react-dom
```

### Yarn

```bash
yarn add https://github.com/speakeasy-sdks/test-sdk
# Install optional peer dependencies if you plan to use React hooks
yarn add @tanstack/react-query react react-dom
```

> [!NOTE]
> This package is published with CommonJS and ES Modules (ESM) support.
<!-- End SDK Installation [installation] -->

<!-- Start Requirements [requirements] -->
## Requirements

For supported JavaScript runtimes, please consult [RUNTIMES.md](RUNTIMES.md).
<!-- End Requirements [requirements] -->

<!-- Start SDK Example Usage [usage] -->
## SDK Example Usage

### Example 1

```typescript
import { openAsBlob } from "node:fs";
import { SDK } from "openapi";

const sdk = new SDK();

async function run() {
  const result = await sdk.postFile({
    upload: await openAsBlob("example.file"),
  });

  console.log(result);
}

run();

```

### Example 2

```typescript
import { openAsBlob } from "node:fs";
import { SDK } from "openapi";

const sdk = new SDK();

async function run() {
  const result = await sdk.tag1.postFileWithEncoding({
    file: await openAsBlob("example.file"),
  });

  console.log(result);
}

run();

```

### Example 3

```typescript
import {
  HeroWidth,
  Icon,
  Int32Enum,
  IntEnum,
  NullableIntEnum,
  NullableStringEnum,
  SDK,
  Type,
} from "openapi";
import { Decimal } from "openapi/types";

const sdk = new SDK({
  deprecatedQueryParam1: "some example query param",
  deprecatedQueryParam2: "some example query param",
  security: {
    myApiKey: {
      myApiKey: process.env["SPEAKEASY_MY_API_KEY"] ?? "",
    },
  },
});

async function run() {
  const result = await sdk.testGroup.tag2.postTest({
    obj: {
      str: "example",
      bool: true,
      integer: 999999,
      int32: 1,
      num: 1.1,
      float32: 8499.3,
      date: new Date("2020-01-01"),
      dateTime: new Date("2020-01-01T00:00:00Z"),
      anything: "<value>",
      boolOpt: true,
      intOptNull: 999999,
      numOptNull: 1.1,
      intEnum: IntEnum.Third,
      int32Enum: Int32Enum.SixtyNine,
      bigint: 593288,
      decimalStr: new Decimal("7028.3"),
      obj: {
        str: "example",
      },
      map: {},
      arr: [],
      any: "<value>",
      nullableIntEnum: NullableIntEnum.Third,
      nullableStringEnum: NullableStringEnum.Second,
      color: "green",
      icon: Icon.Tick,
      heroWidth: HeroWidth.FourHundredAndEighty,
    },
    type: Type.SuperType1,
  });

  console.log(result);
}

run();

```

### A custom readme heading

A custom usage description

```typescript
import { QueryParam2, SDK } from "openapi";

const sdk = new SDK({
  queryParam1: "some example query param",
  security: {
    myApiKey: {
      myApiKey: process.env["SPEAKEASY_MY_API_KEY"] ?? "",
    },
  },
});

async function run() {
  const result = await sdk.tag1.listTest1(
    100,
    QueryParam2.One,
    "some example header param",
  );

  for await (const page of result) {
    console.log(page);
  }
}

run();

```
<!-- End SDK Example Usage [usage] -->

<!-- Start Authentication [security] -->
## Authentication

### Per-Client Security Schemes

This SDK supports multiple security scheme combinations globally. You can choose from one of the alternatives by setting the `security` optional parameter when initializing the SDK client instance. The selected option will be used by default to authenticate with the API for all operations that support it.

#### UserPassAuth

The `UserPassAuth` alternative relies on the following scheme:

| Name                      | Type | Scheme     | Environment Variable                          |
| ------------------------- | ---- | ---------- | --------------------------------------------- |
| `username`<br/>`password` | http | HTTP Basic | `SPEAKEASY_USERNAME`<br/>`SPEAKEASY_PASSWORD` |

Example:
```typescript
import { openAsBlob } from "node:fs";
import { SDK } from "openapi";

const sdk = new SDK({
  security: {
    userPassAuth: {
      username: "<USERNAME>",
      password: "<PASSWORD>",
    },
  },
});

async function run() {
  const result = await sdk.postFile({
    upload: await openAsBlob("example.file"),
  });

  console.log(result);
}

run();

```

#### Option2

All of the following schemes must be satisfied to use the `Option2` alternative:

| Name         | Type   | Scheme      | Environment Variable    |
| ------------ | ------ | ----------- | ----------------------- |
| `bearerAuth` | http   | HTTP Bearer | `SPEAKEASY_BEARER_AUTH` |
| `myApiKey`   | apiKey | API key     | `SPEAKEASY_MY_API_KEY`  |

Example:
```typescript
import { openAsBlob } from "node:fs";
import { SDK } from "openapi";

const sdk = new SDK({
  security: {
    option2: {
      bearerAuth: "<YOUR_JWT>",
      myApiKey: process.env["SPEAKEASY_MY_API_KEY"] ?? "",
    },
  },
});

async function run() {
  const result = await sdk.postFile({
    upload: await openAsBlob("example.file"),
  });

  console.log(result);
}

run();

```

#### Option3

The `Option3` alternative relies on the following scheme:

| Name     | Type   | Scheme       | Environment Variable |
| -------- | ------ | ------------ | -------------------- |
| `oauth2` | oauth2 | OAuth2 token | `SPEAKEASY_OAUTH2`   |

Example:
```typescript
import { openAsBlob } from "node:fs";
import { SDK } from "openapi";

const sdk = new SDK({
  security: {
    option3: {
      oauth2: "Bearer <YOUR_OAUTH2_TOKEN>",
    },
  },
});

async function run() {
  const result = await sdk.postFile({
    upload: await openAsBlob("example.file"),
  });

  console.log(result);
}

run();

```

#### Option4

The `Option4` alternative relies on the following scheme:

| Name                 | Type | Scheme      | Environment Variable                      |
| -------------------- | ---- | ----------- | ----------------------------------------- |
| `appId`<br/>`secret` | http | Custom HTTP | `SPEAKEASY_APP_ID`<br/>`SPEAKEASY_SECRET` |

Example:
```typescript
import { openAsBlob } from "node:fs";
import { SDK } from "openapi";

const sdk = new SDK({
  security: {
    option4: {
      appId: "app-speakeasy-123",
      secret: "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI",
    },
  },
});

async function run() {
  const result = await sdk.postFile({
    upload: await openAsBlob("example.file"),
  });

  console.log(result);
}

run();

```

#### Option5

The `Option5` alternative relies on the following scheme:

| Name         | Type   | Scheme       | Environment Variable    |
| ------------ | ------ | ------------ | ----------------------- |
| `mobileAuth` | oauth2 | OAuth2 token | `SPEAKEASY_MOBILE_AUTH` |

Example:
```typescript
import { openAsBlob } from "node:fs";
import { SDK } from "openapi";

const sdk = new SDK({
  security: {
    option5: {
      mobileAuth: "Bearer <YOUR_OAUTH2_TOKEN>",
    },
  },
});

async function run() {
  const result = await sdk.postFile({
    upload: await openAsBlob("example.file"),
  });

  console.log(result);
}

run();

```

#### Option6

The `Option6` alternative relies on the following scheme:

| Name                                         | Type   | Scheme                         | Environment Variable                                                          |
| -------------------------------------------- | ------ | ------------------------------ | ----------------------------------------------------------------------------- |
| `clientID`<br/>`clientSecret`<br/>`tokenURL` | oauth2 | OAuth2 Client Credentials Flow | `SPEAKEASY_CLIENT_ID`<br/>`SPEAKEASY_CLIENT_SECRET`<br/>`SPEAKEASY_TOKEN_URL` |

Example:
```typescript
import { openAsBlob } from "node:fs";
import { SDK } from "openapi";

const sdk = new SDK({
  security: {
    option6: {
      clientID: process.env["SPEAKEASY_CLIENT_ID"] ?? "",
      clientSecret: process.env["SPEAKEASY_CLIENT_SECRET"] ?? "",
      tokenURL: "/clientcredentials/token",
    },
  },
});

async function run() {
  const result = await sdk.postFile({
    upload: await openAsBlob("example.file"),
  });

  console.log(result);
}

run();

```

#### MyApiKey

The `MyApiKey` alternative relies on the following scheme:

| Name       | Type   | Scheme  | Environment Variable   |
| ---------- | ------ | ------- | ---------------------- |
| `myApiKey` | apiKey | API key | `SPEAKEASY_MY_API_KEY` |

Example:
```typescript
import { openAsBlob } from "node:fs";
import { SDK } from "openapi";

const sdk = new SDK({
  security: {
    myApiKey: {
      myApiKey: process.env["SPEAKEASY_MY_API_KEY"] ?? "",
    },
  },
});

async function run() {
  const result = await sdk.postFile({
    upload: await openAsBlob("example.file"),
  });

  console.log(result);
}

run();

```

### Per-Operation Security Schemes

Some operations in this SDK require the security scheme to be specified at the request level. For example:
```typescript
import { SDK } from "openapi";

const sdk = new SDK();

async function run() {
  await sdk.tag1.auth({
    accessToken: process.env["SPEAKEASY_ACCESS_TOKEN"] ?? "",
  });
}

run();

```
<!-- End Authentication [security] -->

<!-- Start Available Resources and Operations [operations] -->
## Available Resources and Operations

<details open>
<summary>Available methods</summary>

### [SDK](docs/sdks/sdk/README.md)

* [operationWithLeadingAndTrailingUnderscores](docs/sdks/sdk/README.md#operationwithleadingandtrailingunderscores)
* [postFile](docs/sdks/sdk/README.md#postfile) - Post File
* [getPolymorphism](docs/sdks/sdk/README.md#getpolymorphism)
* [getUnionErrors](docs/sdks/sdk/README.md#getunionerrors)
* [getRequestBodyFlattenedAway](docs/sdks/sdk/README.md#getrequestbodyflattenedaway)
* [getFullyFlattenedRequest](docs/sdks/sdk/README.md#getfullyflattenedrequest)
* [createWithUnion](docs/sdks/sdk/README.md#createwithunion) - Create with discriminated union request body
* [testEndpoint](docs/sdks/sdk/README.md#testendpoint)
* [createUser](docs/sdks/sdk/README.md#createuser) - Create User
* [getUser](docs/sdks/sdk/README.md#getuser) - Get User
* [updateUser](docs/sdks/sdk/README.md#updateuser) - Update User
* [deleteUser](docs/sdks/sdk/README.md#deleteuser) - Delete User
* [login](docs/sdks/sdk/README.md#login) - Login
* [validate](docs/sdks/sdk/README.md#validate) - Validate
* [chat](docs/sdks/sdk/README.md#chat)
* [getBinaryDefaultResponse](docs/sdks/sdk/README.md#getbinarydefaultresponse)
* [testEnumFormats](docs/sdks/sdk/README.md#testenumformats) - Test x-speakeasy-enums in different formats
* [binaryAndStringUpload](docs/sdks/sdk/README.md#binaryandstringupload)
* [getErrorInUnion](docs/sdks/sdk/README.md#geterrorinunion)
* [getDuplicateExportCollision](docs/sdks/sdk/README.md#getduplicateexportcollision) - Tests that a spec-defined error type colliding with a built-in SDK error name does not cause TS2308
* [getNamedPrimitiveUnion](docs/sdks/sdk/README.md#getnamedprimitiveunion) - Test named primitive union options using title and x-speakeasy-name-override
* [getEmptyObjectError](docs/sdks/sdk/README.md#getemptyobjecterror) - Get Empty Object Error
* [urlValidationStressTest](docs/sdks/sdk/README.md#urlvalidationstresstest)
* [parenthesesInPathAllowed](docs/sdks/sdk/README.md#parenthesesinpathallowed) - A string with {{ double braces }} and { single braces }
and \{\{ escaped curlies \}\} and `backticks`.
and \`escaped backticks\` and double slashes\\
and 'single quotes' and "double quotes".
and  \'escaped single quotes\' and \"escaped double quotes\".

* [getNestedIntegerString](docs/sdks/sdk/README.md#getnestedintegerstring) - Test nested struct with integer:string tag
* [renderAsset](docs/sdks/sdk/README.md#renderasset) - Render Asset
* [getAsset](docs/sdks/sdk/README.md#getasset) - Get Asset
* [getErrorOnlyExample](docs/sdks/sdk/README.md#geterroronlyexample) - Operation with example only on error response

### [Group](docs/sdks/group/README.md)

* [rootGroupOp](docs/sdks/group/README.md#rootgroupop) - An operation at the group's root level

#### [Group.SubGroup](docs/sdks/subgroup/README.md)

* [subGroupOp](docs/sdks/subgroup/README.md#subgroupop) - An operation at the group's top level

##### [Group.SubGroup.Empty.Tail](docs/sdks/tail/README.md)

* [nestedGroupOp](docs/sdks/tail/README.md#nestedgroupop) - An operation at the group's deepest level

### [NamespaceTests.Conflicts](docs/sdks/conflicts/README.md)

* [getNamespaceConflict](docs/sdks/conflicts/README.md#getnamespaceconflict) - Get Namespace Conflict Test
* [putNamespaceConflict](docs/sdks/conflicts/README.md#putnamespaceconflict) - Put Property Name Conflicts Behind
* [createNamespaceConflict](docs/sdks/conflicts/README.md#createnamespaceconflict) - Create Namespace Conflict Test
* [getTripleNamespaceConflict](docs/sdks/conflicts/README.md#gettriplenamespaceconflict) - Get Triple Namespace Conflict Test
* [getPetOwners](docs/sdks/conflicts/README.md#getpetowners) - Get Pet Owners

### [NamespaceTests.SingleBar](docs/sdks/singlebar/README.md)

* [getSingleNamespaceBarPet](docs/sdks/singlebar/README.md#getsinglenamespacebarpet) - Get Single Namespace Bar Pet

### [NamespaceTests.SingleFoo](docs/sdks/singlefoo/README.md)

* [getSingleNamespaceFooPet](docs/sdks/singlefoo/README.md#getsinglenamespacefoopet) - Get Single Namespace Foo Pet
* [createSingleNamespaceFooPet](docs/sdks/singlefoo/README.md#createsinglenamespacefoopet) - Create Single Namespace Foo Pet

### [NamespaceTests.Types](docs/sdks/types/README.md)

* [getNamespaceTypes](docs/sdks/types/README.md#getnamespacetypes) - Get Namespace Types Test
* [getNamespaceAnimal](docs/sdks/types/README.md#getnamespaceanimal) - Get Namespace Animal (Discriminated Union)
* [getNamespaceVehicle](docs/sdks/types/README.md#getnamespacevehicle) - Get Namespace Vehicle (Non-Discriminated Union)
* [getNamespaceOrganization](docs/sdks/types/README.md#getnamespaceorganization) - Get Namespace Organization (Nested Inline Schemas)

### [~~Obsolete~~](docs/sdks/obsolete/README.md)

* [~~deprecated1~~](docs/sdks/obsolete/README.md#deprecated1) - Deprecated Operation :warning: **Deprecated** Use [getRequestBodyFlattenedAway](docs/sdks/sdk/README.md#getrequestbodyflattenedaway) instead.

### [Tag1](docs/sdks/tag1/README.md)

* [~~deprecated1~~](docs/sdks/tag1/README.md#deprecated1) - Deprecated Operation :warning: **Deprecated** Use [getRequestBodyFlattenedAway](docs/sdks/sdk/README.md#getrequestbodyflattenedaway) instead.
* [auth](docs/sdks/tag1/README.md#auth) - This operation aims at testing available OAuth2 scopes collection:
 - only operation with oauth2 authorizationCode security flow
 - belongs to a subSDK

* [listTest1](docs/sdks/tag1/README.md#listtest1) - Get Test1
* [postFileWithEncoding](docs/sdks/tag1/README.md#postfilewithencoding) - Post File With Encoding

### [TestGroup.Tag2](docs/sdks/tag2/README.md)

* [postTest](docs/sdks/tag2/README.md#posttest) - Post Test2

### [TestGroup.Tag3](docs/sdks/tag3/README.md)

* [postTest](docs/sdks/tag3/README.md#posttest) - Post Test2

</details>
<!-- End Available Resources and Operations [operations] -->

<!-- Start Standalone functions [standalone-funcs] -->
## Standalone functions

All the methods listed above are available as standalone functions. These
functions are ideal for use in applications running in the browser, serverless
runtimes or other environments where application bundle size is a primary
concern. When using a bundler to build your application, all unused
functionality will be either excluded from the final bundle or tree-shaken away.

To read more about standalone functions, check [FUNCTIONS.md](./FUNCTIONS.md).

<details>

<summary>Available standalone functions</summary>

- [`binaryAndStringUpload`](docs/sdks/sdk/README.md#binaryandstringupload)
- [`chat`](docs/sdks/sdk/README.md#chat)
- [`createUser`](docs/sdks/sdk/README.md#createuser) - Create User
- [`createWithUnion`](docs/sdks/sdk/README.md#createwithunion) - Create with discriminated union request body
- [`deleteUser`](docs/sdks/sdk/README.md#deleteuser) - Delete User
- [`getAsset`](docs/sdks/sdk/README.md#getasset) - Get Asset
- [`getBinaryDefaultResponse`](docs/sdks/sdk/README.md#getbinarydefaultresponse)
- [`getDuplicateExportCollision`](docs/sdks/sdk/README.md#getduplicateexportcollision) - Tests that a spec-defined error type colliding with a built-in SDK error name does not cause TS2308
- [`getEmptyObjectError`](docs/sdks/sdk/README.md#getemptyobjecterror) - Get Empty Object Error
- [`getErrorInUnion`](docs/sdks/sdk/README.md#geterrorinunion)
- [`getErrorOnlyExample`](docs/sdks/sdk/README.md#geterroronlyexample) - Operation with example only on error response
- [`getFullyFlattenedRequest`](docs/sdks/sdk/README.md#getfullyflattenedrequest)
- [`getNamedPrimitiveUnion`](docs/sdks/sdk/README.md#getnamedprimitiveunion) - Test named primitive union options using title and x-speakeasy-name-override
- [`getNestedIntegerString`](docs/sdks/sdk/README.md#getnestedintegerstring) - Test nested struct with integer:string tag
- [`getPolymorphism`](docs/sdks/sdk/README.md#getpolymorphism)
- [`getRequestBodyFlattenedAway`](docs/sdks/sdk/README.md#getrequestbodyflattenedaway)
- [`getUnionErrors`](docs/sdks/sdk/README.md#getunionerrors)
- [`getUser`](docs/sdks/sdk/README.md#getuser) - Get User
- [`groupRootGroupOp`](docs/sdks/group/README.md#rootgroupop) - An operation at the group's root level
- [`groupSubGroupEmptyTailNestedGroupOp`](docs/sdks/tail/README.md#nestedgroupop) - An operation at the group's deepest level
- [`groupSubGroupSubGroupOp`](docs/sdks/subgroup/README.md#subgroupop) - An operation at the group's top level
- [`login`](docs/sdks/sdk/README.md#login) - Login
- [`namespaceTestsConflictsCreateNamespaceConflict`](docs/sdks/conflicts/README.md#createnamespaceconflict) - Create Namespace Conflict Test
- [`namespaceTestsConflictsGetNamespaceConflict`](docs/sdks/conflicts/README.md#getnamespaceconflict) - Get Namespace Conflict Test
- [`namespaceTestsConflictsGetPetOwners`](docs/sdks/conflicts/README.md#getpetowners) - Get Pet Owners
- [`namespaceTestsConflictsGetTripleNamespaceConflict`](docs/sdks/conflicts/README.md#gettriplenamespaceconflict) - Get Triple Namespace Conflict Test
- [`namespaceTestsConflictsPutNamespaceConflict`](docs/sdks/conflicts/README.md#putnamespaceconflict) - Put Property Name Conflicts Behind
- [`namespaceTestsSingleBarGetSingleNamespaceBarPet`](docs/sdks/singlebar/README.md#getsinglenamespacebarpet) - Get Single Namespace Bar Pet
- [`namespaceTestsSingleFooCreateSingleNamespaceFooPet`](docs/sdks/singlefoo/README.md#createsinglenamespacefoopet) - Create Single Namespace Foo Pet
- [`namespaceTestsSingleFooGetSingleNamespaceFooPet`](docs/sdks/singlefoo/README.md#getsinglenamespacefoopet) - Get Single Namespace Foo Pet
- [`namespaceTestsTypesGetNamespaceAnimal`](docs/sdks/types/README.md#getnamespaceanimal) - Get Namespace Animal (Discriminated Union)
- [`namespaceTestsTypesGetNamespaceOrganization`](docs/sdks/types/README.md#getnamespaceorganization) - Get Namespace Organization (Nested Inline Schemas)
- [`namespaceTestsTypesGetNamespaceTypes`](docs/sdks/types/README.md#getnamespacetypes) - Get Namespace Types Test
- [`namespaceTestsTypesGetNamespaceVehicle`](docs/sdks/types/README.md#getnamespacevehicle) - Get Namespace Vehicle (Non-Discriminated Union)
- [`operationWithLeadingAndTrailingUnderscores`](docs/sdks/sdk/README.md#operationwithleadingandtrailingunderscores)
- [`parenthesesInPathAllowed`](docs/sdks/sdk/README.md#parenthesesinpathallowed) - A string with {{ double braces }} and { single braces }
and \{\{ escaped curlies \}\} and `backticks`.
and \`escaped backticks\` and double slashes\\
and 'single quotes' and "double quotes".
and  \'escaped single quotes\' and \"escaped double quotes\".

- [`postFile`](docs/sdks/sdk/README.md#postfile) - Post File
- [`renderAsset`](docs/sdks/sdk/README.md#renderasset) - Render Asset
- [`tag1Auth`](docs/sdks/tag1/README.md#auth) - This operation aims at testing available OAuth2 scopes collection:
 - only operation with oauth2 authorizationCode security flow
 - belongs to a subSDK

- [`tag1ListTest1`](docs/sdks/tag1/README.md#listtest1) - Get Test1
- [`tag1PostFileWithEncoding`](docs/sdks/tag1/README.md#postfilewithencoding) - Post File With Encoding
- [`testEndpoint`](docs/sdks/sdk/README.md#testendpoint)
- [`testEnumFormats`](docs/sdks/sdk/README.md#testenumformats) - Test x-speakeasy-enums in different formats
- [`testGroupTag2PostTest`](docs/sdks/tag2/README.md#posttest) - Post Test2
- [`testGroupTag2PostTest`](docs/sdks/tag3/README.md#posttest) - Post Test2
- [`updateUser`](docs/sdks/sdk/README.md#updateuser) - Update User
- [`urlValidationStressTest`](docs/sdks/sdk/README.md#urlvalidationstresstest)
- [`validate`](docs/sdks/sdk/README.md#validate) - Validate
- ~~[`tag1Deprecated1`](docs/sdks/obsolete/README.md#deprecated1)~~ - Deprecated Operation :warning: **Deprecated** Use [`getRequestBodyFlattenedAway`](docs/sdks/sdk/README.md#getrequestbodyflattenedaway) instead.
- ~~[`tag1Deprecated1`](docs/sdks/tag1/README.md#deprecated1)~~ - Deprecated Operation :warning: **Deprecated** Use [`getRequestBodyFlattenedAway`](docs/sdks/sdk/README.md#getrequestbodyflattenedaway) instead.

</details>
<!-- End Standalone functions [standalone-funcs] -->

<!-- Start React hooks with TanStack Query [react-query] -->
## React hooks with TanStack Query

React hooks built on [TanStack Query][tanstack-query] are included in this SDK.
These hooks and the utility functions provided alongside them can be used to
build rich applications that pull data from the API using one of the most
popular asynchronous state management library.

[tanstack-query]: https://tanstack.com/query/v5/docs/framework/react/overview

To learn about this feature and how to get started, check
[REACT_QUERY.md](./REACT_QUERY.md).

> [!WARNING]
>
> This feature is currently in **preview** and is subject to breaking changes
> within the current major version of the SDK as we gather user feedback on it.

<details>

<summary>Available React hooks</summary>

- [`useBinaryAndStringUploadMutation`](docs/sdks/sdk/README.md#binaryandstringupload)
- [`useChatMutation`](docs/sdks/sdk/README.md#chat)
- [`useCreateUserMutation`](docs/sdks/sdk/README.md#createuser) - Create User
- [`useCreateWithUnionMutation`](docs/sdks/sdk/README.md#createwithunion) - Create with discriminated union request body
- [`useDeleteUserMutation`](docs/sdks/sdk/README.md#deleteuser) - Delete User
- [`useFlatRequest`](docs/sdks/sdk/README.md#getfullyflattenedrequest)
- [`useGetAsset`](docs/sdks/sdk/README.md#getasset) - Get Asset
- [`useGetBinaryDefaultResponse`](docs/sdks/sdk/README.md#getbinarydefaultresponse)
- [`useGetDuplicateExportCollision`](docs/sdks/sdk/README.md#getduplicateexportcollision) - Tests that a spec-defined error type colliding with a built-in SDK error name does not cause TS2308
- [`useGetEmptyObjectError`](docs/sdks/sdk/README.md#getemptyobjecterror) - Get Empty Object Error
- [`useGetErrorInUnion`](docs/sdks/sdk/README.md#geterrorinunion)
- [`useGetErrorOnlyExample`](docs/sdks/sdk/README.md#geterroronlyexample) - Operation with example only on error response
- [`useGetNamedPrimitiveUnion`](docs/sdks/sdk/README.md#getnamedprimitiveunion) - Test named primitive union options using title and x-speakeasy-name-override
- [`useGetNestedIntegerString`](docs/sdks/sdk/README.md#getnestedintegerstring) - Test nested struct with integer:string tag
- [`useGetPolymorphism`](docs/sdks/sdk/README.md#getpolymorphism)
- [`useGetRequestBodyFlattenedAway`](docs/sdks/sdk/README.md#getrequestbodyflattenedaway)
- [`useGetUser`](docs/sdks/sdk/README.md#getuser) - Get User
- [`useGroupRootGroupOp`](docs/sdks/group/README.md#rootgroupop) - An operation at the group's root level
- [`useGroupSubGroupEmptyTailNestedGroupOp`](docs/sdks/tail/README.md#nestedgroupop) - An operation at the group's deepest level
- [`useGroupSubGroupSubGroupOp`](docs/sdks/subgroup/README.md#subgroupop) - An operation at the group's top level
- [`useLogin`](docs/sdks/sdk/README.md#login) - Login
- [`useNamespaceTestsConflictsCreateNamespaceConflictMutation`](docs/sdks/conflicts/README.md#createnamespaceconflict) - Create Namespace Conflict Test
- [`useNamespaceTestsConflictsGetNamespaceConflict`](docs/sdks/conflicts/README.md#getnamespaceconflict) - Get Namespace Conflict Test
- [`useNamespaceTestsConflictsGetPetOwners`](docs/sdks/conflicts/README.md#getpetowners) - Get Pet Owners
- [`useNamespaceTestsConflictsGetTripleNamespaceConflict`](docs/sdks/conflicts/README.md#gettriplenamespaceconflict) - Get Triple Namespace Conflict Test
- [`useNamespaceTestsConflictsPutNamespaceConflictMutation`](docs/sdks/conflicts/README.md#putnamespaceconflict) - Put Property Name Conflicts Behind
- [`useNamespaceTestsSingleBarGetSingleNamespaceBarPet`](docs/sdks/singlebar/README.md#getsinglenamespacebarpet) - Get Single Namespace Bar Pet
- [`useNamespaceTestsSingleFooCreateSingleNamespaceFooPetMutation`](docs/sdks/singlefoo/README.md#createsinglenamespacefoopet) - Create Single Namespace Foo Pet
- [`useNamespaceTestsSingleFooGetSingleNamespaceFooPet`](docs/sdks/singlefoo/README.md#getsinglenamespacefoopet) - Get Single Namespace Foo Pet
- [`useNamespaceTestsTypesGetNamespaceAnimal`](docs/sdks/types/README.md#getnamespaceanimal) - Get Namespace Animal (Discriminated Union)
- [`useNamespaceTestsTypesGetNamespaceOrganization`](docs/sdks/types/README.md#getnamespaceorganization) - Get Namespace Organization (Nested Inline Schemas)
- [`useNamespaceTestsTypesGetNamespaceTypes`](docs/sdks/types/README.md#getnamespacetypes) - Get Namespace Types Test
- [`useNamespaceTestsTypesGetNamespaceVehicle`](docs/sdks/types/README.md#getnamespacevehicle) - Get Namespace Vehicle (Non-Discriminated Union)
- [`useOperationWithLeadingAndTrailingUnderscores`](docs/sdks/sdk/README.md#operationwithleadingandtrailingunderscores)
- [`useParenthesesInPathAllowedMutation`](docs/sdks/sdk/README.md#parenthesesinpathallowed) - A string with {{ double braces }} and { single braces }
and \{\{ escaped curlies \}\} and `backticks`.
and \`escaped backticks\` and double slashes\\
and 'single quotes' and "double quotes".
and  \'escaped single quotes\' and \"escaped double quotes\".

- [`usePostFileMutation`](docs/sdks/sdk/README.md#postfile) - Post File
- [`useRenderAssetMutation`](docs/sdks/sdk/README.md#renderasset) - Render Asset
- [`useTag1Auth`](docs/sdks/tag1/README.md#auth) - This operation aims at testing available OAuth2 scopes collection:
 - only operation with oauth2 authorizationCode security flow
 - belongs to a subSDK

- [`useTag1ListTest1`](docs/sdks/tag1/README.md#listtest1) - Get Test1
- [`useTag1PostFileWithEncodingMutation`](docs/sdks/tag1/README.md#postfilewithencoding) - Post File With Encoding
- [`useTestEndpointMutation`](docs/sdks/sdk/README.md#testendpoint)
- [`useTestEnumFormatsMutation`](docs/sdks/sdk/README.md#testenumformats) - Test x-speakeasy-enums in different formats
- [`useTestGroupTag2PostTestMutation`](docs/sdks/tag2/README.md#posttest) - Post Test2
- [`useTestGroupTag2PostTestMutation`](docs/sdks/tag3/README.md#posttest) - Post Test2
- [`useUpdateUserMutation`](docs/sdks/sdk/README.md#updateuser) - Update User
- [`useUrlValidationStressTest`](docs/sdks/sdk/README.md#urlvalidationstresstest)
- [`useValidate`](docs/sdks/sdk/README.md#validate) - Validate
- ~~[`useTag1Deprecated1`](docs/sdks/obsolete/README.md#deprecated1)~~ - Deprecated Operation :warning: **Deprecated** Use [`useGetRequestBodyFlattenedAway`](docs/sdks/sdk/README.md#getrequestbodyflattenedaway) instead.
- ~~[`useTag1Deprecated1`](docs/sdks/tag1/README.md#deprecated1)~~ - Deprecated Operation :warning: **Deprecated** Use [`useGetRequestBodyFlattenedAway`](docs/sdks/sdk/README.md#getrequestbodyflattenedaway) instead.

</details>
<!-- End React hooks with TanStack Query [react-query] -->

<!-- Start Global Parameters [global-parameters] -->
## Global Parameters

Certain parameters are configured globally. These parameters may be set on the SDK client instance itself during initialization. When configured as an option during SDK initialization, These global values will be used as defaults on the operations that use them. When such operations are called, there is a place in each to override the global value, if needed.

For example, you can set `queryParam1` to `"some example query param"` at SDK initialization and then you do not have to pass the same value on calls to operations like `getRequestBodyFlattenedAway`. But if you want to do so you may, which will locally override the global setting. See the example code below for a demonstration.


### Available Globals

The following global parameters are available.
Global parameters can also be set via environment variable.

| Name                  | Type   | Description                                                                        | Environment                       |
| --------------------- | ------ | ---------------------------------------------------------------------------------- | --------------------------------- |
| queryParam1           | string | A long winded, multi-line description<br/>for the query parameter number one.<br/> | SPEAKEASY_QUERY_PARAM1            |
| deprecatedQueryParam1 | string | A deprecated description                                                           | SPEAKEASY_DEPRECATED_QUERY_PARAM1 |
| deprecatedQueryParam2 | string | The deprecatedQueryParam2 parameter.                                               | SPEAKEASY_DEPRECATED_QUERY_PARAM2 |
| loneQueryParam        | string | The loneQueryParam parameter.                                                      | SPEAKEASY_LONE_QUERY_PARAM        |

### Example

```typescript
import { SDK } from "openapi";

const sdk = new SDK({
  loneQueryParam: "<value>",
  queryParam1: "some example query param",
  deprecatedQueryParam1: "some example query param",
  deprecatedQueryParam2: "some example query param",
});

async function run() {
  await sdk.getRequestBodyFlattenedAway();
}

run();

```
<!-- End Global Parameters [global-parameters] -->

<!-- Start Server-sent event streaming [eventstream] -->
## Server-sent event streaming

[Server-sent events][mdn-sse] are used to stream content from certain
operations. These operations will expose the stream as an async iterable that
can be consumed using a [`for await...of`][mdn-for-await-of] loop. The loop will
terminate when the server no longer has any events to send and closes the
underlying connection.

```typescript
import { SDK } from "openapi";

const sdk = new SDK();

async function run() {
  const result = await sdk.chat({
    model: "review-model",
    prompt: "What is the largest city in the world?",
    stream: false,
  });

  console.log(result);
}

run();

```

[mdn-sse]: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events
[mdn-for-await-of]: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Statements/for-await...of
<!-- End Server-sent event streaming [eventstream] -->

<!-- Start Pagination [pagination] -->
## Pagination

Some of the endpoints in this SDK support pagination. To use pagination, you
make your SDK calls as usual, but the returned response object will also be an
async iterable that can be consumed using the [`for await...of`][for-await-of]
syntax.

[for-await-of]: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Statements/for-await...of

Here's an example of one such pagination call:

```typescript
import { QueryParam2, SDK } from "openapi";

const sdk = new SDK({
  queryParam1: "some example query param",
  security: {
    myApiKey: {
      myApiKey: process.env["SPEAKEASY_MY_API_KEY"] ?? "",
    },
  },
});

async function run() {
  const result = await sdk.tag1.listTest1(
    100,
    QueryParam2.One,
    "some example header param",
  );

  for await (const page of result) {
    console.log(page);
  }
}

run();

```
<!-- End Pagination [pagination] -->

<!-- Start File uploads [file-upload] -->
## File uploads

Certain SDK methods accept files as part of a multi-part request. It is possible and typically recommended to upload files as a stream rather than reading the entire contents into memory. This avoids excessive memory consumption and potentially crashing with out-of-memory errors when working with very large files. The following example demonstrates how to attach a file stream to a request.

> [!TIP]
>
> Depending on your JavaScript runtime, there are convenient utilities that return a handle to a file without reading the entire contents into memory:
>
> - **Node.js v20+:** Since v20, Node.js comes with a native `openAsBlob` function in [`node:fs`](https://nodejs.org/docs/latest-v20.x/api/fs.html#fsopenasblobpath-options).
> - **Bun:** The native [`Bun.file`](https://bun.sh/docs/api/file-io#reading-files-bun-file) function produces a file handle that can be used for streaming file uploads.
> - **Browsers:** All supported browsers return an instance to a [`File`](https://developer.mozilla.org/en-US/docs/Web/API/File) when reading the value from an `<input type="file">` element.
> - **Node.js v18:** A file stream can be created using the `fileFrom` helper from [`fetch-blob/from.js`](https://www.npmjs.com/package/fetch-blob).

```typescript
import { openAsBlob } from "node:fs";
import { SDK } from "openapi";

const sdk = new SDK();

async function run() {
  const result = await sdk.tag1.postFileWithEncoding({
    file: await openAsBlob("example.file"),
  });

  console.log(result);
}

run();

```
<!-- End File uploads [file-upload] -->

<!-- Start Retries [retries] -->
## Retries

Some of the endpoints in this SDK support retries.  If you use the SDK without any configuration, it will fall back to the default retry strategy provided by the API.  However, the default retry strategy can be overridden on a per-operation basis, or across the entire SDK.

To change the default retry strategy for a single API call, simply provide a retryConfig object to the call:
```typescript
import { openAsBlob } from "node:fs";
import { SDK } from "openapi";

const sdk = new SDK();

async function run() {
  const result = await sdk.postFile({
    upload: await openAsBlob("example.file"),
  }, {
    retries: {
      strategy: "backoff",
      backoff: {
        initialInterval: 1,
        maxInterval: 50,
        exponent: 1.1,
        maxElapsedTime: 100,
      },
      retryConnectionErrors: false,
    },
  });

  console.log(result);
}

run();

```

If you'd like to override the default retry strategy for all operations that support retries, you can provide a retryConfig at SDK initialization:
```typescript
import { openAsBlob } from "node:fs";
import { SDK } from "openapi";

const sdk = new SDK({
  retryConfig: {
    strategy: "backoff",
    backoff: {
      initialInterval: 1,
      maxInterval: 50,
      exponent: 1.1,
      maxElapsedTime: 100,
    },
    retryConnectionErrors: false,
  },
});

async function run() {
  const result = await sdk.postFile({
    upload: await openAsBlob("example.file"),
  });

  console.log(result);
}

run();

```
<!-- End Retries [retries] -->

<!-- Start Error Handling [errors] -->
## Error Handling

[`SDKBaseError`](./src/models/sdkbaseerror.ts) is the base class for all HTTP error responses. It has the following properties:

| Property            | Type       | Description                                                                             |
| ------------------- | ---------- | --------------------------------------------------------------------------------------- |
| `error.message`     | `string`   | Error message                                                                           |
| `error.statusCode`  | `number`   | HTTP response status code eg `404`                                                      |
| `error.headers`     | `Headers`  | HTTP response headers                                                                   |
| `error.body`        | `string`   | HTTP body. Can be empty string if no body is returned.                                  |
| `error.rawResponse` | `Response` | Raw HTTP response                                                                       |
| `error.data$`       |            | Optional. Some errors may contain structured data. [See Error Classes](#error-classes). |

### Example
```typescript
import * as models from "openapi";
import { SDK } from "openapi";

const sdk = new SDK();

async function run() {
  try {
    const result = await sdk.getUnionErrors(12);

    for await (const page of result) {
      console.log(page);
    }
  } catch (error) {
    // The base class for HTTP error responses
    if (error instanceof models.SDKBaseError) {
      console.log(error.message);
      console.log(error.statusCode);
      console.log(error.body);
      console.log(error.headers);

      // Depending on the method different errors may be thrown
      if (error instanceof models.ErrorsError) {
        console.log(error.data$.error); // string
        console.log(error.data$.code); // number
      }
    }
  }
}

run();

```

### Error Classes
**Primary error:**
* [`SDKBaseError`](./src/models/sdkbaseerror.ts): The base class for HTTP error responses.

<details><summary>Less common errors (15)</summary>

<br />

**Network errors:**
* [`ConnectionError`](./src/models/httpclienterrors.ts): HTTP client was unable to make a request to a server.
* [`RequestTimeoutError`](./src/models/httpclienterrors.ts): HTTP request timed out due to an AbortSignal signal.
* [`RequestAbortedError`](./src/models/httpclienterrors.ts): HTTP request was aborted by the client.
* [`InvalidRequestError`](./src/models/httpclienterrors.ts): Any input used to create a request is invalid.
* [`UnexpectedClientError`](./src/models/httpclienterrors.ts): Unrecognised or unexpected error.


**Inherit from [`SDKBaseError`](./src/models/sdkbaseerror.ts)**:
* [`ErrorsError`](./src/models/errorserror.ts): A not-so-long multi-line error model description. Applicable to 7 of 50 methods.*
* [`BadRequestResponseError`](./src/models/badrequestresponseerror.ts): Bad Request. Status code `400`. Applicable to 2 of 50 methods.*
* [`TaggedError1`](./src/models/taggederror1.ts): Applicable to 2 of 50 methods.*
* [`RequestTimeoutError`](./src/models/httpclienterrors.ts): A spec-defined error that collides with the built-in RequestTimeoutError in httpclienterrors.ts. Status code `408`. Applicable to 1 of 50 methods.*
* [`TaggedError2`](./src/models/taggederror2.ts): Something went wrong. Status code `4XX`. Applicable to 1 of 50 methods.*
* [`ErrorType1`](./src/models/errortype1.ts): An error of type one. Status code `500`. Applicable to 1 of 50 methods.*
* [`ErrorType2`](./src/models/errortype2.ts): Internal Server Error. Status code `500`. Applicable to 1 of 50 methods.*
* [`FailedResponseError`](./src/models/failedresponseerror.ts): An error response with an empty object schema. Status code `500`. Applicable to 1 of 50 methods.*
* [`Test2ResponseError`](./src/models/test2responseerror.ts): Internal Server Error. Status code `500`. Applicable to 1 of 50 methods.*
* [`ResponseValidationError`](./src/models/responsevalidationerror.ts): Type mismatch between the data returned from the server and the structure expected by the SDK. See `error.rawValue` for the raw value and `error.pretty()` for a nicely formatted multi-line string.

</details>

\* Check [the method documentation](#available-resources-and-operations) to see if the error is applicable.
<!-- End Error Handling [errors] -->

<!-- Start Server Selection [server] -->
## Server Selection

### Select Server by Index

You can override the default server globally by passing a server index to the `serverIdx: number` optional parameter when initializing the SDK client instance. The selected server will then be used as the default on the operations that use it. This table lists the indexes associated with the available servers:

| #   | Server                                     | Variables                 | Description                     |
| --- | ------------------------------------------ | ------------------------- | ------------------------------- |
| 0   | `http://localhost:35123`                   |                           | The default server.             |
| 1   | `http://{subdomain}.domain.com/v{version}` | `subdomain`<br/>`version` |                                 |
| 2   | `http://{HostName}:{PORT}`                 | `HostName`<br/>`PORT`     | A server with an enum variable. |

If the selected server has variables, you may override its default values through the additional parameters made available in the SDK constructor:

| Variable    | Parameter                 | Supported Values                      | Default       | Description                              |
| ----------- | ------------------------- | ------------------------------------- | ------------- | ---------------------------------------- |
| `subdomain` | `subdomain: string`       | string                                | `"api"`       |                                          |
| `version`   | `version: string`         | string                                | `"1"`         |                                          |
| `HostName`  | `hostName: string`        | string                                | `"localhost"` | The hostname of the server.              |
| `PORT`      | `port: models.ServerPORT` | - `"80"`<br/>- `"8080"`<br/>- `"443"` | `"8080"`      | The port on which the server is running. |

#### Example

```typescript
import { openAsBlob } from "node:fs";
import { SDK } from "openapi";

const sdk = new SDK({
  serverIdx: 2,
  hostName: "localhost",
  port: "443",
});

async function run() {
  const result = await sdk.postFile({
    upload: await openAsBlob("example.file"),
  });

  console.log(result);
}

run();

```

### Override Server URL Per-Client

The default server can also be overridden globally by passing a URL to the `serverURL: string` optional parameter when initializing the SDK client instance. For example:
```typescript
import { openAsBlob } from "node:fs";
import { SDK } from "openapi";

const sdk = new SDK({
  serverURL: "http://localhost:8080",
});

async function run() {
  const result = await sdk.postFile({
    upload: await openAsBlob("example.file"),
  });

  console.log(result);
}

run();

```

### Override Server URL Per-Operation

The server URL can also be overridden on a per-operation basis, provided a server list was specified for the operation. For example:
```typescript
import { QueryParam2, SDK } from "openapi";

const sdk = new SDK({
  queryParam1: "some example query param",
  security: {
    myApiKey: {
      myApiKey: process.env["SPEAKEASY_MY_API_KEY"] ?? "",
    },
  },
});

async function run() {
  const result = await sdk.tag1.listTest1(
    100,
    QueryParam2.One,
    "some example header param",
    undefined,
    {
      serverURL: "http://localhost:35123",
    },
  );

  for await (const page of result) {
    console.log(page);
  }
}

run();

```
<!-- End Server Selection [server] -->

<!-- Start Custom HTTP Client [http-client] -->
## Custom HTTP Client

The TypeScript SDK makes API calls using an `HTTPClient` that wraps the native
[Fetch API](https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API). This
client is a thin wrapper around `fetch` and provides the ability to attach hooks
around the request lifecycle that can be used to modify the request or handle
errors and response.

The `HTTPClient` constructor takes an optional `fetcher` argument that can be
used to integrate a third-party HTTP client or when writing tests to mock out
the HTTP client and feed in fixtures.

The following example shows how to:
- route requests through a proxy server using [undici](https://www.npmjs.com/package/undici)'s ProxyAgent
- use the `"beforeRequest"` hook to add a custom header and a timeout to requests
- use the `"requestError"` hook to log errors

```typescript
import { SDK } from "openapi";
import { ProxyAgent } from "undici";
import { HTTPClient } from "openapi/lib/http";

const dispatcher = new ProxyAgent("http://proxy.example.com:8080");

const httpClient = new HTTPClient({
  // 'fetcher' takes a function that has the same signature as native 'fetch'.
  fetcher: (input, init) =>
    // 'dispatcher' is specific to undici and not part of the standard Fetch API.
    fetch(input, { ...init, dispatcher } as RequestInit),
});

httpClient.addHook("beforeRequest", (request) => {
  const nextRequest = new Request(request, {
    signal: request.signal || AbortSignal.timeout(5000)
  });

  nextRequest.headers.set("x-custom-header", "custom value");

  return nextRequest;
});

httpClient.addHook("requestError", (error, request) => {
  console.group("Request Error");
  console.log("Reason:", `${error}`);
  console.log("Endpoint:", `${request.method} ${request.url}`);
  console.groupEnd();
});

const sdk = new SDK({ httpClient: httpClient });
```
<!-- End Custom HTTP Client [http-client] -->

<!-- Start Debugging [debug] -->
## Debugging

You can setup your SDK to emit debug logs for SDK requests and responses.

You can pass a logger that matches `console`'s interface as an SDK option.

> [!WARNING]
> Beware that debug logging will reveal secrets, like API tokens in headers, in log messages printed to a console or files. It's recommended to use this feature only during local development and not in production.

```typescript
import { SDK } from "openapi";

const sdk = new SDK({ debugLogger: console });
```

You can also enable a default debug logger by setting an environment variable `SPEAKEASY_DEBUG` to true.
<!-- End Debugging [debug] -->

<!-- Placeholder for Future Speakeasy SDK Sections -->

# Development

## Maturity

This SDK is in beta, and there may be breaking changes between versions without a major version update. Therefore, we recommend pinning usage
to a specific package version. This way, you can install the same version each time without breaking changes unless you are intentionally
looking for the latest version.

## Contributions

While we value open-source contributions to this SDK, this library is generated programmatically. Any manual changes added to internal files will be overwritten on the next generation. 
We look forward to hearing your feedback. Feel free to open a PR or an issue with a proof of concept and we'll do our best to include it in a future release. 

### SDK Created by [Speakeasy](https://www.speakeasy.com/?utm_source=openapi&utm_campaign=typescript)
