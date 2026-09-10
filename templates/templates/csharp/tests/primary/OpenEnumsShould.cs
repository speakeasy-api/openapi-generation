using Xunit;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Web;
using System.Threading.Tasks;
using Openapi;
using Openapi.Models.Operations;
using Openapi.Models.Shared;
using Openapi.Utils;
using Newtonsoft.Json;

public class OpenEnumsShould
{
    [Fact]
    public async Task OpenEnumsRoundTrip()
    {
        CommonHelpers.RecordTest("open-enums-round-trip");
        var sdk = SDK.Builder().WithServerUrl(Helpers.HttpBinUrl).Build();

        // HeroWidth.Values() grows as unrecognized values are observed and the
        // instance cache is shared across tests, so only assert containment —
        // exact counts would make this test order-dependent.
        HeroWidth[] knownHeroWidths = {
            HeroWidth.OneThousandAndEighty,
            HeroWidth.SevenHundredAndTwenty,
            HeroWidth.FourHundredAndEighty
        };
        foreach (var known in knownHeroWidths)
        {
            Assert.Contains(known, HeroWidth.Values());
        }

        // Round trip with Known values
        var resKnown = await sdk.Enums.EnumsPostOpenEnumUnrecognizedAsync(new ThemeRequestOpaque {
                Color = "red",
                Icon = "tick",
                HeroWidth = 720
            });
        Assert.Equal(HttpStatusCode.OK, resKnown.HttpMeta.Response.StatusCode);
        Assert.Equal(Color.Red, resKnown.ThemeResponse.Json.Color);
        Assert.Equal(Icon.Tick, resKnown.ThemeResponse.Json.Icon);
        Assert.Equal(HeroWidth.SevenHundredAndTwenty, resKnown.ThemeResponse.Json.HeroWidth);

        // Round trip with Unknown values
        var resUnKnown = await sdk.Enums.EnumsPostOpenEnumUnrecognizedAsync(new ThemeRequestOpaque {
                Color = "purple",
                Icon = "tick",
                HeroWidth = 2160
            });
        Assert.Equal(HttpStatusCode.OK, resKnown.HttpMeta.Response.StatusCode);
        var updatedHeroWidths = HeroWidth.Values();
        Assert.Contains(HeroWidth.Of(2160), updatedHeroWidths);

        Assert.Equal(Color.Of("purple"), resUnKnown.ThemeResponse.Json.Color);
        Assert.Equal(Icon.Tick, resUnKnown.ThemeResponse.Json.Icon);
        Assert.Equal(HeroWidth.Of(2160), resUnKnown.ThemeResponse.Json.HeroWidth);

        // Serialization/Deserialization of Known values
        var themeKnown = new Theme {
            Color = Color.Red,
            Icon = Icon.Tick,
            HeroWidth = HeroWidth.FourHundredAndEighty
        };

        var jsonKnown = await Helpers.GetSerializedBodyJson(themeKnown);
        Assert.Equal("{\"color\":\"red\",\"heroWidth\":480,\"icon\":\"tick\"}", jsonKnown);

        var deserializedKnown = JsonConvert.DeserializeObject<Theme>(jsonKnown);
        Assert.Equal(themeKnown.Color, deserializedKnown.Color);
        Assert.Equal(themeKnown.Icon, deserializedKnown.Icon);
        Assert.Equal(themeKnown.HeroWidth, deserializedKnown.HeroWidth);

        // Serialization/Deserialization of Unknown values
        var themeUnknown = new Theme {
            Color = Color.Of("yellow"),
            Icon = Icon.Tick,
            HeroWidth = HeroWidth.Of(514)
        };

        Assert.False(themeUnknown.Color.IsKnown());
        Assert.Equal("yellow", themeUnknown.Color.Value);
        Assert.False(themeUnknown.HeroWidth.IsKnown());
        Assert.Equal(514, themeUnknown.HeroWidth.Value);

        var jsonUnknown = await Helpers.GetSerializedBodyJson(themeUnknown);
        Assert.Equal("{\"color\":\"yellow\",\"heroWidth\":514,\"icon\":\"tick\"}", jsonUnknown);

        var deserializedUnknown = JsonConvert.DeserializeObject<Theme>(jsonUnknown);
        Assert.Equal(themeUnknown.Color, deserializedUnknown.Color);
        Assert.Equal(themeUnknown.Icon, deserializedUnknown.Icon);
        Assert.Equal(themeUnknown.HeroWidth, deserializedUnknown.HeroWidth);
    }

