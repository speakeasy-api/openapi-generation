using NUnit.Framework;
using UnityEngine.TestTools;
using System.Collections;
using System.Collections.Generic;
using Company.Product.Feature.Subnamespace;

public class AuthShould
{
    [UnityTest]
    public IEnumerator GlobalSecurityFlattening()
    {
        CommonHelpers.RecordTest("auth-global-security-flattening");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(apiKeyAuth: "Bearer testToken");

            using (
                var res = await sdk.Auth.ApiKeyAuthGlobalAsync()
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
    public IEnumerator GlobalSecurityFlatteningCallback()
    {
        CommonHelpers.RecordTest("auth-global-security-flattening-callback");

        yield return CommonHelpers.Await(async () =>
        {
            var ex = Assert.Throws<System.Exception>(() => new SDK());
            Assert.AreEqual("apiKeyAuth and apiKeyAuthSource cannot both be null", ex.Message);

            var sdk = new SDK(apiKeyAuthSource: () => "Bearer testToken");

            using (
                var res = await sdk.Auth.ApiKeyAuthGlobalAsync()
            )
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
                Assert.True(res.Token.Authenticated);
                Assert.AreEqual("testToken", res.Token.Token);
            }
        });
    }
}
