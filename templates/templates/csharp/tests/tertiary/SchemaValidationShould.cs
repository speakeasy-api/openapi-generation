#nullable enable
using System.Collections.Generic;
using System.IO;
using System.Text;
using System.Threading.Tasks;
using Xunit;
using No_Security.API;
using No_Security.API.Models.Errors;
using No_Security.API.Models.Operations;
using No_Security.API.Models.Shared;
using No_Security.API.Utils;
using No_Security.API.Utils.Sse;

// Round-trip tests for `schemaValidation: lenient`: schema mismatches degrade
// gracefully instead of raising ResponseValidationException. The strict
// counterparts (mismatches raise) are pinned by the primary variant tests.
public class SchemaValidationShould
{
    [Fact]
    public async Task ConstructTypedResponseWhenRequiredFieldMissing()
    {
        CommonHelpers.RecordTest("errors-response-body-validation-lenient");
        var sdk = new SDK();

        var res = await sdk.Errors.GetMalformedErrorResponseAsync(Shape.MissingRequired, 200);

        Assert.NotNull(res.Object);
        Assert.Equal(0, res.Object!.Count);
    }

    [Fact]
    public async Task ConstructTypedResponseWhenKnownFieldHasWrongType()
    {
        var sdk = new SDK();

        var res = await sdk.Errors.GetMalformedErrorResponseAsync(Shape.WrongType, 200);

        Assert.NotNull(res.Object);
        Assert.Equal(0, res.Object!.Count);
    }

    [Fact]
    public async Task ThrowDefaultErrorWhenSuccessBodyIsNotJson()
    {
        var sdk = new SDK();

        var ex = await Assert.ThrowsAsync<SDKException>(
            () => sdk.Errors.GetMalformedErrorResponseAsync(Shape.Html, 200)
        );

        Assert.Equal(200, ex.StatusCode);
        Assert.Contains("502 Bad Gateway", ex.Body);
    }

    [Theory]
    [InlineData(Shape.MalformedJson, 400)]
    [InlineData(Shape.Html, 500)]
    [InlineData(Shape.Plain, 422)]
    public async Task FallBackToDefaultErrorWhenErrorBodyIsMalformed(Shape shape, int statusCode)
    {
        CommonHelpers.RecordTest("errors-response-body-validation-lenient-non-json");
        var sdk = new SDK();

        // Only a body that is not parseable JSON at all degrades to the default
        // SDK error; the typed error class cannot be constructed from it.
        var ex = await Assert.ThrowsAsync<SDKException>(
            () => sdk.Errors.GetMalformedErrorResponseAsync(shape, statusCode)
        );

        Assert.Equal(statusCode, ex.StatusCode);
        Assert.NotEmpty(ex.Body);
        // The underlying deserialization failure is preserved as InnerException.
        Assert.NotNull(ex.InnerException);
        Assert.StartsWith("Newtonsoft.Json", ex.InnerException!.GetType().FullName!);
    }

    [Theory]
    [InlineData(400)]
    [InlineData(500)]
    public async Task ConstructTypedErrorWhenErrorBodyHasWrongTypedField(int statusCode)
    {
        CommonHelpers.RecordTest("errors-error-body-validation-lenient");
        var sdk = new SDK();

        // Valid JSON error body whose `code` is an object where the schema
        // declares a string. Lenient keeps the typed Error, skips the bad
        // field, and records the strict failure on DeserializationException.
        const string rawBody = "{\"code\":{\"invalid\":true},\"message\":\"degraded error\"}";
        var ex = await Assert.ThrowsAsync<Error>(
            () => sdk.Errors.GetMalformedErrorResponseAsync(Shape.WrongTypeError, statusCode)
        );

        // Transport-level context is preserved verbatim.
        Assert.Equal(statusCode, ex.StatusCode);
        Assert.Equal(rawBody, ex.Body);
        Assert.NotNull(ex.RawResponse);
        Assert.Equal(statusCode, (int)ex.RawResponse.StatusCode);

        // The x-speakeasy-error-message accessor still resolves off the
        // best-effort payload rather than the "API error occurred" fallback.
        Assert.Equal("degraded error", ex.Message);

        // Payload is the typed model: valid fields kept, the wrong-typed known
        // field skipped (left default), absent optional field null.
        Assert.NotNull(ex.Payload);
        Assert.Equal("degraded error", ex.Payload.Message);
        Assert.Null(ex.Payload.Code);
        Assert.Null(ex.Payload.Type);

        // The failure strict mode would have raised is captured verbatim: same
        // transport context, and the underlying JSON error as InnerException.
        Assert.NotNull(ex.DeserializationException);
        Assert.Equal(statusCode, ex.DeserializationException!.StatusCode);
        Assert.Equal(rawBody, ex.DeserializationException.Body);
        Assert.NotNull(ex.DeserializationException.InnerException);
        Assert.StartsWith(
            "Newtonsoft.Json",
            ex.DeserializationException.InnerException!.GetType().FullName!
        );
    }

