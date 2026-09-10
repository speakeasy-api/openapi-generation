using Xunit;
using Speakeasy.SecurityOptions;
using Speakeasy.SecurityOptions.Models.Shared;
using System.Net;
using System.Threading.Tasks;

public class AuthAdditionalShould
{
    [Fact]
    public async Task GlobalSecurityBasicHttpSuccess()
    {
        CommonHelpers.RecordTest("auth-basic-http-global-option");

        // Expected to succeed since BasicHTTP takes priority over AccessToken in global security definition
        var sdk = new SDK(
            security: new Security()
            {
                BasicHttp = new BasicHttp()
                {
                    Username = "testUser",
                    Password = "testPass"
                },
                AccessToken = new AccessToken()
                {
                    AccessTokenValue = "Bearer ignored"
                }
            }
        );

        var res = await sdk.Auth.GlobalSecurityOptionBasicHttpAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.BasicAuth);
        Assert.True(res.BasicAuth.Authenticated);
        Assert.Equal("testUser", res.BasicAuth.User);
    }

    [Fact]
    public async Task GlobalSecurityFieldsOrdering()
    {
        CommonHelpers.RecordTest("auth-global-security-option-fields-ordering");

        // When `maintainOpenApiOrder: false` is set, security fields are sorted in alphabetical order.
        // Since the first non-nil field is selected ApiKeyAuth takes precedence over BasicHttp,
        // which is not valid authentication for this endpoint.
        var sdk = new SDK(
            security: new Security()
            {
                ApiKeyAuth = new ApiKeyAuth()
                {
                    ApiKeyAuthValue = "Bearer test_api_key"
                },
                BasicHttp = new BasicHttp()
                {
                    Username = "testUser",
                    Password = "testPass"
                }
            }
        );

        var ex = await Assert.ThrowsAsync<Speakeasy.SecurityOptions.Models.Errors.APIException>(async () =>
        {
            await sdk.Auth.GlobalSecurityOptionBasicHttpAsync();
        });
        Assert.Equal(HttpStatusCode.Unauthorized, ex.Response.StatusCode);
    }

    [Fact]
    public async Task HoistedSecurityAccessTokenFirst()
    {
        CommonHelpers.RecordTest("auth-hoisted-security-option-access-token-first");

        var sdk = new SDK(
            security: new Security()
            {
                ApiKeyAuth = new ApiKeyAuth()
                {
                    ApiKeyAuthValue = "testApiKey"
                },
                BasicHttp = new BasicHttp()
                {
                    Username = "username",
                    Password = "password"
                },
                AccessToken = new AccessToken()
                {
                    AccessTokenValue = "Bearer ghp_xxxx"
                }
            }
        );

        var res = await sdk.Auth.HoistedSecurityOptionAccessTokenFirstAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.TokenAuthResponse);
        Assert.Equal("Bearer ghp_xxxx", res.TokenAuthResponse.Token);
    }

    [Fact]
    public async Task HoistedSecurityApiKeyFirst()
    {
        CommonHelpers.RecordTest("auth-hoisted-security-option-api-key-first");

        var sdk = new SDK(
            security: new Security()
            {
                ApiKeyAuth = new ApiKeyAuth()
                {
                    ApiKeyAuthValue = "testApiKey"
                },
                BasicHttp = new BasicHttp()
                {
                    Username = "username",
                    Password = "password"
                },
                AccessToken = new AccessToken()
                {
                    AccessTokenValue = "Bearer ghp_xxxx"
                }
            }
        );

        var res = await sdk.Auth.HoistedSecurityOptionApiKeyFirstAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.TokenAuthResponse);
        Assert.Equal("testApiKey", res.TokenAuthResponse.Token);
    }

    [Fact]
    public async Task HoistedSecurityBasicHttpOnly()
    {
        CommonHelpers.RecordTest("auth-hoisted-security-option-basic-http-only");

        var sdk = new SDK(
            security: new Security()
            {
                ApiKeyAuth = new ApiKeyAuth()
                {
                    ApiKeyAuthValue = "testApiKey"
                },
                BasicHttp = new BasicHttp()
                {
                    Username = "testUser",
                    Password = "testPass"
                },
                AccessToken = new AccessToken()
                {
                    AccessTokenValue = "Bearer ghp_xxxx"
                }
            }
        );

        var res = await sdk.Auth.HoistedSecurityOptionBasicHttpOnlyAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.BasicAuth);
        Assert.True(res.BasicAuth.Authenticated);
        Assert.Equal("testUser", res.BasicAuth.User);
    }

    [Fact]
    public async Task HoistedSecurityInvalidOption()
    {
        CommonHelpers.RecordTest("auth-hoisted-security-invalid-option");

        var sdk = new SDK(
            security: new Security()
            {
                BasicHttp = new BasicHttp()
                {
                    Username = "username",
                    Password = "password"
                }
            }
        );

        var ex = await Assert.ThrowsAsync<Speakeasy.SecurityOptions.Models.Errors.APIException>(async () =>
        {
            await sdk.Auth.HoistedSecurityOptionAccessTokenFirstAsync();
        });
        Assert.Equal(HttpStatusCode.Unauthorized, ex.Response.StatusCode);
    }
}
