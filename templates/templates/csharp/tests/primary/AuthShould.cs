using Xunit;
using Openapi;
using Openapi.Models.Errors;
using Openapi.Models.Shared;
using Openapi.Models.Operations;
using System.Collections.Generic;
using System.Threading.Tasks;
using System.Net;

public class AuthShould
{
    [Fact]
    public async Task NoAuth()
    {
        CommonHelpers.RecordTest("auth-no-auth");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Auth.NoAuthAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task BasicAuth()
    {
        CommonHelpers.RecordTest("auth-basic-auth");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.AuthNew.BasicAuthNewAsync(
            new BasicAuthNewSecurity() { Username = "testUser", Password = "testPass" },
            new AuthServiceRequestBody()
            {
                BasicAuth = new BasicAuth() { Username = "testUser", Password = "testPass" }
            }
        );

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task BasicAuthEmpty()
    {
        CommonHelpers.RecordTest("auth-basic-auth-empty");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.AuthNew.BasicAuthNewAsync(
            new BasicAuthNewSecurity() { Username = "", Password = "" },
            new AuthServiceRequestBody()
            {
                BasicAuth = new BasicAuth() { Username = "", Password = "" }
            }
        );

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task BasicAuthUsernameOnly()
    {
        CommonHelpers.RecordTest("auth-basic-auth-username-only");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.AuthNew.BasicAuthNewAsync(
            new BasicAuthNewSecurity() { Username = "testUser", Password = "" },
            new AuthServiceRequestBody()
            {
                BasicAuth = new BasicAuth() { Username = "testUser", Password = "" }
            }
        );

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task BasicAuthPasswordOnly()
    {
        CommonHelpers.RecordTest("auth-basic-auth-password-only");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.AuthNew.BasicAuthNewAsync(
            new BasicAuthNewSecurity() { Username = "", Password = "testPass" },
            new AuthServiceRequestBody()
            {
                BasicAuth = new BasicAuth() { Username = "", Password = "testPass" }
            }
        );

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task ApiKeyAuthGlobal()
    {
        CommonHelpers.RecordTest("auth-api-key-auth-global");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, security: new Security() { ApiKeyAuth = "Bearer test_api_key" });

        var res = await sdk.Auth.ApiKeyAuthGlobalAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("test_api_key", res.Token.Token);

        sdk = SDK.Builder().WithServerUrl(Helpers.HttpBinUrl).WithSecurity(new Security() { ApiKeyAuth = "Bearer test_api_key" }).Build();

        res = await sdk.Auth.ApiKeyAuthGlobalAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("test_api_key", res.Token.Token);
    }

    [Fact]
    public async Task BearerAuthOperationWithPrefix()
    {
        CommonHelpers.RecordTest("auth-bearer-auth-operation-with-prefix");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Auth.BearerAuthAsync(
            new BearerAuthSecurity() { BearerAuth = "Bearer testToken" }
        );

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.True(res.Token.Authenticated);
        Assert.Equal("testToken", res.Token.Token);
    }

    [Fact]
    public async Task BearerAuthOperationWithoutPrefix()
    {
        CommonHelpers.RecordTest("auth-bearer-auth-operation-without-prefix");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Auth.BearerAuthAsync(
            new BearerAuthSecurity() { BearerAuth = "testToken" }
        );

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.True(res.Token.Authenticated);
        Assert.Equal("testToken", res.Token.Token);
    }

    [Fact]
    public async Task Oauth2Auth()
    {
        CommonHelpers.RecordTest("auth-oauth2-auth");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, security: new Security() { Oauth2 = "Bearer testToken" });

        var res = await sdk.AuthNew.Oauth2AuthNewAsync(
            new AuthServiceRequestBody()
            {
                HeaderAuth = new List<HeaderAuth>()
                {
                    new HeaderAuth()
                    {
                        HeaderName = "Authorization",
                        ExpectedValue = "Bearer testToken"
                    }
                }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        sdk = SDK.Builder().WithServerUrl(Helpers.HttpBinUrl).WithSecurity(new Security() { Oauth2 = "Bearer testToken" }).Build();

        res = await sdk.AuthNew.Oauth2AuthNewAsync(
            new AuthServiceRequestBody()
            {
                HeaderAuth = new List<HeaderAuth>()
                {
                    new HeaderAuth()
                    {
                        HeaderName = "Authorization",
                        ExpectedValue = "Bearer testToken"
                    }
                }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task OpenIdConnectAuth()
    {
        CommonHelpers.RecordTest("auth-open-id-connect-auth");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.AuthNew.OpenIdConnectAuthNewAsync(
            new OpenIdConnectAuthNewSecurity() { OpenIdConnect = "Bearer testToken" },
            new AuthServiceRequestBody()
            {
                HeaderAuth = new List<HeaderAuth>()
                {
                    new HeaderAuth()
                    {
                        HeaderName = "Authorization",
                        ExpectedValue = "Bearer testToken"
                    }
                }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task MultipleSimpleSchemeAuth()
    {
        CommonHelpers.RecordTest("auth-multiple-simple-scheme-auth");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.AuthNew.MultipleSimpleSchemeAuthAsync(
            new MultipleSimpleSchemeAuthSecurity()
            {
                ApiKeyAuthNew = "test_api_key",
                Oauth2 = "Bearer testToken"
            },
            new AuthServiceRequestBody()
            {
                HeaderAuth = new List<HeaderAuth>()
                {
                    new HeaderAuth()
                    {
                        HeaderName = "x-api-key",
                        ExpectedValue = "test_api_key"
                    },
                    new HeaderAuth()
                    {
                        HeaderName = "Authorization",
                        ExpectedValue = "Bearer testToken"
                    }
                }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task MultipleMixedSchemeAuth()
    {
        CommonHelpers.RecordTest("auth-multiple-mixed-scheme-auth");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.AuthNew.MultipleMixedSchemeAuthAsync(
            new MultipleMixedSchemeAuthSecurity()
            {
                ApiKeyAuthNew = "test_api_key",
                BasicAuth = new SchemeBasicAuth() { Username = "testUser", Password = "testPass" }
            },
            new AuthServiceRequestBody()
            {
                HeaderAuth = new List<HeaderAuth>()
                {
                    new HeaderAuth()
                    {
                        HeaderName = "x-api-key",
                        ExpectedValue = "test_api_key"
                    }
                },
                BasicAuth = new BasicAuth()
                {
                    Username = "testUser",
                    Password = "testPass"
                }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task MultipleSimpleOptionsAuthFirstOption()
    {
        CommonHelpers.RecordTest("auth-multiple-simple-options-auth-first-option");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.AuthNew.MultipleSimpleOptionsAuthAsync(
            new MultipleSimpleOptionsAuthSecurity { ApiKeyAuthNew = "test_api_key" },
            new AuthServiceRequestBody()
            {
                HeaderAuth = new List<HeaderAuth>()
                {
                    new HeaderAuth()
                    {
                        HeaderName = "x-api-key",
                        ExpectedValue = "test_api_key"
                    }
                }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task MultipleSimpleOptionsAuthSecondOption()
    {
        CommonHelpers.RecordTest("auth-multiple-simple-options-auth-second-option");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.AuthNew.MultipleSimpleOptionsAuthAsync(
            new MultipleSimpleOptionsAuthSecurity { Oauth2 = "Bearer testToken" },
            new AuthServiceRequestBody()
            {
                HeaderAuth = new List<HeaderAuth>()
                {
                    new HeaderAuth()
                    {
                        HeaderName = "Authorization",
                        ExpectedValue = "Bearer testToken"
                    }
                }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task MultipleMixedOptionsAuthFirstOption()
    {
        CommonHelpers.RecordTest("auth-multiple-mixed-options-auth-first-option");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.AuthNew.MultipleMixedOptionsAuthAsync(
            new MultipleMixedOptionsAuthSecurity() { ApiKeyAuthNew = "test_api_key" },
            new AuthServiceRequestBody()
            {
                HeaderAuth = new List<HeaderAuth>()
                {
                    new HeaderAuth()
                    {
                        HeaderName = "x-api-key",
                        ExpectedValue = "test_api_key"
                    }
                }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task MultipleMixedOptionsAuthSecondOption()
    {
        CommonHelpers.RecordTest("auth-multiple-mixed-options-auth-second-option");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.AuthNew.MultipleMixedOptionsAuthAsync(
            new MultipleMixedOptionsAuthSecurity()
            {
                BasicAuth = new SchemeBasicAuth() { Username = "testUser", Password = "testPass" }
            },
            new AuthServiceRequestBody()
            {
                BasicAuth = new BasicAuth()
                {
                    Username = "testUser",
                    Password = "testPass"
                }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task MultipleOptionsWithSimpleSchemesAuthFirstOption()
    {
        CommonHelpers.RecordTest("auth-multiple-options-with-simple-schemes-auth-first-option");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.AuthNew.MultipleOptionsWithSimpleSchemesAuthAsync(
            new MultipleOptionsWithSimpleSchemesAuthSecurity()
            {
                Option1 = new MultipleOptionsWithSimpleSchemesAuthSecurityOption1()
                {
                    ApiKeyAuthNew = "test_api_key",
                    Oauth2 = "Bearer testToken"
                }
            },
            new AuthServiceRequestBody()
            {
                HeaderAuth = new List<HeaderAuth>()
                {
                    new HeaderAuth()
                    {
                        HeaderName = "x-api-key",
                        ExpectedValue = "test_api_key"
                    },
                    new HeaderAuth()
                    {
                        HeaderName = "Authorization",
                        ExpectedValue = "Bearer testToken"
                    }
                }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task MultipleOptionsWithSimpleSchemesAuthSecondOption()
    {
        CommonHelpers.RecordTest("auth-multiple-options-with-simple-schemes-auth-second-option");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.AuthNew.MultipleOptionsWithSimpleSchemesAuthAsync(
            new MultipleOptionsWithSimpleSchemesAuthSecurity()
            {
                Option2 = new MultipleOptionsWithSimpleSchemesAuthSecurityOption2()
                {
                    ApiKeyAuthNew = "test_api_key",
                    OpenIdConnect = "Bearer testToken"
                }
            },
            new AuthServiceRequestBody()
            {
                HeaderAuth = new List<HeaderAuth>()
                {
                    new HeaderAuth()
                    {
                        HeaderName = "x-api-key",
                        ExpectedValue = "test_api_key"
                    },
                    new HeaderAuth()
                    {
                        HeaderName = "Authorization",
                        ExpectedValue = "Bearer testToken"
                    }
                }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task MultipleOptionsWithMixedSchemesAuthFirstOption()
    {
        CommonHelpers.RecordTest("auth-multiple-options-with-mixed-schemes-auth-first-option");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.AuthNew.MultipleOptionsWithMixedSchemesAuthAsync(
            new MultipleOptionsWithMixedSchemesAuthSecurity()
            {
                Option1 = new MultipleOptionsWithMixedSchemesAuthSecurityOption1()
                {
                    ApiKeyAuthNew = "test_api_key",
                    Oauth2 = "Bearer testToken"
                }
            },
            new AuthServiceRequestBody()
            {
                HeaderAuth = new List<HeaderAuth>()
                {
                    new HeaderAuth()
                    {
                        HeaderName = "x-api-key",
                        ExpectedValue = "test_api_key"
                    },
                    new HeaderAuth()
                    {
                        HeaderName = "Authorization",
                        ExpectedValue = "Bearer testToken"
                    }
                }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task MultipleOptionsWithMixedSchemesAuthSecondOption()
    {
        CommonHelpers.RecordTest("auth-multiple-options-with-mixed-schemes-auth-second-option");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.AuthNew.MultipleOptionsWithMixedSchemesAuthAsync(
            new MultipleOptionsWithMixedSchemesAuthSecurity()
            {
                Option2 = new MultipleOptionsWithMixedSchemesAuthSecurityOption2()
                {
                    ApiKeyAuthNew = "test_api_key",
                    BasicAuth = new SchemeBasicAuth()
                    {
                        Username = "testUser",
                        Password = "testPass"
                    }
                }
            },
            new AuthServiceRequestBody()
            {
                HeaderAuth = new List<HeaderAuth>()
                {
                    new HeaderAuth()
                    {
                        HeaderName = "x-api-key",
                        ExpectedValue = "test_api_key"
                    }
                },
                BasicAuth = new BasicAuth()
                {
                    Username = "testUser",
                    Password = "testPass"
                }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task FunctionCallbacksOauthGlobalSecurity()
    {
        CommonHelpers.RecordTest("auth-function-callbacks-oauth-global-security");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, securitySource: () => new Security() { Oauth2 = "Bearer global" });

        var res = await sdk.Auth.GlobalBearerAuthAsync();

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("global", res.Token.Token);

        sdk = SDK.Builder().WithServerUrl(Helpers.HttpBinUrl).WithSecuritySource(() => new Security() { Oauth2 = "Bearer global" }).Build();

        res = await sdk.Auth.GlobalBearerAuthAsync();

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("global", res.Token.Token);
    }

    [Fact]
    public async Task CustomSecurityOptionAppId()
    {
        CommonHelpers.RecordTest("auth-custom-security-option-app-id");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, security: new Security() {
            CustomSchemeAppId = new SchemeCustomSchemeAppID() {
                AppId = "testAppID",
                Secret = "testSecret"
            }
        });

        var res = await sdk.AuthNew.CustomSchemeAppIdAsync();

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        sdk = SDK.Builder().WithServerUrl(Helpers.HttpBinUrl).WithSecurity(new Security() {
            CustomSchemeAppId = new SchemeCustomSchemeAppID() {
                AppId = "testAppID",
                Secret = "testSecret"
            }
        }).Build();

        res = await sdk.AuthNew.CustomSchemeAppIdAsync();

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task HoistedSecurityAccessTokenOnly()
    {
        CommonHelpers.RecordTest("auth-hoisted-security-access-token-only");

        var sdk = new SDK(
            security: new Security()
            {
                ApiKeyAuth = "testApiKey",
                BasicHttp = new SchemeBasicHTTP()
                {
                    Username = "testUser",
                    Password = "testPass"
                },
                AccessToken = "Bearer ghp_xxxx"
            }
        );

        var res = await sdk.Auth.HoistedSecurityAccessTokenOnlyAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Token);
        Assert.Equal("Bearer ghp_xxxx", res.Token.Token);
    }

    [Fact]
    public async Task GlobalSecurityFieldsOrdering()
    {
        CommonHelpers.RecordTest("auth-global-security-fields-ordering");

        // When `maintainOpenApiOrder: false` is set, security fields are sorted in alphabetical order.
        // Since the first non-nil field is selected ApiKeyAuth takes precedence over BasicHttp,
        // which is not valid authentication for this endpoint.
        var sdk = new SDK(
            security: new Security()
            {
                ApiKeyAuth = "testApiKey",
                BasicHttp = new SchemeBasicHTTP()
                {
                    Username = "testUser",
                    Password = "testPass"
                }
            }
        );

        var ex = await Assert.ThrowsAsync<APIException>(async () =>
        {
            await sdk.Auth.GlobalSecurityBasicHttpAsync();
        });
        Assert.Equal(HttpStatusCode.Unauthorized, ex.Response.StatusCode);
    }

    [Fact]
    public async Task HoistedSecurityAccessTokenFirst()
    {
        CommonHelpers.RecordTest("auth-hoisted-security-access-token-first");

        var sdk = new SDK(
            security: new Security()
            {
                ApiKeyAuth = "testApiKey",
                AccessToken = "Bearer ghp_xxxx"
            }
        );

        var res = await sdk.Auth.HoistedSecurityAccessTokenFirstAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Token);
        Assert.Equal("Bearer ghp_xxxx", res.Token.Token);
    }

    [Fact]
    public async Task HoistedSecurityApiKeyFirst()
    {
        CommonHelpers.RecordTest("auth-hoisted-security-api-key-first");

        var sdk = new SDK(
            security: new Security()
            {
                ApiKeyAuth = "testApiKey",
                AccessToken = "Bearer ghp_xxxx"
            }
        );

        var res = await sdk.Auth.HoistedSecurityApiKeyFirstAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Token);
        Assert.Equal("testApiKey", res.Token.Token);
    }

    [Fact]
    public async Task HoistedSecurityBasicHttpOnly()
    {
        CommonHelpers.RecordTest("auth-hoisted-security-basic-http-only");

        var sdk = new SDK(
            security: new Security()
            {
                BasicHttp = new SchemeBasicHTTP()
                {
                    Username = "testUser",
                    Password = "testPass"
                },
                ApiKeyAuth = "testApiKey",
                AccessToken = "Bearer ghp_xxxx"
            }
        );

        var res = await sdk.Auth.HoistedSecurityBasicHttpOnlyAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.BasicAuth);
        Assert.True(res.BasicAuth.Authenticated);
        Assert.Equal("testUser", res.BasicAuth.User);
    }

    [Fact]
    public async Task HoistedSecurityInvalidField()
    {
        CommonHelpers.RecordTest("auth-hoisted-security-invalid-field");

        // Provide only BasicHttp — not valid for accessTokenFirst which expects bearer/apiKey
        var sdk = new SDK(
            security: new Security()
            {
                BasicHttp = new SchemeBasicHTTP()
                {
                    Username = "user",
                    Password = "pass"
                }
            }
        );

        var ex = await Assert.ThrowsAsync<APIException>(async () =>
        {
            await sdk.Auth.HoistedSecurityAccessTokenFirstAsync();
        });
        Assert.Equal(HttpStatusCode.Unauthorized, ex.Response.StatusCode);
    }
}
