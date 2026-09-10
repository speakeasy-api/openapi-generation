# Speakeasy.OpenAPI

Developer-friendly & type-safe Csharp SDK specifically catered to leverage *Speakeasy.OpenAPI* API.

[![Built by Speakeasy](https://img.shields.io/badge/Built_by-SPEAKEASY-374151?style=for-the-badge&labelColor=f3f4f6)](https://www.speakeasy.com/?utm_source=speakeasy-open-api&utm_campaign=csharp)
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
* [Speakeasy.OpenAPI](#speakeasyopenapi)
  * [SDK Installation](#sdk-installation)
  * [SDK Example Usage](#sdk-example-usage)
  * [Authentication](#authentication)
  * [Available Resources and Operations](#available-resources-and-operations)
  * [Global Parameters](#global-parameters)
  * [Server-sent event streaming](#server-sent-event-streaming)
  * [Pagination](#pagination)
  * [Retries](#retries)
  * [Error Handling](#error-handling)
  * [Server Selection](#server-selection)
  * [Custom HTTP Client](#custom-http-client)
* [Development](#development)
  * [Maturity](#maturity)
  * [Contributions](#contributions)

<!-- End Table of Contents [toc] -->

<!-- Start SDK Installation [installation] -->
## SDK Installation

To add a reference to a local instance of the SDK in a .NET project:
```bash
dotnet add reference src/Speakeasy/OpenAPI/Speakeasy.OpenAPI.csproj
```
<!-- End SDK Installation [installation] -->

<!-- Start SDK Example Usage [usage] -->
## SDK Example Usage

### Example 1

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

PostFileRequest req = new PostFileRequest() {
    Upload = new Speakeasy.OpenAPI.File() {
        FileName = "example.file",
        Content = System.IO.File.ReadAllBytes("example.file"),
    },
};

var res = await sdk.PostFileAsync(req);

// handle response
```

### Example 2

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

PostFileWithEncodingRequest req = new PostFileWithEncodingRequest() {
    File = new PostFileWithEncodingFile() {
        FileName = "example.file",
        Content = System.IO.File.ReadAllBytes("example.file"),
    },
};

var res = await sdk.Tag1.PostFileWithEncodingAsync(req);

// handle response
```

### Example 3

```csharp
using NodaTime;
using Speakeasy.OpenAPI;
using System;
using System.Collections.Generic;
using System.Numerics;

var sdk = new SDK(
    deprecatedQueryParam1: "some example query param",
    deprecatedQueryParam2: "some example query param",
    security: new Security() {
        MyApiKey = new MyApiKey() {
            MyApiKeyValue = "<YOUR_API_KEY_HERE>",
        },
    }
);

var res = await sdk.TestGroup.Tag2.PostTestAsync(test2Request: new Test2Request() {
    Obj = new ExhaustiveObject() {
        Str = "example",
        Bool = true,
        Integer = 999999,
        Int32 = 1,
        Num = 1.1D,
        Float32 = 8499.3F,
        Date = LocalDate.FromDateTime(System.DateTime.Parse("2020-01-01")),
        DateTime = System.DateTime.Parse("2020-01-01T00:00:00Z").ToUniversalTime(),
        Anything = "<value>",
        BoolOpt = true,
        IntOptNull = 999999,
        NumOptNull = 1.1D,
        IntEnum = IntEnum.Third,
        Int32Enum = Int32Enum.SixtyNine,
        Bigint = BigInteger.Parse("702830"),
        DecimalStr = 3858.6M,
        Obj = new SimpleObject() {
            Str = "example",
        },
        Map = new Dictionary<string, SimpleObject>() {
            { "key", new SimpleObject() {
                 Str = "example",
             } },
        },
        Arr = new List<SimpleObject>() {
            new SimpleObject() {
                Str = "example",
            },
            new SimpleObject() {
                Str = "example",
            },
        },
        Any = Any.CreateSimpleObject(new SimpleObject() {
            Str = "example",
        }),
        NullableIntEnum = NullableIntEnum.Third,
        NullableStringEnum = NullableStringEnum.Second,
        Color = Color.Green,
        Icon = Icon.Tick,
        HeroWidth = HeroWidth.FourHundredAndEighty,
    },
    Type = Speakeasy.OpenAPI.Type.SuperType1,
});

// handle response
```

### A custom readme heading

A custom usage description

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(
    queryParam1: "some example query param",
    security: new Security() {
        MyApiKey = new MyApiKey() {
            MyApiKeyValue = "<YOUR_API_KEY_HERE>",
        },
    }
);

ListTest1Response? res = await sdk.Tag1.ListTest1Async(
    page: 100,
    queryParam2: QueryParam2.One,
    headerParam1: "some example header param"
);

while (res != null)
{
    // handle items

    res = await res.Next!();
}
```
<!-- End SDK Example Usage [usage] -->

<!-- Start Authentication [security] -->
## Authentication

### Per-Client Security Schemes

This SDK supports multiple security scheme combinations globally. You can choose from one of the alternatives through the `security` optional parameter when initializing the SDK client instance. The selected option will be used by default to authenticate with the API for all operations that support it.

#### UserPassAuth

The `UserPassAuth` alternative relies on the following scheme:

| Name                      | Type | Scheme     |
| ------------------------- | ---- | ---------- |
| `Username`<br/>`Password` | http | HTTP Basic |

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(security: new Security() {
    UserPassAuth = new UserPassAuth() {
        Username = "<USERNAME>",
        Password = "<PASSWORD>",
    },
});

PostFileRequest req = new PostFileRequest() {
    Upload = new Speakeasy.OpenAPI.File() {
        FileName = "example.file",
        Content = System.IO.File.ReadAllBytes("example.file"),
    },
};

var res = await sdk.PostFileAsync(req);

// handle response
```

#### Option2

All of the following schemes must be satisfied to use the `Option2` alternative:

| Name         | Type   | Scheme      |
| ------------ | ------ | ----------- |
| `BearerAuth` | http   | HTTP Bearer |
| `MyApiKey`   | apiKey | API key     |

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(security: new Security() {
    Option2 = new SecurityOption2() {
        BearerAuth = "<YOUR_JWT>",
        MyApiKey = "<YOUR_API_KEY_HERE>",
    },
});

PostFileRequest req = new PostFileRequest() {
    Upload = new Speakeasy.OpenAPI.File() {
        FileName = "example.file",
        Content = System.IO.File.ReadAllBytes("example.file"),
    },
};

var res = await sdk.PostFileAsync(req);

// handle response
```

#### Option3

The `Option3` alternative relies on the following scheme:

| Name     | Type   | Scheme       |
| -------- | ------ | ------------ |
| `Oauth2` | oauth2 | OAuth2 token |

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(security: new Security() {
    Option3 = new SecurityOption3() {
        Oauth2 = "Bearer <YOUR_OAUTH2_TOKEN>",
    },
});

PostFileRequest req = new PostFileRequest() {
    Upload = new Speakeasy.OpenAPI.File() {
        FileName = "example.file",
        Content = System.IO.File.ReadAllBytes("example.file"),
    },
};

var res = await sdk.PostFileAsync(req);

// handle response
```

#### Option4

The `Option4` alternative relies on the following scheme:

| Name                 | Type | Scheme      |
| -------------------- | ---- | ----------- |
| `AppId`<br/>`Secret` | http | Custom HTTP |

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(security: new Security() {
    Option4 = new SecurityOption4() {
        AppId = "app-speakeasy-123",
        Secret = "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI",
    },
});

PostFileRequest req = new PostFileRequest() {
    Upload = new Speakeasy.OpenAPI.File() {
        FileName = "example.file",
        Content = System.IO.File.ReadAllBytes("example.file"),
    },
};

var res = await sdk.PostFileAsync(req);

// handle response
```

#### Option5

The `Option5` alternative relies on the following scheme:

| Name         | Type   | Scheme       |
| ------------ | ------ | ------------ |
| `MobileAuth` | oauth2 | OAuth2 token |

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(security: new Security() {
    Option5 = new SecurityOption5() {
        MobileAuth = "Bearer <YOUR_OAUTH2_TOKEN>",
    },
});

PostFileRequest req = new PostFileRequest() {
    Upload = new Speakeasy.OpenAPI.File() {
        FileName = "example.file",
        Content = System.IO.File.ReadAllBytes("example.file"),
    },
};

var res = await sdk.PostFileAsync(req);

// handle response
```

#### Option6

The `Option6` alternative relies on the following scheme:

| Name                                         | Type   | Scheme                         |
| -------------------------------------------- | ------ | ------------------------------ |
| `ClientID`<br/>`ClientSecret`<br/>`TokenURL` | oauth2 | OAuth2 Client Credentials Flow |

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(security: new Security() {
    Option6 = new SecurityOption6() {
        ClientID = "<YOUR_CLIENT_ID_HERE>",
        ClientSecret = "<YOUR_CLIENT_SECRET_HERE>",
        TokenURL = "/clientcredentials/token",
    },
});

PostFileRequest req = new PostFileRequest() {
    Upload = new Speakeasy.OpenAPI.File() {
        FileName = "example.file",
        Content = System.IO.File.ReadAllBytes("example.file"),
    },
};

var res = await sdk.PostFileAsync(req);

// handle response
```

#### MyApiKey

The `MyApiKey` alternative relies on the following scheme:

| Name       | Type   | Scheme  |
| ---------- | ------ | ------- |
| `MyApiKey` | apiKey | API key |

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(security: new Security() {
    MyApiKey = new MyApiKey() {
        MyApiKeyValue = "<YOUR_API_KEY_HERE>",
    },
});

PostFileRequest req = new PostFileRequest() {
    Upload = new Speakeasy.OpenAPI.File() {
        FileName = "example.file",
        Content = System.IO.File.ReadAllBytes("example.file"),
    },
};

var res = await sdk.PostFileAsync(req);

// handle response
```

### Per-Operation Security Schemes

Some operations in this SDK require the security scheme to be specified at the request level. For example:
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

var res = await sdk.Tag1.AuthAsync(security: new AuthSecurity() {
    AccessToken = "<YOUR_ACCESS_TOKEN_HERE>",
});

// handle response
```
<!-- End Authentication [security] -->

<!-- Start Available Resources and Operations [operations] -->
## Available Resources and Operations

<details open>
<summary>Available methods</summary>

### [SDK](docs/sdks/sdk/README.md)

* [OperationWithLeadingAndTrailingUnderscores](docs/sdks/sdk/README.md#operationwithleadingandtrailingunderscores)
* [PostFile](docs/sdks/sdk/README.md#postfile) - Post File
* [GetPolymorphism](docs/sdks/sdk/README.md#getpolymorphism)
* [GetUnionErrors](docs/sdks/sdk/README.md#getunionerrors)
* [GetRequestBodyFlattenedAway](docs/sdks/sdk/README.md#getrequestbodyflattenedaway)
* [GetFullyFlattenedRequest](docs/sdks/sdk/README.md#getfullyflattenedrequest)
* [CreateWithUnion](docs/sdks/sdk/README.md#createwithunion) - Create with discriminated union request body
* [TestEndpoint](docs/sdks/sdk/README.md#testendpoint)
* [CreateUser](docs/sdks/sdk/README.md#createuser) - Create User
* [GetUser](docs/sdks/sdk/README.md#getuser) - Get User
* [UpdateUser](docs/sdks/sdk/README.md#updateuser) - Update User
* [DeleteUser](docs/sdks/sdk/README.md#deleteuser) - Delete User
* [Login](docs/sdks/sdk/README.md#login) - Login
* [Validate](docs/sdks/sdk/README.md#validate) - Validate
* [Chat](docs/sdks/sdk/README.md#chat)
* [GetBinaryDefaultResponse](docs/sdks/sdk/README.md#getbinarydefaultresponse)
* [TestEnumFormats](docs/sdks/sdk/README.md#testenumformats) - Test x-speakeasy-enums in different formats
* [BinaryAndStringUpload](docs/sdks/sdk/README.md#binaryandstringupload)
* [GetErrorInUnion](docs/sdks/sdk/README.md#geterrorinunion)
* [GetDuplicateExportCollision](docs/sdks/sdk/README.md#getduplicateexportcollision) - Tests that a spec-defined error type colliding with a built-in SDK error name does not cause TS2308
* [GetNamedPrimitiveUnion](docs/sdks/sdk/README.md#getnamedprimitiveunion) - Test named primitive union options using title and x-speakeasy-name-override
* [GetEmptyObjectError](docs/sdks/sdk/README.md#getemptyobjecterror) - Get Empty Object Error
* [UrlValidationStressTest](docs/sdks/sdk/README.md#urlvalidationstresstest)
* [ParenthesesInPathAllowed](docs/sdks/sdk/README.md#parenthesesinpathallowed) - A string with {{ double braces }} and { single braces }
and \{\{ escaped curlies \}\} and `backticks`.
and \`escaped backticks\` and double slashes\\
and 'single quotes' and "double quotes".
and  \'escaped single quotes\' and \"escaped double quotes\".

* [GetNestedIntegerString](docs/sdks/sdk/README.md#getnestedintegerstring) - Test nested struct with integer:string tag
* [RenderAsset](docs/sdks/sdk/README.md#renderasset) - Render Asset
* [GetAsset](docs/sdks/sdk/README.md#getasset) - Get Asset
* [GetErrorOnlyExample](docs/sdks/sdk/README.md#geterroronlyexample) - Operation with example only on error response

### [Group](docs/sdks/group/README.md)

* [RootGroupOp](docs/sdks/group/README.md#rootgroupop) - An operation at the group's root level

#### [Group.SubGroup](docs/sdks/subgroup/README.md)

* [SubGroupOp](docs/sdks/subgroup/README.md#subgroupop) - An operation at the group's top level

##### [Group.SubGroup.Empty.Tail](docs/sdks/tail/README.md)

* [NestedGroupOp](docs/sdks/tail/README.md#nestedgroupop) - An operation at the group's deepest level

### [NamespaceTests.Conflicts](docs/sdks/conflicts/README.md)

* [GetNamespaceConflict](docs/sdks/conflicts/README.md#getnamespaceconflict) - Get Namespace Conflict Test
* [PutNamespaceConflict](docs/sdks/conflicts/README.md#putnamespaceconflict) - Put Property Name Conflicts Behind
* [CreateNamespaceConflict](docs/sdks/conflicts/README.md#createnamespaceconflict) - Create Namespace Conflict Test
* [GetTripleNamespaceConflict](docs/sdks/conflicts/README.md#gettriplenamespaceconflict) - Get Triple Namespace Conflict Test
* [GetPetOwners](docs/sdks/conflicts/README.md#getpetowners) - Get Pet Owners

### [NamespaceTests.SingleBar](docs/sdks/singlebar/README.md)

* [GetSingleNamespaceBarPet](docs/sdks/singlebar/README.md#getsinglenamespacebarpet) - Get Single Namespace Bar Pet

### [NamespaceTests.SingleFoo](docs/sdks/singlefoo/README.md)

* [GetSingleNamespaceFooPet](docs/sdks/singlefoo/README.md#getsinglenamespacefoopet) - Get Single Namespace Foo Pet
* [CreateSingleNamespaceFooPet](docs/sdks/singlefoo/README.md#createsinglenamespacefoopet) - Create Single Namespace Foo Pet

### [NamespaceTests.Types](docs/sdks/types/README.md)

* [GetNamespaceTypes](docs/sdks/types/README.md#getnamespacetypes) - Get Namespace Types Test
* [GetNamespaceAnimal](docs/sdks/types/README.md#getnamespaceanimal) - Get Namespace Animal (Discriminated Union)
* [GetNamespaceVehicle](docs/sdks/types/README.md#getnamespacevehicle) - Get Namespace Vehicle (Non-Discriminated Union)
* [GetNamespaceOrganization](docs/sdks/types/README.md#getnamespaceorganization) - Get Namespace Organization (Nested Inline Schemas)

### [~~Obsolete~~](docs/sdks/obsolete/README.md)

* [~~Deprecated1~~](docs/sdks/obsolete/README.md#deprecated1) - Deprecated Operation :warning: **Deprecated** Use [GetRequestBodyFlattenedAway](docs/sdks/sdk/README.md#getrequestbodyflattenedaway) instead.

### [Tag1](docs/sdks/tag1/README.md)

* [~~Deprecated1~~](docs/sdks/tag1/README.md#deprecated1) - Deprecated Operation :warning: **Deprecated** Use [GetRequestBodyFlattenedAway](docs/sdks/sdk/README.md#getrequestbodyflattenedaway) instead.
* [Auth](docs/sdks/tag1/README.md#auth) - This operation aims at testing available OAuth2 scopes collection:
 - only operation with oauth2 authorizationCode security flow
 - belongs to a subSDK

* [ListTest1](docs/sdks/tag1/README.md#listtest1) - Get Test1
* [PostFileWithEncoding](docs/sdks/tag1/README.md#postfilewithencoding) - Post File With Encoding

### [TestGroup.Tag2](docs/sdks/tag2/README.md)

* [PostTest](docs/sdks/tag2/README.md#posttest) - Post Test2

### [TestGroup.Tag3](docs/sdks/tag3/README.md)

* [PostTest](docs/sdks/tag3/README.md#posttest) - Post Test2

</details>
<!-- End Available Resources and Operations [operations] -->

<!-- Start Global Parameters [global-parameters] -->
## Global Parameters

Certain parameters are configured globally. These parameters may be set on the SDK client instance itself during initialization. When configured as an option during SDK initialization, These global values will be used as defaults on the operations that use them. When such operations are called, there is a place in each to override the global value, if needed.

For example, you can set `queryParam1` to `"some example query param"` at SDK initialization and then you do not have to pass the same value on calls to operations like `GetRequestBodyFlattenedAway`. But if you want to do so you may, which will locally override the global setting. See the example code below for a demonstration.


### Available Globals

The following global parameters are available.

| Name                  | Type   | Description                                                                        |
| --------------------- | ------ | ---------------------------------------------------------------------------------- |
| queryParam1           | string | A long winded, multi-line description<br/>for the query parameter number one.<br/> |
| deprecatedQueryParam1 | string | A deprecated description                                                           |
| deprecatedQueryParam2 | string | The DeprecatedQueryParam2 parameter.                                               |
| loneQueryParam        | string | The LoneQueryParam parameter.                                                      |

### Example

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(
    loneQueryParam: "<value>",
    queryParam1: "some example query param",
    deprecatedQueryParam1: "some example query param",
    deprecatedQueryParam2: "some example query param"
);

var res = await sdk.GetRequestBodyFlattenedAwayAsync();

// handle response
```
<!-- End Global Parameters [global-parameters] -->

<!-- Start Server-sent event streaming [eventstream] -->
## Server-sent event streaming

Server-sent events (SSE) are used to stream content from certain operations. SSE is a web standard that allows a server to push data to a client over a single HTTP connection in real-time.

These operations return an instance of `EventStream`, which implements `IAsyncEnumerable<T>`, `IAsyncDisposable` and `IDisposable`. An `EventStream` can only be consumed once and cannot be replayed.

Consume the stream with `await foreach`, which reads events as they arrive and disposes the stream when enumeration ends. To cancel iteration, pass the same `CancellationToken` to the operation and to `eventStream.WithCancellation(cancellationToken)`.

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

ChatRequest req = ChatRequest.CreateChatModelRequest(
    new ChatModelRequest() {
        Prompt = "What is the largest city in the world?",
        Stream = false,
    }
);

var res = await sdk.ChatAsync(req);

// Handle event stream response
var eventStream = res.ChatStream;
if (eventStream != null)
{
    await foreach (ChatStream eventData in eventStream)
    {
        // handle eventData
    }
}
// Handle JSON response
else if (res.Object != null)
{
    // handle JSON response
    var jsonResponse = res.Object;
}
```

For finer-grained control, you may call `Next()` until it returns `null`. `Next()` never disposes the `EventStream`, so wrap it in `using`/`await using` or dispose it explicitly.
<!-- End Server-sent event streaming [eventstream] -->

<!-- Start Pagination [pagination] -->
## Pagination

Some of the endpoints in this SDK support pagination. To use pagination, you make your SDK calls as usual, but the
returned response object will have a `Next` method that can be called to pull down the next group of results. If the
return value of `Next` is `null`, then there are no more pages to be fetched.

Here's an example of one such pagination call:
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(
    queryParam1: "some example query param",
    security: new Security() {
        MyApiKey = new MyApiKey() {
            MyApiKeyValue = "<YOUR_API_KEY_HERE>",
        },
    }
);

ListTest1Response? res = await sdk.Tag1.ListTest1Async(
    page: 100,
    queryParam2: QueryParam2.One,
    headerParam1: "some example header param"
);

while (res != null)
{
    // handle items

    res = await res.Next!();
}
```
<!-- End Pagination [pagination] -->

<!-- Start Retries [retries] -->
## Retries

Some of the endpoints in this SDK support retries. If you use the SDK without any configuration, it will fall back to the default retry strategy provided by the API. However, the default retry strategy can be overridden on a per-operation basis, or across the entire SDK.

To change the default retry strategy for a single API call, simply pass a `RetryConfig` to the call:
```csharp
using NodaTime;
using Speakeasy.OpenAPI;
using System;
using System.Collections.Generic;
using System.Numerics;

var sdk = new SDK(
    deprecatedQueryParam1: "some example query param",
    deprecatedQueryParam2: "some example query param",
    security: new Security() {
        MyApiKey = new MyApiKey() {
            MyApiKeyValue = "<YOUR_API_KEY_HERE>",
        },
    }
);

var res = await sdk.TestGroup.Tag2.PostTestAsync(
    retryConfig: new RetryConfig(
        strategy: RetryConfig.RetryStrategy.BACKOFF,
        backoff: new BackoffStrategy(
            initialIntervalMs: 1L,
            maxIntervalMs: 50L,
            maxElapsedTimeMs: 100L,
            exponent: 1.1
        ),
        retryConnectionErrors: false
    ),
    test2Request: new Test2Request() {
        Obj = new ExhaustiveObject() {
            Str = "example",
            Bool = true,
            Integer = 999999,
            Int32 = 1,
            Num = 1.1D,
            Float32 = 8499.3F,
            Date = LocalDate.FromDateTime(System.DateTime.Parse("2020-01-01")),
            DateTime = System.DateTime.Parse("2020-01-01T00:00:00Z").ToUniversalTime(),
            Anything = "<value>",
            BoolOpt = true,
            IntOptNull = 999999,
            NumOptNull = 1.1D,
            IntEnum = IntEnum.Third,
            Int32Enum = Int32Enum.SixtyNine,
            Bigint = BigInteger.Parse("702830"),
            DecimalStr = 3858.6M,
            Obj = new SimpleObject() {
                Str = "example",
            },
            Map = new Dictionary<string, SimpleObject>() {
                { "key", new SimpleObject() {
                     Str = "example",
                 } },
            },
            Arr = new List<SimpleObject>() {
                new SimpleObject() {
                    Str = "example",
                },
                new SimpleObject() {
                    Str = "example",
                },
            },
            Any = Any.CreateSimpleObject(new SimpleObject() {
                Str = "example",
            }),
            NullableIntEnum = NullableIntEnum.Third,
            NullableStringEnum = NullableStringEnum.Second,
            Color = Color.Green,
            Icon = Icon.Tick,
            HeroWidth = HeroWidth.FourHundredAndEighty,
        },
        Type = Speakeasy.OpenAPI.Type.SuperType1,
    }
);

// handle response
```

If you'd like to override the default retry strategy for all operations that support retries, you can use the `RetryConfig` optional parameter when intitializing the SDK:
```csharp
using NodaTime;
using Speakeasy.OpenAPI;
using System;
using System.Collections.Generic;
using System.Numerics;

var sdk = new SDK(
    retryConfig: new RetryConfig(
        strategy: RetryConfig.RetryStrategy.BACKOFF,
        backoff: new BackoffStrategy(
            initialIntervalMs: 1L,
            maxIntervalMs: 50L,
            maxElapsedTimeMs: 100L,
            exponent: 1.1
        ),
        retryConnectionErrors: false
    ),
    deprecatedQueryParam1: "some example query param",
    deprecatedQueryParam2: "some example query param",
    security: new Security() {
        MyApiKey = new MyApiKey() {
            MyApiKeyValue = "<YOUR_API_KEY_HERE>",
        },
    }
);

var res = await sdk.TestGroup.Tag2.PostTestAsync(test2Request: new Test2Request() {
    Obj = new ExhaustiveObject() {
        Str = "example",
        Bool = true,
        Integer = 999999,
        Int32 = 1,
        Num = 1.1D,
        Float32 = 8499.3F,
        Date = LocalDate.FromDateTime(System.DateTime.Parse("2020-01-01")),
        DateTime = System.DateTime.Parse("2020-01-01T00:00:00Z").ToUniversalTime(),
        Anything = "<value>",
        BoolOpt = true,
        IntOptNull = 999999,
        NumOptNull = 1.1D,
        IntEnum = IntEnum.Third,
        Int32Enum = Int32Enum.SixtyNine,
        Bigint = BigInteger.Parse("702830"),
        DecimalStr = 3858.6M,
        Obj = new SimpleObject() {
            Str = "example",
        },
        Map = new Dictionary<string, SimpleObject>() {
            { "key", new SimpleObject() {
                 Str = "example",
             } },
        },
        Arr = new List<SimpleObject>() {
            new SimpleObject() {
                Str = "example",
            },
            new SimpleObject() {
                Str = "example",
            },
        },
        Any = Any.CreateSimpleObject(new SimpleObject() {
            Str = "example",
        }),
        NullableIntEnum = NullableIntEnum.Third,
        NullableStringEnum = NullableStringEnum.Second,
        Color = Color.Green,
        Icon = Icon.Tick,
        HeroWidth = HeroWidth.FourHundredAndEighty,
    },
    Type = Speakeasy.OpenAPI.Type.SuperType1,
});

// handle response
```
<!-- End Retries [retries] -->

<!-- Start Error Handling [errors] -->
## Error Handling

[`SDKBaseException`](./src/Speakeasy/OpenAPI/Models/SDKBaseException.cs) is the base exception class for all HTTP error responses. It has the following properties:

| Property      | Type                  | Description           |
|---------------|-----------------------|-----------------------|
| `Message`     | *string*              | Error message         |
| `Request`     | *HttpRequestMessage*  | HTTP request object   |
| `Response`    | *HttpResponseMessage* | HTTP response object  |

Some exceptions in this SDK include an additional `Payload` field, which will contain deserialized custom error data when present. Possible exceptions are listed in the [Error Classes](#error-classes) section.

### Example

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

try
{
    GetUnionErrorsResponse? res = await sdk.GetUnionErrorsAsync(page: 12);

    while (res != null)
    {
        // handle items

        res = await res.Next!();
    }
}
catch (SDKBaseException ex) // all SDK exceptions inherit from SDKBaseException
{
    // ex.ToString() provides a detailed error message
    System.Console.WriteLine(ex);

    // Base exception fields
    HttpRequestMessage request = ex.Request;
    HttpResponseMessage response = ex.Response;
    var statusCode = (int)response.StatusCode;
    var responseBody = ex.Body;

    if (ex is ErrorsError) // different exceptions may be thrown depending on the method
    {
        // Check error data fields
        ErrorsErrorPayload payload = ex.Payload;
        string Error = payload.Error;
        long Code = payload.Code;
        // ...

        // If the response body did not match the expected schema, details are made
        // available in the DeserializationException field
        if (ex.DeserializationException != null)
        {
            ResponseValidationException validationEx = ex.DeserializationException;
            System.Console.WriteLine(validationEx.Message);
            Exception? cause = validationEx.InnerException;
        }
    }

    // An underlying cause may be provided
    if (ex.InnerException != null)
    {
        Exception cause = ex.InnerException;
    }
}
catch (OperationCanceledException ex)
{
    // CancellationToken was cancelled
}
catch (System.Net.Http.HttpRequestException ex)
{
    // Check ex.InnerException for Network connectivity errors
}
```

### Error Classes

**Primary exception:**
* [`SDKBaseException`](./src/Speakeasy/OpenAPI/Models/SDKBaseException.cs): The base class for HTTP error responses.

<details><summary>Less common exceptions (11)</summary>

* [`System.Net.Http.HttpRequestException`](https://learn.microsoft.com/en-us/dotnet/api/system.net.http.httprequestexception): Network connectivity error. For more details about the underlying cause, inspect the `ex.InnerException`.

* Inheriting from [`SDKBaseException`](./src/Speakeasy/OpenAPI/Models/SDKBaseException.cs):
  * [`ErrorsError`](./src/Speakeasy/OpenAPI/Models/ErrorsError.cs): A not-so-long multi-line error model description. Applicable to 7 of 50 methods.*
  * [`BadRequestResponseException`](./src/Speakeasy/OpenAPI/Models/BadRequestResponseException.cs): Bad Request. Status code `400`. Applicable to 2 of 50 methods.*
  * [`TaggedError1`](./src/Speakeasy/OpenAPI/Models/TaggedError1.cs): Applicable to 2 of 50 methods.*
  * [`RequestTimeoutError`](./src/Speakeasy/OpenAPI/Models/RequestTimeoutError.cs): A spec-defined error that collides with the built-in RequestTimeoutError in httpclienterrors.ts. Status code `408`. Applicable to 1 of 50 methods.*
  * [`TaggedError2`](./src/Speakeasy/OpenAPI/Models/TaggedError2.cs): Something went wrong. Status code `4XX`. Applicable to 1 of 50 methods.*
  * [`ErrorType1`](./src/Speakeasy/OpenAPI/Models/ErrorType1.cs): An error of type one. Status code `500`. Applicable to 1 of 50 methods.*
  * [`ErrorType2`](./src/Speakeasy/OpenAPI/Models/ErrorType2.cs): Internal Server Error. Status code `500`. Applicable to 1 of 50 methods.*
  * [`FailedResponseException`](./src/Speakeasy/OpenAPI/Models/FailedResponseException.cs): An error response with an empty object schema. Status code `500`. Applicable to 1 of 50 methods.*
  * [`Test2ResponseException`](./src/Speakeasy/OpenAPI/Models/Test2ResponseException.cs): Internal Server Error. Status code `500`. Applicable to 1 of 50 methods.*
  * [`ResponseValidationError`](./src/Speakeasy/OpenAPI/Models/ResponseValidationError.cs): Thrown when the response data could not be deserialized into the expected type.
</details>

\* Refer to the [relevant documentation](#available-resources-and-operations) to determine whether an exception applies to a specific operation.
<!-- End Error Handling [errors] -->

<!-- Start Server Selection [server] -->
## Server Selection

### Select Server by Index

You can override the default server globally by passing a server index to the `serverIndex: int` optional parameter when initializing the SDK client instance. The selected server will then be used as the default on the operations that use it. This table lists the indexes associated with the available servers:

| #   | Server                                     | Variables                 | Description                     |
| --- | ------------------------------------------ | ------------------------- | ------------------------------- |
| 0   | `http://localhost:35123`                   |                           | The default server.             |
| 1   | `http://{subdomain}.domain.com/v{version}` | `subdomain`<br/>`version` |                                 |
| 2   | `http://{HostName}:{PORT}`                 | `HostName`<br/>`PORT`     | A server with an enum variable. |

If the selected server has variables, you may override its default values through the additional parameters made available in the SDK constructor:

| Variable    | Parameter                            | Supported Values                      | Default       | Description                              |
| ----------- | ------------------------------------ | ------------------------------------- | ------------- | ---------------------------------------- |
| `subdomain` | `subdomain: string`                  | string                                | `"api"`       |                                          |
| `version`   | `version: string`                    | string                                | `"1"`         |                                          |
| `HostName`  | `hostName: string`                   | string                                | `"localhost"` | The hostname of the server.              |
| `PORT`      | `port: Speakeasy.OpenAPI.ServerPORT` | - `"80"`<br/>- `"8080"`<br/>- `"443"` | `"8080"`      | The port on which the server is running. |

#### Example

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(
    serverIndex: 2,
    hostName: "localhost",
    port: "443"
);

PostFileRequest req = new PostFileRequest() {
    Upload = new Speakeasy.OpenAPI.File() {
        FileName = "example.file",
        Content = System.IO.File.ReadAllBytes("example.file"),
    },
};

var res = await sdk.PostFileAsync(req);

// handle response
```

### Override Server URL Per-Client

The default server can also be overridden globally by passing a URL to the `serverUrl: string` optional parameter when initializing the SDK client instance. For example:
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(serverUrl: "http://localhost:8080");

PostFileRequest req = new PostFileRequest() {
    Upload = new Speakeasy.OpenAPI.File() {
        FileName = "example.file",
        Content = System.IO.File.ReadAllBytes("example.file"),
    },
};

var res = await sdk.PostFileAsync(req);

// handle response
```

### Override Server URL Per-Operation

The server URL can also be overridden on a per-operation basis, provided a server list was specified for the operation. For example:
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(
    queryParam1: "some example query param",
    security: new Security() {
        MyApiKey = new MyApiKey() {
            MyApiKeyValue = "<YOUR_API_KEY_HERE>",
        },
    }
);

ListTest1Response? res = await sdk.Tag1.ListTest1Async(
    serverUrl: "http://localhost:35123",
    page: 100,
    queryParam2: QueryParam2.One,
    headerParam1: "some example header param"
);

while (res != null)
{
    // handle items

    res = await res.Next!();
}
```
<!-- End Server Selection [server] -->

<!-- Start Custom HTTP Client [http-client] -->
## Custom HTTP Client

The C# SDK makes API calls using an `ISpeakeasyHttpClient` that wraps the native
[HttpClient](https://docs.microsoft.com/en-us/dotnet/api/system.net.http.httpclient). This
client provides the ability to attach hooks around the request lifecycle that can be used to modify the request or handle
errors and response.

The `ISpeakeasyHttpClient` interface allows you to either use the default `SpeakeasyHttpClient` that comes with the SDK,
or provide your own custom implementation with customized configuration such as custom message handlers, timeouts,
connection pooling, and other HTTP client settings.

The following example shows how to create a custom HTTP client with request modification and error handling:

```csharp
using Speakeasy.OpenAPI;
using Speakeasy.OpenAPI.Utils;
using System.Net.Http;
using System.Threading;
using System.Threading.Tasks;

// Create a custom HTTP client
public class CustomHttpClient : ISpeakeasyHttpClient
{
    private readonly ISpeakeasyHttpClient _defaultClient;

    public CustomHttpClient()
    {
        _defaultClient = new SpeakeasyHttpClient();
    }

    public async Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken? cancellationToken = null)
    {
        // Add custom header and timeout
        request.Headers.Add("x-custom-header", "custom value");
        request.Headers.Add("x-request-timeout", "30");
        
        try
        {
            var response = await _defaultClient.SendAsync(request, cancellationToken);
            // Log successful response
            Console.WriteLine($"Request successful: {response.StatusCode}");
            return response;
        }
        catch (Exception error)
        {
            // Log error
            Console.WriteLine($"Request failed: {error.Message}");
            throw;
        }
    }

    public void Dispose()
    {
        _httpClient?.Dispose();
        _defaultClient?.Dispose();
    }
}

// Use the custom HTTP client with the SDK
var customHttpClient = new CustomHttpClient();
var sdk = new SDK(client: customHttpClient);
```

<details>
<summary>You can also provide a completely custom HTTP client with your own configuration:</summary>

```csharp
using Speakeasy.OpenAPI.Utils;
using System.Net.Http;
using System.Threading;
using System.Threading.Tasks;

// Custom HTTP client with custom configuration
public class AdvancedHttpClient : ISpeakeasyHttpClient
{
    private readonly HttpClient _httpClient;

    public AdvancedHttpClient()
    {
        var handler = new HttpClientHandler()
        {
            MaxConnectionsPerServer = 10,
            // ServerCertificateCustomValidationCallback = customCertValidation, // Custom SSL validation if needed
        };

        _httpClient = new HttpClient(handler)
        {
            Timeout = TimeSpan.FromSeconds(30)
        };
    }

    public async Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken? cancellationToken = null)
    {
        return await _httpClient.SendAsync(request, cancellationToken ?? CancellationToken.None);
    }

    public void Dispose()
    {
        _httpClient?.Dispose();
    }
}

var sdk = SDK.Builder()
    .WithClient(new AdvancedHttpClient())
    .Build();
```
</details>

<details>
<summary>For simple debugging, you can enable request/response logging by implementing a custom client:</summary>

```csharp
public class LoggingHttpClient : ISpeakeasyHttpClient
{
    private readonly ISpeakeasyHttpClient _innerClient;

    public LoggingHttpClient(ISpeakeasyHttpClient innerClient = null)
    {
        _innerClient = innerClient ?? new SpeakeasyHttpClient();
    }

    public async Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken? cancellationToken = null)
    {
        // Log request
        Console.WriteLine($"Sending {request.Method} request to {request.RequestUri}");
        
        var response = await _innerClient.SendAsync(request, cancellationToken);
        
        // Log response
        Console.WriteLine($"Received {response.StatusCode} response");
        
        return response;
    }

    public void Dispose() => _innerClient?.Dispose();
}

var sdk = new SDK(client: new LoggingHttpClient());
```
</details>

The SDK also provides built-in hook support through the `SDKConfiguration.Hooks` system, which automatically handles
`BeforeRequestAsync`, `AfterSuccessAsync`, and `AfterErrorAsync` hooks for advanced request lifecycle management.
<!-- End Custom HTTP Client [http-client] -->

<!-- Placeholder for Future Speakeasy SDK Sections -->

# Development

## Maturity

This SDK is in beta, and there may be breaking changes between versions without a major version update. Therefore, we recommend pinning usage
to a specific package version. This way, you can install the same version each time without breaking changes unless you are intentionally
looking for the latest version.

## Contributions

While we value open-source contributions to this SDK, this library is generated programmatically. Any manual changes added to internal files will be overwritten on the next generation. 
We look forward to hearing your feedback. Feel free to open a PR or an issue with a proof of concept and we'll do our best to include it in a future release. 

### SDK Created by [Speakeasy](https://www.speakeasy.com/?utm_source=speakeasy-open-api&utm_campaign=csharp)
