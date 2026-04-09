#nullable enable
using System;
using System.Collections;
using System.Collections.Generic;
using System.Diagnostics.CodeAnalysis;
using System.IO;
using System.Linq;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Reflection;
using System.Text.Json;
using Openapi.Utils;
using Xunit;

// Compat — v2 implementation of the test-side glue defined in
// templates/csharp/tests/primary/Compat.cs. Same public signatures, v2-native
// implementations using System.Text.Json with the source-generated
// JsonTypeInfoResolver chain wired up via JsonOptions.
public static class Compat
{
    public static T DeserializeJson<T>(string json) =>
        JsonSerializer.Deserialize<T>(json, Openapi.Internal.Json.JsonOptions.SerializerOptions)!;

    public static string SerializeJson(object? value) =>
        JsonSerializer.Serialize(value, Openapi.Internal.Json.JsonOptions.SerializerOptions);

    public static void AssertIsJsonParseException(Exception? ex) =>
        Assert.IsType<JsonException>(ex);

    // Variant-specific parser error message for the malformed body "{,}" used in the
    // errors test handler. System.Text.Json emits a different message than Newtonsoft.
    public const string MalformedJsonInnerMessage =
        "',' is an invalid start of a property name. Expected a '\"'. Path: $ | LineNumber: 0 | BytePositionInLine: 1.";

    // Test-only shim for the v1 RequestBodySerializer.Serialize signature.
    // Tests under csharp/tests/primary call this through Compat to keep a single
    // call site across v1 and v2. The v2 production API splits per
    // serialization-method into typed entry points; Compat rebuilds the
    // v1-shaped `(object body, string method, ...)` dispatch on top of those
    // typed methods. Reflection-mode JSON path is AOT-hostile but tests are
    // not AOT-published, so we silence the analyzer here.
    [RequiresUnreferencedCode("Test-only path; production callers use SerializeJsonBody<T>.")]
    [RequiresDynamicCode("Test-only path; production callers use SerializeJsonBody<T>.")]
    public static HttpContent? SerializeRequestBody(
        object? body,
        string requestField,
        string method,
        bool nullable,
        bool optional,
        string format = ""
    )
    {
        if (body == null)
        {
            if (!nullable && !optional)
            {
                throw new ArgumentNullException(nameof(body), "request body is required");
            }
            if (nullable && method == "json")
            {
                return RequestBodySerializer.NullJsonBody();
            }
            return null;
        }

        return method switch
        {
            "json" => RequestBodySerializer.SerializeJsonBodyUntyped(body, format),
            "raw" => RequestBodySerializer.SerializeRaw(body),
            "string" => RequestBodySerializer.SerializeString(body),
            "form" => null,
            "multipart" => SerializeMultipartReflective(body),
            _ => throw new NotSupportedException(
                $"Serialization method '{method}' is not supported by Compat.SerializeRequestBody."
            ),
        };
    }

    // Stringify any value for test assertions. v1 delegates to Utilities.ToString
    // (already reflective via Newtonsoft). v2 implements its own reflective STJ
    // path so production Utilities can stay AOT-clean.
    [RequiresUnreferencedCode("Reflects over user types to stringify class/list/map shapes for tests.")]
    [RequiresDynamicCode("Reflects over user types to stringify class/list/map shapes for tests.")]
    public static string ToString(object? value)
    {
        if (value == null) return "";
        if (value is string s) return s;
        var t = value.GetType();
        if (t.IsPrimitive || t.IsEnum || Utilities.IsDate(value)
            || value is System.Numerics.BigInteger || t == typeof(decimal))
        {
            return Utilities.ValueToString(value);
        }
        return JsonSerializer.Serialize(value, t, Openapi.Internal.Json.JsonOptions.SerializerOptions);
    }

    // Reflective multipart for test inspection only. Mirrors v1 SerializeMultipart
    // on top of v2's typed [MultipartField] attribute. Production multipart path
    // is emitted inline at call sites; this method exists so shared tests can
    // assert on serialized bytes via Compat.
    private const string MultipartFieldAttrFullName = "Openapi.Generators.Serialization.MultipartFieldAttribute";

