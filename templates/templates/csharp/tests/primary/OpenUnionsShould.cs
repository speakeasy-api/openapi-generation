using System.Collections.Generic;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;
using Openapi.Models.Shared;
using Openapi.Utils;
using Xunit;

public class OpenUnionsShould
{
    private static readonly JsonSerializerSettings Settings =
        Utilities.GetDefaultJsonDeserializerSettings();

    private static Vehicle Deserialize(string json)
    {
        return JsonConvert.DeserializeObject<Vehicle>(json, Settings)!;
    }

    [Fact]
    public void ParseKnownVariant()
    {
        CommonHelpers.RecordTest("open-union-known-variant");

        var car = Deserialize("{\"vehicleType\":\"car\",\"wheelsType\":\"four\"}");
        Assert.Equal(VehicleType.Car.ToString(), car.Type.ToString());
        Assert.NotNull(car.Car);
        Assert.False(car.IsUnknown());
        Assert.Null(car.UnknownRaw);

        var bike = Deserialize(
            "{\"vehicleType\":\"bike\",\"wheelsType\":\"two\",\"colour\":\"red\"}"
        );
        Assert.Equal(VehicleType.Bike.ToString(), bike.Type.ToString());
        Assert.NotNull(bike.Bike);
        Assert.Equal("red", bike.Bike!.Colour);
        Assert.False(bike.IsUnknown());

        var restored = Deserialize(JsonConvert.SerializeObject(car));
        Assert.Equal(VehicleType.Car.ToString(), restored.Type.ToString());
        Assert.NotNull(restored.Car);
    }

    [Fact]
    public void PreserveUnknownDiscriminatorAsRaw()
    {
        CommonHelpers.RecordTest("open-union-unknown-discriminator");

        var v = Deserialize(
            "{\"vehicleType\":\"spaceship\",\"thrust\":9000,"
                + "\"launchedAt\":\"2024-01-02T03:04:05+02:00\"}"
        );
        Assert.Equal(VehicleType.Unknown.ToString(), v.Type.ToString());
        Assert.True(v.IsUnknown());
        Assert.Null(v.Car);
        Assert.Null(v.Bike);
        Assert.NotNull(v.UnknownRaw);

        var raw = (JObject)v.UnknownRaw!;
        Assert.Equal("spaceship", (string?)raw["vehicleType"]);
        Assert.Equal(9000, (int?)raw["thrust"]);
        // Date-time-shaped strings stay raw strings, offset untouched.
        Assert.Equal(JTokenType.String, raw["launchedAt"]!.Type);
        Assert.Equal("2024-01-02T03:04:05+02:00", (string?)raw["launchedAt"]);

        var serialized = JsonConvert.SerializeObject(v);
        Assert.Contains("2024-01-02T03:04:05+02:00", serialized);
        // JToken.Parse would date-coerce launchedAt and break DeepEquals; parse
        // through the SDK settings (DateParseHandling.None) like the SDK does.
        var reparsed = JsonConvert.DeserializeObject<JToken>(serialized, Settings)!;
        Assert.True(JToken.DeepEquals(reparsed, raw));
        var restored = Deserialize(serialized);
        Assert.True(restored.IsUnknown());
    }

    [Fact]
    public void TreatMissingDiscriminatorAsUnknown()
    {
        CommonHelpers.RecordTest("open-union-missing-discriminator");

        var v = Deserialize("{\"wheelsType\":\"four\"}");
        Assert.True(v.IsUnknown());
        Assert.Equal(VehicleType.Unknown.ToString(), v.Type.ToString());
    }

    [Fact]
    public void TreatInvalidPayloadShapesAsUnknown()
    {
        CommonHelpers.RecordTest("open-union-invalid-payload");

        foreach (var payload in new[] { "\"hello\"", "42", "[1,2]", "null" })
        {
            var v = Deserialize(payload);
            Assert.True(v.IsUnknown());
            Assert.NotNull(v.UnknownRaw);

            // Primitive raw values must round-trip as valid JSON of the same shape.
            var serialized = JsonConvert.SerializeObject(v);
            Assert.True(JToken.DeepEquals(JToken.Parse(serialized), JToken.Parse(payload)));
        }
    }