    [Fact]
    public void OpenEnumsMethods()
    {
        // 0. Values()
        var knownValues = OpenEnum.Values();
        OpenEnum[] expectedKnownValues = {
            OpenEnum.Of(101),
            OpenEnum.Of(404)
        };
        Assert.Equal(expectedKnownValues, knownValues);

        OpenEnum sixSixSix = 666;

        var updatedValues = OpenEnum.Values();
        OpenEnum[] expectedUpdatedValues = {
            OpenEnum.Of(666),
            OpenEnum.Of(101),
            OpenEnum.Of(404)
        };
        Assert.Equal(expectedUpdatedValues, updatedValues);

        // 1. Known()
        Assert.True(Color.Red.IsKnown());
        Assert.True(Color.Of("red").IsKnown());
        Assert.True(OpenEnum.Of(101).IsKnown());
        Assert.True(OpenEnum.Of(101).IsKnown());
        Assert.False(sixSixSix.IsKnown());


        // 2. ToString()
        Assert.Equal("red", Color.Red.ToString());
        Assert.Equal("101", OpenEnum.OneHundredAndOne.ToString());

        Assert.Equal("BLACK", Color.Of("BLACK").ToString());
        OpenEnum oneTwoThree = 123;
        Assert.Equal("123", oneTwoThree.ToString());


        // 3. Equal()
        Color red = "red";
        Assert.Equal(Color.Red, red);
        Assert.Equal(Color.Of("red"), red);
        Assert.True(Color.Red == red);
        Assert.True(red.Equals(Color.Red));

        Color brown = "brown";
        Assert.Equal(Color.Of("brown"), brown);
        Assert.True(brown.Equals(Color.Of("brown")));
        Assert.True(Color.Of("brown") == brown);

        Color maroon = "Brown";
        Assert.False(maroon.Equals(brown));
        Assert.True(maroon != Color.Of("brown"));
    }

    [Fact]
    public async Task OpenEnumsRoundTripStringUnion()
    {
        CommonHelpers.RecordTest("open-enums-round-trip-string-union");
        var sdk = SDK.Builder().WithServerUrl(Helpers.HttpBinUrl).Build();

        var res = await sdk.Enums.EnumsPostOpenEnumUnrecognizedAsync(new ThemeRequestOpaque {
            Color = "purple",
            Icon = "tick",
            HeroWidth = 2160
        });
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        var theme = res.ThemeResponse!.Json;
        Assert.Equal(Color.Of("purple"), theme.Color);
        Assert.Equal(Icon.Tick, theme.Icon);
        Assert.Equal(HeroWidth.Of(2160), theme.HeroWidth);

        // Round trip: feed the received open-enum values back as raw string/int,
        // exactly as a forward-compatible client would relay unrecognized values.
        var roundTrip = await sdk.Enums.EnumsPostOpenEnumUnrecognizedAsync(new ThemeRequestOpaque {
            Color = theme.Color!.ToString(),
            Icon = theme.Icon!.Value,
            HeroWidth = theme.HeroWidth!.Value
        });
        Assert.Equal(HttpStatusCode.OK, roundTrip.HttpMeta.Response.StatusCode);
        var roundTripTheme = roundTrip.ThemeResponse!.Json;
        Assert.Equal(Color.Of("purple"), roundTripTheme.Color);
        Assert.Equal(Icon.Tick, roundTripTheme.Icon);
        Assert.Equal(HeroWidth.Of(2160), roundTripTheme.HeroWidth);
    }

