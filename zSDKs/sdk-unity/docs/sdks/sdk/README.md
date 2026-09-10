# SDK

## Overview

This document will show case as many of our features as possible in as little operations/models as possible.
This will then generate a SDK that we can more easily review than the test SDKs based on uber.yaml spec.

Speakeasy Docs
<https://speakeasy.com/docs>

### Available Operations

* [OperationWithLeadingAndTrailingUnderscores](#operationwithleadingandtrailingunderscores)
* [PostFile](#postfile) - Post File
* [GetPolymorphism](#getpolymorphism)
* [GetRequestBodyFlattenedAway](#getrequestbodyflattenedaway)
* [GetFullyFlattenedRequest](#getfullyflattenedrequest)
* [CreateWithUnion](#createwithunion) - Create with discriminated union request body
* [TestEndpoint](#testendpoint)
* [CreateUser](#createuser) - Create User
* [GetUser](#getuser) - Get User
* [UpdateUser](#updateuser) - Update User
* [DeleteUser](#deleteuser) - Delete User
* [Login](#login) - Login
* [Validate](#validate) - Validate
* [GetBinaryDefaultResponse](#getbinarydefaultresponse)
* [TestEnumFormats](#testenumformats) - Test x-speakeasy-enums in different formats
* [BinaryAndStringUpload](#binaryandstringupload)
* [GetDuplicateExportCollision](#getduplicateexportcollision) - Tests that a spec-defined error type colliding with a built-in SDK error name does not cause TS2308
* [GetNamedPrimitiveUnion](#getnamedprimitiveunion) - Test named primitive union options using title and x-speakeasy-name-override
* [GetEmptyObjectError](#getemptyobjecterror) - Get Empty Object Error
* [UrlValidationStressTest](#urlvalidationstresstest)
* [ParenthesesInPathAllowed](#parenthesesinpathallowed) - A string with {{ double braces }} and { single braces }
and \{\{ escaped curlies \}\} and `backticks`.
and \`escaped backticks\` and double slashes\\
and 'single quotes' and "double quotes".
and  \'escaped single quotes\' and \"escaped double quotes\".

* [GetNestedIntegerString](#getnestedintegerstring) - Test nested struct with integer:string tag
* [GetErrorOnlyExample](#geterroronlyexample) - Operation with example only on error response

## OperationWithLeadingAndTrailingUnderscores

### Example Usage

<!-- UsageSnippet language="unity" operationID="_operation_with_leading_and_trailing_underscores_" method="get" path="/test_operation_id_with_underscores" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.OperationWithLeadingAndTrailingUnderscoresAsync(qp1: "renamed"))
{
    // handle response
}


```

### Parameters

| Parameter                                                                                                                                        | Type                                                                                                                                             | Required                                                                                                                                         | Description                                                                                                                                      | Example                                                                                                                                          |
| ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `Qp1`                                                                                                                                            | *string*                                                                                                                                         | :heavy_check_mark:                                                                                                                               | This parameter will not be filled in with the queryParam1 global because it uses x-speakeasy-name-override which results in a non-matching name. | renamed                                                                                                                                          |

### Response

**[OperationWithLeadingAndTrailingUnderscoresResponse](../../Models/OperationWithLeadingAndTrailingUnderscoresResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## PostFile

This is a test endpoint.
It has a description.

### Example Usage

<!-- UsageSnippet language="unity" operationID="postFile" method="post" path="/file" -->
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

### Parameters

| Parameter                                          | Type                                               | Required                                           | Description                                        |
| -------------------------------------------------- | -------------------------------------------------- | -------------------------------------------------- | -------------------------------------------------- |
| `request`                                          | [PostFileRequest](../../Models/PostFileRequest.md) | :heavy_check_mark:                                 | The request object to use for the request.         |

### Response

**[PostFileResponse](../../Models/PostFileResponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| ErrorsError      | 415, 4XX         | application/json |
| ErrorsError      | 5XX              | application/json |

## GetPolymorphism

### Example Usage

<!-- UsageSnippet language="unity" operationID="getPolymorphism" method="get" path="/polymorphism" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.GetPolymorphismAsync())
{
    // handle response
}


```

### Response

**[GetPolymorphismResponse](../../Models/GetPolymorphismResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## GetRequestBodyFlattenedAway

### Example Usage

<!-- UsageSnippet language="unity" operationID="getRequestBodyFlattenedAway" method="get" path="/requestBodyFlattenedAway" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(loneQueryParam: "<value>");


using(var res = await sdk.GetRequestBodyFlattenedAwayAsync())
{
    // handle response
}


```

### Parameters

| Parameter          | Type               | Required           | Description        |
| ------------------ | ------------------ | ------------------ | ------------------ |
| `LoneQueryParam`   | *string*           | :heavy_minus_sign: | N/A                |

### Response

**[GetRequestBodyFlattenedAwayResponse](../../Models/GetRequestBodyFlattenedAwayResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## GetFullyFlattenedRequest

### Example Usage

<!-- UsageSnippet language="unity" operationID="getFullyFlattenedRequest" method="post" path="/fullyFlattenedRequest" example="namedExampleThatIsntMatchedAcrossDifferentExamples" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(security: new Security() {
        Option6 = new SecurityOption6() {
            ClientCredentials = "<YOUR_CLIENT_CREDENTIALS_HERE>",
        },
    });


using(var res = await sdk.GetFullyFlattenedRequestAsync(
    lang: "en",
    requestBody: new GetFullyFlattenedRequestRequestBody() {
    Name = "<value>",
}))
{
    // handle response
}


```

### Parameters

| Parameter                                                                                  | Type                                                                                       | Required                                                                                   | Description                                                                                |
| ------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------ |
| `Lang`                                                                                     | *string*                                                                                   | :heavy_check_mark:                                                                         | N/A                                                                                        |
| `RequestBody`                                                                              | [GetFullyFlattenedRequestRequestBody](../../Models/GetFullyFlattenedRequestRequestBody.md) | :heavy_check_mark:                                                                         | N/A                                                                                        |
| `MaxLength`                                                                                | *long*                                                                                     | :heavy_minus_sign:                                                                         | N/A                                                                                        |

### Response

**[GetFullyFlattenedRequestResponse](../../Models/GetFullyFlattenedRequestResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## CreateWithUnion

Test CLI generation for discriminated unions with dot-notation flags

### Example Usage

<!-- UsageSnippet language="unity" operationID="createWithUnion" method="post" path="/unionRequestBody" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.CreateWithUnionAsync(shapeRequest: new ShapeRequest() {
    Name = "<value>",
    Shape = Shape.CreateRectangle(
        new Rectangle() {
            Type = "rectangle",
            Width = 3125.73D,
            Height = 922.51D,
        },
    ),
}))
{
    // handle response
}


```

### Parameters

| Parameter                                    | Type                                         | Required                                     | Description                                  |
| -------------------------------------------- | -------------------------------------------- | -------------------------------------------- | -------------------------------------------- |
| `ShapeRequest`                               | [ShapeRequest](../../Models/ShapeRequest.md) | :heavy_check_mark:                           | N/A                                          |
| `DryRun`                                     | *bool*                                       | :heavy_minus_sign:                           | If true, validates without creating          |

### Response

**[CreateWithUnionResponse](../../Models/CreateWithUnionResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## TestEndpoint

### Example Usage

<!-- UsageSnippet language="unity" operationID="testEndpoint" method="post" path="/test/endpoint/{testName}" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.TestEndpointAsync(
    testName: "<value>",
    requestBody: new TestEndpointRequestBody() {
    Test = "<value>",
}))
{
    // handle response
}


```

### Parameters

| Parameter                                                          | Type                                                               | Required                                                           | Description                                                        |
| ------------------------------------------------------------------ | ------------------------------------------------------------------ | ------------------------------------------------------------------ | ------------------------------------------------------------------ |
| `TestName`                                                         | *string*                                                           | :heavy_check_mark:                                                 | N/A                                                                |
| `RequestBody`                                                      | [TestEndpointRequestBody](../../Models/TestEndpointRequestBody.md) | :heavy_check_mark:                                                 | N/A                                                                |

### Response

**[TestEndpointResponse](../../Models/TestEndpointResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## CreateUser

Creates a new user in the system. Multiple named examples demonstrate
different pairing scenarios for documentation generation.


### Example Usage: paired-example

<!-- UsageSnippet language="unity" operationID="createUser" method="put" path="/user" example="paired-example" -->
```csharp
using Speakeasy.OpenAPI;
using System.Collections.Generic;

var sdk = new SDK();

BaseUser req = new BaseUser() {
    Email = "paired@example.com",
    FirstName = "John",
};


using(var res = await sdk.CreateUserAsync(req))
{
    // handle response
}


```
### Example Usage: request-only

<!-- UsageSnippet language="unity" operationID="createUser" method="put" path="/user" example="request-only" -->
```csharp
using Speakeasy.OpenAPI;
using System.Collections.Generic;

var sdk = new SDK();

BaseUser req = new BaseUser() {
    Email = "request-only@example.com",
};


using(var res = await sdk.CreateUserAsync(req))
{
    // handle response
}


```
### Example Usage: response-only

<!-- UsageSnippet language="unity" operationID="createUser" method="put" path="/user" example="response-only" -->
```csharp
using Speakeasy.OpenAPI;
using System.Collections.Generic;

var sdk = new SDK();

BaseUser req = new BaseUser() {
    Id = "8ffac18c-7d88-4879-b057-e5f45b9ce7de",
    Email = "Virginie47@gmail.com",
    Gender = Gender.Other,
};


using(var res = await sdk.CreateUserAsync(req))
{
    // handle response
}


```

### Parameters

| Parameter                                  | Type                                       | Required                                   | Description                                |
| ------------------------------------------ | ------------------------------------------ | ------------------------------------------ | ------------------------------------------ |
| `request`                                  | [BaseUser](../../Models/BaseUser.md)       | :heavy_check_mark:                         | The request object to use for the request. |

### Response

**[CreateUserResponse](../../Models/CreateUserResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## GetUser

Get User

### Example Usage

<!-- UsageSnippet language="unity" operationID="getUser" method="get" path="/user/{id}" example="success" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.GetUserAsync(id: "<id>"))
{
    // handle response
}


```

### Parameters

| Parameter          | Type               | Required           | Description        |
| ------------------ | ------------------ | ------------------ | ------------------ |
| `Id`               | *string*           | :heavy_check_mark: | N/A                |

### Response

**[GetUserResponse](../../Models/GetUserResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## UpdateUser

Update User

### Example Usage

<!-- UsageSnippet language="unity" operationID="updateUser" method="post" path="/user/{id}" -->
```csharp
using Speakeasy.OpenAPI;
using System.Collections.Generic;

var sdk = new SDK();


using(var res = await sdk.UpdateUserAsync(
    id: "<id>",
    user: new User() {
    Id = "8ffac18c-7d88-4879-b057-e5f45b9ce7de",
    Email = "Joanny.Feeney@gmail.com",
    Gender = Gender.Other,
}))
{
    // handle response
}


```

### Parameters

| Parameter                    | Type                         | Required                     | Description                  |
| ---------------------------- | ---------------------------- | ---------------------------- | ---------------------------- |
| `Id`                         | *string*                     | :heavy_check_mark:           | N/A                          |
| `User`                       | [User](../../Models/User.md) | :heavy_check_mark:           | N/A                          |

### Response

**[UpdateUserResponse](../../Models/UpdateUserResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## DeleteUser

Delete User

### Example Usage

<!-- UsageSnippet language="unity" operationID="deleteUser" method="delete" path="/user/{id}" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.DeleteUserAsync(id: "<id>"))
{
    // handle response
}


```

### Parameters

| Parameter          | Type               | Required           | Description        |
| ------------------ | ------------------ | ------------------ | ------------------ |
| `Id`               | *string*           | :heavy_check_mark: | N/A                |

### Response

**[DeleteUserResponse](../../Models/DeleteUserResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## Login

Login

### Example Usage

<!-- UsageSnippet language="unity" operationID="login" method="get" path="/auth/login" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.LoginAsync())
{
    // handle response
}


```

### Response

**[LoginResponse](../../Models/LoginResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## Validate

Validate

### Example Usage

<!-- UsageSnippet language="unity" operationID="validate" method="get" path="/auth/validate" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.ValidateAsync())
{
    // handle response
}


```

### Response

**[ValidateResponse](../../Models/ValidateResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## GetBinaryDefaultResponse

### Example Usage

<!-- UsageSnippet language="unity" operationID="getBinaryDefaultResponse" method="get" path="/binaryDefaultResponse" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.GetBinaryDefaultResponseAsync())
{
    // handle response
}


```

### Response

**[GetBinaryDefaultResponseResponse](../../Models/GetBinaryDefaultResponseResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## TestEnumFormats

This endpoint tests the x-speakeasy-enums extension in both array and map formats,
including partial map coverage and both string and integer enum types.

### Example Usage

<!-- UsageSnippet language="unity" operationID="testEnumFormats" method="post" path="/enumFormats" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

TestEnumFormatsRequest req = new TestEnumFormatsRequest() {
    StringArrayFormat = StringArrayFormat.AwaitingReviewProcess,
    StringMapFormat = StringMapFormat.ModerateImportanceLevel,
    StringPartialMapFormat = StringPartialMapFormat.InitialDraftVersion,
    IntegerMapFormat = IntegerMapFormat.SuccessfulOperationComplete,
    IntegerPartialMapFormat = IntegerPartialMapFormat.PrimaryFirstOption,
};


using(var res = await sdk.TestEnumFormatsAsync(req))
{
    // handle response
}


```

### Parameters

| Parameter                                                        | Type                                                             | Required                                                         | Description                                                      |
| ---------------------------------------------------------------- | ---------------------------------------------------------------- | ---------------------------------------------------------------- | ---------------------------------------------------------------- |
| `request`                                                        | [TestEnumFormatsRequest](../../Models/TestEnumFormatsRequest.md) | :heavy_check_mark:                                               | The request object to use for the request.                       |

### Response

**[TestEnumFormatsResponse](../../Models/TestEnumFormatsResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## BinaryAndStringUpload

### Example Usage

<!-- UsageSnippet language="unity" operationID="binaryAndStringUpload" method="post" path="/binaryAndStringUpload" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

BinaryAndStringUploadRequest req = new BinaryAndStringUploadRequest() {
    Binary = File.ReadAllBytes("test.json"),
    String = File.ReadAllText("test.json"),
};


using(var res = await sdk.BinaryAndStringUploadAsync(req))
{
    // handle response
}


```

### Parameters

| Parameter                                                                    | Type                                                                         | Required                                                                     | Description                                                                  |
| ---------------------------------------------------------------------------- | ---------------------------------------------------------------------------- | ---------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| `request`                                                                    | [BinaryAndStringUploadRequest](../../Models/BinaryAndStringUploadRequest.md) | :heavy_check_mark:                                                           | The request object to use for the request.                                   |

### Response

**[BinaryAndStringUploadResponse](../../Models/BinaryAndStringUploadResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## GetDuplicateExportCollision

Tests that a spec-defined error type colliding with a built-in SDK error name does not cause TS2308

### Example Usage

<!-- UsageSnippet language="unity" operationID="getDuplicateExportCollision" method="get" path="/duplicateExportCollision" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.GetDuplicateExportCollisionAsync())
{
    // handle response
}


```

### Response

**[GetDuplicateExportCollisionResponse](../../Models/GetDuplicateExportCollisionResponse.md)**

### Errors

| Error Type          | Status Code         | Content Type        |
| ------------------- | ------------------- | ------------------- |
| RequestTimeoutError | 408                 | application/json    |
| SDKException        | 4XX, 5XX            | \*/\*               |

## GetNamedPrimitiveUnion

Test named primitive union options using title and x-speakeasy-name-override

### Example Usage

<!-- UsageSnippet language="unity" operationID="getNamedPrimitiveUnion" method="get" path="/namedPrimitiveUnion" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.GetNamedPrimitiveUnionAsync())
{
    // handle response
}


```

### Response

**[GetNamedPrimitiveUnionResponse](../../Models/GetNamedPrimitiveUnionResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## GetEmptyObjectError

This endpoint tests the behavior when an error response has an empty object schema.

### Example Usage

<!-- UsageSnippet language="unity" operationID="getEmptyObjectError" method="get" path="/emptyObjectError" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.GetEmptyObjectErrorAsync())
{
    // handle response
}


```

### Response

**[GetEmptyObjectErrorResponse](../../Models/GetEmptyObjectErrorResponse.md)**

### Errors

| Error Type              | Status Code             | Content Type            |
| ----------------------- | ----------------------- | ----------------------- |
| FailedResponseException | 500                     | application/json        |
| SDKException            | 4XX, 5XX                | \*/\*                   |

## UrlValidationStressTest

### Example Usage

<!-- UsageSnippet language="unity" operationID="urlValidationStressTest" method="get" path="/AZaz09-._~!$&'*+,;=:@/%20%25/-._~!$&'()*+,;=:@" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.UrlValidationStressTestAsync())
{
    // handle response
}


```

### Response

**[UrlValidationStressTestResponse](../../Models/UrlValidationStressTestResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## ParenthesesInPathAllowed

A string with {{ double braces }} and { single braces }
and \{\{ escaped curlies \}\} and `backticks`.
and \`escaped backticks\` and double slashes\\
and 'single quotes' and "double quotes".
and  \'escaped single quotes\' and \"escaped double quotes\".


### Example Usage

<!-- UsageSnippet language="unity" operationID="parenthesesInPathAllowed" method="post" path="/jobs/job({id})" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.ParenthesesInPathAllowedAsync(
    id: "<id>",
    templateBracesTest: new TemplateBracesTest() {
    FieldWithBracesInDescription = @"A string with {{"{{"}} double braces }} and { single braces }
    and \{\{ escaped curlies \}\} and `backticks`.
    and \`escaped backticks\` and double slashes\\
    and 'single quotes' and \"double quotes\".
    and  \'escaped single quotes\' and \\"escaped double quotes\\".
    ",
    FieldWithBracesInDefault = @"A string with {{"{{"}} double braces }} and { single braces }
    and \{\{ escaped curlies \}\} and `backticks`.
    and \`escaped backticks\` and double slashes\\
    and 'single quotes' and \"double quotes\".
    and  \'escaped single quotes\' and \\"escaped double quotes\\".
    ",
    FieldWithBracesInConst = @"A string with {{"{{"}} double braces }} and { single braces }
    and \{\{ escaped curlies \}\} and `backticks`.
    and \`escaped backticks\` and double slashes\\
    and 'single quotes' and \"double quotes\".
    and  \'escaped single quotes\' and \\"escaped double quotes\\".
    ",
    FieldWithBracesInTitle = @"A string with {{"{{"}} double braces }} and { single braces }
    and \{\{ escaped curlies \}\} and `backticks`.
    and \`escaped backticks\` and double slashes\\
    and 'single quotes' and \"double quotes\".
    and  \'escaped single quotes\' and \\"escaped double quotes\\".
    ",
    FieldWithBracesInExample = @"A string with {{"{{"}} double braces }} and { single braces }
    and \{\{ escaped curlies \}\} and `backticks`.
    and \`escaped backticks\` and double slashes\\
    and 'single quotes' and \"double quotes\".
    and  \'escaped single quotes\' and \\"escaped double quotes\\".
    ",
}))
{
    // handle response
}


```

### Parameters

| Parameter                                                | Type                                                     | Required                                                 | Description                                              |
| -------------------------------------------------------- | -------------------------------------------------------- | -------------------------------------------------------- | -------------------------------------------------------- |
| `Id`                                                     | *string*                                                 | :heavy_check_mark:                                       | N/A                                                      |
| `TemplateBracesTest`                                     | [TemplateBracesTest](../../Models/TemplateBracesTest.md) | :heavy_check_mark:                                       | N/A                                                      |

### Response

**[ParenthesesInPathAllowedResponse](../../Models/ParenthesesInPathAllowedResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## GetNestedIntegerString

This endpoint tests the behavior when a deeply nested struct contains
an integer field that should be unmarshaled from a string.

### Example Usage

<!-- UsageSnippet language="unity" operationID="getNestedIntegerString" method="get" path="/nestedIntegerString" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.GetNestedIntegerStringAsync())
{
    // handle response
}


```

### Response

**[GetNestedIntegerStringResponse](../../Models/GetNestedIntegerStringResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## GetErrorOnlyExample

This endpoint tests that when an operation has a named example only on
an error response (not on the success response), we still generate
a default example for the success response.

### Example Usage

<!-- UsageSnippet language="unity" operationID="getErrorOnlyExample" method="get" path="/errorOnlyExample" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.GetErrorOnlyExampleAsync())
{
    // handle response
}


```

### Response

**[GetErrorOnlyExampleResponse](../../Models/GetErrorOnlyExampleResponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| ErrorsError      | 404              | application/json |
| SDKException     | 4XX, 5XX         | \*/\*            |