    private static (string? Name, bool IsContent, bool IsJson)? ReadMultipartAttr(MemberInfo m)
    {
        var attr = m.GetCustomAttributes(false).FirstOrDefault(a => a.GetType().FullName == MultipartFieldAttrFullName);
        if (attr == null) return null;
        var t = attr.GetType();
        var name = (string?)t.GetProperty("Name")?.GetValue(attr);
        var isContent = (bool?)t.GetProperty("IsContent")?.GetValue(attr) ?? false;
        var isJson = (bool?)t.GetProperty("IsJson")?.GetValue(attr) ?? false;
        return (name, isContent, isJson);
    }

    [RequiresUnreferencedCode("Reflects over user types to build multipart for tests.")]
    [RequiresDynamicCode("Reflects over user types to build multipart for tests.")]
    private static HttpContent SerializeMultipartReflective(object body)
    {
        var form = new MultipartFormDataContent();
        foreach (var prop in body.GetType().GetProperties())
        {
            var meta = ReadMultipartAttr(prop);
            if (meta == null) continue;
            var value = prop.GetValue(body);
            if (value == null) continue;
            var fieldName = meta.Value.Name ?? prop.Name;

            if (meta.Value.IsJson)
            {
                form.Add(new StringContent(JsonSerializer.Serialize(value, value.GetType(), Openapi.Internal.Json.JsonOptions.SerializerOptions)), fieldName);
                continue;
            }

            if (value is IList list && !(value is byte[]))
            {
                var itemType = prop.PropertyType.IsGenericType
                    ? prop.PropertyType.GetGenericArguments()[0]
                    : typeof(object);
                if (IsFileShape(itemType))
                {
                    foreach (var item in list)
                    {
                        if (item == null) continue;
                        AddFile(form, fieldName, item);
                    }
                }
                else
                {
                    foreach (var item in list)
                    {
                        if (item == null) continue;
                        form.Add(new StringContent(Utilities.ValueToString(item)), fieldName);
                    }
                }
                continue;
            }

            if (IsFileShape(prop.PropertyType))
            {
                AddFile(form, fieldName, value);
                continue;
            }

            form.Add(new StringContent(Utilities.ValueToString(value)), fieldName);
        }
        return form;
    }

    private static bool IsFileShape(Type t)
    {
        if (t == null || t.IsPrimitive || t == typeof(string) || t == typeof(byte[])) return false;
        return t.GetProperties().Any(p => ReadMultipartAttr(p)?.IsContent == true);
    }

    [RequiresUnreferencedCode("Reflects file shape for tests.")]
    private static void AddFile(MultipartFormDataContent form, string fieldName, object fileObj)
    {
        string fileName = "";
        byte[]? content = null;
        foreach (var p in fileObj.GetType().GetProperties())
        {
            var m = ReadMultipartAttr(p);
            if (m == null) continue;
            if (m.Value.IsContent) content = (byte[]?)p.GetValue(fileObj);
            else fileName = (p.GetValue(fileObj) as string) ?? "";
        }
        if (content == null) return;
        var bc = new ByteArrayContent(content);
        bc.Headers.ContentType = new MediaTypeHeaderValue(GuessMime(fileName));
        form.Add(bc, fieldName, fileName);
    }

    private static string GuessMime(string fileName)
    {
        var ext = Path.GetExtension(fileName)?.ToLowerInvariant();
        return ext switch
        {
            ".json" => "application/json",
            ".xml" => "application/xml",
            ".txt" => "text/plain",
            ".csv" => "text/csv",
            ".html" or ".htm" => "text/html",
            ".pdf" => "application/pdf",
            ".zip" => "application/zip",
            ".png" => "image/png",
            ".jpg" or ".jpeg" => "image/jpeg",
            ".gif" => "image/gif",
            ".svg" => "image/svg+xml",
            ".webp" => "image/webp",
            _ => "application/octet-stream",
        };
    }
}
