# SDK

## Overview

This document will show case as many of our features as possible in as little operations/models as possible.
This will then generate a SDK that we can more easily review than the test SDKs based on uber.yaml spec.

Speakeasy Docs
<https://speakeasy.com/docs>

### Available Operations

* [operationWithLeadingAndTrailingUnderscores](#operationwithleadingandtrailingunderscores)
* [postFile](#postfile) - Post File
* [getPolymorphism](#getpolymorphism)
* [getRequestBodyFlattenedAway](#getrequestbodyflattenedaway)
* [getFullyFlattenedRequest](#getfullyflattenedrequest)
* [createWithUnion](#createwithunion) - Create with discriminated union request body
* [testEndpoint](#testendpoint)
* [createUser](#createuser) - Create User
* [getUser](#getuser) - Get User
* [updateUser](#updateuser) - Update User
* [deleteUser](#deleteuser) - Delete User
* [login](#login) - Login
* [validate](#validate) - Validate
* [chat](#chat)
* [getBinaryDefaultResponse](#getbinarydefaultresponse)
* [testEnumFormats](#testenumformats) - Test x-speakeasy-enums in different formats
* [binaryAndStringUpload](#binaryandstringupload)
* [getDuplicateExportCollision](#getduplicateexportcollision) - Tests that a spec-defined error type colliding with a built-in SDK error name does not cause TS2308
* [getNamedPrimitiveUnion](#getnamedprimitiveunion) - Test named primitive union options using title and x-speakeasy-name-override
* [getEmptyObjectError](#getemptyobjecterror) - Get Empty Object Error
* [urlValidationStressTest](#urlvalidationstresstest)
* [parenthesesInPathAllowed](#parenthesesinpathallowed) - A string with {{ double braces }} and { single braces }
and \{\{ escaped curlies \}\} and `backticks`.
and \`escaped backticks\` and double slashes\\
and 'single quotes' and "double quotes".
and  \'escaped single quotes\' and \"escaped double quotes\".

* [getNestedIntegerString](#getnestedintegerstring) - Test nested struct with integer:string tag
* [renderAsset](#renderasset) - Render Asset
* [getAsset](#getasset) - Get Asset
* [getErrorOnlyExample](#geterroronlyexample) - Operation with example only on error response

## operationWithLeadingAndTrailingUnderscores

### Example Usage

<!-- UsageSnippet language="java" operationID="_operation_with_leading_and_trailing_underscores_" method="get" path="/test_operation_id_with_underscores" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.OperationWithLeadingAndTrailingUnderscoresResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        OperationWithLeadingAndTrailingUnderscoresResponse res = sdk.operationWithLeadingAndTrailingUnderscores()
                .qp1("renamed")
                .call();

        // handle response
    }
}
```

### Parameters

| Parameter                                                                                                                                        | Type                                                                                                                                             | Required                                                                                                                                         | Description                                                                                                                                      | Example                                                                                                                                          |
| ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `qp1`                                                                                                                                            | *String*                                                                                                                                         | :heavy_check_mark:                                                                                                                               | This parameter will not be filled in with the queryParam1 global because it uses x-speakeasy-name-override which results in a non-matching name. | renamed                                                                                                                                          |

### Response

**[OperationWithLeadingAndTrailingUnderscoresResponse](../../models/operations/OperationWithLeadingAndTrailingUnderscoresResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## postFile

This is a test endpoint.
It has a description.

### Example Usage

<!-- UsageSnippet language="java" operationID="postFile" method="post" path="/file" -->
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

        SDK sdk = SDK.builder()
            .build();

        PostFileRequest req = PostFileRequest.builder()
                .upload(File.builder()
                    .fileName("example.file")
                    .content(Blob.from(Paths.get("example.file")))
                    .build())
                .build();

        PostFileResponse res = sdk.postFile()
                .request(req)
                .call();

    }
}
```

### Parameters

| Parameter                                                     | Type                                                          | Required                                                      | Description                                                   |
| ------------------------------------------------------------- | ------------------------------------------------------------- | ------------------------------------------------------------- | ------------------------------------------------------------- |
| `request`                                                     | [PostFileRequest](../../models/operations/PostFileRequest.md) | :heavy_check_mark:                                            | The request object to use for the request.                    |

### Response

**[PostFileResponse](../../models/operations/PostFileResponse.md)**

### Errors

| Error Type          | Status Code         | Content Type        |
| ------------------- | ------------------- | ------------------- |
| models/errors/Error | 415, 4XX            | application/json    |
| models/errors/Error | 5XX                 | application/json    |

## getPolymorphism

### Example Usage

<!-- UsageSnippet language="java" operationID="getPolymorphism" method="get" path="/polymorphism" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.GetPolymorphismResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        GetPolymorphismResponse res = sdk.getPolymorphism()
                .call();

        if (res.object().isPresent()) {
            System.out.println(res.object().get());
        }
    }
}
```

### Response

**[GetPolymorphismResponse](../../models/operations/GetPolymorphismResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## getRequestBodyFlattenedAway

### Example Usage

<!-- UsageSnippet language="java" operationID="getRequestBodyFlattenedAway" method="get" path="/requestBodyFlattenedAway" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.GetRequestBodyFlattenedAwayResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
                .loneQueryParam("<value>")
            .build();

        GetRequestBodyFlattenedAwayResponse res = sdk.getRequestBodyFlattenedAway()
                .call();

        // handle response
    }
}
```

### Response

**[GetRequestBodyFlattenedAwayResponse](../../models/operations/GetRequestBodyFlattenedAwayResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## getFullyFlattenedRequest

### Example Usage

<!-- UsageSnippet language="java" operationID="getFullyFlattenedRequest" method="post" path="/fullyFlattenedRequest" example="namedExampleThatIsntMatchedAcrossDifferentExamples" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.GetFullyFlattenedRequestRequestBody;
import org.openapis.openapi.models.operations.GetFullyFlattenedRequestResponse;
import org.openapis.openapi.models.shared.Security;
import org.openapis.openapi.models.shared.SecurityOption6;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
                .security(Security.builder()
                    .option6(SecurityOption6.builder()
                        .clientID("<id>")
                        .clientSecret("<value>")
                        .tokenURL("/clientcredentials/token")
                        .build())
                    .build())
            .build();

        GetFullyFlattenedRequestResponse res = sdk.getFullyFlattenedRequest()
                .lang("en")
                .requestBody(GetFullyFlattenedRequestRequestBody.builder()
                    .name("<value>")
                    .build())
                .call();

        // handle response
    }
}
```

### Parameters

| Parameter                                                                                             | Type                                                                                                  | Required                                                                                              | Description                                                                                           |
| ----------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| `lang`                                                                                                | *String*                                                                                              | :heavy_check_mark:                                                                                    | N/A                                                                                                   |
| `maxLength`                                                                                           | @Nullable *long*                                                                                      | :heavy_minus_sign:                                                                                    | N/A                                                                                                   |
| `requestBody`                                                                                         | [GetFullyFlattenedRequestRequestBody](../../models/operations/GetFullyFlattenedRequestRequestBody.md) | :heavy_check_mark:                                                                                    | N/A                                                                                                   |

### Response

**[GetFullyFlattenedRequestResponse](../../models/operations/GetFullyFlattenedRequestResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## createWithUnion

Test CLI generation for discriminated unions with dot-notation flags

### Example Usage

<!-- UsageSnippet language="java" operationID="createWithUnion" method="post" path="/unionRequestBody" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.CreateWithUnionResponse;
import org.openapis.openapi.models.shared.Rectangle;
import org.openapis.openapi.models.shared.ShapeRequest;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        CreateWithUnionResponse res = sdk.createWithUnion()
                .shapeRequest(ShapeRequest.builder()
                    .name("<value>")
                    .shape(Rectangle.builder()
                        .type("rectangle")
                        .width(3125.73)
                        .height(922.51)
                        .build())
                    .build())
                .call();

        if (res.object().isPresent()) {
            System.out.println(res.object().get());
        }
    }
}
```

### Parameters

| Parameter                                           | Type                                                | Required                                            | Description                                         |
| --------------------------------------------------- | --------------------------------------------------- | --------------------------------------------------- | --------------------------------------------------- |
| `dryRun`                                            | @Nullable *boolean*                                 | :heavy_minus_sign:                                  | If true, validates without creating                 |
| `shapeRequest`                                      | [ShapeRequest](../../models/shared/ShapeRequest.md) | :heavy_check_mark:                                  | N/A                                                 |

### Response

**[CreateWithUnionResponse](../../models/operations/CreateWithUnionResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## testEndpoint

### Example Usage

<!-- UsageSnippet language="java" operationID="testEndpoint" method="post" path="/test/endpoint/{testName}" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.TestEndpointRequestBody;
import org.openapis.openapi.models.operations.TestEndpointResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        TestEndpointResponse res = sdk.testEndpoint()
                .testName("<value>")
                .requestBody(TestEndpointRequestBody.builder()
                    .test("<value>")
                    .build())
                .call();

        // handle response
    }
}
```

### Parameters

| Parameter                                                                     | Type                                                                          | Required                                                                      | Description                                                                   |
| ----------------------------------------------------------------------------- | ----------------------------------------------------------------------------- | ----------------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| `testName`                                                                    | *String*                                                                      | :heavy_check_mark:                                                            | N/A                                                                           |
| `requestBody`                                                                 | [TestEndpointRequestBody](../../models/operations/TestEndpointRequestBody.md) | :heavy_check_mark:                                                            | N/A                                                                           |

### Response

**[TestEndpointResponse](../../models/operations/TestEndpointResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## createUser

Creates a new user in the system. Multiple named examples demonstrate
different pairing scenarios for documentation generation.


### Example Usage: paired-example

<!-- UsageSnippet language="java" operationID="createUser" method="put" path="/user" example="paired-example" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.CreateUserResponse;
import org.openapis.openapi.models.shared.BaseUser;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        BaseUser req = BaseUser.builder()
                .email("paired@example.com")
                .firstName("John")
                .build();

        CreateUserResponse res = sdk.createUser()
                .request(req)
                .call();

        if (res.user().isPresent()) {
            System.out.println(res.user().get());
        }
    }
}
```
### Example Usage: request-only

<!-- UsageSnippet language="java" operationID="createUser" method="put" path="/user" example="request-only" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.CreateUserResponse;
import org.openapis.openapi.models.shared.BaseUser;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        BaseUser req = BaseUser.builder()
                .email("request-only@example.com")
                .build();

        CreateUserResponse res = sdk.createUser()
                .request(req)
                .call();

        if (res.user().isPresent()) {
            System.out.println(res.user().get());
        }
    }
}
```
### Example Usage: response-only

<!-- UsageSnippet language="java" operationID="createUser" method="put" path="/user" example="response-only" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.CreateUserResponse;
import org.openapis.openapi.models.shared.BaseUser;
import org.openapis.openapi.models.shared.Gender;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        BaseUser req = BaseUser.builder()
                .email("Virginie47@gmail.com")
                .id("8ffac18c-7d88-4879-b057-e5f45b9ce7de")
                .gender(Gender.OTHER)
                .build();

        CreateUserResponse res = sdk.createUser()
                .request(req)
                .call();

        if (res.user().isPresent()) {
            System.out.println(res.user().get());
        }
    }
}
```

### Parameters

| Parameter                                   | Type                                        | Required                                    | Description                                 |
| ------------------------------------------- | ------------------------------------------- | ------------------------------------------- | ------------------------------------------- |
| `request`                                   | [BaseUser](../../models/shared/BaseUser.md) | :heavy_check_mark:                          | The request object to use for the request.  |

### Response

**[CreateUserResponse](../../models/operations/CreateUserResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## getUser

Get User

### Example Usage

<!-- UsageSnippet language="java" operationID="getUser" method="get" path="/user/{id}" example="success" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.GetUserResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        GetUserResponse res = sdk.getUser()
                .id("<id>")
                .call();

        if (res.user().isPresent()) {
            System.out.println(res.user().get());
        }
    }
}
```

### Parameters

| Parameter          | Type               | Required           | Description        |
| ------------------ | ------------------ | ------------------ | ------------------ |
| `id`               | *String*           | :heavy_check_mark: | N/A                |

### Response

**[GetUserResponse](../../models/operations/GetUserResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## updateUser

Update User

### Example Usage

<!-- UsageSnippet language="java" operationID="updateUser" method="post" path="/user/{id}" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.UpdateUserResponse;
import org.openapis.openapi.models.shared.Gender;
import org.openapis.openapi.models.shared.User;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        UpdateUserResponse res = sdk.updateUser()
                .id("<id>")
                .user(User.builder()
                    .id("8ffac18c-7d88-4879-b057-e5f45b9ce7de")
                    .email("Joanny.Feeney@gmail.com")
                    .gender(Gender.OTHER)
                    .build())
                .call();

        if (res.user().isPresent()) {
            System.out.println(res.user().get());
        }
    }
}
```

### Parameters

| Parameter                           | Type                                | Required                            | Description                         |
| ----------------------------------- | ----------------------------------- | ----------------------------------- | ----------------------------------- |
| `id`                                | *String*                            | :heavy_check_mark:                  | N/A                                 |
| `user`                              | [User](../../models/shared/User.md) | :heavy_check_mark:                  | N/A                                 |

### Response

**[UpdateUserResponse](../../models/operations/UpdateUserResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## deleteUser

Delete User

### Example Usage

<!-- UsageSnippet language="java" operationID="deleteUser" method="delete" path="/user/{id}" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.DeleteUserResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        DeleteUserResponse res = sdk.deleteUser()
                .id("<id>")
                .call();

        // handle response
    }
}
```

### Parameters

| Parameter          | Type               | Required           | Description        |
| ------------------ | ------------------ | ------------------ | ------------------ |
| `id`               | *String*           | :heavy_check_mark: | N/A                |

### Response

**[DeleteUserResponse](../../models/operations/DeleteUserResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## login

Login

### Example Usage

<!-- UsageSnippet language="java" operationID="login" method="get" path="/auth/login" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.LoginResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        LoginResponse res = sdk.login()
                .call();

        if (res.object().isPresent()) {
            System.out.println(res.object().get());
        }
    }
}
```

### Response

**[LoginResponse](../../models/operations/LoginResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## validate

Validate

### Example Usage

<!-- UsageSnippet language="java" operationID="validate" method="get" path="/auth/validate" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.ValidateResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        ValidateResponse res = sdk.validate()
                .call();

        if (res.object().isPresent()) {
            System.out.println(res.object().get());
        }
    }
}
```

### Response

**[ValidateResponse](../../models/operations/ValidateResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## chat

### Example Usage

<!-- UsageSnippet language="java" operationID="chat" method="post" path="/chat" -->
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

        SDK sdk = SDK.builder()
            .build();

        ChatRequest req = ChatRequest.of(ChatModelRequest.builder()
                .prompt("What is the largest city in the world?")
                .stream(false)
                .build());

        ChatResponse res = sdk.chat()
                .request(req)
                .call();

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

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `request`                                             | [ChatRequest](../../models/operations/ChatRequest.md) | :heavy_check_mark:                                    | The request object to use for the request.            |

### Response

**[ChatResponse](../../models/operations/ChatResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## getBinaryDefaultResponse

### Example Usage

<!-- UsageSnippet language="java" operationID="getBinaryDefaultResponse" method="get" path="/binaryDefaultResponse" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.GetBinaryDefaultResponseResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        GetBinaryDefaultResponseResponse res = sdk.getBinaryDefaultResponse()
                .call();

        if (res.bytes().isPresent()) {
            // handle response
        }
    }
}
```

### Response

**[GetBinaryDefaultResponseResponse](../../models/operations/GetBinaryDefaultResponseResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## testEnumFormats

This endpoint tests the x-speakeasy-enums extension in both array and map formats,
including partial map coverage and both string and integer enum types.

### Example Usage

<!-- UsageSnippet language="java" operationID="testEnumFormats" method="post" path="/enumFormats" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.IntegerMapFormat;
import org.openapis.openapi.models.operations.IntegerPartialMapFormat;
import org.openapis.openapi.models.operations.StringArrayFormat;
import org.openapis.openapi.models.operations.StringMapFormat;
import org.openapis.openapi.models.operations.StringPartialMapFormat;
import org.openapis.openapi.models.operations.TestEnumFormatsRequest;
import org.openapis.openapi.models.operations.TestEnumFormatsResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        TestEnumFormatsRequest req = TestEnumFormatsRequest.builder()
                .stringArrayFormat(StringArrayFormat.AwaitingReviewProcess)
                .stringMapFormat(StringMapFormat.ModerateImportanceLevel)
                .stringPartialMapFormat(StringPartialMapFormat.InitialDraftVersion)
                .integerMapFormat(IntegerMapFormat.SuccessfulOperationComplete)
                .integerPartialMapFormat(IntegerPartialMapFormat.PrimaryFirstOption)
                .build();

        TestEnumFormatsResponse res = sdk.testEnumFormats()
                .request(req)
                .call();

        if (res.object().isPresent()) {
            System.out.println(res.object().get());
        }
    }
}
```

### Parameters

| Parameter                                                                   | Type                                                                        | Required                                                                    | Description                                                                 |
| --------------------------------------------------------------------------- | --------------------------------------------------------------------------- | --------------------------------------------------------------------------- | --------------------------------------------------------------------------- |
| `request`                                                                   | [TestEnumFormatsRequest](../../models/operations/TestEnumFormatsRequest.md) | :heavy_check_mark:                                                          | The request object to use for the request.                                  |

### Response

**[TestEnumFormatsResponse](../../models/operations/TestEnumFormatsResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## binaryAndStringUpload

### Example Usage

<!-- UsageSnippet language="java" operationID="binaryAndStringUpload" method="post" path="/binaryAndStringUpload" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.BinaryAndStringUploadRequest;
import org.openapis.openapi.models.operations.BinaryAndStringUploadResponse;
import org.openapis.openapi.utils.Utils;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        BinaryAndStringUploadRequest req = BinaryAndStringUploadRequest.builder()
                .binary(Utils.readBytes("test.json"))
                .string(Utils.readString("test.json"))
                .build();

        BinaryAndStringUploadResponse res = sdk.binaryAndStringUpload()
                .request(req)
                .call();

        // handle response
    }
}
```

### Parameters

| Parameter                                                                               | Type                                                                                    | Required                                                                                | Description                                                                             |
| --------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------- |
| `request`                                                                               | [BinaryAndStringUploadRequest](../../models/operations/BinaryAndStringUploadRequest.md) | :heavy_check_mark:                                                                      | The request object to use for the request.                                              |

### Response

**[BinaryAndStringUploadResponse](../../models/operations/BinaryAndStringUploadResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## getDuplicateExportCollision

Tests that a spec-defined error type colliding with a built-in SDK error name does not cause TS2308

### Example Usage

<!-- UsageSnippet language="java" operationID="getDuplicateExportCollision" method="get" path="/duplicateExportCollision" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.RequestTimeoutError;
import org.openapis.openapi.models.operations.GetDuplicateExportCollisionResponse;

public class Application {

    public static void main(String[] args) throws RequestTimeoutError, Exception {

        SDK sdk = SDK.builder()
            .build();

        GetDuplicateExportCollisionResponse res = sdk.getDuplicateExportCollision()
                .call();

        if (res.object().isPresent()) {
            System.out.println(res.object().get());
        }
    }
}
```

### Response

**[GetDuplicateExportCollisionResponse](../../models/operations/GetDuplicateExportCollisionResponse.md)**

### Errors

| Error Type                        | Status Code                       | Content Type                      |
| --------------------------------- | --------------------------------- | --------------------------------- |
| models/errors/RequestTimeoutError | 408                               | application/json                  |
| models/errors/SDKException        | 4XX, 5XX                          | \*/\*                             |

## getNamedPrimitiveUnion

Test named primitive union options using title and x-speakeasy-name-override

### Example Usage

<!-- UsageSnippet language="java" operationID="getNamedPrimitiveUnion" method="get" path="/namedPrimitiveUnion" -->
```java
package hello.world;

import com.fasterxml.jackson.databind.JsonNode;
import java.lang.Boolean;
import java.lang.Exception;
import java.lang.String;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.GetNamedPrimitiveUnionResponse;
import org.openapis.openapi.models.shared.MyObject;
import org.openapis.openapi.models.shared.SomeUnion;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        GetNamedPrimitiveUnionResponse res = sdk.getNamedPrimitiveUnion()
                .call();

        if (res.someUnion().isPresent()) {
            SomeUnion unionValue = res.someUnion().get();
            if (unionValue.myString().isPresent()) {
                String myStringValue = unionValue.myString().get();
                // Handle myString variant
            } else if (unionValue.myObject().isPresent()) {
                MyObject myObjectValue = unionValue.myObject().get();
                // Handle myObject variant
            } else if (unionValue.asFoo().isPresent()) {
                Boolean fooValue = unionValue.asFoo().get();
                // Handle asFoo variant
            } else if (unionValue.asJson().isPresent()) {
                JsonNode raw = unionValue.asJson().get();
                // Handle unknown variant fallback
            }
        }
    }
}
```

### Response

**[GetNamedPrimitiveUnionResponse](../../models/operations/GetNamedPrimitiveUnionResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## getEmptyObjectError

This endpoint tests the behavior when an error response has an empty object schema.

### Example Usage

<!-- UsageSnippet language="java" operationID="getEmptyObjectError" method="get" path="/emptyObjectError" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.FailedResponseException;
import org.openapis.openapi.models.operations.GetEmptyObjectErrorResponse;

public class Application {

    public static void main(String[] args) throws FailedResponseException, Exception {

        SDK sdk = SDK.builder()
            .build();

        GetEmptyObjectErrorResponse res = sdk.getEmptyObjectError()
                .call();

        if (res.object().isPresent()) {
            System.out.println(res.object().get());
        }
    }
}
```

### Response

**[GetEmptyObjectErrorResponse](../../models/operations/GetEmptyObjectErrorResponse.md)**

### Errors

| Error Type                            | Status Code                           | Content Type                          |
| ------------------------------------- | ------------------------------------- | ------------------------------------- |
| models/errors/FailedResponseException | 500                                   | application/json                      |
| models/errors/SDKException            | 4XX, 5XX                              | \*/\*                                 |

## urlValidationStressTest

### Example Usage

<!-- UsageSnippet language="java" operationID="urlValidationStressTest" method="get" path="/AZaz09-._~!$&'*+,;=:@/%20%25/-._~!$&'()*+,;=:@" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.UrlValidationStressTestResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        UrlValidationStressTestResponse res = sdk.urlValidationStressTest()
                .call();

        // handle response
    }
}
```

### Response

**[UrlValidationStressTestResponse](../../models/operations/UrlValidationStressTestResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## parenthesesInPathAllowed

A string with {{ double braces }} and { single braces }
and \{\{ escaped curlies \}\} and `backticks`.
and \`escaped backticks\` and double slashes\\
and 'single quotes' and "double quotes".
and  \'escaped single quotes\' and \"escaped double quotes\".


### Example Usage

<!-- UsageSnippet language="java" operationID="parenthesesInPathAllowed" method="post" path="/jobs/job({id})" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.ParenthesesInPathAllowedResponse;
import org.openapis.openapi.models.shared.TemplateBracesTest;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        ParenthesesInPathAllowedResponse res = sdk.parenthesesInPathAllowed()
                .id("<id>")
                .templateBracesTest(TemplateBracesTest.builder()
                    .fieldWithBracesInDescription("A string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand 'single quotes' and \"double quotes\".\nand  \\'escaped single quotes\\' and \\\"escaped double quotes\\\".\n")
                    .fieldWithBracesInTitle("A string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand 'single quotes' and \"double quotes\".\nand  \\'escaped single quotes\\' and \\\"escaped double quotes\\\".\n")
                    .fieldWithBracesInExample("A string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand 'single quotes' and \"double quotes\".\nand  \\'escaped single quotes\\' and \\\"escaped double quotes\\\".\n")
                    .build())
                .call();

        if (res.templateBracesTest().isPresent()) {
            System.out.println(res.templateBracesTest().get());
        }
    }
}
```

### Parameters

| Parameter                                                       | Type                                                            | Required                                                        | Description                                                     |
| --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- |
| `id`                                                            | *String*                                                        | :heavy_check_mark:                                              | N/A                                                             |
| `templateBracesTest`                                            | [TemplateBracesTest](../../models/shared/TemplateBracesTest.md) | :heavy_check_mark:                                              | N/A                                                             |

### Response

**[ParenthesesInPathAllowedResponse](../../models/operations/ParenthesesInPathAllowedResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## getNestedIntegerString

This endpoint tests the behavior when a deeply nested struct contains
an integer field that should be unmarshaled from a string.

### Example Usage

<!-- UsageSnippet language="java" operationID="getNestedIntegerString" method="get" path="/nestedIntegerString" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.GetNestedIntegerStringResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        GetNestedIntegerStringResponse res = sdk.getNestedIntegerString()
                .call();

        if (res.taskResponse().isPresent()) {
            System.out.println(res.taskResponse().get());
        }
    }
}
```

### Response

**[GetNestedIntegerStringResponse](../../models/operations/GetNestedIntegerStringResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## renderAsset

Render a media asset from a text prompt.

### Example Usage

<!-- UsageSnippet language="java" operationID="renderAsset" method="post" path="/assets/render" -->
```java
package hello.world;

import java.lang.Exception;
import java.util.stream.Stream;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.AssetStream;
import org.openapis.openapi.models.operations.RenderAssetRequest;
import org.openapis.openapi.models.operations.RenderAssetResponse;
import org.openapis.openapi.utils.EventStream;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        RenderAssetRequest req = RenderAssetRequest.builder()
                .prompt("<value>")
                .build();

        RenderAssetResponse res = sdk.renderAsset()
                .request(req)
                .call();

        // handle event stream, must be closed after use!
        try (EventStream<AssetStream> events = res.events()) {
            // Option 1: Use for-each loop
            for (AssetStream event : events) {
                System.out.println(event);
            }

            // Option 2: Use Stream API
            try (Stream<AssetStream> stream = events.stream()) {
                 stream.forEach(System.out::println);
            }
        }
    }
}
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `request`                                                           | [RenderAssetRequest](../../models/operations/RenderAssetRequest.md) | :heavy_check_mark:                                                  | The request object to use for the request.                          |

### Response

**[RenderAssetResponse](../../models/operations/RenderAssetResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## getAsset

Get the current result of an asset job.

### Example Usage

<!-- UsageSnippet language="java" operationID="getAsset" method="get" path="/assets/{id}" -->
```java
package hello.world;

import java.lang.Exception;
import java.util.stream.Stream;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.AssetStatusStream;
import org.openapis.openapi.models.operations.GetAssetResponse;
import org.openapis.openapi.utils.EventStream;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        GetAssetResponse res = sdk.getAsset()
                .id("<id>")
                .stream(true)
                .call();

        // handle event stream, must be closed after use!
        try (EventStream<AssetStatusStream> events = res.events()) {
            // Option 1: Use for-each loop
            for (AssetStatusStream event : events) {
                System.out.println(event);
            }

            // Option 2: Use Stream API
            try (Stream<AssetStatusStream> stream = events.stream()) {
                 stream.forEach(System.out::println);
            }
        }
    }
}
```

### Parameters

| Parameter           | Type                | Required            | Description         |
| ------------------- | ------------------- | ------------------- | ------------------- |
| `id`                | *String*            | :heavy_check_mark:  | N/A                 |
| `stream`            | @Nullable *boolean* | :heavy_minus_sign:  | N/A                 |

### Response

**[GetAssetResponse](../../models/operations/GetAssetResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## getErrorOnlyExample

This endpoint tests that when an operation has a named example only on
an error response (not on the success response), we still generate
a default example for the success response.

### Example Usage

<!-- UsageSnippet language="java" operationID="getErrorOnlyExample" method="get" path="/errorOnlyExample" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.operations.GetErrorOnlyExampleResponse;

public class Application {

    public static void main(String[] args) throws Error, Exception {

        SDK sdk = SDK.builder()
            .build();

        GetErrorOnlyExampleResponse res = sdk.getErrorOnlyExample()
                .call();

        if (res.object().isPresent()) {
            System.out.println(res.object().get());
        }
    }
}
```

### Response

**[GetErrorOnlyExampleResponse](../../models/operations/GetErrorOnlyExampleResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/Error        | 404                        | application/json           |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |