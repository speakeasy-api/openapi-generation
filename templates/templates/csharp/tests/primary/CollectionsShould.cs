using Xunit;
using Openapi;
using Openapi.Models.Operations;
using System;
using System.Collections.Generic;
using System.Net;
using System.Threading.Tasks;

public class CollectionsShould
{

    [Fact]
    public async Task CollectionsContainingNull()
    {
        CommonHelpers.RecordTest("collections-containing-null");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Collections.CollectionsContainingNullAsync(
            request: new CollectionsContainingNullNullishCollections
            {
                RequiredArray = new List<string> { "foo", null },
                RequiredMap = new Dictionary<string, RequiredMap>
                {
                    { "foo", null },
                    { "bar", RequiredMap.CreateInteger(123) },
                    { "str", RequiredMap.CreateStr("123") }
                },
                OptionalArray = new List<string> { "foo", null },
                OptionalMap = new Dictionary<string, OptionalMap>
                {
                    { "foo", null },
                    { "bar", OptionalMap.CreateInteger(123) },
                    { "str", OptionalMap.CreateStr("123") }
                },
                ArrayOfNullUnion = new List<ArrayOfNullUnion>
                {
                    ArrayOfNullUnion.CreateStr("foo"),
                    ArrayOfNullUnion.CreateBoolean(true),
                    null
                },
                MapOfNullUnion = new Dictionary<string, MapOfNullUnion>
                {
                    { "foo", null },
                    { "bar", MapOfNullUnion.CreateInteger(123) },
                    { "boo", MapOfNullUnion.CreateBoolean(true) }
                },
            }

        );

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Object);

        Assert.Equal(res.Object.Json["requiredArray"], new List<object> { "foo", null });

        var requiredMap = res.Object.Json["requiredMap"] as Dictionary<string, object>;
        Assert.Null(requiredMap["foo"]);
        Assert.Equal(requiredMap["bar"], Convert.ToInt64(123));
        Assert.Equal(requiredMap["str"], "123");

        Assert.Equal(res.Object.Json["optionalArray"], new List<object> { "foo", null });

        var optionalMap = res.Object.Json["optionalMap"] as Dictionary<string, object>;
        Assert.Null(optionalMap["foo"]);
        Assert.Equal(optionalMap["bar"], Convert.ToInt64(123));
        Assert.Equal(optionalMap["str"], "123");

        Assert.Equal(res.Object.Json["arrayOfNullUnion"], new List<object> { "foo", true, null });

        var mapOfNullUnion = res.Object.Json["mapOfNullUnion"] as Dictionary<string, object>;
        Assert.Null(mapOfNullUnion["foo"]);
        Assert.Equal(mapOfNullUnion["bar"], Convert.ToInt64(123));
        Assert.Equal(mapOfNullUnion["boo"], true);
    }
}
