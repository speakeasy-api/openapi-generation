using Xunit;
using Company.Product.Feature.Subnamespace;
using Company.Product.Feature.Subnamespace.Models.Shared;
using Company.Product.Feature.Subnamespace.Models.Operations;
using System.Collections.Generic;
using System.Threading.Tasks;

public class AuthShould
{
    [Fact]
    public async Task TestGlobalSecurityFlattening()
    {
        CommonHelpers.RecordTest("auth-global-security-flattening");

        var sdk = new SDK(serverUrl: CommonHelpers.HttpBinUrl, apiKeyAuth: "Bearer testToken");

        var res = await sdk.Auth.ApiKeyAuthGlobalAsync();

        Assert.NotNull(res);
        Assert.Equal(200, res.StatusCode);
        Assert.True(res.Token.Authenticated);
        Assert.Equal("testToken", res.Token.Token);

        sdk = SDK.Builder().WithServerUrl(CommonHelpers.HttpBinUrl).WithApiKeyAuth("Bearer testToken").Build();

        res = await sdk.Auth.ApiKeyAuthGlobalAsync();

        Assert.NotNull(res);
        Assert.Equal(200, res.StatusCode);
        Assert.True(res.Token.Authenticated);
        Assert.Equal("testToken", res.Token.Token);
    }

    [Fact]
    public async Task TestGlobalSecurityFlatteningCallback()
    {
        CommonHelpers.RecordTest("auth-global-security-flattening-callback");

        var ex = Assert.Throws<System.ArgumentException>(() => new SDK(serverUrl: CommonHelpers.HttpBinUrl));
        Assert.Equal("apiKeyAuth and apiKeyAuthSource cannot both be null", ex.Message);

        var sdk = new SDK(serverUrl: CommonHelpers.HttpBinUrl, apiKeyAuthSource: () => "Bearer testToken");

        var res = await sdk.Auth.ApiKeyAuthGlobalAsync();

        Assert.NotNull(res);
        Assert.Equal(200, res.StatusCode);
        Assert.True(res.Token.Authenticated);
        Assert.Equal("testToken", res.Token.Token);

        ex = Assert.Throws<System.ArgumentException>(() => SDK.Builder().WithServerUrl(CommonHelpers.HttpBinUrl).Build());
        Assert.Equal("securitySource cannot be null. One of `ApiKeyAuth` or `apiKeyAuthSource` needs to be defined.", ex.Message);
        sdk = SDK.Builder().WithServerUrl(CommonHelpers.HttpBinUrl).WithApiKeyAuthSource(() => "Bearer testToken").Build();

        res = await sdk.Auth.ApiKeyAuthGlobalAsync();

        Assert.NotNull(res);
        Assert.Equal(200, res.StatusCode);
        Assert.True(res.Token.Authenticated);
    }
}
