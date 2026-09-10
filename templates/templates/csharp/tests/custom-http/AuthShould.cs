using Xunit;
using Speakeasy.CustomHttp;
using Speakeasy.CustomHttp.Models.Components;
using Speakeasy.CustomHttp.Models.Requests;
using System.Collections.Generic;
using System.Threading.Tasks;
using System.Net;

public class AuthShould
{
    [Fact]
    public async Task CustomHttpSchemeOnly()
    {
        CommonHelpers.RecordTest("auth-custom-security-scheme-only");

        var testScopes = new List<string> { "read:products", "write:products" };

        var sdk = new SDK(
            serverUrl: CommonHelpers.ApiTestServiceUrl,
            customHttp: new SchemeCustomHTTPSecurity()
            {
                UserID = 54321,
                Role = Role.Manager,
                Passphrase = "secure-passphrase-123",
                AccessCode = 104,
                Scopes = testScopes
            }
        );

        var res = await sdk.Auth.CustomHttpOnlyAsync();

        Assert.NotNull(res);
        Assert.NotNull(res.Object);
        Assert.Equal("access_granted", res.Object.Grant);
        Assert.Equal(testScopes, res.Object.Scopes);
    }
}