    [Fact]
    public async Task OpenEnumsResponseUsage()
    {
        CommonHelpers.RecordTest("open-enums-response-usage");
        var sdk = SDK.Builder().WithServerUrl(Helpers.HttpBinUrl).Build();

        // forwardCompatibleEnumsByDefault only opens response-used enums: the
        // request-only enum stays a closed C# enum.
        var requestOnly = await sdk.OpenEnums.OpenEnumsResponseUsageRequestOnlyAsync(
            new ObjectWithEnumInRequestOnly {
                Action = EnumUsedInRequestOnly.Create,
                Data = "test data"
            }
        );
        Assert.Equal(HttpStatusCode.OK, requestOnly.HttpMeta.Response.StatusCode);
        Assert.True(typeof(EnumUsedInRequestOnly).IsEnum);

        // Explicitly open enum (x-speakeasy-unknown-values: allow) accepts an
        // unrecognized value and round-trips it.
        var explicitlyOpen = await sdk.OpenEnums.OpenEnumsResponseUsageExplicitlyOpenAsync(
            new ObjectWithEnumExplicitlyOpen {
                Mode = EnumUsedInRequestExplicitlyOpen.Of("delta"),
                Value = "test value"
            }
        );
        Assert.Equal(HttpStatusCode.OK, explicitlyOpen.HttpMeta.Response.StatusCode);
        Assert.False(EnumUsedInRequestExplicitlyOpen.Of("delta").IsKnown());
        Assert.Equal("delta", explicitlyOpen.Object!.Json["mode"]!.ToString());

        // Enum used in both request and response is auto-opened: a known value
        // round-trips, and an unknown value survives instead of throwing.
        var both = await sdk.OpenEnums.OpenEnumsResponseUsageBothRequestAndResponseAsync(
            new ObjectWithEnumInBoth {
                State = EnumUsedInBothRequestAndResponse.Active,
                Description = "known value test"
            }
        );
        Assert.Equal(HttpStatusCode.OK, both.HttpMeta.Response.StatusCode);
        Assert.Equal(EnumUsedInBothRequestAndResponse.Active, both.Object!.Json.State);

        var bothUnknown = await sdk.OpenEnums.OpenEnumsResponseUsageBothRequestAndResponseAsync(
            new ObjectWithEnumInBoth {
                State = EnumUsedInBothRequestAndResponse.Of("hibernating"),
                Description = "unknown value test"
            }
        );
        Assert.Equal(HttpStatusCode.OK, bothUnknown.HttpMeta.Response.StatusCode);
        var bothState = bothUnknown.Object!.Json.State;
        Assert.Equal(EnumUsedInBothRequestAndResponse.Of("hibernating"), bothState);
        Assert.False(bothState.IsKnown());

        // Response-only enum is auto-opened: an unknown value coming back from
        // the server deserializes instead of terminating the response.
        var responseOnly = await sdk.OpenEnums.OpenEnumsResponseUsageResponseOnlyAsync(
            new Dictionary<string, object> {
                ["message"] = "test",
                ["status"] = "unknown_new_value"
            }
        );
        Assert.Equal(HttpStatusCode.OK, responseOnly.HttpMeta.Response.StatusCode);
        var status = responseOnly.Object!.Json.Status;
        Assert.Equal(EnumUsedInResponseOnly.Of("unknown_new_value"), status);
        Assert.False(status.IsKnown());
        Assert.Equal("test", responseOnly.Object!.Json.Message);

        var known = await sdk.OpenEnums.OpenEnumsResponseUsageResponseOnlyAsync(
            new Dictionary<string, object> {
                ["message"] = "foobar",
                ["status"] = "pending"
            }
        );
        Assert.Equal(HttpStatusCode.OK, known.HttpMeta.Response.StatusCode);
        Assert.Equal(EnumUsedInResponseOnly.Pending, known.Object!.Json.Status);
        Assert.True(known.Object!.Json.Status.IsKnown());
    }

    [Fact]
    public void PreserveDateShapedUnknownValues()
    {
        // Presence-aware SDKs read with DateParseHandling.None, so ISO-shaped unknown
        // values reach OpenEnumConverter as raw strings and round-trip unchanged.
        var parsed = JsonConvert.DeserializeObject<Color>(
            "\"2024-01-02T03:04:05Z\"",
            Utilities.GetDefaultJsonDeserializerSettings());
        Assert.NotNull(parsed);
        Assert.False(parsed!.IsKnown());
        Assert.Equal("2024-01-02T03:04:05Z", parsed.Value);
    }


    [Fact]
    public void FractionalValuesForIntegralOpenEnums()
    {
        // long-typed and int-typed open enums must reject fractional wire values
        Assert.Throws<JsonSerializationException>(() =>
            JsonConvert.DeserializeObject<HeroWidth>(
                "719.5", Utilities.GetDefaultJsonDeserializerSettings()));
        Assert.Throws<JsonSerializationException>(() =>
            JsonConvert.DeserializeObject<Int32Enum>(
                "1.7", Utilities.GetDefaultJsonDeserializerSettings()));

        // Integral-valued doubles and long-to-int narrowing still coerce.
        var wholeDouble = JsonConvert.DeserializeObject<HeroWidth>(
            "2160.0", Utilities.GetDefaultJsonDeserializerSettings());
        Assert.Equal(HeroWidth.Of(2160), wholeDouble);
        var narrowed = JsonConvert.DeserializeObject<Int32Enum>(
            "7", Utilities.GetDefaultJsonDeserializerSettings());
        Assert.Equal(Int32Enum.Of(7), narrowed);
    }
}
