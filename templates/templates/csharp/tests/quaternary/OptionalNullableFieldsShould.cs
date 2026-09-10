using System.Net;
using System.Threading.Tasks;
using Company.Product.Feature.Subnamespace;
using Company.Product.Feature.Subnamespace.Utils;
using Newtonsoft.Json;
using Xunit;

// Legacy-mode counterfactuals (presenceAwareJsonSerialization: false) for the four
// optional/nullable quadrants covered by the primary OptionalNullableFieldsShould tests.
//
// With the flag off there is no OptionalNullable<T> wrapper and no strict Required deserialization:
// every field gets a bare [JsonProperty("name")] with no Required argument (Required.Default).
public class OptionalNullableFieldsShould
{
    // Primary: "object-with-optional-nullable-*" — tri-state IsSet/IsNull.
    // Legacy: plain T?, absent == null, no throws.
    [Fact]
    public async Task ObjectWithOptionalTrueNullableTrue_Legacy()
    {
        var sdk = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var full = await sdk.OptionalNullableFields.ObjectWithOptionalTrueNullableTrueFieldAsync("full");
        Assert.Equal(HttpStatusCode.OK, full.RawResponse.StatusCode);
        Assert.Equal(1L, full.ObjectWithOptionalTrueNullableTrueField!.Id);
        Assert.Equal("fluffy is sick", full.ObjectWithOptionalTrueNullableTrueField.MedicalRecord!.Diagnosis);

        var absent = await sdk.OptionalNullableFields.ObjectWithOptionalTrueNullableTrueFieldAsync("medicalRecordAbsent");
        Assert.Equal(1L, absent.ObjectWithOptionalTrueNullableTrueField!.Id);
        Assert.Null(absent.ObjectWithOptionalTrueNullableTrueField.MedicalRecord);

        var nulled = await sdk.OptionalNullableFields.ObjectWithOptionalTrueNullableTrueFieldAsync("medicalRecordNull");
        Assert.Equal(1L, nulled.ObjectWithOptionalTrueNullableTrueField!.Id);
        Assert.Null(nulled.ObjectWithOptionalTrueNullableTrueField.MedicalRecord);
    }

    // Primary: "object-with-optional-false-nullable-true-*" — required, absent throws.
    // Legacy: no Required validation, so absent and null both deserialize to null.
    [Fact]
    public async Task ObjectWithOptionalFalseNullableTrue_Legacy()
    {
        var sdk = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var full = await sdk.OptionalNullableFields.ObjectWithOptionalFalseNullableTrueFieldAsync("full");
        Assert.Equal(HttpStatusCode.OK, full.RawResponse.StatusCode);
        Assert.Equal(1L, full.ObjectWithOptionalFalseNullableTrueField!.Id);
        Assert.Equal("fluffy is sick", full.ObjectWithOptionalFalseNullableTrueField.MedicalRecord!.Diagnosis);

        // Primary throws here (Required.AllowNull, missing key); legacy tolerates it.
        var absent = await sdk.OptionalNullableFields.ObjectWithOptionalFalseNullableTrueFieldAsync("medicalRecordAbsent");
        Assert.Equal(1L, absent.ObjectWithOptionalFalseNullableTrueField!.Id);
        Assert.Null(absent.ObjectWithOptionalFalseNullableTrueField.MedicalRecord);

        var nulled = await sdk.OptionalNullableFields.ObjectWithOptionalFalseNullableTrueFieldAsync("medicalRecordNull");
        Assert.Equal(1L, nulled.ObjectWithOptionalFalseNullableTrueField!.Id);
        Assert.Null(nulled.ObjectWithOptionalFalseNullableTrueField.MedicalRecord);
    }