    [Fact]
    public void PreserveStrictErrorForKnownDiscriminatorWithInvalidSchema()
    {
        CommonHelpers.RecordTest("open-union-known-disc-invalid-schema");

        // Known "bike" discriminator with the required "colour" field missing: the
        // known variant keeps strict validation and must not be masked as Unknown.
        Assert.ThrowsAny<JsonException>(
            () => Deserialize("{\"vehicleType\":\"bike\",\"wheelsType\":\"two\"}")
        );
    }

    [Fact]
    public void HandleMixedKnownAndUnknownInCollections()
    {
        CommonHelpers.RecordTest("open-union-embedded");

        var vehicles = JsonConvert.DeserializeObject<List<Vehicle>>(
            "[{\"vehicleType\":\"car\",\"wheelsType\":\"four\"},"
                + "{\"vehicleType\":\"spaceship\",\"thrust\":9000},"
                + "{\"vehicleType\":\"bike\",\"wheelsType\":\"two\",\"colour\":\"green\"}]",
            Settings
        )!;
        Assert.Equal(3, vehicles.Count);

        Assert.False(vehicles[0].IsUnknown());
        Assert.NotNull(vehicles[0].Car);

        Assert.True(vehicles[1].IsUnknown());
        Assert.Equal("spaceship", (string?)vehicles[1].UnknownRaw!["vehicleType"]);

        Assert.False(vehicles[2].IsUnknown());
        Assert.Equal("green", vehicles[2].Bike!.Colour);
    }

    [Fact]
    public void ThrowOnUnknownWithoutRawPayload()
    {
        // In C#, the Unknown sentinel is reachable without a captured payload by
        // bypassing CreateUnknown; serialization must fail loud since there is
        // no raw token to replay.
        var vehicle = new Vehicle(VehicleType.FromString("UNKNOWN"));
        Assert.True(vehicle.IsUnknown());
        Assert.Null(vehicle.UnknownRaw);

        var ex = Assert.Throws<System.InvalidOperationException>(
            () => JsonConvert.SerializeObject(vehicle)
        );
        Assert.Equal(
            "Unknown union value has no raw payload; construct via CreateUnknown(JToken).",
            ex.Message
        );
    }

    [Fact]
    public void OpenUnionHelperNameConflicts()
    {
        // Spec-derived variant properties keep their names. The coordinated
        // open-union helper family moves aside together.
        var known = OpenUnionHelperNameCollision.CreateUnknown(
            new Unknown { Value = "known variant" }
        );
        Assert.NotNull(known.Unknown);
        Assert.False(known.IsUnknown1());

        var rawVariant = OpenUnionHelperNameCollision.CreateUnknownRaw(
            new UnknownRaw { Raw = "known raw variant" }
        );
        Assert.NotNull(rawVariant.UnknownRaw);
        var predicateVariant = OpenUnionHelperNameCollision.CreateIsUnknown(
            new IsUnknown { Flag = true }
        );
        Assert.NotNull(predicateVariant.IsUnknown);

        var raw = JObject.Parse("{\"future\":true}");
        var unknown = OpenUnionHelperNameCollision.CreateUnknown(raw);
        Assert.True(unknown.IsUnknown1());
        Assert.Same(raw, unknown.Unknown1Raw);
        Assert.True(
            JToken.DeepEquals(
                raw,
                JToken.Parse(JsonConvert.SerializeObject(unknown))
            )
        );

    }

    [Fact]
    public void LoseToGenuineSiblingMatchInSmartUnion()
    {
        CommonHelpers.RecordTest("open-union-smart-union-interop");

        // An open discriminated union candidate can always "succeed" by producing
        // Unknown, so scoring demotes an Unknown result below any sibling with a
        // genuine match.
        var pool = new UnionCandidatePool<int>();
        pool.Add<Vehicle>("A", 0, 0);
        pool.Add<string>("B", 1, 1);

        var (stringWinner, stringValue) = pool.PickBest(JToken.Parse("\"hello\""), Settings);
        Assert.Equal(typeof(string), stringWinner?.Type);
        Assert.Equal("hello", stringValue);

        // A known discriminator still wins for the union candidate.
        var (vehicleWinner, vehicleValue) = pool.PickBest(
            JToken.Parse("{\"vehicleType\":\"car\",\"wheelsType\":\"four\"}"),
            Settings
        );
        Assert.Equal(typeof(Vehicle), vehicleWinner?.Type);
        Assert.False(((Vehicle)vehicleValue!).IsUnknown());
    }
}
