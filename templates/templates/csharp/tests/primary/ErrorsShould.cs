using System.Threading.Tasks;
using Openapi;
using Openapi.Models.Errors;
using Openapi.Models.Operations;
using Xunit;
using System.Net;

public class ErrorsShould
{
    [Fact]
    public async Task TestStatusGetError_DefaultErrorCodes()
    {
        CommonHelpers.RecordTest("errors-status-get-error-default-error-codes");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        // When `clientServerStatusCodesAsErrors: true` is set (default), 4XX and 5XX ranges
        // are automatically treated as errors.

        // 400 and 500 responses are explicitly defined
        var ex400 = await Assert.ThrowsAsync<APIException>(() => sdk.Errors.StatusGetErrorAsync(400));
        Assert.Equal(HttpStatusCode.BadRequest, ex400.Response.StatusCode);
        var ex500 = await Assert.ThrowsAsync<APIException>(() => sdk.Errors.StatusGetErrorAsync(500));
        Assert.Equal(HttpStatusCode.InternalServerError, ex500.Response.StatusCode);

        // 404 and 503 responses are undefined but still treated as errors by default
        var ex404 = await Assert.ThrowsAsync<APIException>(() => sdk.Errors.StatusGetErrorAsync(404));
        Assert.Equal(HttpStatusCode.NotFound, ex404.Response.StatusCode);
        var ex503 = await Assert.ThrowsAsync<APIException>(() => sdk.Errors.StatusGetErrorAsync(503));
        Assert.Equal(HttpStatusCode.ServiceUnavailable, ex503.Response.StatusCode);
    }