    [Fact]
    public async Task StillThrowTypedErrorWhenErrorBodyIsWellFormed()
    {
        var sdk = new SDK();

        // /errors/{statusCode} returns a canonical, schema-valid error body.
        var ex = await Assert.ThrowsAsync<Error>(
            () => sdk.Errors.StatusGetXSpeakeasyErrorsAsync(500)
        );

        Assert.NotNull(ex.Payload);
        Assert.Equal("an error occurred", ex.Payload.Message);
        Assert.Null(ex.DeserializationException);
    }

    [Fact]
    public async Task YieldLenientFrameWithoutAbortingStream()
    {
        CommonHelpers.RecordTest("event-stream-malformed-frame-lenient");
        var sdk = new SDK();

        var res = await sdk.Eventstreams.MalformedStreamAsync();

        var events = new List<JsonEvent>();
        using (var stream = res.JsonEvent)
        {
            Assert.NotNull(stream);
            JsonEvent? e;
            while ((e = await stream!.Next()) != null)
            {
                events.Add(e);
            }
        }

        // Server sends a valid frame followed by one whose data is missing the
        // required `content` field; lenient keeps both and the stream completes.
        Assert.Equal(2, events.Count);
        Assert.Equal("Hello", events[0].Data.Content);
        Assert.NotNull(events[1].Data);
        Assert.Null(events[1].Data.Content);
    }

    [Fact]
    public async Task SkipPlainTextFrameWithoutAbortingStream()
    {
        // A frame whose data is plain text (not JSON) for an object event type
        // has no best-effort typed form: constructing one would fabricate a
        // default-initialized frame indistinguishable from a genuine empty
        // object. Lenient skips it and the stream continues to the next frame.
        var sse = "data: not json at all\n\ndata: {\"content\":\"after\"}\n\n";
        using (var stream = new MemoryStream(Encoding.UTF8.GetBytes(sse)))
        using (var events = new EventStream<JsonEvent>(stream))
        {
            var e = await events.Next();

            Assert.NotNull(e);
            Assert.Equal("after", e!.Data.Content);
            Assert.Null(await events.Next());
        }
    }

    [Fact]
    public async Task KeepTypedVariantWhenKnownDiscriminatorPayloadIsInvalid()
    {
        CommonHelpers.RecordTest("errors-response-body-validation-lenient-union-variant");
        var sdk = new SDK();

        // The service returns a 5XX discriminated error union whose `tag` selects
        // the mapped variant TaggedError1 (error: string), but the body carries
        // `error` as an object. Lenient surfaces the typed TaggedError1 and
        // records the strict failure instead of degrading to a generic error.
        var ex = await Assert.ThrowsAsync<TaggedError1>(
            () => sdk.Errors.GetMalformedErrorUnionResponseAsync()
        );

        Assert.NotNull(ex.Payload);
        Assert.Equal("tag1", ex.Payload.Tag);
        Assert.Null(ex.Payload.Error);
        Assert.NotNull(ex.DeserializationException);
        Assert.NotNull(ex.DeserializationException!.InnerException);
        Assert.StartsWith(
            "Newtonsoft.Json",
            ex.DeserializationException.InnerException!.GetType().FullName!
        );
    }

    [Fact]
    public void KeepTypedVariantOfNonErrorUnionDuringLenientDeserialize()
    {
        // `imageURL` is required on ConstObject1; lenient still constructs the
        // mapped variant of this (non-error) response union instead of
        // degrading to the open-union Unknown fallback. Unlike error unions,
        // no ValidationFailure is recorded on the salvaged variant.
        var result = ResponseBodyDeserializer.DeserializeNotNull<ConstDiscriminatedOneOf>(
            "{\"tag\":\"tag1\"}"
        );

        Assert.Equal(ConstDiscriminatedOneOfType.Tag1, result.Type);
        Assert.False(result.IsUnknown());
        Assert.NotNull(result.ConstObject1);
        Assert.Null(result.ConstObject1!.ImageURL);
    }

    [Fact]
    public void DegradeToUnknownOnlyForUnknownDiscriminator()
    {
        var result = ResponseBodyDeserializer.DeserializeNotNull<ConstDiscriminatedOneOf>(
            "{\"tag\":\"tag99\"}"
        );

        Assert.True(result.IsUnknown());
        Assert.Null(result.ConstObject1);
        Assert.Null(result.ConstObject2);
    }

