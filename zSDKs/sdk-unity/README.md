# Speakeasy.OpenAPI

Developer-friendly & type-safe Unity SDK specifically catered to leverage *Speakeasy.OpenAPI* API.

[![Built by Speakeasy](https://img.shields.io/badge/Built_by-SPEAKEASY-374151?style=for-the-badge&labelColor=f3f4f6)](https://www.speakeasy.com/?utm_source=speakeasy-open-api&utm_campaign=unity)
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
  * [Pagination](#pagination)
  * [Error Handling](#error-handling)
  * [Server Selection](#server-selection)
* [Development](#development)
  * [Maturity](#maturity)
  * [Contributions](#contributions)

<!-- End Table of Contents [toc] -->

<!-- Start SDK Installation [installation] -->
## SDK Installation

The SDK can either be compiled using `dotnet build` and the resultant `.dll` file can be copied into your Unity project's `Assets` folder, or you can copy the source code directly into your project.

The SDK relies on Newtonsoft's JSON.NET Package which can be installed via the Unity Package Manager.

To do so open the Package Manager via `Window > Package Manager` and click the `+` button then `Add package from git URL...` and enter `com.unity.nuget.newtonsoft-json` and click `Add`.
<!-- End SDK Installation [installation] -->

<!-- Start SDK Example Usage [usage] -->
## SDK Example Usage

### Example 1

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

PostFileRequest req = new PostFileRequest() {
    Upload = new File() {
        FileName = "example.file",
        Content = File.ReadAllBytes("example.file"),
    },
};


using(var res = await sdk.PostFileAsync(req))
{
    // handle response
}


```

### Example 2

```csharp
using Speakeasy.OpenAPI;
using System.Collections.Generic;

var sdk = new SDK();

PostFileWithEncodingRequest req = new PostFileWithEncodingRequest() {
    File = new PostFileWithEncodingFile() {
        FileName = "example.file",
        Content = File.ReadAllBytes("example.file"),
    },
};


using(var res = await sdk.Tag1.PostFileWithEncodingAsync(req))
{
    // handle response
}


```

### Example 3

```csharp
using Speakeasy.OpenAPI;
using SDK.Utils;
using System.Collections.Generic;

var sdk = new SDK(
    deprecatedQueryParam1: "some example query param",
    deprecatedQueryParam2: "some example query param",
    security: new Security() {
        MyApiKey = new MyApiKey() {
            MyApiKey = "<YOUR_API_KEY_HERE>",
        },
    });


using(var res = await sdk.TestGroup.Tag2.PostTestAsync(test2Request: new Test2Request() {
    Obj = new ExhaustiveObject() {
        Str = "example",
        Bool = true,
        Integer = 999999,
        Int32 = 1,
        Num = 1.1D,
        Float32 = 8499.3F,
        EnumProp = .Enum.First,
        Date = DateOnly.FromDateTime(System.DateTime.Parse("2020-01-01")),
        DateTime = System.DateTime.Parse("2020-01-01T00:00:00Z"),
        Anything = "<value>",
        BoolOpt = true,
        IntOptNull = 999999,
        NumOptNull = 1.1D,
        IntEnum = IntEnum.Third,
        Int32Enum = Int32Enum.SixtyNine,
        Bigint = 702830,
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
        Any = Any.CreateSimpleObject(
            new SimpleObject() {
                Str = "example",
            },
        ),
        Type = "<value>",
        NullableIntEnum = NullableIntEnum.Third,
        NullableStringEnum = NullableStringEnum.Second,
        Color = Color.Green,
        Icon = Icon.Tick,
        HeroWidth = HeroWidth.FourHundredAndEighty,
    },
    Type = Type.SuperType1,
}))
{
    // handle response
}


```

### A custom readme heading

A custom usage description

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(
    queryParam1: "some example query param",
    security: new Security() {
        MyApiKey = new MyApiKey() {
            MyApiKey = "<YOUR_API_KEY_HERE>",
        },
    });


using(var res = await sdk.Tag1.ListTest1Async(
    page: 100,
    queryParam2: QueryParam2.One,
    headerParam1: "some example header param"))
{
    while(true)
    {
        res = await res.Next();
        // handle items
        if (res != null)
        {
            break;
        }
    }
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
    Upload = new File() {
        FileName = "example.file",
        Content = File.ReadAllBytes("example.file"),
    },
};


using(var res = await sdk.PostFileAsync(req))
{
    // handle response
}


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
    Upload = new File() {
        FileName = "example.file",
        Content = File.ReadAllBytes("example.file"),
    },
};


using(var res = await sdk.PostFileAsync(req))
{
    // handle response
}


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
    Upload = new File() {
        FileName = "example.file",
        Content = File.ReadAllBytes("example.file"),
    },
};


using(var res = await sdk.PostFileAsync(req))
{
    // handle response
}


```

#### Option4

The `Option4` alternative relies on the following scheme:

| Name                      | Type | Scheme      |
| ------------------------- | ---- | ----------- |
| `Username`<br/>`Password` | http | Custom HTTP |

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(security: new Security() {
        Option4 = new SecurityOption4() {
            Username = "",
            Password = "",
        },
    });

PostFileRequest req = new PostFileRequest() {
    Upload = new File() {
        FileName = "example.file",
        Content = File.ReadAllBytes("example.file"),
    },
};


using(var res = await sdk.PostFileAsync(req))
{
    // handle response
}


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
    Upload = new File() {
        FileName = "example.file",
        Content = File.ReadAllBytes("example.file"),
    },
};


using(var res = await sdk.PostFileAsync(req))
{
    // handle response
}


```

#### Option6

The `Option6` alternative relies on the following scheme:

| Name                | Type   | Scheme       |
| ------------------- | ------ | ------------ |
| `ClientCredentials` | oauth2 | OAuth2 token |

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(security: new Security() {
        Option6 = new SecurityOption6() {
            ClientCredentials = "<YOUR_CLIENT_CREDENTIALS_HERE>",
        },
    });

PostFileRequest req = new PostFileRequest() {
    Upload = new File() {
        FileName = "example.file",
        Content = File.ReadAllBytes("example.file"),
    },
};


using(var res = await sdk.PostFileAsync(req))
{
    // handle response
}


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
            MyApiKey = "<YOUR_API_KEY_HERE>",
        },
    });

PostFileRequest req = new PostFileRequest() {
    Upload = new File() {
        FileName = "example.file",
        Content = File.ReadAllBytes("example.file"),
    },
};


using(var res = await sdk.PostFileAsync(req))
{
    // handle response
}


```

### Per-Operation Security Schemes

Some operations in this SDK require the security scheme to be specified at the request level. For example:
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.Tag1.AuthAsync(security: new AuthSecurity() {
        AccessToken = "<YOUR_ACCESS_TOKEN_HERE>",
    }))
{
    // handle response
}


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
* [GetBinaryDefaultResponse](docs/sdks/sdk/README.md#getbinarydefaultresponse)
* [TestEnumFormats](docs/sdks/sdk/README.md#testenumformats) - Test x-speakeasy-enums in different formats
* [BinaryAndStringUpload](docs/sdks/sdk/README.md#binaryandstringupload)
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
| QueryParam1           | string | A long winded, multi-line description<br/>for the query parameter number one.<br/> |
| DeprecatedQueryParam1 | string | A deprecated description                                                           |
| DeprecatedQueryParam2 | string | The DeprecatedQueryParam2 parameter.                                               |
| LoneQueryParam        | string | The LoneQueryParam parameter.                                                      |

### Example

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(
    loneQueryParam: "<value>",
    queryParam1: "some example query param",
    deprecatedQueryParam1: "some example query param",
    deprecatedQueryParam2: "some example query param");


using(var res = await sdk.GetRequestBodyFlattenedAwayAsync())
{
    // handle response
}


```
<!-- End Global Parameters [global-parameters] -->

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
            MyApiKey = "<YOUR_API_KEY_HERE>",
        },
    });


using(var res = await sdk.Tag1.ListTest1Async(
    page: 100,
    queryParam2: QueryParam2.One,
    headerParam1: "some example header param"))
{
    while(true)
    {
        res = await res.Next();
        // handle items
        if (res != null)
        {
            break;
        }
    }
}


```
<!-- End Pagination [pagination] -->

<!-- Start Error Handling [errors] -->
## Error Handling

Handling errors in this SDK should largely match your expectations. All operations return a response object or throw an exception.

By default, an API error will raise a `SDKException` exception, which has the following properties:

| Property      | Type                  | Description           |
|---------------|-----------------------|-----------------------|
| `Message`     | *string*              | The error message     |
| `StatusCode`  | *int*                 | The raw HTTP response |
| `RawResponse` | *HttpResponseMessage* | The raw HTTP response |
| `Body`        | *string*              | The response content  |

When custom error responses are specified for an operation, the SDK may also throw their associated exception. You can refer to respective *Errors* tables in SDK docs for more details on possible exception types for each operation. For example, the `PostFileAsync` method throws the following exceptions:

| Error Type  | Status Code | Content Type     |
| ----------- | ----------- | ---------------- |
| ErrorsError | 415, 4XX    | application/json |
| ErrorsError | 5XX         | application/json |

### Example

```csharp
using Speakeasy.OpenAPI;
using System;

var sdk = new SDK();

PostFileRequest req = new PostFileRequest() {
    Upload = new File() {
        FileName = "example.file",
        Content = File.ReadAllBytes("example.file"),
    },
};

try
{
    using(var res = await sdk.PostFileAsync(req))
    {
            // handle response
    }
}
catch (Exception ex)
{
    if (ex is ErrorsError)
    {
        // handle exception
    }
    else if (ex is ErrorsError)
    {
        // handle exception
    }
    else if (ex is Speakeasy.OpenAPI.SDKException)
    {
        // handle exception
    }
}

```
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

| Variable    | Parameter           | Supported Values                      | Default       | Description                              |
| ----------- | ------------------- | ------------------------------------- | ------------- | ---------------------------------------- |
| `subdomain` | `subdomain: string` | string                                | `"api"`       |                                          |
| `version`   | `version: string`   | string                                | `"1"`         |                                          |
| `HostName`  | `hostName: string`  | string                                | `"localhost"` | The hostname of the server.              |
| `PORT`      | `port: ServerPORT`  | - `"80"`<br/>- `"8080"`<br/>- `"443"` | `"8080"`      | The port on which the server is running. |

#### Example

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(
    serverIndex: 2,
    hostName: "localhost",
    port: "443");

PostFileRequest req = new PostFileRequest() {
    Upload = new File() {
        FileName = "example.file",
        Content = File.ReadAllBytes("example.file"),
    },
};


using(var res = await sdk.PostFileAsync(req))
{
    // handle response
}


```

### Override Server URL Per-Client

The default server can also be overridden globally by passing a URL to the `serverUrl: string` optional parameter when initializing the SDK client instance. For example:
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(serverUrl: "http://localhost:8080");

PostFileRequest req = new PostFileRequest() {
    Upload = new File() {
        FileName = "example.file",
        Content = File.ReadAllBytes("example.file"),
    },
};


using(var res = await sdk.PostFileAsync(req))
{
    // handle response
}


```

### Override Server URL Per-Operation

The server URL can also be overridden on a per-operation basis, provided a server list was specified for the operation. For example:
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(
    queryParam1: "some example query param",
    security: new Security() {
        MyApiKey = new MyApiKey() {
            MyApiKey = "<YOUR_API_KEY_HERE>",
        },
    });


using(var res = await sdk.Tag1.ListTest1Async(
    serverUrl: "http://localhost:35123",
    page: 100,
    queryParam2: QueryParam2.One,
    headerParam1: "some example header param"))
{
    while(true)
    {
        res = await res.Next();
        // handle items
        if (res != null)
        {
            break;
        }
    }
}


```
<!-- End Server Selection [server] -->

<!-- Placeholder for Future Speakeasy SDK Sections -->

# Development

## Maturity

This SDK is in beta, and there may be breaking changes between versions without a major version update. Therefore, we recommend pinning usage
to a specific package version. This way, you can install the same version each time without breaking changes unless you are intentionally
looking for the latest version.

## Contributions

While we value open-source contributions to this SDK, this library is generated programmatically. Any manual changes added to internal files will be overwritten on the next generation. 
We look forward to hearing your feedback. Feel free to open a PR or an issue with a proof of concept and we'll do our best to include it in a future release. 

### SDK Created by [Speakeasy](https://www.speakeasy.com/?utm_source=speakeasy-open-api&utm_campaign=unity)
