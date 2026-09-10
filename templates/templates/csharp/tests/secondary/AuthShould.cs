using HoistedSecurity;
using HoistedSecurity.Models.Operations;
using HoistedSecurity.Models.Shared;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Threading.Tasks;
using Xunit;

public class AuthShould
{
    [Fact]
    public async Task NoAuth()
    {
        CommonHelpers.RecordTest("auth-hoisted-no-auth-retained");

        var sdk = new SDK(serverUrl: CommonHelpers.HttpBinUrl);

        await sdk.Auth.NoAuthAsync();

        sdk = SDK.Builder().WithServerUrl(CommonHelpers.HttpBinUrl).Build();

        await sdk.Auth.NoAuthAsync();
    }

    [Fact]
    public async Task BasicAuth()
    {
        CommonHelpers.RecordTest("auth-hoisted-basic-auth");

        var sdk = new SDK(serverUrl: CommonHelpers.HttpBinUrl, security: new Security() { Username = "testUser", Password = "testPass" });

        var res = await sdk.Auth.BasicAuthAsync("testPass", "testUser");

        Assert.NotNull(res);
        Assert.True(res.Authenticated);

        sdk = SDK.Builder().WithServerUrl(CommonHelpers.HttpBinUrl).WithSecurity(new Security() { Username = "testUser", Password = "testPass" }).Build();

        res = await sdk.Auth.BasicAuthAsync("testPass", "testUser");

        Assert.NotNull(res);
        Assert.True(res.Authenticated);
    }

    [Fact]
    public async Task MultipleMixedOptionsAuth()
    {
        CommonHelpers.RecordTest("auth-hoisted-operation-auth-retained");

        var sdk = new SDK(serverUrl: CommonHelpers.HttpBinUrl);

        await sdk.AuthNew.MultipleMixedOptionsAuthAsync(
            new MultipleMixedOptionsAuthSecurity()
            {
                BasicAuth = new SchemeBasicAuth() { Username = "testUser", Password = "testPass" },
            },
            new AuthServiceRequestBody()
            {
                BasicAuth = new BasicAuth() { Username = "testUser", Password = "testPass" },
            }
        );
    }

    [Fact]
    public async Task OperationLevelOauth2()
    {
        CommonHelpers.RecordTest("auth-operation-level-oauth2");

        var log = new List<CommonHelpers.RequestLogEntry>();
        var sdk = new SDK(serverUrl: CommonHelpers.HttpBinUrl, client: new CommonHelpers.RequestRecorderClient(log));
        var clientID = "speakeasy-sdks";
        var clientSecret = $"supersecret-{CommonHelpers.RandSeq(10)}";

        // A token should be requested with 'read', 'write' and 'erase' scopes.
        await sdk.Hooks.AuthenticatedRequestAsync(
            new AuthenticatedRequestSecurity()
            {
                ClientID = clientID,
                ClientSecret = clientSecret,
                Audience = ""
            }
        );

        // Requires 'read' and 'write' scopes. The same token should be reused
        // since [read, write, erase] is a superset of [read, write].
        await sdk.Hooks.AuthenticatedRequestUnflattenedAsync(
            new AuthenticatedRequestUnflattenedSecurity()
            {
                ClientCredentials = new SchemeClientCredentials
                {
                    ClientID = clientID,
                    ClientSecret = clientSecret,
                    Audience = ""
                }
            }
        );

        var tokenUrl = "/clientcredentials/token?expires_in=90";
        var tokenRequests = log.Where(e => e.RequestUri?.PathAndQuery == tokenUrl).ToList();
        Assert.Single(tokenRequests);
        Assert.Contains("scope=read+write+erase", tokenRequests.First().RequestBody);
    }

    [Fact]
    public async Task CustomSecuritySchemeAppId()
    {
        CommonHelpers.RecordTest("auth-custom-security-scheme-app-id");

        var sdk = new SDK(serverUrl: CommonHelpers.HttpBinUrl);

        await sdk.AuthNew.CustomSchemeAppIdAsync(security: new CustomSchemeAppIdSecurity() {
            AppId = "testAppID",
            Secret = "testSecret"
        });
    }

}
