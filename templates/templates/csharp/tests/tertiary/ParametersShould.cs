using System;
using System.Net.Http;
using System.Threading.Tasks;
using System.Collections.Generic;
using Xunit;
using No_Security.API;
using No_Security.API.Models.Operations;
using System.Net.Http.Headers;

public class ParametersShould
{
    [Fact]
    public async Task ParametersOrderingBodyFirst()
    {

        CommonHelpers.RecordTest("parameters-ordering-body-first");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.FlatParametersOrderingAsync(
            new FlatParametersOrderingRequestBody() {},
            "requiredHeaderParam",
            true
        );
        Assert.Equal(200, res.StatusCode);

        var res2 = await sdk.Parameters.FlatParametersOrderingAsync(
            new FlatParametersOrderingRequestBody() {},
            "requiredHeaderParam",
            true,
            "requiredPathParam",
            true,
            1
        );
        Assert.Equal(200, res2.StatusCode);
    }

    [Fact]
    public async Task FlatParametersOrderingUsingOptionalRequestBodyAsync()
    {

        CommonHelpers.RecordTest("parameters-ordering-using-optional-request-body-body-first");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.FlatParametersOrderingUsingOptionalRequestBodyAsync(
            "requiredHeaderParam",
            true
        );
        Assert.Equal(200, res.StatusCode);

        // TODO: When flatteningOrder: body-first  is used we should have the body param before
        // path param. This tests catches this discrepancy. In this method pathParam comes before RequestBody
        var res2 = await sdk.Parameters.FlatParametersOrderingUsingOptionalRequestBodyAsync(
            "requiredHeaderParam",
            true,
            "requiredPathParam",
            null,
            true,
            1
        );
        Assert.Equal(200, res2.StatusCode);
    }
}
