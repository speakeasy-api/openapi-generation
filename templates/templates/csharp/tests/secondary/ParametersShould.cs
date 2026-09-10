using System;
using System.Net.Http;
using System.Threading.Tasks;
using System.Collections.Generic;
using Xunit;
using HoistedSecurity;
using HoistedSecurity.Models.Operations;
using HoistedSecurity.Models.Shared;
using System.Net.Http.Headers;


public class ParametersShould
{
    [Fact]
    public async Task ParametersOrderingParametersFirst()
    {

        CommonHelpers.RecordTest("parameters-ordering-parameters-first");

        var sdk = new SDK(serverUrl: CommonHelpers.HttpBinUrl, security: new Security() {
            Password = "YOUR_PASSWORD",
            Username = "YOUR_USERNAME",
        });

        await sdk.Parameters.FlatParametersOrderingAsync(
            "requiredHeaderParam",
            true,
            new FlatParametersOrderingRequestBody() {}
        );

        await sdk.Parameters.FlatParametersOrderingAsync(
            "requiredHeaderParam",
            true,
            new FlatParametersOrderingRequestBody() {},
            "requiredPathParam",
            false,
            1
        );
    }


    [Fact]
    public async Task ParametersOrderingUsingOptionalRequestBodyParametersFirst()
    {
        CommonHelpers.RecordTest("parameters-ordering-using-optional-request-body-parameters-first");

        var sdk = new SDK(serverUrl: CommonHelpers.HttpBinUrl, security: new Security() {
            Password = "YOUR_PASSWORD",
            Username = "YOUR_USERNAME",
        });

        await sdk.Parameters.FlatParametersOrderingUsingOptionalRequestBodyAsync(
            "requiredHeaderParam",
            true
        );

        await sdk.Parameters.FlatParametersOrderingUsingOptionalRequestBodyAsync(
            "requiredHeaderParam",
            true,
            "requiredPathParam",
            false,
            1,
            null
        );
    }
}