    // Primary: "object-with-optional-true-nullable-false-*" — non-nullable, explicit null throws.
    // Legacy: no DisallowNull validation, so explicit null deserializes to null.
    [Fact]
    public async Task ObjectWithOptionalTrueNullableFalse_Legacy()
    {
        var sdk = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var full = await sdk.OptionalNullableFields.ObjectWithOptionalTrueNullableFalseFieldAsync("full");
        Assert.Equal(HttpStatusCode.OK, full.RawResponse.StatusCode);
        Assert.Equal(1L, full.ObjectWithOptionalTrueNullableFalseField!.Id);
        Assert.Equal("fluffy is sick", full.ObjectWithOptionalTrueNullableFalseField.MedicalRecord!.Diagnosis);

        var absent = await sdk.OptionalNullableFields.ObjectWithOptionalTrueNullableFalseFieldAsync("medicalRecordAbsent");
        Assert.Equal(1L, absent.ObjectWithOptionalTrueNullableFalseField!.Id);
        Assert.Null(absent.ObjectWithOptionalTrueNullableFalseField.MedicalRecord);

        // Primary throws here (Required.DisallowNull, explicit null); legacy tolerates it.
        var nulled = await sdk.OptionalNullableFields.ObjectWithOptionalTrueNullableFalseFieldAsync("medicalRecordNull");
        Assert.Equal(1L, nulled.ObjectWithOptionalTrueNullableFalseField!.Id);
        Assert.Null(nulled.ObjectWithOptionalTrueNullableFalseField.MedicalRecord);
    }

    // Primary: "object-with-optional-false-nullable-false-*" — required non-nullable,
    // both absent and explicit null THROW. Legacy: neither throws, both become null.
    [Fact]
    public async Task ObjectWithOptionalFalseNullableFalse_Legacy()
    {
        var sdk = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var full = await sdk.OptionalNullableFields.ObjectWithOptionalFalseNullableFalseFieldAsync("full");
        Assert.Equal(HttpStatusCode.OK, full.RawResponse.StatusCode);
        Assert.Equal(1L, full.ObjectWithOptionalFalseNullableFalseField!.Id);
        Assert.Equal("fluffy is sick", full.ObjectWithOptionalFalseNullableFalseField.MedicalRecord!.Diagnosis);

        // Primary throws here (Required.Always, missing key); legacy tolerates it.
        var absent = await sdk.OptionalNullableFields.ObjectWithOptionalFalseNullableFalseFieldAsync("medicalRecordAbsent");
        Assert.Equal(1L, absent.ObjectWithOptionalFalseNullableFalseField!.Id);
        Assert.Null(absent.ObjectWithOptionalFalseNullableFalseField.MedicalRecord);

        // Primary throws here (Required.Always, explicit null); legacy tolerates it.
        var nulled = await sdk.OptionalNullableFields.ObjectWithOptionalFalseNullableFalseFieldAsync("medicalRecordNull");
        Assert.Equal(1L, nulled.ObjectWithOptionalFalseNullableFalseField!.Id);
        Assert.Null(nulled.ObjectWithOptionalFalseNullableFalseField.MedicalRecord);
    }

    // pre-existing legacy behavior (presenceAwareJsonSerialization: false) with no
    // DateParseHandling.None in JsonSerializerSettings(): an untyped `object` field coerces
    // a date-time-shaped string into a DateTime instead of preserving the raw string.
    [Fact]
    public void Legacy_UntypedField_CoercesDateShapedString()
    {
        var model = JsonConvert.DeserializeObject<UntypedFieldModel>(
            "{\"val\":\"2024-01-02T03:04:05Z\"}",
            Utilities.GetDefaultJsonDeserializerSettings()
        )!;

        // Coercion: the value comes back as a boxed DateTime, not the raw string.
        Assert.IsType<System.DateTime>(model.Val);
        Assert.NotEqual("2024-01-02T03:04:05Z", model.Val);
        Assert.Equal(
            new System.DateTime(2024, 1, 2, 3, 4, 5, System.DateTimeKind.Utc),
            ((System.DateTime)model.Val!).ToUniversalTime()
        );
    }

    private class UntypedFieldModel
    {
        [JsonProperty("val")]
        public object? Val { get; set; }
    }
}
