# openapi

Developer-friendly & type-safe Java SDK specifically catered to leverage *openapi* API.

[![Built by Speakeasy](https://img.shields.io/badge/Built_by-SPEAKEASY-374151?style=for-the-badge&labelColor=f3f4f6)](https://www.speakeasy.com/?utm_source=openapi&utm_campaign=java)
[![License: test-1](https://img.shields.io/badge/LICENSE_//_test--1-3b5bdb?style=for-the-badge&labelColor=eff6ff)](testurl.com)


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
  * [Asynchronous Support](#asynchronous-support)
  * [Authentication](#authentication)
  * [Available Resources and Operations](#available-resources-and-operations)
  * [Global Parameters](#global-parameters)
  * [Server-sent event streaming](#server-sent-event-streaming)
  * [Pagination](#pagination)
  * [File uploads](#file-uploads)
  * [Retries](#retries)
  * [Error Handling](#error-handling)
  * [Server Selection](#server-selection)
  * [Custom HTTP Client](#custom-http-client)
  * [Debugging](#debugging)
  * [Jackson Configuration](#jackson-configuration)
* [Development](#development)
  * [Maturity](#maturity)
  * [Contributions](#contributions)

<!-- End Table of Contents [toc] -->

<!-- Start SDK Installation [installation] -->
## SDK Installation

### Getting started

JDK 17 or later is required.

The samples below show how a published SDK artifact is used:

Gradle:
```groovy
implementation 'org.openapis.review:openapi:0.0.1'
```

Maven:
```xml
<dependency>
    <groupId>org.openapis.review</groupId>
    <artifactId>openapi</artifactId>
    <version>0.0.1</version>
</dependency>
```

### How to build
After cloning the git repository to your file system you can build the SDK artifact from source to the `build` directory by running `./gradlew build` on *nix systems or `gradlew.bat` on Windows systems.

If you wish to build from source and publish the SDK artifact to your local Maven repository (on your filesystem) then use the following command (after cloning the git repo locally):

On *nix:
```bash
./gradlew publishToMavenLocal -Pskip.signing
```
On Windows:
```bash
gradlew.bat publishToMavenLocal -Pskip.signing
```
<!-- End SDK Installation [installation] -->

<!-- Start SDK Example Usage [usage] -->
## SDK Example Usage

### Example 1

```java
package hello.world;

import java.lang.Exception;
import java.nio.file.Paths;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.operations.PostFileRequest;
import org.openapis.openapi.models.operations.PostFileResponse;
import org.openapis.openapi.models.shared.File;
import org.openapis.openapi.utils.Blob;

public class Application {

    public static void main(String[] args) throws Error, Exception {

        SDK sdk = SDK.builder().build();

        PostFileRequest req = PostFileRequest.builder()
                .upload(File.builder()
                        .fileName("example.file")
                        .content(Blob.from(Paths.get("example.file")))
                        .build())
                .build();

        PostFileResponse res = sdk.postFile().request(req).call();
    }
}

```

### Example 2

```java
package hello.world;

import java.lang.Exception;
import java.nio.file.Paths;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.operations.File;
import org.openapis.openapi.models.operations.PostFileWithEncodingRequest;
import org.openapis.openapi.models.operations.PostFileWithEncodingResponse;
import org.openapis.openapi.utils.Blob;

public class Application {

    public static void main(String[] args) throws Error, Exception {

        SDK sdk = SDK.builder().build();

        PostFileWithEncodingRequest req = PostFileWithEncodingRequest.builder()
                .file(File.builder()
                        .fileName("example.file")
                        .content(Blob.from(Paths.get("example.file")))
                        .build())
                .build();

        PostFileWithEncodingResponse res =
                sdk.tag1().postFileWithEncoding().request(req).call();
    }
}

```

### Example 3

```java
package hello.world;

import java.lang.Exception;
import java.math.BigDecimal;
import java.math.BigInteger;
import java.time.LocalDate;
import java.time.OffsetDateTime;
import java.util.List;
import java.util.Map;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.BadRequestResponseException;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.errors.Test2ResponseException;
import org.openapis.openapi.models.operations.PostTest2Response;
import org.openapis.openapi.models.shared.Any;
import org.openapis.openapi.models.shared.Color;
import org.openapis.openapi.models.shared.ExhaustiveObject;
import org.openapis.openapi.models.shared.HeroWidth;
import org.openapis.openapi.models.shared.Icon;
import org.openapis.openapi.models.shared.Int32Enum;
import org.openapis.openapi.models.shared.IntEnum;
import org.openapis.openapi.models.shared.MyApiKey;
import org.openapis.openapi.models.shared.NullableIntEnum;
import org.openapis.openapi.models.shared.NullableStringEnum;
import org.openapis.openapi.models.shared.Security;
import org.openapis.openapi.models.shared.SimpleObject;
import org.openapis.openapi.models.shared.Test2Request;
import org.openapis.openapi.models.shared.Type;

public class Application {

    public static void main(String[] args)
            throws BadRequestResponseException, Error, Test2ResponseException, Exception {

        SDK sdk = SDK.builder()
                .deprecatedQueryParam1("some example query param")
                .deprecatedQueryParam2("some example query param")
                .security(Security.builder()
                        .myApiKey(MyApiKey.builder().myApiKey("<value>").build())
                        .build())
                .build();

        PostTest2Response res = sdk.testGroup()
                .tag2()
                .postTest()
                .test2Request(Test2Request.builder()
                        .obj(ExhaustiveObject.builder()
                                .str("example")
                                .bool(true)
                                .integer(999999L)
                                .int32(1)
                                .num(1.1)
                                .float32(8499.3f)
                                .date(LocalDate.parse("2020-01-01"))
                                .dateTime(OffsetDateTime.parse("2020-01-01T00:00:00Z"))
                                .anything("<value>")
                                .int32Enum(Int32Enum.SIXTY_NINE)
                                .bigint(new BigInteger("593288"))
                                .decimalStr(new BigDecimal("7028.3"))
                                .obj(SimpleObject.builder().str("example").build())
                                .map(Map.ofEntries())
                                .arr(List.of())
                                .any(Any.of("<value>"))
                                .nullableStringEnum(NullableStringEnum.SECOND)
                                .icon(Icon.TICK)
                                .boolOpt(true)
                                .intOptNull(999999L)
                                .numOptNull(1.1)
                                .intEnum(IntEnum.Third)
                                .nullableIntEnum(NullableIntEnum.Third)
                                .color(Color.GREEN)
                                .heroWidth(HeroWidth.FOUR_HUNDRED_AND_EIGHTY)
                                .build())
                        .type(Type.SuperType1)
                        .build())
                .call();

        if (res.body().isPresent()) {
            System.out.println(res.body().get());
        }
    }
}

```

### A custom readme heading

A custom usage description

```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.BadRequestResponseException;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.operations.QueryParam2;
import org.openapis.openapi.models.operations.ResultArray;
import org.openapis.openapi.models.shared.MyApiKey;
import org.openapis.openapi.models.shared.Security;

public class Application {

    public static void main(String[] args) throws BadRequestResponseException, Error, Exception {

        SDK sdk = SDK.builder()
                .queryParam1("some example query param")
                .security(Security.builder()
                        .myApiKey(MyApiKey.builder().myApiKey("<value>").build())
                        .build())
                .build();

        sdk.tag1()
                .listTest1()
                .page(100L)
                .queryParam2(QueryParam2.ONE)
                .headerParam1("some example header param")
                .callAsStreamUnwrapped()
                .forEach((ResultArray item) -> {
                    // handle item
                });
    }
}

```
#### Asynchronous Call
An asynchronous SDK client is also available that returns a [`CompletableFuture<T>`][comp-fut]. See [Asynchronous Support](#asynchronous-support) for more details on async benefits and reactive library integration.
```java
package hello.world;

import java.nio.file.Paths;
import java.util.concurrent.CompletableFuture;
import org.openapis.openapi.AsyncSDK;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.PostFileRequest;
import org.openapis.openapi.models.operations.async.PostFileResponse;
import org.openapis.openapi.models.shared.File;
import org.openapis.openapi.utils.Blob;

public class Application {

    public static void main(String[] args) {

        AsyncSDK sdk = SDK.builder().build().toAsync();

        PostFileRequest req = PostFileRequest.builder()
                .upload(File.builder()
                        .fileName("example.file")
                        .content(Blob.from(Paths.get("example.file")))
                        .build())
                .build();

        CompletableFuture<PostFileResponse> resFut = sdk.postFile().request(req).call();
    }
}

```

[comp-fut]: https://docs.oracle.com/javase/8/docs/api/java/util/concurrent/CompletableFuture.html

#### Union Consumption Patterns

When a response field is a union model:

- Discriminated unions: branch on the discriminator (`switch`) and then narrow to the concrete type.
- Non-discriminated unions: use generated accessors (for example `string()`, `asLong()`, `simpleObject()`) to determine the active variant.

For full model-specific examples (including Java 11/16/21 variants), see each union model's **Supported Types** section in the generated model docs.
<!-- End SDK Example Usage [usage] -->

<!-- Start Asynchronous Support [async-support] -->
## Asynchronous Support

The SDK provides comprehensive asynchronous support using Java's [`CompletableFuture<T>`][comp-fut] and [Reactive Streams `Publisher<T>`][reactive-streams] APIs. This design makes no assumptions about your choice of reactive toolkit, allowing seamless integration with any reactive library.

<details>
<summary>Why Use Async?</summary>

Asynchronous operations provide several key benefits:

- **Non-blocking I/O**: Your threads stay free for other work while operations are in flight
- **Better resource utilization**: Handle more concurrent operations with fewer threads
- **Improved scalability**: Build highly responsive applications that can handle thousands of concurrent requests
- **Reactive integration**: Works seamlessly with reactive streams and backpressure handling

</details>

<details>
<summary>Reactive Library Integration</summary>

The SDK returns [Reactive Streams `Publisher<T>`][reactive-streams] instances for operations dealing with streams involving multiple I/O interactions. We use Reactive Streams instead of JDK Flow API to provide broader compatibility with the reactive ecosystem, as most reactive libraries natively support Reactive Streams.

**Why Reactive Streams over JDK Flow?**
- **Broader ecosystem compatibility**: Most reactive libraries (Project Reactor, RxJava, Akka Streams, etc.) natively support Reactive Streams
- **Industry standard**: Reactive Streams is the de facto standard for reactive programming in Java
- **Better interoperability**: Seamless integration without additional adapters for most use cases

**Integration with Popular Libraries:**
- **Project Reactor**: Use `Flux.from(publisher)` to convert to Reactor types
- **RxJava**: Use `Flowable.fromPublisher(publisher)` for RxJava integration
- **Akka Streams**: Use `Source.fromPublisher(publisher)` for Akka Streams integration
- **Vert.x**: Use `ReadStream.fromPublisher(vertx, publisher)` for Vert.x reactive streams
- **Mutiny**: Use `Multi.createFrom().publisher(publisher)` for Quarkus Mutiny integration

**For JDK Flow API Integration:**
If you need JDK Flow API compatibility (e.g., for Quarkus/Mutiny 2), you can use adapters:
```java
// Convert Reactive Streams Publisher to Flow Publisher
Flow.Publisher<T> flowPublisher = FlowAdapters.toFlowPublisher(reactiveStreamsPublisher);

// Convert Flow Publisher to Reactive Streams Publisher
Publisher<T> reactiveStreamsPublisher = FlowAdapters.toPublisher(flowPublisher);
```

For standard single-response operations, the SDK returns `CompletableFuture<T>` for straightforward async execution.

</details>

<details>
<summary>Supported Operations</summary>

Async support is available for:

- **[Server-sent Events](#server-sent-event-streaming)**: Stream real-time events with Reactive Streams `Publisher<T>`
- **[JSONL Streaming](#jsonl-streaming)**: Process streaming JSON lines asynchronously
- **[Pagination](#pagination)**: Iterate through paginated results using `callAsPublisher()` and `callAsPublisherUnwrapped()`
- **[File Uploads](#file-uploads)**: Upload files asynchronously with progress tracking
- **[File Downloads](#file-downloads)**: Download files asynchronously with streaming support
- **[Standard Operations](#example)**: All regular API calls return `CompletableFuture<T>` for async execution

</details>

[comp-fut]: https://docs.oracle.com/javase/8/docs/api/java/util/concurrent/CompletableFuture.html
[reactive-streams]: https://www.reactive-streams.org/
<!-- End Asynchronous Support [async-support] -->

<!-- Start Authentication [security] -->
## Authentication

### Per-Client Security Schemes

This SDK supports multiple security scheme combinations globally. You can choose from one of the alternatives through the `security` builder method when initializing the SDK client instance. The selected option will be used by default to authenticate with the API for all operations that support it.

#### UserPassAuth

The `UserPassAuth` alternative relies on the following scheme:

| Name                      | Type | Scheme     |
| ------------------------- | ---- | ---------- |
| `username`<br/>`password` | http | HTTP Basic |

```java
package hello.world;

import java.lang.Exception;
import java.nio.file.Paths;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.operations.PostFileRequest;
import org.openapis.openapi.models.operations.PostFileResponse;
import org.openapis.openapi.models.shared.File;
import org.openapis.openapi.models.shared.Security;
import org.openapis.openapi.models.shared.UserPassAuth;
import org.openapis.openapi.utils.Blob;

public class Application {

    public static void main(String[] args) throws Error, Exception {

        SDK sdk = SDK.builder()
                .security(Security.builder()
                        .userPassAuth(UserPassAuth.builder()
                                .username("<USERNAME>")
                                .password("<PASSWORD>")
                                .build())
                        .build())
                .build();

        PostFileRequest req = PostFileRequest.builder()
                .upload(File.builder()
                        .fileName("example.file")
                        .content(Blob.from(Paths.get("example.file")))
                        .build())
                .build();

        PostFileResponse res = sdk.postFile().request(req).call();
    }
}

```

#### Option2

All of the following schemes must be satisfied to use the `Option2` alternative:

| Name         | Type   | Scheme      |
| ------------ | ------ | ----------- |
| `bearerAuth` | http   | HTTP Bearer |
| `myApiKey`   | apiKey | API key     |

```java
package hello.world;

import java.lang.Exception;
import java.nio.file.Paths;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.operations.PostFileRequest;
import org.openapis.openapi.models.operations.PostFileResponse;
import org.openapis.openapi.models.shared.File;
import org.openapis.openapi.models.shared.Security;
import org.openapis.openapi.models.shared.SecurityOption2;
import org.openapis.openapi.utils.Blob;

public class Application {

    public static void main(String[] args) throws Error, Exception {

        SDK sdk = SDK.builder()
                .security(Security.builder()
                        .option2(SecurityOption2.builder()
                                .bearerAuth("<YOUR_JWT>")
                                .myApiKey("<value>")
                                .build())
                        .build())
                .build();

        PostFileRequest req = PostFileRequest.builder()
                .upload(File.builder()
                        .fileName("example.file")
                        .content(Blob.from(Paths.get("example.file")))
                        .build())
                .build();

        PostFileResponse res = sdk.postFile().request(req).call();
    }
}

```

#### Option3

The `Option3` alternative relies on the following scheme:

| Name     | Type   | Scheme       |
| -------- | ------ | ------------ |
| `oauth2` | oauth2 | OAuth2 token |

```java
package hello.world;

import java.lang.Exception;
import java.nio.file.Paths;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.operations.PostFileRequest;
import org.openapis.openapi.models.operations.PostFileResponse;
import org.openapis.openapi.models.shared.File;
import org.openapis.openapi.models.shared.Security;
import org.openapis.openapi.models.shared.SecurityOption3;
import org.openapis.openapi.utils.Blob;

public class Application {

    public static void main(String[] args) throws Error, Exception {

        SDK sdk = SDK.builder()
                .security(Security.builder()
                        .option3(SecurityOption3.builder()
                                .oauth2("Bearer <YOUR_OAUTH2_TOKEN>")
                                .build())
                        .build())
                .build();

        PostFileRequest req = PostFileRequest.builder()
                .upload(File.builder()
                        .fileName("example.file")
                        .content(Blob.from(Paths.get("example.file")))
                        .build())
                .build();

        PostFileResponse res = sdk.postFile().request(req).call();
    }
}

```

#### Option4

The `Option4` alternative relies on the following scheme:

| Name                 | Type | Scheme      |
| -------------------- | ---- | ----------- |
| `appId`<br/>`secret` | http | Custom HTTP |

```java
package hello.world;

import java.lang.Exception;
import java.nio.file.Paths;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.operations.PostFileRequest;
import org.openapis.openapi.models.operations.PostFileResponse;
import org.openapis.openapi.models.shared.File;
import org.openapis.openapi.models.shared.Security;
import org.openapis.openapi.models.shared.SecurityOption4;
import org.openapis.openapi.utils.Blob;

public class Application {

    public static void main(String[] args) throws Error, Exception {

        SDK sdk = SDK.builder()
                .security(Security.builder()
                        .option4(SecurityOption4.builder()
                                .appId("app-speakeasy-123")
                                .secret("MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI")
                                .build())
                        .build())
                .build();

        PostFileRequest req = PostFileRequest.builder()
                .upload(File.builder()
                        .fileName("example.file")
                        .content(Blob.from(Paths.get("example.file")))
                        .build())
                .build();

        PostFileResponse res = sdk.postFile().request(req).call();
    }
}

```

#### Option5

The `Option5` alternative relies on the following scheme:

| Name         | Type   | Scheme       |
| ------------ | ------ | ------------ |
| `mobileAuth` | oauth2 | OAuth2 token |

```java
package hello.world;

import java.lang.Exception;
import java.nio.file.Paths;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.operations.PostFileRequest;
import org.openapis.openapi.models.operations.PostFileResponse;
import org.openapis.openapi.models.shared.File;
import org.openapis.openapi.models.shared.Security;
import org.openapis.openapi.models.shared.SecurityOption5;
import org.openapis.openapi.utils.Blob;

public class Application {

    public static void main(String[] args) throws Error, Exception {

        SDK sdk = SDK.builder()
                .security(Security.builder()
                        .option5(SecurityOption5.builder()
                                .mobileAuth("Bearer <YOUR_OAUTH2_TOKEN>")
                                .build())
                        .build())
                .build();

        PostFileRequest req = PostFileRequest.builder()
                .upload(File.builder()
                        .fileName("example.file")
                        .content(Blob.from(Paths.get("example.file")))
                        .build())
                .build();

        PostFileResponse res = sdk.postFile().request(req).call();
    }
}

```

#### Option6

The `Option6` alternative relies on the following scheme:

| Name                                         | Type   | Scheme                         |
| -------------------------------------------- | ------ | ------------------------------ |
| `clientID`<br/>`clientSecret`<br/>`tokenURL` | oauth2 | OAuth2 Client Credentials Flow |

```java
package hello.world;

import java.lang.Exception;
import java.nio.file.Paths;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.operations.PostFileRequest;
import org.openapis.openapi.models.operations.PostFileResponse;
import org.openapis.openapi.models.shared.File;
import org.openapis.openapi.models.shared.Security;
import org.openapis.openapi.models.shared.SecurityOption6;
import org.openapis.openapi.utils.Blob;

public class Application {

    public static void main(String[] args) throws Error, Exception {

        SDK sdk = SDK.builder()
                .security(Security.builder()
                        .option6(SecurityOption6.builder()
                                .clientID("<id>")
                                .clientSecret("<value>")
                                .tokenURL("/clientcredentials/token")
                                .build())
                        .build())
                .build();

        PostFileRequest req = PostFileRequest.builder()
                .upload(File.builder()
                        .fileName("example.file")
                        .content(Blob.from(Paths.get("example.file")))
                        .build())
                .build();

        PostFileResponse res = sdk.postFile().request(req).call();
    }
}

```

#### MyApiKey

The `MyApiKey` alternative relies on the following scheme:

| Name       | Type   | Scheme  |
| ---------- | ------ | ------- |
| `myApiKey` | apiKey | API key |

```java
package hello.world;

import java.lang.Exception;
import java.nio.file.Paths;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.operations.PostFileRequest;
import org.openapis.openapi.models.operations.PostFileResponse;
import org.openapis.openapi.models.shared.File;
import org.openapis.openapi.models.shared.MyApiKey;
import org.openapis.openapi.models.shared.Security;
import org.openapis.openapi.utils.Blob;

public class Application {

    public static void main(String[] args) throws Error, Exception {

        SDK sdk = SDK.builder()
                .security(Security.builder()
                        .myApiKey(MyApiKey.builder().myApiKey("<value>").build())
                        .build())
                .build();

        PostFileRequest req = PostFileRequest.builder()
                .upload(File.builder()
                        .fileName("example.file")
                        .content(Blob.from(Paths.get("example.file")))
                        .build())
                .build();

        PostFileResponse res = sdk.postFile().request(req).call();
    }
}

```

### Per-Operation Security Schemes

Some operations in this SDK require the security scheme to be specified at the request level. For example:
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.AuthResponse;
import org.openapis.openapi.models.operations.AuthSecurity;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder().build();

        AuthResponse res = sdk.tag1()
                .auth()
                .security(AuthSecurity.builder()
                        .accessToken(System.getenv().getOrDefault("ACCESS_TOKEN", ""))
                        .build())
                .call();

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

* [operationWithLeadingAndTrailingUnderscores](docs/sdks/sdk/README.md#operationwithleadingandtrailingunderscores)
* [postFile](docs/sdks/sdk/README.md#postfile) - Post File
* [getPolymorphism](docs/sdks/sdk/README.md#getpolymorphism)
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

<!-- Start Global Parameters [global-parameters] -->
## Global Parameters

Certain parameters are configured globally. These parameters may be set on the SDK client instance itself during initialization. When configured as an option during SDK initialization, These global values will be used as defaults on the operations that use them. When such operations are called, there is a place in each to override the global value, if needed.

For example, you can set `queryParam1` to `"some example query param"` at SDK initialization and then you do not have to pass the same value on calls to operations like `getRequestBodyFlattenedAway`. But if you want to do so you may, which will locally override the global setting. See the example code below for a demonstration.


### Available Globals

The following global parameters are available.

| Name                  | Type             | Description                                                                        |
| --------------------- | ---------------- | ---------------------------------------------------------------------------------- |
| queryParam1           | java.lang.String | A long winded, multi-line description<br/>for the query parameter number one.<br/> |
| deprecatedQueryParam1 | java.lang.String | A deprecated description                                                           |
| deprecatedQueryParam2 | java.lang.String | The deprecatedQueryParam2 parameter.                                               |
| loneQueryParam        | java.lang.String | The loneQueryParam parameter.                                                      |

### Example

```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.GetRequestBodyFlattenedAwayResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
                .loneQueryParam("<value>")
                .queryParam1("some example query param")
                .deprecatedQueryParam1("some example query param")
                .deprecatedQueryParam2("some example query param")
                .build();

        GetRequestBodyFlattenedAwayResponse res =
                sdk.getRequestBodyFlattenedAway().call();

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

```java
package hello.world;

import java.lang.Exception;
import java.util.stream.Stream;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.ChatRequest;
import org.openapis.openapi.models.operations.ChatResponse;
import org.openapis.openapi.models.operations.ChatStream;
import org.openapis.openapi.models.shared.ChatModelRequest;
import org.openapis.openapi.utils.EventStream;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder().build();

        ChatRequest req = ChatRequest.of(
                ChatModelRequest.builder()
                        .prompt("What is the largest city in the world?")
                        .stream(false)
                        .build());

        ChatResponse res = sdk.chat().request(req).call();

        // handle event stream, must be closed after use!
        try (EventStream<ChatStream> events = res.events()) {
            // Option 1: Use for-each loop
            for (ChatStream event : events) {
                System.out.println(event);
            }

            // Option 2: Use Stream API
            try (Stream<ChatStream> stream = events.stream()) {
                stream.forEach(System.out::println);
            }
        }
    }
}

```
#### Reactive Streams Interoperability
An asynchronous SDK client is also available for event streaming that returns a [`Flow.Publisher<T>`][flow-pub]. See [Asynchronous Support](#asynchronous-support) for more details on async benefits and reactive library integration.
```java
package hello.world;

import org.openapis.openapi.AsyncSDK;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.ChatRequest;
import org.openapis.openapi.models.operations.ChatStream;
import org.openapis.openapi.models.operations.async.ChatResponse;
import org.openapis.openapi.models.shared.ChatModelRequest;
import org.openapis.openapi.utils.reactive.EventStream;
import reactor.core.publisher.Flux;

public class Application {

    public static void main(String[] args) {

        AsyncSDK sdk = SDK.builder().build().toAsync();

        ChatRequest req = ChatRequest.of(
                ChatModelRequest.builder()
                        .prompt("What is the largest city in the world?")
                        .stream(false)
                        .build());

        EventStream<ChatResponse, ChatStream> res = sdk.chat().request(req).call();

        // handle async event stream using Reactive Streams Publisher
        // EventStream is a Reactive Streams Publisher, providing broad compatibility with reactive libraries

        // Example using Project Reactor (illustrative)
        Flux<ChatStream> flux = Flux.from(res);
        flux.subscribe(
                event -> System.out.println(event),
                error -> error.printStackTrace(),
                () -> System.out.println("Event stream completed"));
    }
}

```

[mdn-sse]: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events
[flow-pub]: https://docs.oracle.com/javase/9/docs/api/java/util/concurrent/Flow.Publisher.html
<!-- End Server-sent event streaming [eventstream] -->

<!-- Start Pagination [pagination] -->
## Pagination

Some of the endpoints in this SDK support pagination. To use pagination, you can make your SDK calls using the `callAsIterable` or `callAsStream` methods.
For certain operations, you can also use the `callAsStreamUnwrapped` method that streams individual page items directly.

Here's an example depicting the different ways to use pagination:

```java
package hello.world;

import java.lang.Exception;
import java.lang.Iterable;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.BadRequestResponseException;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.operations.ListTest1Response;
import org.openapis.openapi.models.operations.QueryParam2;
import org.openapis.openapi.models.operations.ResultArray;
import org.openapis.openapi.models.shared.MyApiKey;
import org.openapis.openapi.models.shared.Security;

public class Application {

    public static void main(String[] args) throws BadRequestResponseException, Error, Exception {

        SDK sdk = SDK.builder()
                .queryParam1("some example query param")
                .security(Security.builder()
                        .myApiKey(MyApiKey.builder().myApiKey("<value>").build())
                        .build())
                .build();

        var b = sdk.tag1()
                .listTest1()
                .page(100L)
                .queryParam2(QueryParam2.ONE)
                .headerParam1("some example header param");

        // Iterate through all pages using a traditional for-each loop
        // Each iteration returns a complete page response
        Iterable<ListTest1Response> iterable = b.callAsIterable();
        for (ListTest1Response page : iterable) {
            // handle page
        }

        // Stream through all pages and process individual items
        // callAsStreamUnwrapped() flattens pages into individual items
        b.callAsStreamUnwrapped().forEach((ResultArray item) -> {
            // handle item
        });

        // Stream through pages without unwrapping (each item is a complete page)
        b.callAsStream().forEach((ListTest1Response page) -> {
            // handle page
        });
    }
}

```
#### Asynchronous Pagination
An asynchronous SDK client is also available for pagination that returns a [`Flow.Publisher<T>`][flow-pub]. For async pagination, you can use `callAsPublisher()` to get pages as a publisher, or `callAsPublisherUnwrapped()` to get individual items directly. See [Asynchronous Support](#asynchronous-support) for more details on async benefits and reactive library integration.
```java
package hello.world;

import org.openapis.openapi.AsyncSDK;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.QueryParam2;
import org.openapis.openapi.models.operations.async.ListTest1Response;
import org.openapis.openapi.models.operations.async.ResultArray;
import org.openapis.openapi.models.shared.MyApiKey;
import org.openapis.openapi.models.shared.Security;
import reactor.core.publisher.Flux;

public class Application {

    public static void main(String[] args) {

        AsyncSDK sdk = SDK.builder()
                .queryParam1("some example query param")
                .security(Security.builder()
                        .myApiKey(MyApiKey.builder().myApiKey("<value>").build())
                        .build())
                .build()
                .toAsync();

        var b = sdk.tag1()
                .listTest1()
                .page(100L)
                .queryParam2(QueryParam2.ONE)
                .headerParam1("some example header param");

        // Example using Project Reactor (illustrative) - pages
        Flux<ListTest1Response> pageFlux = Flux.from(b.callAsPublisher());
        pageFlux.subscribe(
                page -> System.out.println(page),
                error -> error.printStackTrace(),
                () -> System.out.println("Pagination completed"));
        // Example using Project Reactor (illustrative) - individual items
        Flux<ResultArray> itemFlux = Flux.from(b.callAsPublisherUnwrapped());
        itemFlux.subscribe(
                item -> System.out.println(item),
                error -> error.printStackTrace(),
                () -> System.out.println("Items completed"));
    }
}

```

[flow-pub]: https://docs.oracle.com/javase/9/docs/api/java/util/concurrent/Flow.Publisher.html
<!-- End Pagination [pagination] -->

<!-- Start File uploads [file-upload] -->
## File uploads

Certain SDK methods accept file objects as part of a request body or multi-part request. It is possible and typically recommended to upload files as a stream rather than reading the entire contents into memory. This avoids excessive memory consumption and potentially crashing with out-of-memory errors when working with very large files.

The SDK provides a [`Blob`](src/main/java/org/openapis/openapi/utils/Blob.java) utility class for efficient file handling. It supports various input sources including file paths, streams, strings, and byte arrays, while providing memory-efficient streaming and reactive processing.

```java
// Recommended for large files - streams data efficiently
Blob fileBlob = Blob.from(Paths.get("large-document.pdf"));

// For in-memory data
Blob textBlob = Blob.from("Hello, World!");
Blob dataBlob = Blob.from(myByteArray);
```

> [!TIP]
> For comprehensive documentation including all factory methods, consumption patterns, and advanced usage examples, see the [Blob Utility Documentation](docs/utils/Blob.md).

The following example demonstrates how to attach a file to a request:
```java
package hello.world;

import java.lang.Exception;
import java.nio.file.Paths;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.operations.File;
import org.openapis.openapi.models.operations.PostFileWithEncodingRequest;
import org.openapis.openapi.models.operations.PostFileWithEncodingResponse;
import org.openapis.openapi.utils.Blob;

public class Application {

    public static void main(String[] args) throws Error, Exception {

        SDK sdk = SDK.builder().build();

        PostFileWithEncodingRequest req = PostFileWithEncodingRequest.builder()
                .file(File.builder()
                        .fileName("example.file")
                        .content(Blob.from(Paths.get("example.file")))
                        .build())
                .build();

        PostFileWithEncodingResponse res =
                sdk.tag1().postFileWithEncoding().request(req).call();
    }
}

```
<!-- End File uploads [file-upload] -->

<!-- Start Retries [retries] -->
## Retries

Some of the endpoints in this SDK support retries. If you use the SDK without any configuration, it will fall back to the default retry strategy provided by the API. However, the default retry strategy can be overridden on a per-operation basis, or across the entire SDK.

To change the default retry strategy for a single API call, you can provide a `RetryConfig` object through the `retryConfig` builder method:
```java
package hello.world;

import java.lang.Exception;
import java.math.BigDecimal;
import java.math.BigInteger;
import java.time.LocalDate;
import java.time.OffsetDateTime;
import java.util.List;
import java.util.Map;
import java.util.concurrent.TimeUnit;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.BadRequestResponseException;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.errors.Test2ResponseException;
import org.openapis.openapi.models.operations.PostTest2Response;
import org.openapis.openapi.models.shared.Any;
import org.openapis.openapi.models.shared.Color;
import org.openapis.openapi.models.shared.ExhaustiveObject;
import org.openapis.openapi.models.shared.HeroWidth;
import org.openapis.openapi.models.shared.Icon;
import org.openapis.openapi.models.shared.Int32Enum;
import org.openapis.openapi.models.shared.IntEnum;
import org.openapis.openapi.models.shared.MyApiKey;
import org.openapis.openapi.models.shared.NullableIntEnum;
import org.openapis.openapi.models.shared.NullableStringEnum;
import org.openapis.openapi.models.shared.Security;
import org.openapis.openapi.models.shared.SimpleObject;
import org.openapis.openapi.models.shared.Test2Request;
import org.openapis.openapi.models.shared.Type;
import org.openapis.openapi.utils.BackoffStrategy;
import org.openapis.openapi.utils.RetryConfig;

public class Application {

    public static void main(String[] args)
            throws BadRequestResponseException, Error, Test2ResponseException, Exception {

        SDK sdk = SDK.builder()
                .deprecatedQueryParam1("some example query param")
                .deprecatedQueryParam2("some example query param")
                .security(Security.builder()
                        .myApiKey(MyApiKey.builder().myApiKey("<value>").build())
                        .build())
                .build();

        PostTest2Response res = sdk.testGroup()
                .tag2()
                .postTest()
                .retryConfig(RetryConfig.builder()
                        .backoff(BackoffStrategy.builder()
                                .initialInterval(1L, TimeUnit.MILLISECONDS)
                                .maxInterval(50L, TimeUnit.MILLISECONDS)
                                .maxElapsedTime(1000L, TimeUnit.MILLISECONDS)
                                .baseFactor(1.1)
                                .jitterFactor(0.15)
                                .retryConnectError(false)
                                .build())
                        .build())
                .test2Request(Test2Request.builder()
                        .obj(ExhaustiveObject.builder()
                                .str("example")
                                .bool(true)
                                .integer(999999L)
                                .int32(1)
                                .num(1.1)
                                .float32(8499.3f)
                                .date(LocalDate.parse("2020-01-01"))
                                .dateTime(OffsetDateTime.parse("2020-01-01T00:00:00Z"))
                                .anything("<value>")
                                .int32Enum(Int32Enum.SIXTY_NINE)
                                .bigint(new BigInteger("593288"))
                                .decimalStr(new BigDecimal("7028.3"))
                                .obj(SimpleObject.builder().str("example").build())
                                .map(Map.ofEntries())
                                .arr(List.of())
                                .any(Any.of("<value>"))
                                .nullableStringEnum(NullableStringEnum.SECOND)
                                .icon(Icon.TICK)
                                .boolOpt(true)
                                .intOptNull(999999L)
                                .numOptNull(1.1)
                                .intEnum(IntEnum.Third)
                                .nullableIntEnum(NullableIntEnum.Third)
                                .color(Color.GREEN)
                                .heroWidth(HeroWidth.FOUR_HUNDRED_AND_EIGHTY)
                                .build())
                        .type(Type.SuperType1)
                        .build())
                .call();

        if (res.body().isPresent()) {
            System.out.println(res.body().get());
        }
    }
}

```

If you'd like to override the default retry strategy for all operations that support retries, you can provide a configuration at SDK initialization:
```java
package hello.world;

import java.lang.Exception;
import java.math.BigDecimal;
import java.math.BigInteger;
import java.time.LocalDate;
import java.time.OffsetDateTime;
import java.util.List;
import java.util.Map;
import java.util.concurrent.TimeUnit;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.BadRequestResponseException;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.errors.Test2ResponseException;
import org.openapis.openapi.models.operations.PostTest2Response;
import org.openapis.openapi.models.shared.Any;
import org.openapis.openapi.models.shared.Color;
import org.openapis.openapi.models.shared.ExhaustiveObject;
import org.openapis.openapi.models.shared.HeroWidth;
import org.openapis.openapi.models.shared.Icon;
import org.openapis.openapi.models.shared.Int32Enum;
import org.openapis.openapi.models.shared.IntEnum;
import org.openapis.openapi.models.shared.MyApiKey;
import org.openapis.openapi.models.shared.NullableIntEnum;
import org.openapis.openapi.models.shared.NullableStringEnum;
import org.openapis.openapi.models.shared.Security;
import org.openapis.openapi.models.shared.SimpleObject;
import org.openapis.openapi.models.shared.Test2Request;
import org.openapis.openapi.models.shared.Type;
import org.openapis.openapi.utils.BackoffStrategy;
import org.openapis.openapi.utils.RetryConfig;

public class Application {

    public static void main(String[] args)
            throws BadRequestResponseException, Error, Test2ResponseException, Exception {

        SDK sdk = SDK.builder()
                .retryConfig(RetryConfig.builder()
                        .backoff(BackoffStrategy.builder()
                                .initialInterval(1L, TimeUnit.MILLISECONDS)
                                .maxInterval(50L, TimeUnit.MILLISECONDS)
                                .maxElapsedTime(1000L, TimeUnit.MILLISECONDS)
                                .baseFactor(1.1)
                                .jitterFactor(0.15)
                                .retryConnectError(false)
                                .build())
                        .build())
                .deprecatedQueryParam1("some example query param")
                .deprecatedQueryParam2("some example query param")
                .security(Security.builder()
                        .myApiKey(MyApiKey.builder().myApiKey("<value>").build())
                        .build())
                .build();

        PostTest2Response res = sdk.testGroup()
                .tag2()
                .postTest()
                .test2Request(Test2Request.builder()
                        .obj(ExhaustiveObject.builder()
                                .str("example")
                                .bool(true)
                                .integer(999999L)
                                .int32(1)
                                .num(1.1)
                                .float32(8499.3f)
                                .date(LocalDate.parse("2020-01-01"))
                                .dateTime(OffsetDateTime.parse("2020-01-01T00:00:00Z"))
                                .anything("<value>")
                                .int32Enum(Int32Enum.SIXTY_NINE)
                                .bigint(new BigInteger("593288"))
                                .decimalStr(new BigDecimal("7028.3"))
                                .obj(SimpleObject.builder().str("example").build())
                                .map(Map.ofEntries())
                                .arr(List.of())
                                .any(Any.of("<value>"))
                                .nullableStringEnum(NullableStringEnum.SECOND)
                                .icon(Icon.TICK)
                                .boolOpt(true)
                                .intOptNull(999999L)
                                .numOptNull(1.1)
                                .intEnum(IntEnum.Third)
                                .nullableIntEnum(NullableIntEnum.Third)
                                .color(Color.GREEN)
                                .heroWidth(HeroWidth.FOUR_HUNDRED_AND_EIGHTY)
                                .build())
                        .type(Type.SuperType1)
                        .build())
                .call();

        if (res.body().isPresent()) {
            System.out.println(res.body().get());
        }
    }
}

```
<!-- End Retries [retries] -->

<!-- Start Error Handling [errors] -->
## Error Handling

Handling errors in this SDK should largely match your expectations. All operations return a response object or raise an exception.


[`SDKBaseException`](./src/main/java/models/errors/SDKBaseException.java) is the base class for all HTTP error responses. It has the following properties:

| Method           | Type                        | Description                                                              |
| ---------------- | --------------------------- | ------------------------------------------------------------------------ |
| `message()`      | `String`                    | Error message                                                            |
| `code()`         | `int`                       | HTTP response status code eg `404`                                       |
| `headers`        | `Map<String, List<String>>` | HTTP response headers                                                    |
| `body()`         | `byte[]`                    | HTTP body as a byte array. Can be empty array if no body is returned.    |
| `bodyAsString()` | `String`                    | HTTP body as a UTF-8 string. Can be empty string if no body is returned. |
| `rawResponse()`  | `HttpResponse<?>`           | Raw HTTP response (body already read and not available for re-read)      |

### Example
```java
package hello.world;

import java.io.UncheckedIOException;
import java.lang.Exception;
import java.lang.String;
import java.nio.file.Paths;
import java.util.Optional;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.errors.SDKBaseException;
import org.openapis.openapi.models.operations.PostFileRequest;
import org.openapis.openapi.models.operations.PostFileResponse;
import org.openapis.openapi.models.shared.File;
import org.openapis.openapi.utils.Blob;

public class Application {

    public static void main(String[] args) throws Error, Exception {

        SDK sdk = SDK.builder().build();
        try {

            PostFileRequest req = PostFileRequest.builder()
                    .upload(File.builder()
                            .fileName("example.file")
                            .content(Blob.from(Paths.get("example.file")))
                            .build())
                    .build();

            PostFileResponse res = sdk.postFile().request(req).call();

        } catch (SDKBaseException ex) { // all SDK exceptions inherit from SDKBaseException

            // ex.ToString() provides a detailed error message including
            // HTTP status code, headers, and error payload (if any)
            System.out.println(ex);

            // Base exception fields
            var rawResponse = ex.rawResponse();
            var headers = ex.headers();
            var contentType = headers.first("Content-Type");
            int statusCode = ex.code();
            Optional<byte[]> responseBody = ex.body();

            // different error subclasses may be thrown
            // depending on the service call
            if (ex instanceof Error) {
                var e = (Error) ex;
                // Check error data fields
                e.data().ifPresent(payload -> {
                    String error = payload.error();
                    long code = payload.code();
                });
            }

            // An underlying cause may be provided. If the error payload
            // cannot be deserialized then the deserialization exception
            // will be set as the cause.
            if (ex.getCause() != null) {
                var cause = ex.getCause();
            }
        } catch (UncheckedIOException ex) {
            // handle IO error (connection, timeout, etc)
        }
    }
}

```

### Error Classes
**Primary error:**
* [`SDKBaseException`](./src/main/java/models/errors/SDKBaseException.java): The base class for HTTP error responses.

<details><summary>Less common errors (11)</summary>

<br />

**Network errors:**
* `java.io.IOException` (always wrapped by `java.io.UncheckedIOException`). Commonly encountered subclasses of
`IOException` include `java.net.ConnectException`, `java.net.SocketTimeoutException`, `EOFException` (there are
many more subclasses in the JDK platform).

**Inherit from [`SDKBaseException`](./src/main/java/models/errors/SDKBaseException.java)**:
* [`org.openapis.openapi.models.errors.Error`](./src/main/java/models/errors/org.openapis.openapi.models.errors.Error.java): A not-so-long multi-line error model description. Applicable to 5 of 48 methods.*
* [`org.openapis.openapi.models.errors.BadRequestResponseException`](./src/main/java/models/errors/org.openapis.openapi.models.errors.BadRequestResponseException.java): Bad Request. Status code `400`. Applicable to 2 of 48 methods.*
* [`org.openapis.openapi.models.errors.RequestTimeoutError`](./src/main/java/models/errors/org.openapis.openapi.models.errors.RequestTimeoutError.java): A spec-defined error that collides with the built-in RequestTimeoutError in httpclienterrors.ts. Status code `408`. Applicable to 1 of 48 methods.*
* [`org.openapis.openapi.models.errors.FailedResponseException`](./src/main/java/models/errors/org.openapis.openapi.models.errors.FailedResponseException.java): An error response with an empty object schema. Status code `500`. Applicable to 1 of 48 methods.*
* [`org.openapis.openapi.models.errors.Test2ResponseException`](./src/main/java/models/errors/org.openapis.openapi.models.errors.Test2ResponseException.java): Internal Server Error. Status code `500`. Applicable to 1 of 48 methods.*


</details>

\* Check [the method documentation](#available-resources-and-operations) to see if the error is applicable.
<!-- End Error Handling [errors] -->

<!-- Start Server Selection [server] -->
## Server Selection

### Select Server by Index

You can override the default server globally using the `.serverIndex(int serverIdx)` builder method when initializing the SDK client instance. The selected server will then be used as the default on the operations that use it. This table lists the indexes associated with the available servers:

| #   | Server                                     | Variables                 | Description                     |
| --- | ------------------------------------------ | ------------------------- | ------------------------------- |
| 0   | `http://localhost:35123`                   |                           | The default server.             |
| 1   | `http://{subdomain}.domain.com/v{version}` | `subdomain`<br/>`version` |                                 |
| 2   | `http://{HostName}:{PORT}`                 | `HostName`<br/>`PORT`     | A server with an enum variable. |

If the selected server has variables, you may override its default values using the associated builder method(s):

| Variable    | BuilderMethod                 | Supported Values                      | Default       | Description                              |
| ----------- | ----------------------------- | ------------------------------------- | ------------- | ---------------------------------------- |
| `subdomain` | `subdomain(String subdomain)` | java.lang.String                      | `"api"`       |                                          |
| `version`   | `version(String version)`     | java.lang.String                      | `"1"`         |                                          |
| `HostName`  | `hostName(String hostName)`   | java.lang.String                      | `"localhost"` | The hostname of the server.              |
| `PORT`      | `port(ServerPORT port)`       | - `"80"`<br/>- `"8080"`<br/>- `"443"` | `"8080"`      | The port on which the server is running. |

#### Example

```java
package hello.world;

import java.lang.Exception;
import java.nio.file.Paths;
import org.openapis.openapi.SDK;
import org.openapis.openapi.SDK.Builder.ServerPORT;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.operations.PostFileRequest;
import org.openapis.openapi.models.operations.PostFileResponse;
import org.openapis.openapi.models.shared.File;
import org.openapis.openapi.utils.Blob;

public class Application {

    public static void main(String[] args) throws Error, Exception {

        SDK sdk = SDK.builder()
                .serverIndex(2)
                .hostName("heavy-bowler.org")
                .port(ServerPORT.FOUR_HUNDRED_AND_FORTY_THREE)
                .build();

        PostFileRequest req = PostFileRequest.builder()
                .upload(File.builder()
                        .fileName("example.file")
                        .content(Blob.from(Paths.get("example.file")))
                        .build())
                .build();

        PostFileResponse res = sdk.postFile().request(req).call();
    }
}

```

### Override Server URL Per-Client

The default server can also be overridden globally using the `.serverURL(String serverUrl)` builder method when initializing the SDK client instance. For example:
```java
package hello.world;

import java.lang.Exception;
import java.nio.file.Paths;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.operations.PostFileRequest;
import org.openapis.openapi.models.operations.PostFileResponse;
import org.openapis.openapi.models.shared.File;
import org.openapis.openapi.utils.Blob;

public class Application {

    public static void main(String[] args) throws Error, Exception {

        SDK sdk = SDK.builder().serverURL("http://localhost:8080").build();

        PostFileRequest req = PostFileRequest.builder()
                .upload(File.builder()
                        .fileName("example.file")
                        .content(Blob.from(Paths.get("example.file")))
                        .build())
                .build();

        PostFileResponse res = sdk.postFile().request(req).call();
    }
}

```

### Override Server URL Per-Operation

The server URL can also be overridden on a per-operation basis, provided a server list was specified for the operation. For example:
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.BadRequestResponseException;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.operations.QueryParam2;
import org.openapis.openapi.models.operations.ResultArray;
import org.openapis.openapi.models.shared.MyApiKey;
import org.openapis.openapi.models.shared.Security;

public class Application {

    public static void main(String[] args) throws BadRequestResponseException, Error, Exception {

        SDK sdk = SDK.builder()
                .queryParam1("some example query param")
                .security(Security.builder()
                        .myApiKey(MyApiKey.builder().myApiKey("<value>").build())
                        .build())
                .build();

        sdk.tag1()
                .listTest1()
                .serverURL("http://localhost:35123")
                .page(100L)
                .queryParam2(QueryParam2.ONE)
                .headerParam1("some example header param")
                .callAsStreamUnwrapped()
                .forEach((ResultArray item) -> {
                    // handle item
                });
    }
}

```
<!-- End Server Selection [server] -->

<!-- Start Custom HTTP Client [http-client] -->
## Custom HTTP Client

The Java SDK makes API calls using an `HTTPClient` that wraps the native
[HttpClient](https://docs.oracle.com/en/java/javase/11/docs/api/java.net.http/java/net/http/HttpClient.html). This
client provides the ability to attach hooks around the request lifecycle that can be used to modify the request or handle
errors and response.

The `HTTPClient` interface allows you to either use the default `SpeakeasyHTTPClient` that comes with the SDK,
or provide your own custom implementation with customized configuration such as custom executors, SSL context,
connection pools, and other HTTP client settings.

The interface provides synchronous (`send`) methods and asynchronous (`sendAsync`) methods. The `sendAsync` method
is used to power the async SDK methods and returns a `CompletableFuture<HttpResponse<Blob>>` for non-blocking operations.

The following example shows how to add a custom header and handle errors:

```java
import org.openapis.openapi.SDK;
import org.openapis.openapi.utils.HTTPClient;
import org.openapis.openapi.utils.SpeakeasyHTTPClient;
import org.openapis.openapi.utils.Utils;

import java.io.IOException;
import java.net.URISyntaxException;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.io.InputStream;
import java.time.Duration;

public class Application {
    public static void main(String[] args) {
        // Create a custom HTTP client with hooks
        HTTPClient httpClient = new HTTPClient() {
            private final HTTPClient defaultClient = new SpeakeasyHTTPClient();
            
            @Override
            public HttpResponse<InputStream> send(HttpRequest request) throws IOException, URISyntaxException, InterruptedException {
                // Add custom header and timeout using Utils.copy()
                HttpRequest modifiedRequest = Utils.copy(request)
                    .header("x-custom-header", "custom value")
                    .timeout(Duration.ofSeconds(30))
                    .build();
                    
                try {
                    HttpResponse<InputStream> response = defaultClient.send(modifiedRequest);
                    // Log successful response
                    System.out.println("Request successful: " + response.statusCode());
                    return response;
                } catch (Exception error) {
                    // Log error
                    System.err.println("Request failed: " + error.getMessage());
                    throw error;
                }
            }
        };

        SDK sdk = SDK.builder()
            .client(httpClient)
            .build();
    }
}
```

<details>
<summary>Custom HTTP Client Configuration</summary>

You can also provide a completely custom HTTP client with your own configuration:

```java
import org.openapis.openapi.SDK;
import org.openapis.openapi.utils.HTTPClient;
import org.openapis.openapi.utils.Blob;
import org.openapis.openapi.utils.ResponseWithBody;

import java.io.IOException;
import java.net.URISyntaxException;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.io.InputStream;
import java.time.Duration;
import java.util.concurrent.Executors;
import java.util.concurrent.CompletableFuture;

public class Application {
    public static void main(String[] args) {
        // Custom HTTP client with custom configuration
        HTTPClient customHttpClient = new HTTPClient() {
            private final HttpClient client = HttpClient.newBuilder()
                .executor(Executors.newFixedThreadPool(10))
                .connectTimeout(Duration.ofSeconds(30))
                // .sslContext(customSslContext) // Add custom SSL context if needed
                .build();

            @Override
            public HttpResponse<InputStream> send(HttpRequest request) throws IOException, URISyntaxException, InterruptedException {
                return client.send(request, HttpResponse.BodyHandlers.ofInputStream());
            }

            @Override
            public CompletableFuture<HttpResponse<Blob>> sendAsync(HttpRequest request) {
                // Convert response to HttpResponse<Blob> for async operations
                return client.sendAsync(request, HttpResponse.BodyHandlers.ofPublisher())
                    .thenApply(resp -> new ResponseWithBody<>(resp, Blob::from));
            }
        };

        SDK sdk = SDK.builder()
            .client(customHttpClient)
            .build();
    }
}
```

</details>

You can also enable debug logging on the default `SpeakeasyHTTPClient`:

```java
import org.openapis.openapi.SDK;
import org.openapis.openapi.utils.SpeakeasyHTTPClient;

public class Application {
    public static void main(String[] args) {
        SpeakeasyHTTPClient httpClient = new SpeakeasyHTTPClient();
        httpClient.enableDebugLogging(true);

        SDK sdk = SDK.builder()
            .client(httpClient)
            .build();
    }
}
```
<!-- End Custom HTTP Client [http-client] -->

<!-- Start Debugging [debug] -->
## Debugging

### Debug & Logging

#### SLF4j Logging
This SDK uses [SLF4j](https://www.slf4j.org/) for structured logging across HTTP requests, retries, pagination, streaming, and hooks. SLF4j provides comprehensive visibility into SDK operations.

**Log Levels:**
- **DEBUG**: High-level operations (HTTP requests/responses, retry attempts, page fetches, hook execution, stream lifecycle)
- **TRACE**: Detailed information (request/response bodies, backoff calculations, individual items processed)

**Configuration:**

Add your preferred SLF4j implementation to your project. For example, using Logback:

```gradle
dependencies {
    implementation 'ch.qos.logback:logback-classic:1.4.14'
}
```

Configure logging levels in your `logback.xml`:

```xml
<configuration>
    <appender name="STDOUT" class="ch.qos.logback.core.ConsoleAppender">
        <encoder>
            <pattern>%d{HH:mm:ss.SSS} [%thread] %-5level %logger{36} - %msg%n</pattern>
        </encoder>
    </appender>

    <!-- SDK-wide logging -->
    <logger name="org.openapis.openapi" level="DEBUG"/>
    
    <!-- Component-specific logging -->
    <logger name="org.openapis.openapi.utils.SpeakeasyHTTPClient" level="DEBUG"/>
    <logger name="org.openapis.openapi.utils.Retries" level="DEBUG"/>
    <logger name="org.openapis.openapi.utils.pagination" level="DEBUG"/>
    <logger name="org.openapis.openapi.utils.Hooks" level="TRACE"/>
    
    <root level="INFO">
        <appender-ref ref="STDOUT"/>
    </root>
</configuration>
```

**What Gets Logged:**
- **HTTP Client**: Request/response details, headers (with sensitive headers redacted), bodies (at TRACE level)
- **Retries**: Retry attempts, backoff delays, exhaustion, non-retryable exceptions
- **Pagination**: Page fetches, pagination state, errors
- **Streaming**: Stream initialization, item processing, closure
- **Hooks**: Hook execution counts, operation IDs, exceptions

#### Legacy Debug Logging
For backward compatibility, you can still use the legacy debug logging method:

```java
SDK.builder()
    .enableHTTPDebugLogging(true)
    .build();
```
Example output:
```
Sending request: http://localhost:35123/bearer#global GET
Request headers: {Accept=[application/json], Authorization=[******], Client-Level-Header=[added by client], Idempotency-Key=[some-key], x-speakeasy-user-agent=[speakeasy-sdk/java 0.0.1 internal 0.1.0 org.openapis.openapi]}
Received response: (GET http://localhost:35123/bearer#global) 200
Response headers: {access-control-allow-credentials=[true], access-control-allow-origin=[*], connection=[keep-alive], content-length=[50], content-type=[application/json], date=[Wed, 09 Apr 2025 01:43:29 GMT], server=[gunicorn/19.9.0]}
Response body:
{
  "authenticated": true, 
  "token": "global"
}
```
__WARNING__: Debug logging should only be used for temporary debugging purposes. Leaving this option on in a production system could expose credentials/secrets in logs. <i>Authorization</i> headers are redacted by default. You can specify additional redacted header names via `SpeakeasyHTTPClient.setRedactedHeaders`.

__NOTE__: This is a convenience method that calls `HTTPClient.enableDebugLogging()`. The `SpeakeasyHTTPClient` honors this setting. If you are using a custom HTTP client, it is up to the custom client to honor this setting.


#### JDK HTTP Client Logging
Another option is to set the System property `-Djdk.httpclient.HttpClient.log=all`. However, this option does not log request/response bodies.
<!-- End Debugging [debug] -->

<!-- Start Jackson Configuration [jackson] -->
## Jackson Configuration

The SDK ships with a pre-configured Jackson [`ObjectMapper`][jackson-databind] accessible via
`JSON.getMapper()`. It is set up with type modules, strict deserializers, and the feature flags
needed for full SDK compatibility (including ISO-8601 `OffsetDateTime` serialization):

```java
import org.openapis.openapi.utils.JSON;

String json = JSON.getMapper().writeValueAsString(response);
```

To compose with your own `ObjectMapper`, register the provided `OpenapiJacksonModule`, which
bundles all the same modules and feature flags as a single plug-and-play module:

```java
import org.openapis.openapi.utils.OpenapiJacksonModule;
import com.fasterxml.jackson.databind.ObjectMapper;

ObjectMapper myMapper = new ObjectMapper()
    .registerModule(new OpenapiJacksonModule());

String json = myMapper.writeValueAsString(response);
```

[jackson-databind]: https://github.com/FasterXML/jackson-databind
[jackson-jsr310]: https://github.com/FasterXML/jackson-modules-java8/tree/master/datetime
<!-- End Jackson Configuration [jackson] -->

<!-- Placeholder for Future Speakeasy SDK Sections -->

# Development

## Maturity

This SDK is in beta, and there may be breaking changes between versions without a major version update. Therefore, we recommend pinning usage
to a specific package version. This way, you can install the same version each time without breaking changes unless you are intentionally
looking for the latest version.

## Contributions

While we value open-source contributions to this SDK, this library is generated programmatically. Any manual changes added to internal files will be overwritten on the next generation. 
We look forward to hearing your feedback. Feel free to open a PR or an issue with a proof of concept and we'll do our best to include it in a future release. 

### SDK Created by [Speakeasy](https://www.speakeasy.com/?utm_source=openapi&utm_campaign=java)
