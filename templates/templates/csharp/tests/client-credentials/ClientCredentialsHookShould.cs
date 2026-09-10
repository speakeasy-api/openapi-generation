using Polly;
using RestSharp;
using Xunit;
using System;
using System.Linq;
using System.Net;
using System.Threading.Tasks;
using System.Collections.Generic;
using Speakeasy.Client.Credentials;
using Speakeasy.Client.Credentials.Hooks;
using Speakeasy.Client.Credentials.Models.Components;

public class ClientCredentialsHookShould
{
    [Fact]
    public void TestAdditionalDependencies()
    {
        RestClient client = null;
        var policy = Policy.Handle<Exception>().Retry(1);

        policy.Execute(() => {
            client = new RestClient("http://example.com");
        });

        Assert.NotNull(client);
    }

    [Fact]
    public async Task TestClientCredentialsHookSuccessfullyAuthenticates()
    {
        CommonHelpers.RecordTest("hooks-client-credentials-success");

        var sdk = new SDK(serverUrl: CommonHelpers.ApiTestServiceUrl, security: new Security() {
            ClientID = "speakeasy-sdks",
            ClientSecret = $"supersecret-{CommonHelpers.RandSeq(10)}",
        });

        var res = await sdk.Hooks.AuthenticatedRequestAsync(null);
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.False(res.HttpMeta.Request.Headers.Contains("clientId"));
        Assert.False(res.HttpMeta.Request.Headers.Contains("clientSecret"));
    }

    [Fact]
    public async Task TestClientCredentialsHookSuccessfullyAuthenticatesGlobalServer()
    {
        CommonHelpers.RecordTest("hooks-client-credentials-success-global-server");

        var sdk = new SDK(security: new Security() {
            ClientID = "speakeasy-sdks",
            ClientSecret = $"supersecret-{CommonHelpers.RandSeq(10)}",
        });

        var res = await sdk.Hooks.AuthenticatedRequestGlobalServerAsync(null);
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.False(res.HttpMeta.Request.Headers.Contains("clientId"));
        Assert.False(res.HttpMeta.Request.Headers.Contains("clientSecret"));
    }

    [Fact]
    public async Task TestClientCredentialsHookSuccessfullyAuthenticatesWithAltTokenURL()
    {
        CommonHelpers.RecordTest("hooks-client-credentials-success-alt-token-url");

        var log = new List<CommonHelpers.RequestLogEntry>();
        var tokenUrl = "/clientcredentials/alt/token";
        var sdk = new SDK(
            serverUrl: CommonHelpers.ApiTestServiceUrl,
            client: new CommonHelpers.RequestRecorderClient(log),
            security: new Security() {
                ClientID = "speakeasy-sdks",
                ClientSecret = $"supersecret-{CommonHelpers.RandSeq(10)}",
                TokenURL = tokenUrl,
                Scopes = new List<string> { "alt:one", ClientCredentialsOAuth2Scope.AltTwo.Value() }
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
    }

    [Fact]
    public async Task TestClientCredentialsHookNoScopes()
    {
        CommonHelpers.RecordTest("hooks-client-credentials-no-scopes");

        var clientID = "speakeasy-sdks";
        var clientSecret = $"supersecret-{CommonHelpers.RandSeq(10)}";

        var sdk = new SDK(
            serverUrl: CommonHelpers.ApiTestServiceUrl,
            security: new Security() {
                ClientID = clientID,
                ClientSecret = clientSecret
            }
        );

        // expected to fail since the token endpoint requires 'read' and 'write' scopes
        // but the authenticatedRequestNoScopes operation does not specify any.
        var ex = await Assert.ThrowsAsync<Exception>(
            () => sdk.Hooks.AuthenticatedRequestNoScopesAsync(null)
        );
        Assert.NotNull(ex);
        Assert.Equal("An error occurred while calling BeforeRequestAsync hook", ex.Message);
        Assert.Equal("Unexpected status code BadRequest: empty_scopes\n", ex.InnerException?.Message);

        // same check but this time we override the default scopes with an empty list
        sdk = new SDK(
            serverUrl: CommonHelpers.ApiTestServiceUrl,
            security: new Security() {
                ClientID = clientID,
                ClientSecret = clientSecret,
                Scopes = new List<string>()  // overrides global scopes
            }
        );

        ex = await Assert.ThrowsAsync<Exception>(
            () => sdk.Hooks.AuthenticatedRequestAsync(null) // would normally require [read, write]
        );
        Assert.NotNull(ex);
        Assert.Equal("An error occurred while calling BeforeRequestAsync hook", ex.Message);
        Assert.Equal("Unexpected status code BadRequest: empty_scopes\n", ex.InnerException?.Message);

        // now use a different tokenUrl that will allow no scopes to be requested
        var tokenUrl = "/clientcredentials/token?expires_in=90&skip_scopes=true";
        var log = new List<CommonHelpers.RequestLogEntry>();
        sdk = new SDK(
            serverUrl: CommonHelpers.ApiTestServiceUrl,
            client: new CommonHelpers.RequestRecorderClient(log),
            security: new Security() {
                ClientID = clientID,
                ClientSecret = clientSecret,
                TokenURL = tokenUrl,
            }
        );

        var res = await sdk.Hooks.AuthenticatedRequestNoScopesAsync(null);
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        // since the token is not expired, it should be reused on subsequent call
        res = await sdk.Hooks.AuthenticatedRequestNoScopesAsync(null);
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var tokenRequests = log.Where(e => e.RequestUri?.PathAndQuery == tokenUrl).ToList();
        Assert.Single(tokenRequests);
        Assert.DoesNotContain("scope=", tokenRequests.First().RequestBody);
    }

    [Fact]
    public async Task TestClientCredentialsHookConcurrentRequests()
    {
        var sdk = new SDK(serverUrl: CommonHelpers.ApiTestServiceUrl, security: new Security() {
                ClientID = "speakeasy-sdks",
                ClientSecret = $"supersecret-{CommonHelpers.RandSeq(10)}",
        });

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
    }

    [Fact]
    public async Task TestClientCredentialsHookLowercaseBearerToken()
    {
        CommonHelpers.RecordTest("hooks-client-credentials-lowercase-bearer");

        var sdk = new SDK(serverUrl: CommonHelpers.ApiTestServiceUrl, security: new Security() {
            ClientID = "speakeasy-sdks",
            ClientSecret = $"supersecret-{CommonHelpers.RandSeq(10)}",
            TokenURL = "/clientcredentials/token?token_type=bearer",
        });

        var res = await sdk.Hooks.AuthenticatedRequestAsync(null);
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }
}
