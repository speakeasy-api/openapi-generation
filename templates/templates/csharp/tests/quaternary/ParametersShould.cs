using System;
using System.Net.Http;
using System.Threading.Tasks;
using System.Collections.Generic;
using Xunit;
using Company.Product.Feature.Subnamespace;
using Company.Product.Feature.Subnamespace.Models.Operations;
using Company.Product.Feature.Subnamespace.Models.Shared;
using System.Net.Http.Headers;

public class ParametersShould
{
    [Fact]
    public async Task ParametersOrderingWithLegacyFlatteningOrder()
    {
        CommonHelpers.RecordTest("parameters-ordering-with-legacy-flattening-order");

        var sdk = new SDK(serverUrl: CommonHelpers.HttpBinUrl, apiKeyAuth: "Token YOUR_API_KEY");

        var res = await sdk.Parameters.FlatParametersOrderingAsync(
            new FlatParametersOrderingRequestBody() {},
            "requiredHeaderParam",
            true,
            "requiredPathParam",
            true,
            1
        );
        Assert.Equal(200, res.StatusCode);

        var res2 = await sdk.Parameters.FlatParametersOrderingAsync(
            new FlatParametersOrderingRequestBody() {},
            "requiredHeaderParam",
            true
        );
        Assert.Equal(200, res2.StatusCode);

    }

    [Fact]
    public async Task ParametersOrderingUsingOptionalRequestBodyWithLegacyFlatteningOrder()
    {
        CommonHelpers.RecordTest("parameters-ordering-using-optional-request-body-with-legacy-flattening-order");

        var sdk = new SDK(serverUrl: CommonHelpers.HttpBinUrl, apiKeyAuth: "Token YOUR_API_KEY");

        var res = await sdk.Parameters.FlatParametersOrderingUsingOptionalRequestBodyAsync(
            "requiredHeaderParam",
            true
        );
        Assert.Equal(200, res.StatusCode);

        var res2 = await sdk.Parameters.FlatParametersOrderingUsingOptionalRequestBodyAsync(
            "requiredHeaderParam",
            true,
            "requiredPathParam",
            null,
            false,
            2
        );
        Assert.Equal(200, res2.StatusCode);
    }
}
