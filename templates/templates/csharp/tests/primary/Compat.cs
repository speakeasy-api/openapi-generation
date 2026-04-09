#nullable enable
using System;
using System.Net.Http;
using Newtonsoft.Json;
using Openapi.Utils;
using Xunit;

// Compat — v1 ⇄ v2 test-side glue.
//
// Tests under tests/primary are inherited by both csharp (v1) and csharpv2.
// Where the two SDKs expose different internal helpers (Newtonsoft-based in v1,
// System.Text.Json + source-generated context in v2), tests must not call those
// helpers directly. Instead they call Compat.* — each variant ships its own
// Compat.cs with the same signatures but variant-native implementations.
//
// v2's override lives at templates/csharpv2/tests/primary/Compat.cs.
public static class Compat
{
    public static T DeserializeJson<T>(string json) =>
        ResponseBodyDeserializer.DeserializeNotNull<T>(json, NullValueHandling.Ignore);

    public static string SerializeJson(object? value) =>
        JsonConvert.SerializeObject(value);

    public static void AssertIsJsonParseException(Exception? ex) {
        var typed = Assert.IsType<JsonReaderException>(ex);
        Assert.Equal(
            "Invalid property identifier character: ,. Path '', line 1, position 1.",
            typed.Message
        );
    }

    public static HttpContent? SerializeRequestBody(
        object? body,
        string requestField,
        string method,
        bool nullable,
        bool optional,
        string format = ""
    ) => RequestBodySerializer.Serialize(body, requestField, method, nullable, optional, format);

    // Stringify any value for test assertions. v1 delegates to Utilities.ToString
    // (already reflective via Newtonsoft). v2 reimplements with reflective STJ
    // path so production Utilities can stay AOT-clean.
    public static string ToString(object? value) => Utilities.ToString(value);
}
