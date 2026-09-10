using Xunit;
using System;
using System.Linq;
using System.Net;
using System.Threading.Tasks;
using System.Collections.Generic;
using Speakeasy.Client.Credentials;
using Speakeasy.Client.Credentials.Hooks;
using Speakeasy.Client.Credentials.Models.Components;

public class ClientCredentialsBasicHookShould
{
    [Fact]
    public async Task TestClientCredentialsBasicHookSuccessfullyAuthenticates()
    {
        CommonHelpers.RecordTest("hooks-client-credentials-basic-success");

        var log = new List<CommonHelpers.RequestLogEntry>();
        var sdk = new SDK(
            serverUrl: CommonHelpers.ApiTestServiceUrl,
            client: new CommonHelpers.RequestRecorderClient(log),
            security: new Security()
            {
                ClientCredentials = new SchemeClientCredentials()
                {
                    ClientID = "speakeasy-sdks",
                    ClientSecret = $"supersecret-{CommonHelpers.RandSeq(10)}",
                    Audience = "",
                }
            }
        );

        var res = await sdk.Hooks.AuthenticatedRequestAsync(null);
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        // Verify that client_id and client_secret are NOT in the request body
        var tokenRequests = log.Where(e => e.RequestUri?.PathAndQuery.Contains("/clientcredentials/token") == true).ToList();
        Assert.Single(tokenRequests);

        var tokenRequest = tokenRequests.First();
        Assert.DoesNotContain("client_id=", tokenRequest.RequestBody);
        Assert.DoesNotContain("client_secret=", tokenRequest.RequestBody);

        // Verify that Authorization header with Basic auth is present
        Assert.True(tokenRequest.RequestHeaders.ContainsKey("Authorization"), "Authorization header should be present");
        Assert.StartsWith("Basic ", tokenRequest.RequestHeaders["Authorization"].First());
    }

    [Fact]
    public async Task TestClientCredentialsBasicHookSuccessfullyAuthenticatesGlobalServer()
    {
        CommonHelpers.RecordTest("hooks-client-credentials-basic-success-global-server");

        var log = new List<CommonHelpers.RequestLogEntry>();
        var sdk = new SDK(
            client: new CommonHelpers.RequestRecorderClient(log),
            security: new Security()
            {
                ClientCredentials = new SchemeClientCredentials()
                {
                    ClientID = "speakeasy-sdks",
                    ClientSecret = $"supersecret-{CommonHelpers.RandSeq(10)}",
                    Audience = "",
                }
            }
        );

        var res = await sdk.Hooks.AuthenticatedRequestGlobalServerAsync(null);
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        // Verify Basic auth is used
        var tokenRequests = log.Where(e => e.RequestUri?.PathAndQuery.Contains("/clientcredentials/token") == true).ToList();
        Assert.Single(tokenRequests);

        var tokenRequest = tokenRequests.First();
        Assert.DoesNotContain("client_id=", tokenRequest.RequestBody);
        Assert.DoesNotContain("client_secret=", tokenRequest.RequestBody);
        Assert.True(tokenRequest.RequestHeaders.ContainsKey("Authorization"), "Authorization header should be present");
        Assert.StartsWith("Basic ", tokenRequest.RequestHeaders["Authorization"].First());
    }

    [Fact]
    public async Task TestClientCredentialsBasicHookSuccessfullyAuthenticatesWithAltTokenURL()
    {
        CommonHelpers.RecordTest("hooks-client-credentials-basic-success-alt-token-url");

        var log = new List<CommonHelpers.RequestLogEntry>();
        var tokenUrl = "/clientcredentials/alt/token";
        var sdk = new SDK(
            serverUrl: CommonHelpers.ApiTestServiceUrl,
            client: new CommonHelpers.RequestRecorderClient(log),
            security: new Security()
            {
                ClientCredentials = new SchemeClientCredentials()
                {
                    ClientID = "speakeasy-sdks",
                    ClientSecret = $"supersecret-{CommonHelpers.RandSeq(10)}",
                    TokenURL = tokenUrl,
                    Scopes = new List<string> { "alt:one", ClientCredentialsOAuth2Scope.AltTwo.Value() },
                    Audience = "",
                }
            }
        );

        var res = await sdk.Hooks.AuthenticatedRequestAsync(null);
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        // since the token is already expired, a new one should be requested
        res = await sdk.Hooks.AuthenticatedRequestAsync(null);
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var tokenRequests = log.Where(e => e.RequestUri?.PathAndQuery == tokenUrl && e.RequestBody.Contains("scope=alt%3Aone+alt%3Atwo")).ToList();
        Assert.Equal(2, tokenRequests.Count());

        // Verify all token requests use Basic auth
        foreach (var tokenRequest in tokenRequests)
        {
            Assert.DoesNotContain("client_id=", tokenRequest.RequestBody);
            Assert.DoesNotContain("client_secret=", tokenRequest.RequestBody);
            Assert.True(tokenRequest.RequestHeaders.ContainsKey("Authorization"), "Authorization header should be present");
            Assert.StartsWith("Basic ", tokenRequest.RequestHeaders["Authorization"].First());
        }
    }

    [Fact]
    public async Task TestClientCredentialsBasicHookConcurrentRequests()
    {
        var log = new List<CommonHelpers.RequestLogEntry>();
        var sdk = new SDK(
            serverUrl: CommonHelpers.ApiTestServiceUrl,
            client: new CommonHelpers.RequestRecorderClient(log),
            security: new Security()
            {
                ClientCredentials = new SchemeClientCredentials()
                {
                    ClientID = "speakeasy-sdks",
                    ClientSecret = $"supersecret-{CommonHelpers.RandSeq(10)}",
                    Audience = "",
                }
            }
        );

        int numConcurrentRequests = 10;
        var tasks = new Task[numConcurrentRequests];

        for (int i = 0; i < numConcurrentRequests; i++)
        {
            tasks[i] = Task.Run(async () =>
            {
                var res = await sdk.Hooks.AuthenticatedRequestAsync(null);
                Assert.NotNull(res);
                Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
            });
        }

        await Task.WhenAll(tasks);

        // Verify that Basic auth was used in token requests
        var tokenRequests = log.Where(e => e.RequestUri?.PathAndQuery.Contains("/clientcredentials/token") == true).ToList();
        Assert.NotEmpty(tokenRequests);

        foreach (var tokenRequest in tokenRequests)
        {
            Assert.DoesNotContain("client_id=", tokenRequest.RequestBody);
            Assert.DoesNotContain("client_secret=", tokenRequest.RequestBody);
            Assert.True(tokenRequest.RequestHeaders.ContainsKey("Authorization"), "Authorization header should be present");
            Assert.StartsWith("Basic ", tokenRequest.RequestHeaders["Authorization"].First());
        }
    }
}
