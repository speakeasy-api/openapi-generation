using System.Collections.Generic;
using NUnit.Framework;
using UnityEngine.TestTools;
using Openapi;
using Openapi.Models.Operations;
using Openapi.Models.Shared;

using System.Collections;

public class AuthShould
{
    [UnityTest]
    public IEnumerator NoAuth()
    {
        CommonHelpers.RecordTest("auth-no-auth");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
              var res = await sdk.Auth.NoAuthAsync()
            )
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator BasicAuth()
    {
        CommonHelpers.RecordTest("auth-basic-auth");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.AuthNew.BasicAuthNewAsync(
                    new BasicAuthNewSecurity() { Username = "testUser", Password = "testPass" },
                    new AuthServiceRequestBody()
                    {
                        BasicAuth = new BasicAuth() { Username = "testUser", Password = "testPass" }
                    }
                )
            )
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator BasicAuthEmpty()
    {
        CommonHelpers.RecordTest("auth-basic-auth-empty");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.AuthNew.BasicAuthNewAsync(
                    new BasicAuthNewSecurity() { Username = "", Password = "" },
                    new AuthServiceRequestBody()
                    {
                        BasicAuth = new BasicAuth() { Username = "", Password = "" }
                    }
                )
            )
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator BasicAuthUsernameOnly()
    {
        CommonHelpers.RecordTest("auth-basic-auth-username-only");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.AuthNew.BasicAuthNewAsync(
                    new BasicAuthNewSecurity() { Username = "testUser", Password = "" },
                    new AuthServiceRequestBody()
                    {
                        BasicAuth = new BasicAuth() { Username = "testUser", Password = "" }
                    }
                )
            )
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator BasicAuthPasswordOnly()
    {
        CommonHelpers.RecordTest("auth-basic-auth-password-only");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.AuthNew.BasicAuthNewAsync(
                    new BasicAuthNewSecurity() { Username = "", Password = "testPass" },
                    new AuthServiceRequestBody()
                    {
                        BasicAuth = new BasicAuth() { Username = "", Password = "testPass" }
                    }
                )
            )
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator ApiKeyAuthGlobal()
    {
        CommonHelpers.RecordTest("auth-api-key-auth-global");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(security: new Security() { ApiKeyAuth = "Bearer test_api_key" });

            using (
                var res = await sdk.Auth.ApiKeyAuthGlobalAsync()
            )
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual("test_api_key", res.Token.Token);
            }
        });
    }

    [UnityTest]
    public IEnumerator BearerAuthOperationWithPrefix()
    {
        CommonHelpers.RecordTest("auth-bearer-auth-operation-with-prefix");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Auth.BearerAuthAsync(
                    new BearerAuthSecurity() { BearerAuth = "Bearer testToken" }
                )
            )
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
                Assert.True(res.Token.Authenticated);
                Assert.AreEqual("testToken", res.Token.Token);
            }
        });
    }

    [UnityTest]
    public IEnumerator BearerAuthOperationWithoutPrefix()
    {
        CommonHelpers.RecordTest("auth-bearer-auth-operation-without-prefix");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Auth.BearerAuthAsync(
                    new BearerAuthSecurity() { BearerAuth = "testToken" }
                )
            )
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
                Assert.True(res.Token.Authenticated);
                Assert.AreEqual("testToken", res.Token.Token);
            }
        });
    }

    [UnityTest]
    public IEnumerator Oauth2Auth()
    {
        CommonHelpers.RecordTest("auth-oauth2-auth");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(security: new Security() { Oauth2 = "Bearer testToken" });

            using (
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
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator OpenIdConnectAuth()
    {
        CommonHelpers.RecordTest("auth-open-id-connect-auth");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
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
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator MultipleSimpleSchemeAuth()
    {
        CommonHelpers.RecordTest("auth-multiple-simple-scheme-auth");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
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
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator MultipleMixedSchemeAuth()
    {
        CommonHelpers.RecordTest("auth-multiple-mixed-scheme-auth");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.AuthNew.MultipleMixedSchemeAuthAsync(
                    new MultipleMixedSchemeAuthSecurity()
                    {
                        ApiKeyAuthNew = "test_api_key",
                        BasicAuth = new SchemeBasicAuth()
                        {
                            Username = "testUser",
                            Password = "testPass"
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
                        BasicAuth = new BasicAuth() { Username = "testUser", Password = "testPass" }
                    }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator MultipleSimpleOptionsAuthFirstOption()
    {
        CommonHelpers.RecordTest("auth-multiple-simple-options-auth-first-option");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
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
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator MultipleSimpleOptionsAuthSecondOption()
    {
        CommonHelpers.RecordTest("auth-multiple-simple-options-auth-second-option");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
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
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator MultipleMixedOptionsAuthFirstOption()
    {
        CommonHelpers.RecordTest("auth-multiple-mixed-options-auth-first-option");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
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
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator MultipleMixedOptionsAuthSecondOption()
    {
        CommonHelpers.RecordTest("auth-multiple-mixed-options-auth-second-option");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.AuthNew.MultipleMixedOptionsAuthAsync(
                    new MultipleMixedOptionsAuthSecurity()
                    {
                        BasicAuth = new SchemeBasicAuth()
                        {
                            Username = "testUser",
                            Password = "testPass"
                        }
                    },
                    new AuthServiceRequestBody()
                    {
                        BasicAuth = new BasicAuth() { Username = "testUser", Password = "testPass" }
                    }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator MultipleOptionsWithSimpleSchemesAuthFirstOption()
    {
        CommonHelpers.RecordTest("auth-multiple-options-with-simple-schemes-auth-first-option");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
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
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator MultipleOptionsWithSimpleSchemesAuthSecondOption()
    {
        CommonHelpers.RecordTest("auth-multiple-options-with-simple-schemes-auth-second-option");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
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
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator MultipleOptionsWithMixedSchemesAuthFirstOption()
    {
        CommonHelpers.RecordTest("auth-multiple-options-with-mixed-schemes-auth-first-option");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
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
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator MultipleOptionsWithMixedSchemesAuthSecondOption()
    {
        CommonHelpers.RecordTest("auth-multiple-options-with-mixed-schemes-auth-second-option");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
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
                        BasicAuth = new BasicAuth() { Username = "testUser", Password = "testPass" }
                    }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator FunctionCallbacksOauthGlobalSecurity()
    {
        CommonHelpers.RecordTest("auth-function-callbacks-oauth-global-security");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(securitySource: () => new Security() { Oauth2 = "Bearer global" });

            using (
                var res = await sdk.Auth.GlobalBearerAuthAsync()
            )
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
                Assert.True(res.Token.Authenticated);
                Assert.AreEqual("global", res.Token.Token);
            }
        });
    }
}