    [Fact]
    public async Task TestStatusGetError_300_NonError()
    {
        CommonHelpers.RecordTest("errors-status-get-error300-non-error");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Errors.StatusGetErrorAsync(300);
        Assert.Equal(HttpStatusCode.Ambiguous, res.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task TestStatusGetErrorXSpeakeasyErrors()
    {
        CommonHelpers.RecordTest("errors-status-get-error-x-speakeasy-errors");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        // Note: Trailing "\n" are expected in response bodies because json.NewEncoder(w).Encode always appends a new line.

        // 400 response is explicitly defined and is marked as an error in `x-speakeasy-errors`
        var ex400 = await Assert.ThrowsAsync<APIException>(
        () => sdk.Errors.StatusGetXSpeakeasyErrorsAsync(400)
        );
        var body400 = "{\"message\":\"an error occurred\",\"code\":\"400\",\"type\":\"internal\"}\n";
        var msg400 = $"API error occurred. Body: {body400}.";
        Assert.Equal(msg400, ex400.Message);
        Assert.Equal(HttpStatusCode.BadRequest, ex400.Response.StatusCode);
        Assert.Equal(body400, ex400.Body);
        Assert.Equal($"Status: BadRequest. {msg400}", ex400.ToString());

        // 401 response is undefined but it is marked as an error in `x-speakeasy-errors`
        var ex401 = await Assert.ThrowsAsync<APIException>(
            () => sdk.Errors.StatusGetXSpeakeasyErrorsAsync(401)
        );
        var body401 = "{\"message\":\"an error occurred\",\"code\":\"401\",\"type\":\"internal\"}\n";
        var msg401 = $"API error occurred. Body: {body401}.";
        Assert.Equal(msg401, ex401.Message);
        Assert.Equal(HttpStatusCode.Unauthorized, ex401.Response.StatusCode);
        Assert.Equal(body401, ex401.Body);
        Assert.Equal($"Status: Unauthorized. {msg401}", ex401.ToString());

        // 402 response is undefined and is not treated as an API error since it's not listed in `x-speakeasy-errors`.
        // Instead we raise a "Unknown status code received" exception.
        var ex402 = await Assert.ThrowsAsync<APIException>(
            () => sdk.Errors.StatusGetXSpeakeasyErrorsAsync(402)
        );
        var body402 = "{\"message\":\"an error occurred\",\"code\":\"402\",\"type\":\"internal\"}\n";
        var msg402 = $"Unknown status code received. Body: {body402}.";
        Assert.Equal(msg402, ex402.Message);
        Assert.Equal(HttpStatusCode.PaymentRequired, ex402.Response.StatusCode);
        Assert.Equal(body402, ex402.Body);
        Assert.Equal($"Status: PaymentRequired. {msg402}", ex402.ToString());

        // Both 500 and 501 responses are marked as errors since `5XX` is listed `x-speakeasy-errors`.
        var ex500 = await Assert.ThrowsAsync<Openapi.Models.Errors.Error>(
            () => sdk.Errors.StatusGetXSpeakeasyErrorsAsync(500)
        );
        var body500 = "{\"message\":\"an error occurred\",\"code\":\"500\",\"type\":\"internal\"}\n";
        // The response body message field is expected to override the Exception message due to
        // `x-speakeasy-error-message: true` being explicitly set to true on the Error.Message field.
        // c.f. `getErrorMessageFieldDeep` in common/utils.ts
        var msg500 = $"an error occurred. Body: {body500}.";
        Assert.Equal(msg500, ex500.Message);
        Assert.Equal(HttpStatusCode.InternalServerError, ex500.Response.StatusCode);
        Assert.Equal(body500, ex500.Body);
        Assert.Equal($"Status: InternalServerError. {msg500}", ex500.ToString());
        Assert.Equal("500", ex500.Code);  // Legacy field
        Assert.Equal("500", ex500.Payload.Code);

        var ex501 = await Assert.ThrowsAsync<StatusGetXSpeakeasyErrorsResponseBody>(
            () => sdk.Errors.StatusGetXSpeakeasyErrorsAsync(501)
        );
        var body501 = "{\"message\":\"an error occurred\",\"code\":\"501\",\"type\":\"internal\"}\n";
        // The response body message field is expected to override the Exception message due to
        // the message field being implicitly selected by the getOrInferErrorMessageFieldDeep function
        // c.f. `inferErrorMessageField` in common/utils.ts
        var msg501 = $"an error occurred. Body: {body501}.";
        Assert.Equal(msg501, ex501.Message);
        Assert.Equal(HttpStatusCode.NotImplemented, ex501.HttpMeta.Response.StatusCode);  // Legacy field
        Assert.Equal(HttpStatusCode.NotImplemented, ex501.Response.StatusCode);
        Assert.Equal(body501, ex501.Body);
        Assert.Equal($"Status: NotImplemented. {msg501}", ex501.ToString());
        Assert.Equal("501", ex501.Code);  // Legacy field
        Assert.Equal("501", ex501.Payload.Code);
        Assert.Equal(Openapi.Models.Shared.ErrorType.Internal, ex501.Type);
    }

    [Fact]
    public async Task TestOptionalNullableErrorMessage()
    {
        CommonHelpers.RecordTest("errors-optional-nullable-error-message");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        // The optional+nullable x-speakeasy-error-message field is templated
        // as OptionalNullable<string?> (presenceAwareJsonSerialization: true).
        // The generated ErrorMessage(...) accessor must unwrap it.
        var ex = await Assert.ThrowsAsync<NullableOptionalErrorMessage>(
            () => sdk.Errors.GetNullableOptionalErrorMessageAsync()
        );
        var body = "{\"message\":\"an error occurred\",\"code\":\"512\",\"type\":\"internal\"}\n";
        Assert.Equal($"an error occurred. Body: {body}.", ex.Message);
        Assert.True(ex.Payload.Message.IsSet);
        Assert.Equal("an error occurred", ex.Payload.Message.Value);
    }

    [Fact]
    public async Task TestOptionalNullableNestedErrorMessage()
    {
        CommonHelpers.RecordTest("errors-optional-nullable-error-message-nested");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        // The x-speakeasy-error-message field is nested inside an optional+nullable
        // `cause` object, templated as OptionalNullable<Cause?>. ErrorMessage(...)
        // must peel off the intermediate wrapper before accessing `.Message`;
        var ex = await Assert.ThrowsAsync<NestedNullableOptionalErrorMessage>(
            () => sdk.Errors.GetNestedNullableOptionalErrorMessageAsync(cause: true)
        );
        // The exception message resolves through cause.Message specifically —
        // a value distinct from the top-level "an error occurred" proves the
        // nested field won, not the outer one.
        Assert.StartsWith("underlying cause.", ex.Message);
        Assert.True(ex.Payload.Cause.IsSet);
        Assert.Equal("underlying cause", ex.Payload.Cause.Value?.Message);
    }

    [Fact]
    public async Task TestStatusGetSuccessXSpeakeasyErrorsEmptyStatusCodeList()
    {
        CommonHelpers.RecordTest("errors-status-get-error-x-speakeasy-errors-none");
        var sdk = new SDK();

        // `x-speakeasy-errors` overrides the `clientServerStatusCodesAsErrors` setting
        // with a dummy list, meaning all responses are treated as non-errors.

        var res200 = await sdk.Errors.StatusGetNonErrorAsync(200);
        Assert.NotNull(res200);
        Assert.Equal(HttpStatusCode.OK, res200.HttpMeta.Response.StatusCode);

        var res400 = await sdk.Errors.StatusGetNonErrorAsync(400);
        Assert.NotNull(res400);
        Assert.Equal(HttpStatusCode.BadRequest, res400.HttpMeta.Response.StatusCode);

        var res500 = await sdk.Errors.StatusGetNonErrorAsync(500);
        Assert.NotNull(res500);
        Assert.Equal(HttpStatusCode.InternalServerError, res500.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task TestStatusGetSuccessXSpeakeasyErrorsUnspecifiedResponses()
    {
        CommonHelpers.RecordTest("errors-status-get-error-x-speakeasy-errors-default");
        var sdk = new SDK();

        // `x-speakeasy-errors` overrides the `clientServerStatusCodesAsErrors` setting
        // by marking all *unspecified* responses as errors.

        // 200 and 400 responses are explicitly defined, so they are treated as non-errors.
        var res200 = await sdk.Errors.StatusGetDefaultErrorAsync(200);
        Assert.NotNull(res200);
        Assert.Equal(HttpStatusCode.OK, res200.HttpMeta.Response.StatusCode);

        var res400 = await sdk.Errors.StatusGetDefaultErrorAsync(400);
        Assert.NotNull(res400);
        Assert.Equal(HttpStatusCode.BadRequest, res400.HttpMeta.Response.StatusCode);

        // 404 and 500 responses are undefined, so they are treated as errors.
        var ex404 = await Assert.ThrowsAsync<APIException>(() => sdk.Errors.StatusGetDefaultErrorAsync(404));
        Assert.Equal(HttpStatusCode.NotFound, ex404.Response.StatusCode);

        var ex500 = await Assert.ThrowsAsync<APIException>(() => sdk.Errors.StatusGetDefaultErrorAsync(500));
        Assert.Equal(HttpStatusCode.InternalServerError, ex500.Response.StatusCode);

        // To make sure the catch-all "default" code gets properly applied in `templateErrorStatusCodesCheck`,
        // an AfterError hook was added to Hooks/TestHook.cs to recover from 418 error.
        var res418 = await sdk.Errors.StatusGetDefaultErrorAsync(418);
        Assert.NotNull(res418);
        Assert.Equal(HttpStatusCode.OK, res418.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task TestResponseValidationException()
    {
        CommonHelpers.RecordTest("errors-response-body-validation-enabled");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        // An AfterSuccess hook was added to Hooks/TestHook.cs to set the response body to a
        // malformed json, which causes ResponseBodyDeserializer.Deserialize to throw.
        var ex = await Assert.ThrowsAsync<ResponseValidationException>(
            () => sdk.Errors.StatusGetXSpeakeasyErrorsAsync(418)
        );
        var exMsg = "Failed to deserialize response body into ErrorPayload. Body: {,}.";
        var innerMsg = "Invalid property identifier character: ,. Path '', line 1, position 1.";
        Assert.Equal(exMsg, ex.Message);
        Assert.NotNull(ex.InnerException);
        Assert.IsType<Newtonsoft.Json.JsonReaderException>(ex.InnerException);
        Assert.Equal(innerMsg, ex.InnerException.Message);
        Assert.Equal($"Status: 418. {exMsg}\n{innerMsg}", ex.ToString());
    }

    [Theory]
    [InlineData(Shape.MalformedJson, 400, "{\"error\":")]
    [InlineData(Shape.Html, 400, "502 Bad Gateway")]
    [InlineData(Shape.Plain, 500, "internal server timeout")]
    public async Task TestErrorBodyValidationEnabled(Shape shape, int statusCode, string expectedSubstring)
    {
        CommonHelpers.RecordTest("errors-error-body-validation-enabled");
        var sdk = new SDK();

        // schemaValidation strict (default): a malformed 4xx/5xx body raises
        // ResponseValidationException with the raw body and status preserved.
        var ex = await Assert.ThrowsAsync<ResponseValidationException>(
            () => sdk.Errors.GetMalformedErrorResponseAsync(shape, statusCode)
        );
        Assert.Equal(statusCode, (int)ex.Response.StatusCode);
        Assert.Contains(expectedSubstring, ex.Body);
    }

    [Fact]
    public async Task TestConnectionErrorGet()
    {
        CommonHelpers.RecordTest("errors-connection-error");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var ex = await Assert.ThrowsAsync<System.Net.Http.HttpRequestException>(
            () => sdk.Errors.ConnectionErrorGetAsync()
        );

        Assert.Contains("not known (somebrokenapi.broken:80)", ex.Message);
        Assert.IsType<System.Net.Sockets.SocketException>(ex.InnerException);
    }

    [Fact]
    public async Task TestUnionOfErrors()
    {
        CommonHelpers.RecordTest("errors-union-of-errors");
        var sdk = new Openapi.SDK();

        var req1 = ErrorUnionPostRequestBody.CreateErrorType1RequestBody(
            new ErrorType1RequestBody { Error = "Error1" }
        );
        var ex1 = await Assert.ThrowsAsync<ErrorType1>(() => sdk.Errors.ErrorUnionPostAsync(req1));
        Assert.Equal(HttpStatusCode.InternalServerError, ex1.Response.StatusCode);
        Assert.Equal("Error1", ex1.Error); // Legacy field
        Assert.Equal("Error1", ex1.Payload.Error);

        var req2 = ErrorUnionPostRequestBody.CreateErrorType2RequestBody(
            new ErrorType2RequestBody { ErrorType2Message = new ErrorType2Message { Message = "Error2" } }
        );
        var ex2 = await Assert.ThrowsAsync<ErrorType2>(() => sdk.Errors.ErrorUnionPostAsync(req2));
        Assert.Equal(HttpStatusCode.InternalServerError, ex2.Response.StatusCode);
        Assert.Equal("Error2", ex2.Error.Message);  // Legacy field
        Assert.Equal("Error2", ex2.Payload.Error.Message);
    }

    [Fact]
    public async Task TestDiscriminatedUnionOfErrors()
    {
        CommonHelpers.RecordTest("errors-union-of-errors-discriminated");
        var sdk = new Openapi.SDK();

        var req1 = ErrorUnionDiscriminatedPostRequestBody.CreateTaggedError1RequestBody(
            new TaggedError1RequestBody
            {
                Tag = "tag1",
                Error = "Error1"
            }
        );
        var ex1 = await Assert.ThrowsAsync<TaggedError1>(
            () => sdk.Errors.ErrorUnionDiscriminatedPostAsync(req1)
        );
        Assert.Equal(HttpStatusCode.BadRequest, ex1.Response.StatusCode);
        Assert.Equal(Tag.Tag1, ex1.Tag);  // Legacy field
        Assert.Equal("Error1", ex1.Error); // Legacy field
        var payload1 = ex1.Payload;
        Assert.Equal(Tag.Tag1, payload1.Tag);
        Assert.Equal("Error1", payload1.Error);

        // Make sure CreateTag1 properly overrides the payload tag to `Tag1`
        payload1.Tag = Tag.Tag0;
        var tag1 = ErrorUnionDiscriminatedPostResponseBody.CreateTag1(payload1);
        Assert.Equal(Tag.Tag1, tag1.TaggedError1Payload.Tag);

        var req2 = ErrorUnionDiscriminatedPostRequestBody.CreateTaggedError2RequestBody(
            new TaggedError2RequestBody
            {
                Tag = "tag2",
                TaggedError2Message = new TaggedError2Message { Message = "Error2" }
            }
        );
        var ex2 = await Assert.ThrowsAsync<TaggedError2>(
            () => sdk.Errors.ErrorUnionDiscriminatedPostAsync(req2)
        );
        Assert.Equal(HttpStatusCode.BadRequest, ex2.Payload.HttpMeta.Response.StatusCode);
        Assert.Equal(HttpStatusCode.BadRequest, ex2.Response.StatusCode);
        Assert.Equal("tag2", ex2.Tag); // Legacy field
        Assert.Equal("Error2", ex2.Error.Message); // Legacy field
        Assert.Equal("tag2", ex2.Payload.Tag);
        Assert.Equal("Error2", ex2.Payload.Error.Message);
    }

}
