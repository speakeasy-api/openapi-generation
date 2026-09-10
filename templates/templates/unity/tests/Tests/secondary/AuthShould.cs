using System;
using System.Collections;
using HoistedSecurity;
using HoistedSecurity.Models.Operations;
using HoistedSecurity.Models.Shared;
using NUnit.Framework;
using UnityEngine.TestTools;

public class AuthShould
{
    [UnityTest]
    public IEnumerator NoAuth()
    {
        CommonHelpers.RecordTest("auth-hoisted-no-auth-retained");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (var res = await sdk.Auth.NoAuthAsync())
            {
                Assert.NotNull(res);
            }
        });
    }

    [UnityTest]
    public IEnumerator BasicAuth()
    {
        CommonHelpers.RecordTest("auth-hoisted-basic-auth");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(
                security: new Security() { Username = "testUser", Password = "testPass" }
            );

            using (var res = await sdk.Auth.BasicAuthAsync("testPass", "testUser"))
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
                Assert.True(res.User.Authenticated);
            }
        });
    }

    [UnityTest]
    public IEnumerator TestMultipleMixedOptionsAuth()
    {
        CommonHelpers.RecordTest("auth-hoisted-operation-auth-retained");

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
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }
}