    [Fact]
    public void PreserveValidationFailureOnErrorUnionVariant()
    {
        // Discriminator "tag1" maps to TaggedError1 (error: string); the body
        // carries error as an object, failing strict validation. Lenient keeps
        // the typed variant and records the failure on the union's
        // ValidationFailure so the error-union throw can surface it.
        var result = ResponseBodyDeserializer.DeserializeNotNull<ErrorUnionDiscriminatedPostResponseBody>(
            "{\"tag\":\"tag1\",\"error\":{\"message\":\"boom\"}}"
        );

        Assert.Equal(ErrorUnionDiscriminatedPostResponseBodyType.Tag1, result.Type);
        Assert.False(result.IsUnknown());
        Assert.NotNull(result.TaggedError1Payload);
        Assert.NotNull(result.ValidationFailure);
        Assert.StartsWith(
            "Newtonsoft.Json",
            result.ValidationFailure!.GetType().FullName!
        );
    }

    [Fact]
    public async Task SurfaceDeserializationExceptionOnThrownErrorUnionVariant()
    {
        var sdk = new SDK();

        // The service echoes the request body back as the 4XX response. Sending a
        // tag1-discriminated body whose `error` is an object (not the string the
        // TaggedError1 variant declares) makes the known variant fail strict
        // validation; lenient throws the typed TaggedError1 and preserves the
        // strict failure on DeserializationException.
        var req = ErrorUnionDiscriminatedPostRequestBody.CreateTaggedError2RequestBody(
            new TaggedError2RequestBody()
            {
                Tag = "tag1",
                TaggedError2Message = new TaggedError2Message() { Message = "boom" },
            }
        );

        var ex = await Assert.ThrowsAsync<TaggedError1>(
            () => sdk.Errors.ErrorUnionDiscriminatedPostAsync(req)
        );

        Assert.NotNull(ex.Payload);
        Assert.NotNull(ex.DeserializationException);
        Assert.NotNull(ex.DeserializationException!.InnerException);
        Assert.StartsWith(
            "Newtonsoft.Json",
            ex.DeserializationException.InnerException!.GetType().FullName!
        );
    }

    [Fact]
    public void RecoverNestedWrongTypedFieldDuringLenientConstruct()
    {
        // A nested scalar (`error.message`, declared string) carries an object.
        // Lenient must recurse: the outer object and sibling fields are kept and
        // only the deeply-nested wrong-typed field is skipped, not the whole body.
        var obj = ResponseBodyDeserializer.ConstructUnvalidated<TaggedError2Payload>(
            "{\"tag\":\"tag2\",\"error\":{\"message\":{\"deeply\":\"nested\"}}}"
        );

        Assert.NotNull(obj);
        Assert.Equal("tag2", obj!.Tag);
        Assert.NotNull(obj.Error);
        Assert.Null(obj.Error.Message);
    }

    [Fact]
    public void SkipScalarWhereObjectExpectedDuringLenientConstruct()
    {
        // Inverse mismatch: `error` is declared an object but the body carries a
        // scalar. Lenient skips the field (leaving its default) and keeps the
        // sibling `tag`, rather than failing the whole construct.
        var obj = ResponseBodyDeserializer.ConstructUnvalidated<TaggedError2Payload>(
            "{\"tag\":\"tag2\",\"error\":\"boom\"}"
        );

        Assert.NotNull(obj);
        Assert.Equal("tag2", obj!.Tag);
        Assert.Null(obj.Error.Message);
    }

    [Fact]
    public void PreserveSiblingAdditionalPropertiesDuringLenientConstruct()
    {
        // A wrong-typed known field (`normalField`, declared string) forces the
        // lenient path; the sibling additionalProperties map must survive the
        // best-effort construct rather than being lost along with the bad field.
        var obj = ResponseBodyDeserializer.ConstructUnvalidated<ObjWithStringAdditionalProperties>(
            "{\"normalField\":{\"unexpected\":true},\"additionalProperties\":{\"extra\":\"kept\"}}"
        );

        Assert.NotNull(obj);
        Assert.Null(obj!.NormalField);
        Assert.NotNull(obj.AdditionalProperties);
        Assert.Equal("kept", obj.AdditionalProperties!["extra"]);
    }

    [Theory]
    [InlineData("")]
    [InlineData("   ")]
    [InlineData("{\"code\":")]
    [InlineData("<html>x</html>")]
    public void ReturnFalseFromLenientConstructForUnparseableBody(string body)
    {
        // Phase-1 guard: a body that is not complete, parseable JSON has no
        // best-effort representation and must not yield a partial object.
        var ok = ResponseBodyDeserializer.TryConstructUnvalidated<ErrorPayload>(body, out var value);

        Assert.False(ok);
        Assert.Null(value);
    }

    [Theory]
    [InlineData("null")]
    [InlineData("42")]
    [InlineData("true")]
    [InlineData("\"scalar\"")]
    [InlineData("[1,2]")]
    public void ReturnFalseFromLenientConstructForUnmappableRoot(string body)
    {
        // Phase-2 guard: parseable JSON whose root cannot map to the target
        // object (null, scalar, array) yields no instance either; the caller
        // degrades to the default error exactly as for unparseable bodies.
        var ok = ResponseBodyDeserializer.TryConstructUnvalidated<ErrorPayload>(body, out var value);

        Assert.False(ok);
        Assert.Null(value);
    }
}
