# openapi

Developer-friendly & type-safe Go SDK specifically catered to leverage *openapi* API.

[![Built by Speakeasy](https://img.shields.io/badge/Built_by-SPEAKEASY-374151?style=for-the-badge&labelColor=f3f4f6)](https://www.speakeasy.com/?utm_source=openapi&utm_campaign=go)
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
  * [Special Types](#special-types)
* [Development](#development)
  * [Maturity](#maturity)
  * [Contributions](#contributions)

<!-- End Table of Contents [toc] -->

<!-- Start SDK Installation [installation] -->
## SDK Installation

To add the SDK as a dependency to your project:
```bash
go get example.com/openapi-go-sdk
```
<!-- End SDK Installation [installation] -->

<!-- Start SDK Example Usage [usage] -->
## SDK Example Usage

### Example 1

```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	s := examplealias.New()

	example, fileErr := os.Open("example.file")
	if fileErr != nil {
		panic(fileErr)
	}

	res, err := s.PostFile(ctx, examplealias.PostFileRequest{
		Upload: examplealias.File{
			FileName: "example.file",
			Content:  example,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if res.File != nil {
		// handle response
	}
}

```

### Example 2

```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	s := examplealias.New()

	example, fileErr := os.Open("example.file")
	if fileErr != nil {
		panic(fileErr)
	}

	res, err := s.Tag1.PostFileWithEncoding(ctx, examplealias.PostFileWithEncodingRequest{
		File: examplealias.PostFileWithEncodingFile{
			FileName: "example.file",
			Content:  example,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if res.Object != nil {
		// handle response
	}
}

```

### Example 3

```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"example.com/openapi-go-sdk/optionalnullable"
	"example.com/openapi-go-sdk/types"
	"log"
	"math/big"
	"os"
)

func main() {
	ctx := context.Background()

	s := examplealias.New(
		examplealias.WithDeprecatedQueryParam1("some example query param"),
		examplealias.WithDeprecatedQueryParam2("some example query param"),
		examplealias.WithSecurity(examplealias.Security{
			MyAPIKey: &examplealias.MyAPIKey{
				MyAPIKey: os.Getenv("SPEAKEASY_MY_API_KEY"),
			},
		}),
	)

	res, err := s.TestGroup.Tag2.PostTest(ctx, examplealias.Test2Request{
		Obj: examplealias.ExhaustiveObject{
			Str:        "example",
			Bool:       true,
			Integer:    999999,
			Int32:      1,
			Num:        1.1,
			Float32:    8499.3,
			Date:       types.MustDateFromString("2020-01-01"),
			DateTime:   types.MustTimeFromString("2020-01-01T00:00:00Z"),
			Anything:   "<value>",
			BoolOpt:    examplealias.Pointer(true),
			IntOptNull: examplealias.Pointer[int64](999999),
			NumOptNull: examplealias.Pointer[float64](1.1),
			IntEnum:    examplealias.IntEnumThird.ToPointer(),
			Int32Enum:  examplealias.Int32EnumSixtyNine,
			Bigint:     big.NewInt(593288),
			DecimalStr: types.MustNewDecimalFromString("7028.3"),
			Obj: examplealias.SimpleObject{
				Str: "example",
			},
			Map: map[string]examplealias.SimpleObject{},
			Arr: []examplealias.SimpleObject{},
			Any: examplealias.NewAny(
				"<value>",
			),
			NullableIntEnum:    optionalnullable.From(examplealias.Pointer(examplealias.NullableIntEnumThird)),
			NullableStringEnum: examplealias.NullableStringEnumSecond.ToPointer(),
			Color:              examplealias.ColorGreen.ToPointer(),
			Icon:               examplealias.IconTick,
			HeroWidth:          examplealias.HeroWidthFourHundredAndEighty.ToPointer(),
		},
		Type: examplealias.TypeSuperType1.ToPointer(),
	})
	if err != nil {
		log.Fatal(err)
	}
	if res.Body != nil {
		// handle response
	}
}

```

### A custom readme heading

A custom usage description

```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	s := examplealias.New(
		examplealias.WithQueryParam1("some example query param"),
		examplealias.WithSecurity(examplealias.Security{
			MyAPIKey: &examplealias.MyAPIKey{
				MyAPIKey: os.Getenv("SPEAKEASY_MY_API_KEY"),
			},
		}),
	)

	res, err := s.Tag1.ListTest1(ctx, 100, examplealias.QueryParam2One, "some example header param")
	if err != nil {
		log.Fatal(err)
	}
	if res.Object != nil {
		for {
			// handle items

			res, err = res.Next()

			if err != nil {
				// handle error
			}

			if res == nil {
				break
			}
		}
	}
}

```
<!-- End SDK Example Usage [usage] -->

<!-- Start Authentication [security] -->
## Authentication

### Per-Client Security Schemes

This SDK supports multiple security scheme combinations globally. You can choose from one of the alternatives by using the `WithSecurity` option when initializing the SDK client instance. The selected option will be used by default to authenticate with the API for all operations that support it.

#### UserPassAuth

The `UserPassAuth` alternative relies on the following scheme:

| Name                      | Type | Scheme     | Environment Variable                          |
| ------------------------- | ---- | ---------- | --------------------------------------------- |
| `Username`<br/>`Password` | http | HTTP Basic | `SPEAKEASY_USERNAME`<br/>`SPEAKEASY_PASSWORD` |

Example:
```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	s := examplealias.New(
		examplealias.WithSecurity(examplealias.Security{
			UserPassAuth: &examplealias.UserPassAuth{
				Username: "<USERNAME>",
				Password: "<PASSWORD>",
			},
		}),
	)

	example, fileErr := os.Open("example.file")
	if fileErr != nil {
		panic(fileErr)
	}

	res, err := s.PostFile(ctx, examplealias.PostFileRequest{
		Upload: examplealias.File{
			FileName: "example.file",
			Content:  example,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if res.File != nil {
		// handle response
	}
}

```

#### Option2

All of the following schemes must be satisfied to use the `Option2` alternative:

| Name         | Type   | Scheme      | Environment Variable    |
| ------------ | ------ | ----------- | ----------------------- |
| `BearerAuth` | http   | HTTP Bearer | `SPEAKEASY_BEARER_AUTH` |
| `MyAPIKey`   | apiKey | API key     | `SPEAKEASY_MY_API_KEY`  |

Example:
```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	s := examplealias.New(
		examplealias.WithSecurity(examplealias.Security{
			Option2: &examplealias.SecurityOption2{
				BearerAuth: "<YOUR_JWT>",
				MyAPIKey:   os.Getenv("SPEAKEASY_MY_API_KEY"),
			},
		}),
	)

	example, fileErr := os.Open("example.file")
	if fileErr != nil {
		panic(fileErr)
	}

	res, err := s.PostFile(ctx, examplealias.PostFileRequest{
		Upload: examplealias.File{
			FileName: "example.file",
			Content:  example,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if res.File != nil {
		// handle response
	}
}

```

#### Option3

The `Option3` alternative relies on the following scheme:

| Name     | Type   | Scheme       | Environment Variable |
| -------- | ------ | ------------ | -------------------- |
| `Oauth2` | oauth2 | OAuth2 token | `SPEAKEASY_OAUTH2`   |

Example:
```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	s := examplealias.New(
		examplealias.WithSecurity(examplealias.Security{
			Option3: &examplealias.SecurityOption3{
				Oauth2: "Bearer <YOUR_OAUTH2_TOKEN>",
			},
		}),
	)

	example, fileErr := os.Open("example.file")
	if fileErr != nil {
		panic(fileErr)
	}

	res, err := s.PostFile(ctx, examplealias.PostFileRequest{
		Upload: examplealias.File{
			FileName: "example.file",
			Content:  example,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if res.File != nil {
		// handle response
	}
}

```

#### Option4

The `Option4` alternative relies on the following scheme:

| Name                 | Type | Scheme      | Environment Variable                      |
| -------------------- | ---- | ----------- | ----------------------------------------- |
| `AppID`<br/>`Secret` | http | Custom HTTP | `SPEAKEASY_APP_ID`<br/>`SPEAKEASY_SECRET` |

Example:
```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	s := examplealias.New(
		examplealias.WithSecurity(examplealias.Security{
			Option4: &examplealias.SecurityOption4{
				AppID:  "app-speakeasy-123",
				Secret: "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI",
			},
		}),
	)

	example, fileErr := os.Open("example.file")
	if fileErr != nil {
		panic(fileErr)
	}

	res, err := s.PostFile(ctx, examplealias.PostFileRequest{
		Upload: examplealias.File{
			FileName: "example.file",
			Content:  example,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if res.File != nil {
		// handle response
	}
}

```

#### Option5

The `Option5` alternative relies on the following scheme:

| Name         | Type   | Scheme       | Environment Variable    |
| ------------ | ------ | ------------ | ----------------------- |
| `MobileAuth` | oauth2 | OAuth2 token | `SPEAKEASY_MOBILE_AUTH` |

Example:
```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	s := examplealias.New(
		examplealias.WithSecurity(examplealias.Security{
			Option5: &examplealias.SecurityOption5{
				MobileAuth: "Bearer <YOUR_OAUTH2_TOKEN>",
			},
		}),
	)

	example, fileErr := os.Open("example.file")
	if fileErr != nil {
		panic(fileErr)
	}

	res, err := s.PostFile(ctx, examplealias.PostFileRequest{
		Upload: examplealias.File{
			FileName: "example.file",
			Content:  example,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if res.File != nil {
		// handle response
	}
}

```

#### Option6

The `Option6` alternative relies on the following scheme:

| Name                | Type   | Scheme       | Environment Variable           |
| ------------------- | ------ | ------------ | ------------------------------ |
| `ClientCredentials` | oauth2 | OAuth2 token | `SPEAKEASY_CLIENT_CREDENTIALS` |

Example:
```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	s := examplealias.New(
		examplealias.WithSecurity(examplealias.Security{
			Option6: &examplealias.SecurityOption6{
				ClientCredentials: os.Getenv("SPEAKEASY_CLIENT_CREDENTIALS"),
			},
		}),
	)

	example, fileErr := os.Open("example.file")
	if fileErr != nil {
		panic(fileErr)
	}

	res, err := s.PostFile(ctx, examplealias.PostFileRequest{
		Upload: examplealias.File{
			FileName: "example.file",
			Content:  example,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if res.File != nil {
		// handle response
	}
}

```

#### MyApiKey

The `MyApiKey` alternative relies on the following scheme:

| Name       | Type   | Scheme  | Environment Variable   |
| ---------- | ------ | ------- | ---------------------- |
| `MyAPIKey` | apiKey | API key | `SPEAKEASY_MY_API_KEY` |

Example:
```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	s := examplealias.New(
		examplealias.WithSecurity(examplealias.Security{
			MyAPIKey: &examplealias.MyAPIKey{
				MyAPIKey: os.Getenv("SPEAKEASY_MY_API_KEY"),
			},
		}),
	)

	example, fileErr := os.Open("example.file")
	if fileErr != nil {
		panic(fileErr)
	}

	res, err := s.PostFile(ctx, examplealias.PostFileRequest{
		Upload: examplealias.File{
			FileName: "example.file",
			Content:  example,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if res.File != nil {
		// handle response
	}
}

```

### Per-Operation Security Schemes

Some operations in this SDK require the security scheme to be specified at the request level. For example:
```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	s := examplealias.New()

	res, err := s.Tag1.Auth(ctx, examplealias.AuthSecurity{
		AccessToken: os.Getenv("SPEAKEASY_ACCESS_TOKEN"),
	})
	if err != nil {
		log.Fatal(err)
	}
	if res != nil {
		// handle response
	}
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
* [URLValidationStressTest](docs/sdks/sdk/README.md#urlvalidationstresstest)
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
Global parameters can also be set via environment variable.

| Name                  | Type   | Description                                                                        | Environment                       |
| --------------------- | ------ | ---------------------------------------------------------------------------------- | --------------------------------- |
| QueryParam1           | string | A long winded, multi-line description<br/>for the query parameter number one.<br/> | SPEAKEASY_QUERY_PARAM1            |
| DeprecatedQueryParam1 | string | A deprecated description                                                           | SPEAKEASY_DEPRECATED_QUERY_PARAM1 |
| DeprecatedQueryParam2 | string | The DeprecatedQueryParam2 parameter.                                               | SPEAKEASY_DEPRECATED_QUERY_PARAM2 |
| LoneQueryParam        | string | The LoneQueryParam parameter.                                                      | SPEAKEASY_LONE_QUERY_PARAM        |

### Example

```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
	ctx := context.Background()

	s := examplealias.New(
		examplealias.WithLoneQueryParam("<value>"),
		examplealias.WithQueryParam1("some example query param"),
		examplealias.WithDeprecatedQueryParam1("some example query param"),
		examplealias.WithDeprecatedQueryParam2("some example query param"),
	)

	res, err := s.GetRequestBodyFlattenedAway(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if res != nil {
		// handle response
	}
}

```
<!-- End Global Parameters [global-parameters] -->

<!-- Start Server-sent event streaming [eventstream] -->
## Server-sent event streaming

[Server-sent events][mdn-sse] are used to stream content from certain
operations. These operations will expose the stream as an iterable that
can be consumed using a simple `for` loop. The loop will
terminate when the server no longer has any events to send and closes the
underlying connection.

```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
	ctx := context.Background()

	s := examplealias.New()

	res, err := s.Chat(ctx, examplealias.NewChatRequest(
		examplealias.ChatModelRequest{
			Prompt: "What is the largest city in the world?",
			Stream: examplealias.Pointer(false),
		},
	))
	if err != nil {
		log.Fatal(err)
	}
	if res.Object != nil {
		defer res.ChatStream.Close()

		for res.ChatStream.Next() {
			event := res.ChatStream.Value()
			log.Print(event)
			// Handle the event
		}
	}
}

```

[mdn-sse]: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events
<!-- End Server-sent event streaming [eventstream] -->

<!-- Start Pagination [pagination] -->
## Pagination

Some of the endpoints in this SDK support pagination. To use pagination, you make your SDK calls as usual, but the
returned response object will have a `Next` method that can be called to pull down the next group of results. If the
return value of `Next` is `nil`, then there are no more pages to be fetched.

Here's an example of one such pagination call:
```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	s := examplealias.New(
		examplealias.WithQueryParam1("some example query param"),
		examplealias.WithSecurity(examplealias.Security{
			MyAPIKey: &examplealias.MyAPIKey{
				MyAPIKey: os.Getenv("SPEAKEASY_MY_API_KEY"),
			},
		}),
	)

	res, err := s.Tag1.ListTest1(ctx, 100, examplealias.QueryParam2One, "some example header param")
	if err != nil {
		log.Fatal(err)
	}
	if res.Object != nil {
		for {
			// handle items

			res, err = res.Next()

			if err != nil {
				// handle error
			}

			if res == nil {
				break
			}
		}
	}
}

```
<!-- End Pagination [pagination] -->

<!-- Start Retries [retries] -->
## Retries

Some of the endpoints in this SDK support retries. If you use the SDK without any configuration, it will fall back to the default retry strategy provided by the API. However, the default retry strategy can be overridden on a per-operation basis, or across the entire SDK.

To change the default retry strategy for a single API call, simply provide a `retry.Config` object to the call by using the `WithRetries` option:
```go
package main

import (
	""
	"context"
	examplealias "example.com/openapi-go-sdk"
	"example.com/openapi-go-sdk/retry"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	s := examplealias.New()

	example, fileErr := os.Open("example.file")
	if fileErr != nil {
		panic(fileErr)
	}

	res, err := s.PostFile(ctx, examplealias.PostFileRequest{
		Upload: examplealias.File{
			FileName: "example.file",
			Content:  example,
		},
	}, examplealias.WithRetries(
		retry.Config{
			Strategy: "backoff",
			Backoff: &retry.BackoffStrategy{
				InitialInterval: 1,
				MaxInterval:     50,
				Exponent:        1.1,
				MaxElapsedTime:  100,
			},
			RetryConnectionErrors: false,
		}))
	if err != nil {
		log.Fatal(err)
	}
	if res.File != nil {
		// handle response
	}
}

```

If you'd like to override the default retry strategy for all operations that support retries, you can use the `WithRetryConfig` option at SDK initialization:
```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"example.com/openapi-go-sdk/retry"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	s := examplealias.New(
		examplealias.WithRetryConfig(
			retry.Config{
				Strategy: "backoff",
				Backoff: &retry.BackoffStrategy{
					InitialInterval: 1,
					MaxInterval:     50,
					Exponent:        1.1,
					MaxElapsedTime:  100,
				},
				RetryConnectionErrors: false,
			}),
	)

	example, fileErr := os.Open("example.file")
	if fileErr != nil {
		panic(fileErr)
	}

	res, err := s.PostFile(ctx, examplealias.PostFileRequest{
		Upload: examplealias.File{
			FileName: "example.file",
			Content:  example,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if res.File != nil {
		// handle response
	}
}

```
<!-- End Retries [retries] -->

<!-- Start Error Handling [errors] -->
## Error Handling

Handling errors in this SDK should largely match your expectations. All operations return a response object or an error, they will never return both.

By Default, an API error will return `examplealias.SDKError`. When custom error responses are specified for an operation, the SDK may also return their associated error. You can refer to respective *Errors* tables in SDK docs for more details on possible error types for each operation.

For example, the `GetUnionErrors` function may return the following errors:

| Error Type                                     | Status Code | Content Type     |
| ---------------------------------------------- | ----------- | ---------------- |
| examplealias.ErrorsError                       | 404         | application/json |
| examplealias.GetUnionErrorsInternalServerError | 500         | application/json |
| examplealias.ClientError                       | 4XX         | application/json |
| examplealias.SDKError                          | 5XX         | \*/\*            |

### Example

```go
package main

import (
	"context"
	"errors"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
	ctx := context.Background()

	s := examplealias.New()

	res, err := s.GetUnionErrors(ctx, 12)
	if err != nil {

		var e *ErrorsError
		if errors.As(err, &e) {
			// handle error
			log.Fatal(e.Error())
		}

		var e *GetUnionErrorsInternalServerError
		if errors.As(err, &e) {
			// handle error
			log.Fatal(e.Error())
		}

		var e *ClientError
		if errors.As(err, &e) {
			// handle error
			log.Fatal(e.Error())
		}

		var e *examplealias.SDKError
		if errors.As(err, &e) {
			// handle error
			log.Fatal(e.Error())
		}
	}
}

```
<!-- End Error Handling [errors] -->

<!-- Start Server Selection [server] -->
## Server Selection

### Select Server by Index

You can override the default server globally using the `WithServerIndex(serverIndex int)` option when initializing the SDK client instance. The selected server will then be used as the default on the operations that use it. This table lists the indexes associated with the available servers:

| #   | Server                                     | Variables                 | Description                     |
| --- | ------------------------------------------ | ------------------------- | ------------------------------- |
| 0   | `http://localhost:35123`                   |                           | The default server.             |
| 1   | `http://{subdomain}.domain.com/v{version}` | `subdomain`<br/>`version` |                                 |
| 2   | `http://{HostName}:{PORT}`                 | `HostName`<br/>`PORT`     | A server with an enum variable. |

If the selected server has variables, you may override its default values using the associated option(s):

| Variable    | Option                            | Supported Values                      | Default       | Description                              |
| ----------- | --------------------------------- | ------------------------------------- | ------------- | ---------------------------------------- |
| `subdomain` | `WithSubdomain(subdomain string)` | string                                | `"api"`       |                                          |
| `version`   | `WithVersion(version string)`     | string                                | `"1"`         |                                          |
| `HostName`  | `WithHostName(hostName string)`   | string                                | `"localhost"` | The hostname of the server.              |
| `PORT`      | `WithPort(port ServerPORT)`       | - `"80"`<br/>- `"8080"`<br/>- `"443"` | `"8080"`      | The port on which the server is running. |

#### Example

```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	s := examplealias.New(
		examplealias.WithServerIndex(2),
		examplealias.WithHostName("localhost"),
		examplealias.WithPort("443"),
	)

	example, fileErr := os.Open("example.file")
	if fileErr != nil {
		panic(fileErr)
	}

	res, err := s.PostFile(ctx, examplealias.PostFileRequest{
		Upload: examplealias.File{
			FileName: "example.file",
			Content:  example,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if res.File != nil {
		// handle response
	}
}

```

### Override Server URL Per-Client

The default server can also be overridden globally using the `WithServerURL(serverURL string)` option when initializing the SDK client instance. For example:
```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	s := examplealias.New(
		examplealias.WithServerURL("http://localhost:8080"),
	)

	example, fileErr := os.Open("example.file")
	if fileErr != nil {
		panic(fileErr)
	}

	res, err := s.PostFile(ctx, examplealias.PostFileRequest{
		Upload: examplealias.File{
			FileName: "example.file",
			Content:  example,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if res.File != nil {
		// handle response
	}
}

```

### Override Server URL Per-Operation

The server URL can also be overridden on a per-operation basis, provided a server list was specified for the operation. For example:
```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	s := examplealias.New(
		examplealias.WithQueryParam1("some example query param"),
		examplealias.WithSecurity(examplealias.Security{
			MyAPIKey: &examplealias.MyAPIKey{
				MyAPIKey: os.Getenv("SPEAKEASY_MY_API_KEY"),
			},
		}),
	)

	res, err := s.Tag1.ListTest1(ctx, 100, examplealias.QueryParam2One, "some example header param", examplealias.WithMethodServerURL("http://localhost:35123"))
	if err != nil {
		log.Fatal(err)
	}
	if res.Object != nil {
		for {
			// handle items

			res, err = res.Next()

			if err != nil {
				// handle error
			}

			if res == nil {
				break
			}
		}
	}
}

```
<!-- End Server Selection [server] -->

<!-- Start Custom HTTP Client [http-client] -->
## Custom HTTP Client

The Go SDK makes API calls that wrap an internal HTTP client. The requirements for the HTTP client are very simple. It must match this interface:

```go
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
```

The built-in `net/http` client satisfies this interface and a default client based on the built-in is provided by default. To replace this default with a client of your own, you can implement this interface yourself or provide your own client configured as desired. Here's a simple example, which adds a client with a 30 second timeout.

```go
import (
	"net/http"
	"time"

	"example.com/openapi-go-sdk"
)

var (
	httpClient = &http.Client{Timeout: 30 * time.Second}
	sdkClient  = examplealias.New(examplealias.WithClient(httpClient))
)
```

This can be a convenient way to configure timeouts, cookies, proxies, custom headers, and other low-level configuration.
<!-- End Custom HTTP Client [http-client] -->

<!-- Start Special Types [types] -->
## Special Types

This SDK defines the following custom types to assist with marshalling and unmarshalling data.

### Date

`types.Date` is a wrapper around time.Time that allows for JSON marshaling a date string formatted as "2006-01-02".

#### Usage

```go
d1 := types.NewDate(time.Now()) // returns *types.Date

d2 := types.DateFromTime(time.Now()) // returns types.Date

d3, err := types.NewDateFromString("2019-01-01") // returns *types.Date, error

d4, err := types.DateFromString("2019-01-01") // returns types.Date, error

d5 := types.MustNewDateFromString("2019-01-01") // returns *types.Date and panics on error

d6 := types.MustDateFromString("2019-01-01") // returns types.Date and panics on error
```
<!-- End Special Types [types] -->

<!-- Placeholder for Future Speakeasy SDK Sections -->

# Development

## Maturity

This SDK is in beta, and there may be breaking changes between versions without a major version update. Therefore, we recommend pinning usage
to a specific package version. This way, you can install the same version each time without breaking changes unless you are intentionally
looking for the latest version.

## Contributions

While we value open-source contributions to this SDK, this library is generated programmatically. Any manual changes added to internal files will be overwritten on the next generation. 
We look forward to hearing your feedback. Feel free to open a PR or an issue with a proof of concept and we'll do our best to include it in a future release. 

### SDK Created by [Speakeasy](https://www.speakeasy.com/?utm_source=openapi&utm_campaign=go)
